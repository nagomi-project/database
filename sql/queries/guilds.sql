-- name: RegisterGuildIfMissing :exec
-- Inserts a guild into the registy if it is not already registered.
INSERT INTO guilds_registry (guild_id)
VALUES (@guild_id)
ON CONFLICT (guild_id) DO NOTHING;

-- name: GetRegisteredGuild :one
-- Fetches a guild's registry information.
SELECT * FROM guilds_registry
WHERE
    guild_id = @guild_id;

-- name: UpdateGuildRegistryTime :exec
-- Updates when a guild settings were modified.
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