-- +goose Up

CREATE TABLE IF NOT EXISTS infraction_settings (
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    guild_id SNOWFLAKE NOT NULL,

    muted_role_id SNOWFLAKE,
    appeal_duration SMALLINT NOT NULL DEFAULT 31, -- this is in days

    infraction_proof_id SNOWFLAKE,
    request_infraction_proof BOOLEAN NOT NULL DEFAULT FALSE,

    PRIMARY KEY (guild_id),
    FOREIGN KEY (guild_id) REFERENCES guilds_registry (guild_id) ON DELETE CASCADE
);

CREATE TYPE infraction_action AS ENUM (
    'note',
    'warn',
    'mute',
    'unmute',
    'kick',
    'ban',
    'unban',
    'softban'
);

CREATE TABLE IF NOT EXISTS next_infraction_ids (
    guild_id SNOWFLAKE NOT NULL,
    next_id BIGINT NOT NULL DEFAULT 1,

    PRIMARY KEY (guild_id),
    FOREIGN KEY (guild_id) REFERENCES guilds_registry (guild_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS infraction_log (
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,

    guild_id SNOWFLAKE NOT NULL,
    case_number INT NOT NULL,
    member_id SNOWFLAKE NOT NULL,
    moderator_id SNOWFLAKE NOT NULL,

    hidden BOOLEAN NOT NULL DEFAULT FALSE,
    action infraction_action NOT NULL,
    reason VARCHAR(512),

    PRIMARY KEY (guild_id, case_number),
    FOREIGN KEY (guild_id) REFERENCES guilds_registry (guild_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS infraction_proof_messages (
    guild_id SNOWFLAKE NOT NULL,
    case_number INT NOT NULL,
    message_url TEXT NOT NULL,

    PRIMARY KEY (guild_id, case_number),
    FOREIGN KEY (guild_id, case_number) REFERENCES infraction_log (guild_id, case_number) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS infraction_expiry_schedule (
    expires_at TIMESTAMPTZ,

    guild_id SNOWFLAKE NOT NULL,
    case_number INT NOT NULL,
    member_id SNOWFLAKE NOT NULL,
    action infraction_action NOT NULL,

    PRIMARY KEY (guild_id, case_number),
    UNIQUE (guild_id, member_id, action),
    FOREIGN KEY (guild_id) REFERENCES guilds_registry (guild_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS infraction_active_bans (
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    guild_id SNOWFLAKE NOT NULL,
    case_number INT NOT NULL,
    member_id SNOWFLAKE NOT NULL,

    can_submit_appeal BOOLEAN NOT NULL DEFAULT TRUE,
    appeal_pending BOOLEAN NOT NULL DEFAULT FALSE,
    appealable_at TIMESTAMPTZ NOT NULL,

    PRIMARY KEY (guild_id, case_number),
    UNIQUE (guild_id, member_id),
    FOREIGN KEY (guild_id) REFERENCES guilds_registry (guild_id) ON DELETE CASCADE
);

CREATE TYPE appeal_status AS ENUM (
    'submitted',
    'approved',
    'denied',
    'blocked',
    'unblocked'
);

CREATE TABLE IF NOT EXISTS ban_appeal_logs (
    log_id BIGSERIAL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    guild_id SNOWFLAKE NOT NULL,
    case_number INT NOT NULL,
    actor_id SNOWFLAKE NOT NULL,

    status appeal_status NOT NULL,

    PRIMARY KEY (log_id),
    FOREIGN KEY (guild_id) REFERENCES guilds_registry (guild_id) ON DELETE CASCADE
);

CREATE VIEW infraction_details AS
SELECT
    infraction_log.created_at,
    infraction_log.updated_at,
    infraction_log.expires_at,
    infraction_log.guild_id,
    infraction_log.case_number,
    infraction_log.member_id,
    infraction_log.moderator_id,
    infraction_log.hidden,
    infraction_log.action,
    infraction_log.reason,
    CASE
        WHEN infraction_expiry_schedule.case_number IS NOT NULL THEN TRUE
        ELSE FALSE
    END AS active,
    COALESCE(infraction_active_bans.can_submit_appeal, FALSE) AS appealable,
    infraction_proof_messages.message_url
FROM infraction_log
LEFT JOIN infraction_expiry_schedule ON
    infraction_expiry_schedule.guild_id = infraction_log.guild_id
    AND infraction_expiry_schedule.member_id = infraction_log.member_id
    AND infraction_expiry_schedule.case_number = infraction_log.case_number
LEFT JOIN infraction_proof_messages ON
    infraction_proof_messages.guild_id = infraction_log.guild_id
    AND infraction_proof_messages.case_number = infraction_log.case_number
LEFT JOIN infraction_active_bans ON
    infraction_active_bans.guild_id = infraction_log.guild_id
    AND infraction_active_bans.case_number = infraction_log.case_number;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION infraction_expiry_schedule_events()
RETURNS trigger AS $$
DECLARE
    payload json;
BEGIN
    -- When a new infraction expiry schedule entry is inserted.
    IF TG_OP = 'INSERT' THEN
        payload := json_build_object(
            'inserted', TRUE,
            'expires_at', NEW.expires_at,
            'guild_id', NEW.guild_id,
            'case_number', NEW.case_number,
            'member_id', NEW.member_id,
            'action', NEW.action
        );

        PERFORM pg_notify('infraction_expiry_schedule_events', payload::text);
        RETURN NEW;
    END IF;

    -- When the details of an infraction expiry schedule entry are updated.
    -- Will only really happen if the expires_at field is updated.
    IF TG_OP = 'UPDATE' THEN
        payload := json_build_object(
            'updated', TRUE,
            'expires_at', NEW.expires_at,
            'guild_id', NEW.guild_id,
            'case_number', NEW.case_number,
            'member_id', NEW.member_id,
            'action', NEW.action
        );

        PERFORM pg_notify('infraction_expiry_schedule_events', payload::text);
        RETURN NEW;
    END IF;

    -- When an existing infraction expiry schedule entry is removed.
    IF TG_OP = 'DELETE' THEN
        payload := json_build_object(
            'removed', TRUE,
            'expires_at', OLD.expires_at,
            'guild_id', OLD.guild_id,
            'case_number', OLD.case_number,
            'member_id', OLD.member_id,
            'action', OLD.action
        );

        PERFORM pg_notify('infraction_expiry_schedule_events', payload::text);
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER infraction_expiry_schedule_insert_notify
AFTER INSERT ON infraction_expiry_schedule
FOR EACH ROW
EXECUTE FUNCTION infraction_expiry_schedule_events();

CREATE TRIGGER infraction_expiry_schedule_update_notify
AFTER UPDATE ON infraction_expiry_schedule
FOR EACH ROW
EXECUTE FUNCTION infraction_expiry_schedule_events();

CREATE TRIGGER infraction_expiry_schedule_delete_notify
AFTER DELETE ON infraction_expiry_schedule
FOR EACH ROW
EXECUTE FUNCTION infraction_expiry_schedule_events();

-- +goose Down

DROP FUNCTION IF EXISTS infraction_expiry_schedule_events() CASCADE;

DROP VIEW IF EXISTS infraction_details;

DROP TABLE IF EXISTS ban_appeal_logs;
DROP TABLE IF EXISTS infraction_active_bans;
DROP TABLE IF EXISTS infraction_expiry_schedule;
DROP TABLE IF EXISTS infraction_proof_messages;
DROP TABLE IF EXISTS infraction_log;
DROP TABLE IF EXISTS next_infraction_ids;
DROP TABLE IF EXISTS infraction_settings;

DROP TYPE IF EXISTS appeal_status;
DROP TYPE IF EXISTS infraction_action;
