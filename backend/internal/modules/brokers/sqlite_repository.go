package brokers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gerador-contratos/backend/internal/modules/common"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) List(ctx context.Context, tenantID string) ([]Broker, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, nome, COALESCE(cpf,''), COALESCE(creci,''), COALESCE(banco,''), COALESCE(agencia,''), COALESCE(conta,''), COALESCE(pix,''), created_at, updated_at
		FROM brokers
		WHERE tenant_id = ?
		ORDER BY nome ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list brokers: %w", err)
	}
	defer rows.Close()

	result := make([]Broker, 0)
	for rows.Next() {
		var b Broker
		if err := rows.Scan(&b.ID, &b.TenantID, &b.Nome, &b.CPF, &b.CRECI, &b.Banco, &b.Agencia, &b.Conta, &b.Pix, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan broker: %w", err)
		}
		result = append(result, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate brokers: %w", err)
	}

	return result, nil
}

func (r *SQLiteRepository) GetByID(ctx context.Context, tenantID, brokerID string) (*Broker, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, nome, COALESCE(cpf,''), COALESCE(creci,''), COALESCE(banco,''), COALESCE(agencia,''), COALESCE(conta,''), COALESCE(pix,''), created_at, updated_at
		FROM brokers
		WHERE tenant_id = ? AND id = ?
	`, tenantID, brokerID)

	var b Broker
	if err := row.Scan(&b.ID, &b.TenantID, &b.Nome, &b.CPF, &b.CRECI, &b.Banco, &b.Agencia, &b.Conta, &b.Pix, &b.CreatedAt, &b.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewNotFound("broker_not_found", "corretor nao encontrado")
		}
		return nil, fmt.Errorf("get broker: %w", err)
	}
	return &b, nil
}

func (r *SQLiteRepository) Create(ctx context.Context, broker Broker) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO brokers (id, tenant_id, nome, cpf, creci, banco, agencia, conta, pix)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, broker.ID, broker.TenantID, broker.Nome, nullIfEmpty(broker.CPF), nullIfEmpty(broker.CRECI), nullIfEmpty(broker.Banco), nullIfEmpty(broker.Agencia), nullIfEmpty(broker.Conta), nullIfEmpty(broker.Pix))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return common.NewConflict("broker_exists", "corretor com este nome ja existe")
		}
		return fmt.Errorf("create broker: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) Update(ctx context.Context, broker Broker) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE brokers
		SET nome = ?, cpf = ?, creci = ?, banco = ?, agencia = ?, conta = ?, pix = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND tenant_id = ?
	`, broker.Nome, nullIfEmpty(broker.CPF), nullIfEmpty(broker.CRECI), nullIfEmpty(broker.Banco), nullIfEmpty(broker.Agencia), nullIfEmpty(broker.Conta), nullIfEmpty(broker.Pix), broker.ID, broker.TenantID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return common.NewConflict("broker_exists", "corretor com este nome ja existe")
		}
		return fmt.Errorf("update broker: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return common.NewNotFound("broker_not_found", "corretor nao encontrado")
	}
	return nil
}

func (r *SQLiteRepository) Delete(ctx context.Context, tenantID, brokerID string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM brokers
		WHERE tenant_id = ? AND id = ?
	`, tenantID, brokerID)
	if err != nil {
		return fmt.Errorf("delete broker: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return common.NewNotFound("broker_not_found", "corretor nao encontrado")
	}
	return nil
}

func nullIfEmpty(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}
