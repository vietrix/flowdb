-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	is_admin BOOLEAN NOT NULL DEFAULT FALSE,
	mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
	mfa_secret_enc BYTEA,
	mfa_verified_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS external_identities (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	provider TEXT NOT NULL,
	subject TEXT NOT NULL,
	email TEXT,
	groups JSONB,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE(provider, subject)
);

CREATE TABLE IF NOT EXISTS groups (
	id UUID PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS group_members (
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
	PRIMARY KEY(user_id, group_id)
);

CREATE TABLE IF NOT EXISTS roles (
	id UUID PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	permissions JSONB NOT NULL DEFAULT '[]',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS role_bindings (
	id UUID PRIMARY KEY,
	role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
	user_id UUID REFERENCES users(id) ON DELETE CASCADE,
	group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
	resource TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS policies (
	id UUID PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	doc JSONB NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS settings (
	id INT PRIMARY KEY,
	security_mode TEXT NOT NULL,
	flags JSONB NOT NULL,
	config JSONB NOT NULL,
	ip_allowlist TEXT[] NOT NULL DEFAULT '{}',
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS connection_secrets (
	id UUID PRIMARY KEY,
	encrypted BYTEA NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS connections (
	id UUID PRIMARY KEY,
	name TEXT NOT NULL,
	type TEXT NOT NULL,
	host TEXT NOT NULL,
	port INT NOT NULL,
	database TEXT NOT NULL,
	username TEXT NOT NULL,
	secret_ref UUID NOT NULL REFERENCES connection_secrets(id) ON DELETE RESTRICT,
	tls JSONB NOT NULL DEFAULT '{}'::jsonb,
	tags JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS query_history (
	id UUID PRIMARY KEY,
	user_id UUID REFERENCES users(id) ON DELETE SET NULL,
	connection_id UUID REFERENCES connections(id) ON DELETE SET NULL,
	statement_hash TEXT NOT NULL,
	status TEXT NOT NULL,
	row_count INT NOT NULL DEFAULT 0,
	duration_ms BIGINT NOT NULL DEFAULT 0,
	started_at TIMESTAMPTZ NOT NULL,
	ended_at TIMESTAMPTZ,
	action TEXT,
	resource TEXT,
	approval_id UUID
);

CREATE TABLE IF NOT EXISTS audit_log (
	id UUID PRIMARY KEY,
	event_type TEXT NOT NULL,
	actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
	details JSONB,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	prev_hash TEXT,
	hash TEXT,
	error_id TEXT
);

CREATE TABLE IF NOT EXISTS query_approvals (
	id UUID PRIMARY KEY,
	connection_id UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	statement TEXT NOT NULL,
	status TEXT NOT NULL,
	environment TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	approved_by UUID REFERENCES users(id),
	approved_at TIMESTAMPTZ,
	denied_by UUID REFERENCES users(id),
	denied_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS pii_rules (
	id UUID PRIMARY KEY,
	connection_id UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
	resource TEXT NOT NULL,
	field TEXT NOT NULL,
	mask_type TEXT NOT NULL DEFAULT 'mask',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	csrf_token TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	expires_at TIMESTAMPTZ NOT NULL,
	last_auth_at TIMESTAMPTZ NOT NULL,
	last_mfa_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS oidc_states (
	state TEXT PRIMARY KEY,
	nonce TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS oidc_states;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS pii_rules;
DROP TABLE IF EXISTS query_approvals;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS query_history;
DROP TABLE IF EXISTS connections;
DROP TABLE IF EXISTS connection_secrets;
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS policies;
DROP TABLE IF EXISTS role_bindings;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS external_identities;
DROP TABLE IF EXISTS users;
