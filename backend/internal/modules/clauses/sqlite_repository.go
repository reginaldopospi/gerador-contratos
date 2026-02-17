package clauses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gerador-contratos/backend/internal/modules/common"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) List(ctx context.Context, tenantID string) ([]ClauseTemplate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(tenant_id,''), clause_key, title, content, is_active, updated_at
		FROM clause_templates
		WHERE tenant_id = ? OR tenant_id IS NULL
		ORDER BY clause_key ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list clauses: %w", err)
	}
	defer rows.Close()

	out := make([]ClauseTemplate, 0)
	for rows.Next() {
		var (
			c ClauseTemplate
			active int
		)
		if err := rows.Scan(&c.ID, &c.TenantID, &c.ClauseKey, &c.Title, &c.Content, &active, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan clause: %w", err)
		}
		c.IsActive = active == 1
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate clauses: %w", err)
	}
	return out, nil
}

func (r *SQLiteRepository) Upsert(ctx context.Context, clause ClauseTemplate) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO clause_templates (id, tenant_id, clause_key, title, content, is_active, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(tenant_id, clause_key)
		DO UPDATE SET title = excluded.title, content = excluded.content, is_active = excluded.is_active, updated_at = CURRENT_TIMESTAMP
	`, clause.ID, nullIfEmpty(clause.TenantID), clause.ClauseKey, clause.Title, clause.Content, boolToInt(clause.IsActive))
	if err != nil {
		return fmt.Errorf("upsert clause: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetByID(ctx context.Context, tenantID, clauseID string) (*ClauseTemplate, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(tenant_id,''), clause_key, title, content, is_active, updated_at
		FROM clause_templates
		WHERE id = ? AND (tenant_id = ? OR tenant_id IS NULL)
	`, clauseID, tenantID)

	var (
		c ClauseTemplate
		active int
	)
	if err := row.Scan(&c.ID, &c.TenantID, &c.ClauseKey, &c.Title, &c.Content, &active, &c.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewNotFound("clause_not_found", "clausula nao encontrada")
		}
		return nil, fmt.Errorf("get clause: %w", err)
	}
	c.IsActive = active == 1
	return &c, nil
}

func (r *SQLiteRepository) Delete(ctx context.Context, tenantID, clauseID string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM clause_templates
		WHERE id = ? AND tenant_id = ?
	`, clauseID, tenantID)
	if err != nil {
		return fmt.Errorf("delete clause: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return common.NewNotFound("clause_not_found", "clausula nao encontrada")
	}
	return nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}
