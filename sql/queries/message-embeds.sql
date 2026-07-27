-- name: RegisterMessageEmbedSettingsIfMissing :exec
-- Inserts message embed settings if they are not already created.
INSERT INTO message_embed_settings (guild_id)
VALUES (@guild_id)
ON CONFLICT (guild_id) DO NOTHING;

-- name: UpsertMessageEmbedSettings :one
INSERT INTO message_embed_settings (
    guild_id,
    ignored_channels,
    ignored_roles,
    allow_cross_guild_embeds
)
VALUES (
    @guild_id,
    @ignored_channels,
    @ignored_roles,
    @allow_cross_guild_embeds
)
ON CONFLICT (guild_id) DO UPDATE SET
    updated_at = now(),
    ignored_channels = EXCLUDED.ignored_channels,
    ignored_roles = EXCLUDED.ignored_roles,
    allow_cross_guild_embeds = EXCLUDED.allow_cross_guild_embeds
RETURNING *;

-- name: GetMessageEmbedSettings :one
SELECT * FROM message_embed_settings
WHERE
    guild_id = @guild_id;