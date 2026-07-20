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

-- name: FindRegisteredGuilds :many
-- Fetches the guilds that are registered from a provided list.
SELECT * FROM guilds_registry
WHERE
    guild_id = ANY(@guild_ids::TEXT[]);

-- name: UpdateGuildRegistryTime :exec
-- Updates when a guild settings were modified.
INSERT INTO guilds_registry (guild_id)
VALUES (@guild_id)
ON CONFLICT (guild_id) DO
UPDATE SET
    updated_at = now()
RETURNING *;

-- name: RemoveLogChannel :one
-- Removes an existing log channel.
DELETE FROM event_log_channels
WHERE
    type = @type
    AND guild_id = @guild_id
RETURNING *;