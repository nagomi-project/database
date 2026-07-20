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
