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

// migrations list registered in sequential order.
var migrations = []migration{
	{
		version: 1,
		name:    "initial_core_schema",
		sql: `
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS services (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    deploy_token TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT 'image',
    source_config TEXT NOT NULL DEFAULT '{}',
    image TEXT NOT NULL DEFAULT '',
    command TEXT NOT NULL DEFAULT '',
    args TEXT NOT NULL DEFAULT '[]',
    env_vars TEXT NOT NULL DEFAULT '[]',
    ports TEXT NOT NULL DEFAULT '[]',
    volumes TEXT NOT NULL DEFAULT '[]',
    domains TEXT NOT NULL DEFAULT '[]',
    replicas INTEGER NOT NULL DEFAULT 1,
    cpu_limit REAL NOT NULL DEFAULT 0.0,
    memory_limit INTEGER NOT NULL DEFAULT 0,
    restart_policy TEXT NOT NULL DEFAULT 'unless-stopped',
    health_check TEXT NOT NULL DEFAULT '{}',
    cron_jobs TEXT NOT NULL DEFAULT '[]',
    labels TEXT NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'stopped',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(project_id, name)
);

CREATE INDEX IF NOT EXISTS idx_services_project_id ON services(project_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_services_deploy_token ON services(deploy_token) WHERE deploy_token != '';

CREATE TABLE IF NOT EXISTS domains (
    id TEXT PRIMARY KEY,
    service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    domain_name TEXT NOT NULL UNIQUE,
    port INTEGER NOT NULL,
    path TEXT NOT NULL DEFAULT '',
    https INTEGER NOT NULL DEFAULT 1,
    cert_mode TEXT NOT NULL DEFAULT 'letsencrypt',
    middlewares TEXT NOT NULL DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_domains_service_id ON domains(service_id);

CREATE TABLE IF NOT EXISTS deployments (
    id TEXT PRIMARY KEY,
    service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    trigger TEXT NOT NULL DEFAULT 'manual',
    commit_hash TEXT NOT NULL DEFAULT '',
    commit_message TEXT NOT NULL DEFAULT '',
    logs TEXT NOT NULL DEFAULT '',
    started_at DATETIME NOT NULL,
    finished_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_deployments_service_id ON deployments(service_id);

CREATE TABLE IF NOT EXISTS backups (
    id TEXT PRIMARY KEY,
    service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    file_name TEXT NOT NULL,
    size_bytes INTEGER NOT NULL DEFAULT 0,
    started_at DATETIME NOT NULL,
    finished_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_backups_service_id ON backups(service_id);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'viewer',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    val TEXT NOT NULL
);
`,
	},
	{
		version: 2,
		name:    "add_backups_and_cron_jobs",
		sql: `
CREATE TABLE IF NOT EXISTS backups (
    id TEXT PRIMARY KEY,
    service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    file_name TEXT NOT NULL,
    size_bytes INTEGER NOT NULL DEFAULT 0,
    started_at DATETIME NOT NULL,
    finished_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_backups_service_id ON backups(service_id);
`,
	},
	{
		version: 3,
		name:    "add_sessions",
		sql: `
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);
`,
	},
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
