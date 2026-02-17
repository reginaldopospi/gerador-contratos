package contracts

import (
	"context"
	"database/sql"
	"encoding/json"
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

func (r *SQLiteRepository) CreateContract(ctx context.Context, contract Contract) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO contracts (id, tenant_id, numero, tipo, status, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, contract.ID, contract.TenantID, contract.Numero, contract.Tipo, contract.Status, contract.CreatedBy)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return common.NewConflict("contract_exists", "numero de contrato ja cadastrado")
		}
		return fmt.Errorf("create contract: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetContract(ctx context.Context, tenantID, contractID string) (*Contract, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, numero, tipo, status, created_by, created_at, updated_at
		FROM contracts
		WHERE id = ? AND tenant_id = ?
	`, contractID, tenantID)

	var c Contract
	if err := row.Scan(&c.ID, &c.TenantID, &c.Numero, &c.Tipo, &c.Status, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewNotFound("contract_not_found", "contrato nao encontrado")
		}
		return nil, fmt.Errorf("get contract: %w", err)
	}
	return &c, nil
}

func (r *SQLiteRepository) ListContracts(ctx context.Context, tenantID string, filter ListContractsFilter) ([]Contract, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	query := `
		SELECT id, tenant_id, numero, tipo, status, created_by, created_at, updated_at
		FROM contracts
		WHERE tenant_id = ?
	`
	args := []any{tenantID}

	if strings.TrimSpace(filter.Status) != "" {
		query += ` AND status = ?`
		args = append(args, strings.TrimSpace(filter.Status))
	}
	if strings.TrimSpace(filter.Query) != "" {
		query += ` AND (LOWER(numero) LIKE LOWER(?) OR LOWER(tipo) LIKE LOWER(?))`
		q := "%" + strings.TrimSpace(filter.Query) + "%"
		args = append(args, q, q)
	}

	query += ` ORDER BY updated_at DESC LIMIT ? OFFSET ?`
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list contracts: %w", err)
	}
	defer rows.Close()

	contracts := make([]Contract, 0)
	for rows.Next() {
		var c Contract
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Numero, &c.Tipo, &c.Status, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan contract: %w", err)
		}
		contracts = append(contracts, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate contracts: %w", err)
	}

	return contracts, nil
}

func (r *SQLiteRepository) CreateVersion(ctx context.Context, version ContractVersion) error {
	payload, err := json.Marshal(version.Data)
	if err != nil {
		return fmt.Errorf("marshal version data: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO contract_versions (id, contract_id, version_number, data_json, created_by)
		VALUES (?, ?, ?, ?, ?)
	`, version.ID, version.ContractID, version.VersionNumber, string(payload), version.CreatedBy)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return common.NewConflict("version_exists", "versao de contrato ja existe")
		}
		return fmt.Errorf("create version: %w", err)
	}

	_, _ = r.db.ExecContext(ctx, `UPDATE contracts SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, version.ContractID)
	return nil
}

func (r *SQLiteRepository) GetLatestVersion(ctx context.Context, contractID string) (*ContractVersion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, contract_id, version_number, data_json, created_by, created_at
		FROM contract_versions
		WHERE contract_id = ?
		ORDER BY version_number DESC
		LIMIT 1
	`, contractID)

	return scanVersionRow(row)
}

func (r *SQLiteRepository) GetVersion(ctx context.Context, contractID string, versionNumber int) (*ContractVersion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, contract_id, version_number, data_json, created_by, created_at
		FROM contract_versions
		WHERE contract_id = ? AND version_number = ?
	`, contractID, versionNumber)

	return scanVersionRow(row)
}

func (r *SQLiteRepository) ListVersions(ctx context.Context, contractID string) ([]ContractVersion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, contract_id, version_number, data_json, created_by, created_at
		FROM contract_versions
		WHERE contract_id = ?
		ORDER BY version_number DESC
	`, contractID)
	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	defer rows.Close()

	versions := make([]ContractVersion, 0)
	for rows.Next() {
		var (
			v         ContractVersion
			dataJSON  string
		)
		if err := rows.Scan(&v.ID, &v.ContractID, &v.VersionNumber, &dataJSON, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan version: %w", err)
		}
		if err := json.Unmarshal([]byte(dataJSON), &v.Data); err != nil {
			return nil, fmt.Errorf("unmarshal version data: %w", err)
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate versions: %w", err)
	}
	return versions, nil
}

func scanVersionRow(row *sql.Row) (*ContractVersion, error) {
	var (
		v        ContractVersion
		dataJSON string
	)
	if err := row.Scan(&v.ID, &v.ContractID, &v.VersionNumber, &dataJSON, &v.CreatedBy, &v.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewNotFound("contract_version_not_found", "versao nao encontrada")
		}
		return nil, fmt.Errorf("scan version: %w", err)
	}
	if err := json.Unmarshal([]byte(dataJSON), &v.Data); err != nil {
		return nil, fmt.Errorf("unmarshal version data: %w", err)
	}
	return &v, nil
}
