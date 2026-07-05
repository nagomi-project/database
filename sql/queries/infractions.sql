-- name: RegisterInfractionSettingsIfMissing :exec
-- Inserts guild infraction settings if they are not already created.
INSERT INTO infraction_settings (guild_id)
VALUES (@guild_id)
ON CONFLICT (guild_id) DO NOTHING;

-- name: GetGuildInfractionSettings :one
-- Fetches the current infraction settings set for the guild.
SELECT *
FROM infraction_settings
WHERE
    guild_id = @guild_id;

-- name: GetExpiringInfractionCases :many
-- Fetch a list of infractions that will be expiring within a specified cutoff time.
SELECT * FROM infraction_details
WHERE expires_at <= @cutoff;

-- name: InsertInfractionProofMessage :one
-- Inserts a message url for an infraction's proof.
INSERT INTO infraction_proof_messages (
    guild_id,
    case_number,
    message_url
)
VALUES (@guild_id, @case_id, @message_url)
ON CONFLICT DO NOTHING
RETURNING *;

-- name: GetInfractionByCaseId :one
-- Fetches an infraction's information based on the case id.
SELECT * FROM infraction_details
WHERE
    guild_id = @guild_id
    AND case_number = @case_id;

-- name: GetMemberInfractions :many
-- Fetches all of the infractions for a member
SELECT * FROM infraction_details
WHERE guild_id = @guild_id AND member_id = @member_id
ORDER BY case_number DESC;

-- name: GetActiveMuteInfraction :one
-- Fetches an active mute infraction for a member.
SELECT * FROM infraction_details
WHERE
    guild_id = @guild_id
    AND member_id = @member_id
    AND active = TRUE
    AND action = 'mute'::infraction_action;

-- name: GetActiveBanInfraction :one
-- Fetches an active ban infraction for a member.
SELECT * FROM infraction_details
WHERE
    guild_id = @guild_id
    AND member_id = @member_id
    AND active = TRUE
    AND action = 'ban'::infraction_action;

-- name: InfractMember :one
-- Infracts a member.
WITH next_infraction AS (
    INSERT INTO next_infraction_ids (guild_id, next_id)
    VALUES (@guild_id, 2)
    ON CONFLICT (guild_id) DO UPDATE SET
        next_id = next_infraction_ids.next_id + 1
    RETURNING next_id - 1 AS case_number
)
INSERT INTO infraction_log (
    expires_at,
    guild_id,
    case_number,
    member_id,
    moderator_id,
    action,
    reason
)
SELECT
    @expiry,
    @guild_id,
    next_infraction.case_number,
    @member_id,
    @moderator_id,
    @action,
    @reason
FROM next_infraction
RETURNING *;

-- name: UpdateInfractionCaseDetails :one
-- Updates the information regarding an infraction case and returns the new infraction.
WITH updated_case AS (
    UPDATE infraction_log SET
        updated_at = now(),
        hidden = COALESCE(sqlc.narg('hidden'), infraction_log.hidden),
        reason = COALESCE(sqlc.narg('reason'), infraction_log.reason)
    WHERE
        infraction_log.guild_id = @guild_id
        AND infraction_log.case_number = @case_id
    RETURNING *
)
SELECT infraction_details.*
FROM infraction_details
JOIN updated_case ON
    updated_case.guild_id = infraction_details.guild_id
    AND updated_case.case_number = infraction_details.case_number;

-- name: ScheduleInfraction :one
-- Schedules an infraction.
INSERT INTO infraction_expiry_schedule (
    expires_at,
    guild_id,
    case_number,
    member_id,
    action
)
VALUES (@expiry, @guild_id, @case_id, @member_id, @action)
ON CONFLICT DO NOTHING
RETURNING *;

-- name: InsertActiveBan :exec
-- Inserts a ban.
INSERT INTO infraction_active_bans (
    guild_id,
    case_number,
    member_id,
    appealable
)
VALUES (@guild_id, @case_id, @member_id, COALESCE(sqlc.narg('appealable')::BOOLEAN, true))
ON CONFLICT (guild_id, case_number) DO NOTHING;

-- name: GetActiveBan :one
-- Fetches the active ban.
SELECT * FROM infraction_active_bans
WHERE
    guild_id = @guild_id
    AND member_id = @member_id;

-- name: UpdateActiveBan :one
-- Updates the details for an active ban.
UPDATE infraction_active_bans SET
    updated_at = now(),
    can_submit_appeal = COALESCE(sqlc.narg('appealable'), can_submit_appeal),
    appeal_pending = COALESCE(sqlc.narg('is_pending'), appeal_pending),
    appealable_at = COALESCE(sqlc.narg('can_appeal_on'), appealable_at)
WHERE
    guild_id = @guild_id
    AND member_id = @member_id
RETURNING *;

-- name: RemoveActiveBan :exec
-- Removes an active ban.
DELETE FROM infraction_active_bans
WHERE
    guild_id = @guild_id
    AND member_id = @member_id;


-- name: GetBanAppealLogsByCaseId :many
-- Fetch all the logs for a ban appeal based on its relating case.
SELECT * FROM ban_appeal_logs
WHERE
    guild_id = @guild_id
    AND case_number = @case_id
ORDER BY
    log_id DESC;

-- name: LogBanAppealStatus :one
-- Will log inforamtion for an active ban. This log is used for keeping track of updates to a ban appeal.
INSERT INTO ban_appeal_logs (
    guild_id,
    case_number,
    actor_id,
    status
)
VALUES (
    @guild_id,
    @case_id,
    @actor_id,
    @appeal_status
)
RETURNING *;

-- name: UnscheduleInfractionByCaseId :exec
-- Removes a scheduled infraction.
DELETE FROM infraction_expiry_schedule
WHERE
    guild_id = @guild_id
    AND case_number = @case_id;

-- name: UnscheduleInfractionByType :exec
-- Unschedules an infraction based on its type.
DELETE FROM infraction_expiry_schedule
WHERE
    guild_id = @guild_id
    AND member_id = @member_id
    AND action = @action;
