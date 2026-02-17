package clauses

import "time"

type ClauseTemplate struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id,omitempty"`
	ClauseKey string    `json:"clause_key"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"is_active"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpsertClauseInput struct {
	ClauseKey string `json:"clause_key"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	IsActive  bool   `json:"is_active"`
}
