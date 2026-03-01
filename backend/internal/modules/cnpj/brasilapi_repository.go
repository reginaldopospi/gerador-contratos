package cnpj

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gerador-contratos/backend/internal/modules/common"
)

const defaultBrasilAPIBaseURL = "https://brasilapi.com.br/api/cnpj/v1"

type BrasilAPIRepository struct {
	client  *http.Client
	baseURL string
}

type brasilAPIResponse struct {
	CNPJ                    string `json:"cnpj"`
	RazaoSocial             string `json:"razao_social"`
	DescricaoTipoLogradouro string `json:"descricao_tipo_de_logradouro"`
	Logradouro              string `json:"logradouro"`
	Numero                  string `json:"numero"`
	Complemento             string `json:"complemento"`
	Bairro                  string `json:"bairro"`
	CEP                     string `json:"cep"`
	Municipio               string `json:"municipio"`
	UF                      string `json:"uf"`
}

func NewBrasilAPIRepository(client *http.Client, baseURL string) *BrasilAPIRepository {
	repoClient := client
	if repoClient == nil {
		repoClient = &http.Client{Timeout: 10 * time.Second}
	}

	resolvedBaseURL := strings.TrimSpace(baseURL)
	if resolvedBaseURL == "" {
		resolvedBaseURL = defaultBrasilAPIBaseURL
	}

	return &BrasilAPIRepository{
		client:  repoClient,
		baseURL: strings.TrimRight(resolvedBaseURL, "/"),
	}
}

// LookupCompanyByCNPJ consulta o provedor externo e mapeia os campos necessarios do formulario.
func (r *BrasilAPIRepository) LookupCompanyByCNPJ(ctx context.Context, cnpjDigits string) (*Company, error) {
	url := fmt.Sprintf("%s/%s", r.baseURL, strings.TrimSpace(cnpjDigits))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create cnpj request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "gerador-contratos/1.0")

	res, err := r.client.Do(req)
	if err != nil {
		return nil, common.NewInternal("cnpj_lookup_failed", "falha ao consultar dados da Receita Federal")
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= http.StatusInternalServerError {
		return nil, common.NewInternal(
			"cnpj_lookup_unavailable",
			"servico da Receita Federal indisponivel no momento",
		)
	}
	if res.StatusCode != http.StatusOK {
		return nil, common.NewInternal("cnpj_lookup_failed", "falha ao consultar dados da Receita Federal")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, common.NewInternal("cnpj_lookup_failed", "falha ao ler dados da Receita Federal")
	}

	var payload brasilAPIResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, common.NewInternal("cnpj_lookup_failed", "falha ao interpretar dados da Receita Federal")
	}

	logradouro := strings.TrimSpace(payload.Logradouro)
	tipoLogradouro := strings.TrimSpace(payload.DescricaoTipoLogradouro)
	if tipoLogradouro != "" && logradouro != "" {
		logradouro = tipoLogradouro + " " + logradouro
	} else if tipoLogradouro != "" {
		logradouro = tipoLogradouro
	}

	return &Company{
		CNPJ:        strings.TrimSpace(payload.CNPJ),
		RazaoSocial: strings.TrimSpace(payload.RazaoSocial),
		Endereco: CompanyAddress{
			CEP:         strings.TrimSpace(payload.CEP),
			Logradouro:  logradouro,
			Numero:      strings.TrimSpace(payload.Numero),
			Complemento: strings.TrimSpace(payload.Complemento),
			Bairro:      strings.TrimSpace(payload.Bairro),
			Cidade:      strings.TrimSpace(payload.Municipio),
			UF:          strings.TrimSpace(payload.UF),
		},
	}, nil
}
