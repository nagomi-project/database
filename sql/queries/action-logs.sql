-- name: LogAction :one
-- Logs an action that was done.
WITH next_log AS (
    INSERT INTO next_log_ids (guild_id, id)
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
    next_log.next_id,
    @guild_id,
    @actor_id,
    @action_type,
    @action_source,
    @action_reason
FROM next_log
RETURNING *;