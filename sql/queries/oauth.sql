-- name: CreateSession :one
-- Creates a new panel login session for a user.
INSERT INTO discord_oauth_sessions (
    session_hash,
    client_id,
    expires_at,
    access_token,
    refresh_token
)
VALUES (
    digest(@session, 'sha256'),
    @client_id,
    @expiry,
    pgp_sym_encrypt(@access_token, @encryption_key),
    pgp_sym_encrypt(@refresh_token, @encryption_key)
)
RETURNING 
    created_at,
    updated_at,
    expires_at,
    revoked_at,
    client_id,
    pgp_sym_decrypt(access_token, @encryption_key)::TEXT AS access_token,
    pgp_sym_decrypt(refresh_token, @encryption_key)::TEXT AS refresh_token;

-- name: ValidateSession :one
-- Validates a session of an existing login for a user.
SELECT
    created_at,
    updated_at,
    expires_at,
    revoked_at,
    client_id,
    pgp_sym_decrypt(access_token, @encryption_key)::TEXT AS access_token,
    pgp_sym_decrypt(refresh_token, @encryption_key)::TEXT AS refresh_token
FROM discord_oauth_sessions
WHERE
    session_hash = digest(@session, 'sha256')
    AND revoked_at IS NULL
LIMIT 1;

-- name: UpdateSession :one
-- Updates the information for a specific session
UPDATE discord_oauth_sessions
SET 
    updated_at = now(),
    expires_at = @expiry,
    access_token = pgp_sym_encrypt(@access_token, @encryption_key),
    refresh_token = pgp_sym_encrypt(@refresh_token, @encryption_key)
WHERE
    session_hash = digest(@session, 'sha256')
    AND revoked_at IS NULL
RETURNING
    created_at,
    updated_at,
    expires_at,
    revoked_at,
    client_id,
    pgp_sym_decrypt(access_token, @encryption_key)::TEXT AS access_token,
    pgp_sym_decrypt(refresh_token, @encryption_key)::TEXT AS refresh_token;

-- name: RevokeSession :exec
-- Revokes an existing session.
UPDATE discord_oauth_sessions
SET
    updated_at = now(),
    revoked_at = now()
WHERE
    session_hash = digest(@session, 'sha256');

-- name: DeleteExpiredSessions :exec
-- Deletes all expired sessions.
DELETE FROM discord_oauth_sessions
WHERE expires_at <= now();