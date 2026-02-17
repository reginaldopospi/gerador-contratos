package brokers

import "time"

type Broker struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Nome      string    `json:"nome"`
	CPF       string    `json:"cpf,omitempty"`
	CRECI     string    `json:"creci,omitempty"`
	Banco     string    `json:"banco,omitempty"`
	Agencia   string    `json:"agencia,omitempty"`
	Conta     string    `json:"conta,omitempty"`
	Pix       string    `json:"pix,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateBrokerInput struct {
	Nome    string `json:"nome"`
	CPF     string `json:"cpf"`
	CRECI   string `json:"creci"`
	Banco   string `json:"banco"`
	Agencia string `json:"agencia"`
	Conta   string `json:"conta"`
	Pix     string `json:"pix"`
}

type UpdateBrokerInput struct {
	Nome    string `json:"nome"`
	CPF     string `json:"cpf"`
	CRECI   string `json:"creci"`
	Banco   string `json:"banco"`
	Agencia string `json:"agencia"`
	Conta   string `json:"conta"`
	Pix     string `json:"pix"`
}
