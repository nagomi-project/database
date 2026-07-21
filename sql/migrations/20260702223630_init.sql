-- +goose Up

CREATE DOMAIN SNOWFLAKE AS TEXT CHECK (VALUE ~ '^[0-9]{17,}$'); -- This is used to ensure that Discord IDs are actually snowflakes.

CREATE TABLE IF NOT EXISTS guilds_registry (
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    guild_id SNOWFLAKE NOT NULL,

    PRIMARY KEY (guild_id)
);

CREATE TYPE guild_module_type AS ENUM (
    'infractions',
    'ban_appeals',

    'event_logs',
    'tickets',
    'mod_mail',

    'voice_rooms',
    'activity_tracking'
);

CREATE TABLE IF NOT EXISTS guild_modules (
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    guild_id SNOWFLAKE NOT NULL,
    module_type guild_module_type NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,

    PRIMARY KEY (guild_id, module_type),
    FOREIGN KEY (guild_id)
        REFERENCES guilds_registry (guild_id)
        ON DELETE CASCADE
);

CREATE TYPE action_log_type AS ENUM (
    'guild_settings_update'
);

CREATE TYPE action_log_source AS ENUM (
    -- Action taken through the website panel
    'panel',
    -- Action taken through the bot directly (interactions, commands, etc)
    'discord'
);

-- Used to keep track of action log ids while also staying concurrency safe.
CREATE TABLE IF NOT EXISTS next_log_ids (
    guild_id SNOWFLAKE NOT NULL,
    next_id BIGINT NOT NULL DEFAULT 1,

    PRIMARY KEY (guild_id),
    FOREIGN KEY (guild_id) REFERENCES guilds_registry (guild_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS action_logs (
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    id BIGINT NOT NULL,
    guild_id SNOWFLAKE NOT NULL,
    actor_id SNOWFLAKE NOT NULL,

    type action_log_type NOT NULL,
    source action_log_source NOT NULL,
    action TEXT NOT NULL,

    PRIMARY KEY (guild_id, id),
    FOREIGN KEY (guild_id) REFERENCES guilds_registry (guild_id) ON DELETE CASCADE
);

CREATE TYPE event_log_type AS ENUM (
    'user_join',
    'user_leave',
    'user_kick',
    'user_ban',
    'user_unban',
    'user_timeout',
    'user_roles_update',
    'user_nickname_update',
    'user_voice_join',
    'user_voice_move',
    'user_voice_leave',

    'message_edit',
    'message_delete',
    'message_image_remove',

    'channel_create',
    'channel_update',
    'channel_delete',

    'role_create',
    'role_update',
    'role_delete',

    'emoji_create',
    'emoji_update',
    'emoji_delete'
);

CREATE TABLE IF NOT EXISTS event_log_settings (
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    guild_id SNOWFLAKE NOT NULL,

    ignored_channels SNOWFLAKE[],
    ignore_roles SNOWFLAKE[],

    PRIMARY KEY (guild_id),
    FOREIGN KEY (guild_id) REFERENCES guilds_registry (guild_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS event_log_channels (
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    type event_log_type NOT NULL,
    guild_id SNOWFLAKE NOT NULL,
    channel_id SNOWFLAKE NOT NULL,

    PRIMARY KEY (guild_id, type),
    FOREIGN KEY (guild_id) REFERENCES guilds_registry (guild_id) ON DELETE CASCADE
);

-- +goose Down

DROP TABLE IF EXISTS action_logs;
DROP TABLE IF EXISTS next_log_ids;
DROP TABLE IF EXISTS event_log_channels;
DROP TABLE IF EXISTS guilds_registry;

DROP TYPE IF EXISTS action_log_source;
DROP TYPE IF EXISTS action_log_type;
DROP TYPE IF EXISTS log_channel_type;

DROP DOMAIN IF EXISTS SNOWFLAKE;
