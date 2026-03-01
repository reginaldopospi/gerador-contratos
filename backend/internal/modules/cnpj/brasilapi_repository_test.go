package cnpj

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gerador-contratos/backend/internal/modules/common"
)

func TestBrasilAPIRepositoryLookupMapsProviderPayload(t *testing.T) {
	// Garante que o mapeamento mantenha o formato esperado pelo formulario.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/58132597000108" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"cnpj":                         "58132597000108",
			"razao_social":                 "Empresa Exemplo LTDA",
			"descricao_tipo_de_logradouro": "Rua",
			"logradouro":                   "das Flores",
			"numero":                       "123",
			"complemento":                  "Sala 2",
			"bairro":                       "Centro",
			"cep":                          "08663040",
			"municipio":                    "Suzano",
			"uf":                           "sp",
		})
	}))
	defer server.Close()

	repo := NewBrasilAPIRepository(server.Client(), server.URL)

	company, err := repo.LookupCompanyByCNPJ(context.Background(), "58132597000108")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if company == nil {
		t.Fatalf("expected company payload")
	}
	if company.RazaoSocial != "Empresa Exemplo LTDA" {
		t.Fatalf("unexpected razao social: %s", company.RazaoSocial)
	}
	if company.Endereco.Logradouro != "Rua das Flores" {
		t.Fatalf("unexpected logradouro: %s", company.Endereco.Logradouro)
	}
}

func TestBrasilAPIRepositoryLookupReturnsNilFor404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	repo := NewBrasilAPIRepository(server.Client(), server.URL)
	company, err := repo.LookupCompanyByCNPJ(context.Background(), "58132597000108")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if company != nil {
		t.Fatalf("expected nil company for 404 status")
	}
}

func TestBrasilAPIRepositoryLookupReturnsAppErrorForRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	repo := NewBrasilAPIRepository(server.Client(), server.URL)
	_, err := repo.LookupCompanyByCNPJ(context.Background(), "58132597000108")
	if err == nil {
		t.Fatalf("expected error for rate limit")
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got: %v", err)
	}
	if appErr.Code != "cnpj_lookup_unavailable" {
		t.Fatalf("expected cnpj_lookup_unavailable code, got: %s", appErr.Code)
	}
}
