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
