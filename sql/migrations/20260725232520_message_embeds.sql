-- +goose Up

ALTER TYPE guild_module_type
ADD VALUE IF NOT EXISTS 'message_embeds';

CREATE TABLE IF NOT EXISTS message_embed_settings (
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    guild_id SNOWFLAKE NOT NULL,

    ignored_channels SNOWFLAKE[],
    ignored_roles SNOWFLAKE[],

    -- If enabled, this will allow other guilds to embed messages and
    -- vice versa, assuming the bot shares the server.
    -- It still will ignore embedding from channels from the shared
    -- guild.
    allow_cross_guild_embeds BOOLEAN NOT NULL DEFAULT FALSE,

    PRIMARY KEY (guild_id),
    FOREIGN KEY (guild_id)
        REFERENCES guilds_registry (guild_id)
        ON DELETE CASCADE
);

-- +goose Down

-- ALTER TYPE guild_module_type
-- REMOVE VALUE IF EXISTS 'message_embeds';

DROP TABLE IF EXISTS message_embed_settings;