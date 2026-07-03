-- +goose Up

CREATE DOMAIN SNOWFLAKE AS TEXT CHECK (VALUE ~ '^[0-9]{17,}$'); -- This is used to ensure that Discord IDs are actually snowflakes.

CREATE TABLE IF NOT EXISTS guilds_registry (
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    guild_id SNOWFLAKE NOT NULL,

    PRIMARY KEY (guild_id)
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

CREATE TYPE log_channel_type AS ENUM (
    -- All events that take place in the server.
    --
    -- If there is a channel set for certain events,
    -- it will not be sent to the "all" channel.
    'all',
    -- Message related events that take place in the server.
    -- deleted, edited
    'message',
    -- User / Member related events that take place in the server.
    -- join, leave, ban, unban, kick,
    'user',
    -- Where infractions will be logged.
    -- kick, ban, unban, softban, mute.
    'infraction_log'
);

CREATE TABLE IF NOT EXISTS log_channels (
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    type log_channel_type NOT NULL,
    guild_id SNOWFLAKE NOT NULL,
    channel_id SNOWFLAKE NOT NULL,

    PRIMARY KEY (guild_id, type),
    FOREIGN KEY (guild_id) REFERENCES guilds_registry (guild_id) ON DELETE CASCADE
);

-- +goose Down

DROP TABLE IF EXISTS action_logs;
DROP TABLE IF EXISTS next_log_ids;
DROP TABLE IF EXISTS log_channels;
DROP TABLE IF EXISTS guilds_registry;

DROP TYPE IF EXISTS action_log_source;
DROP TYPE IF EXISTS action_log_type;
DROP TYPE IF EXISTS log_channel_type;

DROP DOMAIN IF EXISTS SNOWFLAKE;
