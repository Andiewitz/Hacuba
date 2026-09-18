-- 001_init.sql — auth-db schema.
--
-- OWNERSHIP: this database is private to the auth service.
-- Only the auth service role connects (see docker-compose.yml).
-- No other service gets credentials or network routes to auth-db.
--
-- UUIDs are generated in Go (uuid v7) and sent as primary keys —
-- Postgres never mints identities, it only enforces uniqueness.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT email_lowercase CHECK (email = lower(email))
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

CREATE TABLE IF NOT EXISTS refresh_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- SHA-256 hex of the opaque refresh token. The raw token only ever
    -- lives in the HttpOnly cookie + server memory, never in the DB.
    token_hash TEXT NOT NULL UNIQUE,
    -- SHA-256 hex of the CSRF token bound to this session.
    csrf_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_refresh_sessions_user_id ON refresh_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_sessions_expires_at ON refresh_sessions (expires_at);

-- Lock down: revoke default public access. The migration runner / auth
-- role is the only writer. Adjust role name to match POSTGRES_USER.
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;
