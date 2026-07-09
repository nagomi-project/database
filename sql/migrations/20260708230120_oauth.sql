-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS discord_oauth_sessions (
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,

    session_hash BYTEA NOT NULL,
    client_id SNOWFLAKE NOT NULL,
    access_token BYTEA NOT NULL,
    refresh_token BYTEA NOT NULL,

    PRIMARY KEY (session_hash)
);

-- +goose Down
DROP TABLE IF EXISTS discord_oauth_sessions;

DROP EXTENSION IF EXISTS pgcrypto;
