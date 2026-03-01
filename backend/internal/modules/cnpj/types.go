package cnpj

// CompanyAddress representa os campos de endereco retornados pela consulta de CNPJ.
type CompanyAddress struct {
	CEP         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Numero      string `json:"numero"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Cidade      string `json:"cidade"`
	UF          string `json:"uf"`
}

// Company representa os dados principais de pessoa juridica para o formulario.
type Company struct {
	CNPJ        string         `json:"cnpj"`
	RazaoSocial string         `json:"razao_social"`
	Endereco    CompanyAddress `json:"endereco"`
}
