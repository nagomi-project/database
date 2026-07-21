-- name: RegisterEventLogSettingsIfMissing :exec
-- Inserts guild event log settings if they are not already created.
INSERT INTO event_log_settings (guild_id)
VALUES (@guild_id)
ON CONFLICT (guild_id) DO NOTHING;

-- name: GetEventLogSettings :one
-- Fetch the options for the event logs.
SELECT * FROM event_log_settings
WHERE
    guild_id = @guild_id;

-- name: GetEventLogChannels :many
-- Fetch all of the log channels for the guild.
SELECT * FROM event_log_channels
WHERE
    guild_id = @guild_id;

-- name: UpsertLogChannel :one
-- Creates a new log channel or modifies the id of an existing one.
INSERT INTO event_log_channels (type, guild_id, channel_id)
VALUES (@type, @guild_id, @channel_id)
ON CONFLICT (guild_id, type) DO UPDATE SET
    updated_at = now(),
    channel_id = @channel_id
RETURNING *;

-- name: UpsertManyLogChannels :many
-- Creates a new log channel or modifies the id of existing ones.
INSERT INTO event_log_channels (type, guild_id, channel_id)
SELECT types.type, @guild_id, channels.channel_id
FROM unnest(CAST(@types AS event_log_type[]))
    WITH ORDINALITY AS types(type, position)
JOIN unnest(CAST(@channel_ids AS SNOWFLAKE[]))
    WITH ORDINALITY AS channels(channel_id, position)
    USING (position)
ON CONFLICT (guild_id, type) DO UPDATE SET
    updated_at = now(),
    channel_id = EXCLUDED.channel_id
RETURNING *;
