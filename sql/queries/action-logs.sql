-- name: LogAction :one
-- Logs an action that was done.
WITH next_log AS (
    INSERT INTO next_log_ids (guild_id, next_id)
    VALUES (@guild_id, 2)
    ON CONFLICT (guild_id) DO UPDATE SET
        next_id = next_log_ids.next_id + 1
    RETURNING next_id - 1 AS id
)
INSERT INTO action_logs (
    id,
    guild_id,
    actor_id,
    type,
    source,
    action
)
SELECT
    next_log.id,
    @guild_id,
    @actor_id,
    @action_type,
    @action_source,
    @action_reason
FROM next_log
RETURNING *;

-- name: GetActionLogPage :many
-- Gets the most recent actions done in the server.
SELECT * FROM action_logs
WHERE
    guild_id = @guild_id
ORDER BY
    created_at DESC
OFFSET (GREATEST(@page::SMALLINT, 1) - 1) * COALESCE(sqlc.narg('page_size'), 5)
LIMIT COALESCE(sqlc.narg('page_size'), 5);

-- name: GetActionLogPageDetails :one
-- Fetches pagination details for an action log page.
SELECT
    COUNT(*)::INTEGER AS total_entries,
    CEIL(COUNT(*)::NUMERIC / COALESCE(sqlc.narg('page_size'), 5))::INTEGER AS total_pages
FROM action_logs
WHERE
    guild_id = @guild_id;
