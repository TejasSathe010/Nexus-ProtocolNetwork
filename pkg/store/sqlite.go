package store

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return nil, fmt.Errorf("enable foreign_keys: %w", err)
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS tenants (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			secret TEXT NOT NULL UNIQUE,
			label TEXT,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY(tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS routes (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			match_type TEXT NOT NULL,    -- EXACT or PREFIX
			match_value TEXT NOT NULL,   -- event type or prefix
			target_channel TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY(tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		);`,
	}

	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	return nil
}
