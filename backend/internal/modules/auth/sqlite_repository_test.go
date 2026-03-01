package auth

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func newSQLiteRepositoryTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	// Keep schema minimal for repository unit tests.
	schema := `
		PRAGMA foreign_keys = ON;
		CREATE TABLE tenants (
			id TEXT PRIMARY KEY,
			nome_fantasia TEXT NOT NULL,
			cnpj TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestGetUserByEmailCaseInsensitiveLookup(t *testing.T) {
	db := newSQLiteRepositoryTestDB(t)
	repo := NewSQLiteRepository(db)

	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `
		INSERT INTO tenants (id, nome_fantasia, cnpj)
		VALUES (?, ?, ?)
	`, "tenant-1", "Imobiliaria Teste", ""); err != nil {
		t.Fatalf("insert tenant: %v", err)
	}

	// Insert legacy user with uppercase email to validate case-insensitive search.
	if _, err := db.ExecContext(ctx, `
		INSERT INTO users (id, tenant_id, name, email, password_hash, role, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "user-1", "tenant-1", "Admin", "ADMIN@TESTE.COM", "hash", "admin", 1); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	user, err := repo.GetUserByEmail(ctx, "admin@teste.com")
	if err != nil {
		t.Fatalf("get user by email: %v", err)
	}
	if user.ID != "user-1" {
		t.Fatalf("unexpected user id: %s", user.ID)
	}
}

func TestListTenantsReturnsSummaryWithUserCounters(t *testing.T) {
	db := newSQLiteRepositoryTestDB(t)
	repo := NewSQLiteRepository(db)

	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `
		INSERT INTO tenants (id, nome_fantasia, cnpj, created_at)
		VALUES
			('tenant-1', 'Imobiliaria A', '11.111.111/0001-11', '2026-02-01 10:00:00'),
			('tenant-2', 'Imobiliaria B', '22.222.222/0001-22', '2026-02-02 10:00:00')
	`); err != nil {
		t.Fatalf("insert tenants: %v", err)
	}

	// Cobre combinacao de usuario admin, ativo e inativo para validar agregacoes.
	if _, err := db.ExecContext(ctx, `
		INSERT INTO users (id, tenant_id, name, email, password_hash, role, is_active)
		VALUES
			('user-1', 'tenant-1', 'Admin A', 'admin-a@imob.com', 'hash', 'admin', 1),
			('user-2', 'tenant-1', 'Operador A', 'operador-a@imob.com', 'hash', 'operador', 0),
			('user-3', 'tenant-2', 'Admin B', 'admin-b@imob.com', 'hash', 'admin', 1)
	`); err != nil {
		t.Fatalf("insert users: %v", err)
	}

	items, err := repo.ListTenants(ctx)
	if err != nil {
		t.Fatalf("list tenants: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 tenants, got %d", len(items))
	}

	// A ordenacao deve respeitar created_at DESC para manter os mais novos no topo.
	if items[0].TenantID != "tenant-2" {
		t.Fatalf("expected tenant-2 first, got %s", items[0].TenantID)
	}
	if items[1].TenantID != "tenant-1" {
		t.Fatalf("expected tenant-1 second, got %s", items[1].TenantID)
	}

	if items[0].AdminEmail != "admin-b@imob.com" {
		t.Fatalf("unexpected admin email for tenant-2: %s", items[0].AdminEmail)
	}
	if items[0].TotalUsers != 1 || items[0].ActiveUsers != 1 {
		t.Fatalf("unexpected counters for tenant-2: total=%d active=%d", items[0].TotalUsers, items[0].ActiveUsers)
	}

	if items[1].AdminEmail != "admin-a@imob.com" {
		t.Fatalf("unexpected admin email for tenant-1: %s", items[1].AdminEmail)
	}
	if items[1].TotalUsers != 2 || items[1].ActiveUsers != 1 {
		t.Fatalf("unexpected counters for tenant-1: total=%d active=%d", items[1].TotalUsers, items[1].ActiveUsers)
	}
}
