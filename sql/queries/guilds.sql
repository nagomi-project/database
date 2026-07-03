-- name: UpsertGuildRegistry :one
-- Crates a new guild registry entry or updates the time of an existing one.
INSERT INTO guilds_registry (guild_id)
VALUES (@guild_id)
ON CONFLICT (guild_id) DO
UPDATE SET
    updated_at = now()
RETURNING *;

-- name: UpsertLogChannel :one
-- Creates a new log channel or modifies the id of an existing one.
INSERT INTO log_channels (type, guild_id, channel_id)
VALUES (@type, @guild_id, @channel_id)
ON CONFLICT (guild_id, type) DO UPDATE SET
    updated_at = now(),
    channel_id = @channel_id
RETURNING *;

-- name: RemoveLogChannel :one
-- Removes an existing log channel.
DELETE FROM log_channels
WHERE
    type = @type
    AND guild_id = @guild_id
RETURNING *;