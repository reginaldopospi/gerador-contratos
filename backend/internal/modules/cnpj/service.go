package cnpj

import (
	"context"
	"strings"

	"gerador-contratos/backend/internal/modules/auth"
	"gerador-contratos/backend/internal/modules/common"
)

const cnpjDigitsLength = 14

type Service struct {
	repo ProviderRepository
}

func NewService(repo ProviderRepository) *Service {
	return &Service{repo: repo}
}

// LookupByCNPJ valida o documento e retorna os dados juridicos para autopreenchimento.
func (s *Service) LookupByCNPJ(
	ctx context.Context,
	_ auth.AuthClaims,
	rawCNPJ string,
) (*Company, error) {
	cnpjDigits := onlyDigits(rawCNPJ)
	if len(cnpjDigits) != cnpjDigitsLength {
		return nil, common.NewBadRequest("invalid_cnpj", "cnpj invalido")
	}

	company, err := s.repo.LookupCompanyByCNPJ(ctx, cnpjDigits)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, common.NewNotFound("cnpj_not_found", "cnpj nao encontrado na Receita Federal")
	}

	company.CNPJ = formatCNPJ(cnpjDigits)
	company.RazaoSocial = strings.TrimSpace(company.RazaoSocial)
	company.Endereco.UF = strings.ToUpper(strings.TrimSpace(company.Endereco.UF))
	return company, nil
}

func onlyDigits(value string) string {
	var out strings.Builder
	out.Grow(len(value))
	for _, ch := range value {
		if ch >= '0' && ch <= '9' {
			out.WriteRune(ch)
		}
	}
	return out.String()
}

func formatCNPJ(cnpjDigits string) string {
	if len(cnpjDigits) != cnpjDigitsLength {
		return strings.TrimSpace(cnpjDigits)
	}
	return cnpjDigits[0:2] + "." +
		cnpjDigits[2:5] + "." +
		cnpjDigits[5:8] + "/" +
		cnpjDigits[8:12] + "-" +
		cnpjDigits[12:14]
}
