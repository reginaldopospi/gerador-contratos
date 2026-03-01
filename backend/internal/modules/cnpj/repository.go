package cnpj

import "context"

// ProviderRepository encapsula a consulta externa de dados cadastrais por CNPJ.
type ProviderRepository interface {
	LookupCompanyByCNPJ(ctx context.Context, cnpjDigits string) (*Company, error)
}
