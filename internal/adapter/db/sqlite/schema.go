package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type migration struct {
	version int
	name    string
	sql     string
}

func runMigrations(ctx context.Context, q queryer) error {
	pragmas := `
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = -2000;
PRAGMA temp_store = MEMORY;
`
	if _, err := q.ExecContext(ctx, pragmas); err != nil {
		return fmt.Errorf("failed to apply pragmas: %w", err)
	}

	createMeta := `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at DATETIME NOT NULL
);
`
	if _, err := q.ExecContext(ctx, createMeta); err != nil {
		return fmt.Errorf("failed to create schema_migrations: %w", err)
	}

	applied := make(map[int]bool)
	rows, err := q.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to query schema_migrations: %w", err)
	}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var v int
			if err := rows.Scan(&v); err == nil {
				applied[v] = true
			}
		}
	}

	for _, m := range migrations {
		if applied[m.version] {
			continue
		}
		if _, err := q.ExecContext(ctx, m.sql); err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w", m.version, m.name, err)
		}
		recordSQL := "INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)"
		if _, err := q.ExecContext(ctx, recordSQL, m.version, m.name, time.Now().UTC()); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", m.version, err)
		}
	}

	return nil
}
