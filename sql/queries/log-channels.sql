-- name: GetGuildLogChannels :many
-- Fetch all of the log channels for the guild.
SELECT * FROM log_channels
WHERE
    guild_id = @guild_id;