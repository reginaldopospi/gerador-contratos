package contracts

import "time"

type Contract struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Numero    string    `json:"numero"`
	Tipo      string    `json:"tipo"`
	Status    string    `json:"status"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ContractVersion struct {
	ID            string         `json:"id"`
	ContractID    string         `json:"contract_id"`
	VersionNumber int            `json:"version_number"`
	Data          map[string]any `json:"data"`
	CreatedBy     string         `json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
}

type ContractPreview struct {
	Title    string   `json:"title"`
	Sections []string `json:"sections"`
	FullText string   `json:"full_text,omitempty"`
}

type CreateContractInput struct {
	Numero string         `json:"numero"`
	Tipo   string         `json:"tipo"`
	Status string         `json:"status"`
	Data   map[string]any `json:"data"`
}

type AddVersionInput struct {
	Data map[string]any `json:"data"`
}

type ListContractsFilter struct {
	Status string
	Query  string
	Limit  int
	Offset int
}

type ContractDetails struct {
	Contract      Contract          `json:"contract"`
	LatestVersion *ContractVersion  `json:"latest_version,omitempty"`
	Versions      []ContractVersion `json:"versions"`
}
