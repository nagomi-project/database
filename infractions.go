package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nagomi-project/database/internal/gen"
)

type infractions struct {
	db *Database

	Notifications *infractionsNotifier
}

func newInfractions(db *Database) *infractions {
	return &infractions{
		db:            db,
		Notifications: newInfractionsNotifier(db),
	}
}

type InfractionEntry struct {
	IssuedAt   time.Time
	ModifiedAt time.Time
	ExpiresAt  *time.Time

	CaseID      int32
	MemberID    string
	ModeratorID string

	Action   InfractionAction
	Reason   *string
	ProofURL *string

	Active     bool
	Appealable bool // This really would only be applicable to bans, no other infraction is appealable.
	Hidden     bool
}

// newInfractionEntryFromDetails formats data from the database into a public, usable InfractionEntry structure.
func newInfractionEntryFromDetails(d gen.InfractionDetail) *InfractionEntry {
	e := InfractionEntry{
		IssuedAt:   d.CreatedAt.Time,
		ModifiedAt: d.UpdatedAt.Time,

		CaseID:      d.CaseNumber,
		MemberID:    d.MemberID,
		ModeratorID: d.ModeratorID,
		Action:      d.Action,
		Appealable:  d.Appealable,
		Hidden:      d.Hidden,
	}

	if d.ExpiresAt.Valid {
		e.ExpiresAt = &d.CreatedAt.Time
	}

	if d.Reason.Valid {
		e.Reason = &d.Reason.String
	}

	if d.MessageUrl.Valid {
		e.ProofURL = &d.MessageUrl.String
	}

	return &e
}

// InfractMemberWithCallback will add a member infraction and require passing in a callback before the database will commit.
//
// This is for when the Discord API errors, it allows rolling back the transaction so unnecessary data is not stored.
func (i *infractions) InfractMemberWithCallback(ctx context.Context, guildId, memberId, moderatorId string, action InfractionAction, duration *time.Duration, reason *string, appealable *bool, cb func(e InfractionEntry) error) (*InfractionEntry, error) {
	if _, err := i.db.GuildSettings.GetOrCreateGuildSettings(ctx, guildId); err != nil {
		return nil, err
	}

	entry := InfractionEntry{
		MemberID:    memberId,
		ModeratorID: moderatorId,
		Action:      action,
		Reason:      reason,
	}

	if err := i.db.withTx(ctx, func(ctx context.Context, txDb *Database) error {
		var expiry pgtype.Timestamptz
		if duration != nil {
			expiresAt := time.Now().Add(*duration)
			entry.ExpiresAt = &expiresAt
		}

		infReason := NullableStringToText(reason)
		expiry = NullableTimeToTimestamptz(entry.ExpiresAt)

		infraction, err := txDb.queries.InfractMember(ctx, txDb.dbtx, gen.InfractMemberParams{
			Expiry:      expiry,
			GuildID:     guildId,
			MemberID:    memberId,
			ModeratorID: moderatorId,
			Action:      action,
			Reason:      infReason,
		})
		if err != nil {
			return err
		}

		entry.CaseID = infraction.CaseNumber
		entry.Hidden = infraction.Hidden

		switch action {
		case InfractionActionMute:
			if _, err := txDb.queries.ScheduleInfraction(ctx, txDb.dbtx, gen.ScheduleInfractionParams{
				Expiry:   expiry,
				GuildID:  guildId,
				CaseID:   infraction.CaseNumber,
				MemberID: memberId,
				Action:   action,
			}); err != nil {
				return err
			}
		case InfractionActionBan:
			canAppealBan := NullableBoolToBool(appealable)

			if _, err := txDb.queries.ScheduleInfraction(ctx, txDb.dbtx, gen.ScheduleInfractionParams{
				Expiry:   expiry,
				GuildID:  guildId,
				CaseID:   infraction.CaseNumber,
				MemberID: memberId,
				Action:   action,
			}); err != nil {
				return err
			}

			if err := txDb.queries.InsertActiveBan(ctx, txDb.dbtx, gen.InsertActiveBanParams{
				GuildID:    guildId,
				CaseID:     infraction.CaseNumber,
				MemberID:   memberId,
				Appealable: canAppealBan,
			}); err != nil {
				return err
			}

			entry.Appealable = canAppealBan.Bool
		case InfractionActionUnmute:
			if err := txDb.queries.UnscheduleInfractionByType(ctx, txDb.dbtx, gen.UnscheduleInfractionByTypeParams{
				GuildID:  guildId,
				MemberID: memberId,
				Action:   InfractionActionMute,
			}); err != nil {
				return err
			}
		case InfractionActionUnban:
			if err := txDb.queries.UnscheduleInfractionByType(ctx, txDb.dbtx, gen.UnscheduleInfractionByTypeParams{
				GuildID:  guildId,
				MemberID: memberId,
				Action:   InfractionActionBan,
			}); err != nil {
				return err
			}

			if err := txDb.queries.RemoveActiveBan(ctx, txDb.dbtx, gen.RemoveActiveBanParams{
				GuildID:  guildId,
				MemberID: memberId,
			}); err != nil {
				return err
			}
		}

		if err := cb(entry); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &entry, nil
}

// UpdateInfractionCaseReason will update the reason for an existing case.
func (i *infractions) UpdateInfractionCaseReason(ctx context.Context, guildId string, caseId int32, reason string) (*InfractionEntry, error) {
	updatedEntry, err := i.db.queries.UpdateInfractionCaseDetails(ctx, i.db.dbtx, gen.UpdateInfractionCaseDetailsParams{
		GuildID: guildId,
		CaseID:  caseId,
		Reason:  NullableStringToText(&reason),
	})
	if err != nil {
		return nil, err
	}

	return newInfractionEntryFromDetails(updatedEntry), nil
}

// UpdateInfractionCaseVisibility will update the visibility of an existing infraction.
//
// Instead of deleting infractions, they're instead labeled as "hidden" so they no longer appear on a member's log history. This is ideal for
// when moderation action is reversed or incorrectly done and should not be held against the user.
func (i *infractions) UpdateInfractionCaseVisibility(ctx context.Context, guildId string, caseId int32, hidden bool) (*InfractionEntry, error) {
	updatedEntry, err := i.db.queries.UpdateInfractionCaseDetails(ctx, i.db.dbtx, gen.UpdateInfractionCaseDetailsParams{
		GuildID: guildId,
		CaseID:  caseId,
		Hidden:  NullableBoolToBool(&hidden),
	})
	if err != nil {
		return nil, err
	}

	return newInfractionEntryFromDetails(updatedEntry), nil
}

type BanAppealLog struct {
	ActorID string
	Status  AppealStatus
}

type ActiveBanDetails struct {
	AppliedAt time.Time
	CaseID    int32

	IsAppealable  bool
	AppealPending bool
	AppealableOn  time.Time

	ActionLog []BanAppealLog
}

// GetActiveBanDetails will fetch all of the information on a member's active ban.
func (i *infractions) GetActiveBanDetails(ctx context.Context, guildId, memberId string) (*ActiveBanDetails, error) {
	activeBan, err := i.db.queries.GetActiveBan(ctx, i.db.dbtx, gen.GetActiveBanParams{
		GuildID:  guildId,
		MemberID: memberId,
	})
	if err != nil {
		return nil, err
	}

	appealLogs, err := i.db.queries.GetBanAppealLogsByCaseId(ctx, i.db.dbtx, gen.GetBanAppealLogsByCaseIdParams{
		GuildID: guildId,
		CaseID:  activeBan.CaseNumber,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	details := &ActiveBanDetails{
		AppliedAt:     activeBan.CreatedAt.Time,
		CaseID:        activeBan.CaseNumber,
		IsAppealable:  activeBan.CanSubmitAppeal,
		AppealPending: activeBan.AppealPending,

		ActionLog: make([]BanAppealLog, len(appealLogs)),
	}

	for idx, log := range appealLogs {
		details.ActionLog[idx] = BanAppealLog{
			ActorID: log.ActorID,
			Status:  log.Status,
		}
	}

	return details, nil
}

// UpdateBanAppealStatus will update the status of a ban appeal.
func (i *infractions) UpdateBanAppealStatus(ctx context.Context, guildId, memberId, actorId string, status AppealStatus) error {
	if err := i.db.withTx(ctx, func(ctx context.Context, txDb *Database) error {
		settings, err := txDb.queries.GetGuildInfractionSettings(ctx, txDb.dbtx, guildId)
		if err != nil {
			return err
		}

		infraction, err := txDb.queries.GetActiveBanInfraction(ctx, txDb.dbtx, gen.GetActiveBanInfractionParams{
			GuildID:  guildId,
			MemberID: memberId,
		})
		if err != nil {
			return err
		}

		switch status {
		case AppealStatusSubmitted:
			if _, err := txDb.queries.UpdateActiveBan(ctx, txDb.dbtx, gen.UpdateActiveBanParams{
				IsPending: NullableBoolToBool(new(true)),

				GuildID:  guildId,
				MemberID: memberId,
			}); err != nil {
				return err
			}
		case AppealStatusApproved:
			if err := txDb.queries.RemoveActiveBan(ctx, txDb.dbtx, gen.RemoveActiveBanParams{
				GuildID:  guildId,
				MemberID: memberId,
			}); err != nil {
				return err
			}
		case AppealStatusDenied:
			defaultNextAppealTime := time.Now().Add(time.Duration(settings.AppealDuration) * (24 * time.Hour))

			if _, err := txDb.queries.UpdateActiveBan(ctx, txDb.dbtx, gen.UpdateActiveBanParams{
				IsPending:   NullableBoolToBool(new(false)),
				CanAppealOn: NullableTimeToTimestamptz(&defaultNextAppealTime),

				GuildID:  guildId,
				MemberID: memberId,
			}); err != nil {
				return err
			}
		case AppealStatusBlocked:
			if _, err := txDb.queries.UpdateActiveBan(ctx, txDb.dbtx, gen.UpdateActiveBanParams{
				Appealable: NullableBoolToBool(new(false)),

				GuildID:  guildId,
				MemberID: memberId,
			}); err != nil {
				return err
			}
		case AppealStatusUnblocked:
			if _, err := txDb.queries.UpdateActiveBan(ctx, txDb.dbtx, gen.UpdateActiveBanParams{
				Appealable: NullableBoolToBool(new(true)),

				GuildID:  guildId,
				MemberID: memberId,
			}); err != nil {
				return err
			}
		}

		if _, err := txDb.queries.LogBanAppealStatus(ctx, txDb.dbtx, gen.LogBanAppealStatusParams{
			GuildID:      guildId,
			CaseID:       infraction.CaseNumber,
			ActorID:      actorId,
			AppealStatus: status,
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

type infractionsNotifier struct {
	db *Database
}

// In the future, this should maybe be a general notifier structure that will reuse a connection.
// It doesn't seem like a wise idea to open multiple new connections for notification events.
//
// For the time being, I'm too lazy to do that (since I already had this code) and I'm not sure
// if there will be other usecases for a listener. I'll cross that road if and when it comes to it.
func newInfractionsNotifier(db *Database) *infractionsNotifier {
	return &infractionsNotifier{db}
}

type InfractionsNotificationEvent struct {
	Inserted   *bool            `json:"inserted,omitempty"`
	Updated    *bool            `json:"updated,omitempty"`
	Removed    *bool            `json:"removed,omitempty"`
	CaseNumber int32            `json:"case_number"`
	ExpiresAt  *time.Time       `json:"expires_at"`
	GuildID    string           `json:"guild_id"`
	MemberID   string           `json:"member_id"`
	Action     InfractionAction `json:"action"`
}

// Listen will create a new connection that listens for notifications for infraction expirations.
func (n *infractionsNotifier) Listen(ctx context.Context, data chan<- InfractionsNotificationEvent) error {
	conn, err := pgx.Connect(ctx, n.db.pool.Config().ConnString())
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	if _, err := conn.Exec(ctx, "LISTEN infraction_expiry_schedule_events"); err != nil {
		return err
	}
	defer conn.Exec(context.Background(), "UNLISTEN infraction_expiry_schedule_events") //nolint:errcheck

	for {
		notif, err := conn.WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			return err
		}

		var event InfractionsNotificationEvent
		if err := json.Unmarshal([]byte(notif.Payload), &event); err != nil {
			return err
		}

		select {
		case data <- event:
		case <-ctx.Done():
			return nil
		}
	}
}
