package clauses

import "context"

type Repository interface {
	List(ctx context.Context, tenantID string) ([]ClauseTemplate, error)
	Upsert(ctx context.Context, clause ClauseTemplate) error
	GetByID(ctx context.Context, tenantID, clauseID string) (*ClauseTemplate, error)
	Delete(ctx context.Context, tenantID, clauseID string) error
}
