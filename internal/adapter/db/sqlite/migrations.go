package sqlite

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
	{
		version: 4,
		name:    "add_service_project_name_and_deploy_script",
		sql: `
ALTER TABLE services ADD COLUMN project_name TEXT NOT NULL DEFAULT '';
ALTER TABLE services ADD COLUMN deploy_script TEXT NOT NULL DEFAULT '';
`,
	},
	{
		version: 5,
		name:    "add_database_config_and_redirects",
		sql: `
ALTER TABLE services ADD COLUMN database_config TEXT NOT NULL DEFAULT '{}';
ALTER TABLE services ADD COLUMN redirects TEXT NOT NULL DEFAULT '[]';
`,
	},
	{
		version: 6,
		name:    "add_actions_storage_providers_and_service_options",
		sql: `
ALTER TABLE services ADD COLUMN primary_domain_id TEXT NOT NULL DEFAULT '';
ALTER TABLE services ADD COLUMN zero_downtime INTEGER NOT NULL DEFAULT 1;

CREATE TABLE IF NOT EXISTS actions (
    id TEXT PRIMARY KEY,
    project_name TEXT NOT NULL DEFAULT '',
    service_name TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    description TEXT NOT NULL DEFAULT '',
    no_kill INTEGER NOT NULL DEFAULT 0,
    no_logs INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    user_id TEXT NOT NULL DEFAULT '',
    is_api_action INTEGER NOT NULL DEFAULT 0,
    is_system_action INTEGER NOT NULL DEFAULT 0,
    meta TEXT NOT NULL DEFAULT '{}'
);
CREATE INDEX IF NOT EXISTS idx_actions_project_service ON actions (project_name, service_name);
CREATE INDEX IF NOT EXISTS idx_actions_type ON actions (type);

CREATE TABLE IF NOT EXISTS storage_providers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'local',
    path TEXT NOT NULL DEFAULT '',
    endpoint TEXT NOT NULL DEFAULT '',
    bucket TEXT NOT NULL DEFAULT '',
    region TEXT NOT NULL DEFAULT '',
    access_key TEXT NOT NULL DEFAULT '',
    secret_key TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
`,
	},
	{
		version: 7,
		name:    "add_domain_project_and_service_names",
		sql: `
ALTER TABLE domains ADD COLUMN project_name TEXT NOT NULL DEFAULT '';
ALTER TABLE domains ADD COLUMN service_name TEXT NOT NULL DEFAULT '';
`,
	},
	{
		version: 8,
		name:    "add_project_env_and_service_notes_error",
		sql: `
ALTER TABLE projects ADD COLUMN env TEXT NOT NULL DEFAULT '';
ALTER TABLE services ADD COLUMN notes TEXT NOT NULL DEFAULT '';
ALTER TABLE services ADD COLUMN last_error TEXT NOT NULL DEFAULT '';
`,
	},
}
