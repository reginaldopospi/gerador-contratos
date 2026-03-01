package cnpj

import (
	"context"
	"errors"
	"testing"

	"gerador-contratos/backend/internal/modules/auth"
	"gerador-contratos/backend/internal/modules/common"
)

type serviceRepoStub struct {
	result *Company
	err    error
}

func (s serviceRepoStub) LookupCompanyByCNPJ(_ context.Context, _ string) (*Company, error) {
	return s.result, s.err
}

func TestServiceLookupByCNPJRejectsInvalidDocument(t *testing.T) {
	service := NewService(serviceRepoStub{})

	_, err := service.LookupByCNPJ(context.Background(), auth.AuthClaims{}, "123")
	if err == nil {
		t.Fatalf("expected error for invalid cnpj")
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got: %v", err)
	}
	if appErr.Code != "invalid_cnpj" {
		t.Fatalf("expected invalid_cnpj code, got: %s", appErr.Code)
	}
}

func TestServiceLookupByCNPJReturnsNotFoundWhenProviderHasNoRecord(t *testing.T) {
	service := NewService(serviceRepoStub{result: nil})

	_, err := service.LookupByCNPJ(context.Background(), auth.AuthClaims{}, "58132597000108")
	if err == nil {
		t.Fatalf("expected error when cnpj is not found")
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got: %v", err)
	}
	if appErr.Code != "cnpj_not_found" {
		t.Fatalf("expected cnpj_not_found code, got: %s", appErr.Code)
	}
}

func TestServiceLookupByCNPJFormatsCNPJAndNormalizesUF(t *testing.T) {
	service := NewService(serviceRepoStub{
		result: &Company{
			CNPJ:        "58132597000108",
			RazaoSocial: "  EMPRESA TESTE LTDA  ",
			Endereco: CompanyAddress{
				UF: "sp",
			},
		},
	})

	company, err := service.LookupByCNPJ(context.Background(), auth.AuthClaims{}, "58132597000108")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if company.CNPJ != "58.132.597/0001-08" {
		t.Fatalf("expected formatted cnpj, got: %s", company.CNPJ)
	}
	if company.RazaoSocial != "EMPRESA TESTE LTDA" {
		t.Fatalf("expected trimmed razao social, got: %s", company.RazaoSocial)
	}
	if company.Endereco.UF != "SP" {
		t.Fatalf("expected normalized uf, got: %s", company.Endereco.UF)
	}
}
