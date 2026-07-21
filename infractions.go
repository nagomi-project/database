package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nagomi-project/database/internal/gen"
)

type infractions struct {
	db *Database
}

func newInfractions(db *Database) *infractions {
	return &infractions{db}
}

type InfractionEntry struct {
	IssuedAt   time.Time
	ModifiedAt time.Time
	ExpiresAt  *time.Time

	CaseID      int32
	GuildID     string
	MemberID    string
	ModeratorID string

	Action   ModerationAction
	Reason   *string
	ProofURL *string

	Active     bool
	Appealable bool // This really would only be applicable to bans, no other infraction is appealable.
	Hidden     bool
}

// newInfractionEntryFromDetails formats data from the database into a public, usable InfractionEntry structure.
func newInfractionEntryFromDetails(d gen.ModerationCaseDetail) *InfractionEntry {
	e := InfractionEntry{
		IssuedAt:   d.CreatedAt.Time,
		ModifiedAt: d.UpdatedAt.Time,

		CaseID:      d.CaseNumber,
		GuildID:     d.GuildID,
		MemberID:    d.MemberID,
		ModeratorID: d.ModeratorID,

		Action: d.Action,

		Active:     d.Active,
		Appealable: d.Appealable,
		Hidden:     d.Hidden,
	}

	if d.ExpiresAt.Valid {
		e.ExpiresAt = &d.ExpiresAt.Time
	}

	if d.Reason.Valid {
		e.Reason = &d.Reason.String
	}

	if d.MessageUrl.Valid {
		e.ProofURL = &d.MessageUrl.String
	}

	return &e
}

type InfractionSettings struct {
	MutedRoleID                  *string
	AppealDuration               int16
	ShouldRequestInfractionProof bool
	InfractionProofChannelID     *string
}

func (i *infractions) getConfiguration(ctx context.Context, guildId string) (*InfractionSettings, error) {
	config, err := i.db.queries.GetGuildInfractionSettings(ctx, i.db.dbtx, guildId)
	if err != nil {
		return nil, err
	}

	ic := &InfractionSettings{
		AppealDuration:               config.AppealDuration,
		ShouldRequestInfractionProof: config.RequestInfractionProof,
	}

	if config.MutedRoleID.Valid {
		ic.MutedRoleID = &config.MutedRoleID.String
	}

	if config.InfractionProofID.Valid {
		ic.InfractionProofChannelID = &config.InfractionProofID.String
	}

	return ic, nil
}

func (i *infractions) GetOrCreateConfiguration(ctx context.Context, guildId string) (*InfractionSettings, error) {
	config, err := i.getConfiguration(ctx, guildId)
	if err != nil {
		if err == pgx.ErrNoRows {
			if err := i.db.queries.RegisterInfractionSettingsIfMissing(ctx, i.db.dbtx, guildId); err != nil {
				return nil, err
			}

			return i.getConfiguration(ctx, guildId)
		}

		return nil, err
	}

	return config, nil
}

func (i *infractions) GetExpiringInfractions(ctx context.Context, cutoff time.Time) ([]InfractionEntry, error) {
	expiring, err := i.db.queries.GetExpiringInfractionCases(ctx, i.db.dbtx, NullableTimeToTimestamptz(&cutoff))
	if err != nil {
		return nil, err
	}

	entries := make([]InfractionEntry, len(expiring))
	for i, inf := range expiring {
		entries[i] = *newInfractionEntryFromDetails(inf)
	}

	return entries, nil
}

// InfractMemberWithCallback will add a member infraction and require passing in a callback before the database will commit.
//
// This is for when the Discord API errors, it allows rolling back the transaction so unnecessary data is not stored.
func (i *infractions) InfractMemberWithCallback(ctx context.Context, guildId, memberId, moderatorId string, action ModerationAction, duration *time.Duration, reason *string, appealable *bool, cb func(e InfractionEntry) error) (*InfractionEntry, error) {
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

		entry.IssuedAt = infraction.CreatedAt.Time
		entry.ModifiedAt = infraction.UpdatedAt.Time
		entry.CaseID = infraction.CaseNumber
		entry.Hidden = infraction.Hidden

		switch action {
		case ModerationActionMute:
			if _, err := txDb.queries.ScheduleInfraction(ctx, txDb.dbtx, gen.ScheduleInfractionParams{
				Expiry:   expiry,
				GuildID:  guildId,
				CaseID:   infraction.CaseNumber,
				MemberID: memberId,
				Action:   action,
			}); err != nil {
				if err == pgx.ErrNoRows {
					return ErrUserAlreadyMuted
				}

				return err
			}

			entry.Active = true
		case ModerationActionBan:
			canAppealBan := NullableBoolToBool(appealable)

			if _, err := txDb.queries.ScheduleInfraction(ctx, txDb.dbtx, gen.ScheduleInfractionParams{
				Expiry:   expiry,
				GuildID:  guildId,
				CaseID:   infraction.CaseNumber,
				MemberID: memberId,
				Action:   action,
			}); err != nil {
				if err == pgx.ErrNoRows {
					return ErrUserAlreadyBanned
				}

				return err
			}

			entry.Active = true

			if err := txDb.queries.InsertActiveBan(ctx, txDb.dbtx, gen.InsertActiveBanParams{
				GuildID:    guildId,
				CaseID:     infraction.CaseNumber,
				MemberID:   memberId,
				Appealable: canAppealBan,
			}); err != nil {
				return err
			}

			entry.Appealable = canAppealBan.Bool
		case ModerationActionUnmute:
			if _, err := txDb.queries.UnscheduleInfractionByType(ctx, txDb.dbtx, gen.UnscheduleInfractionByTypeParams{
				GuildID:  guildId,
				MemberID: memberId,
				Action:   ModerationActionMute,
			}); err != nil {
				if err == pgx.ErrNoRows {
					return ErrUserNotMuted
				}

				return err
			}
		case ModerationActionUnban:
			if _, err := txDb.queries.UnscheduleInfractionByType(ctx, txDb.dbtx, gen.UnscheduleInfractionByTypeParams{
				GuildID:  guildId,
				MemberID: memberId,
				Action:   ModerationActionBan,
			}); err != nil {
				if err == pgx.ErrNoRows {
					return ErrUserNotBanned
				}

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

// InfractMember will add a member infraction.
func (i *infractions) InfractMember(ctx context.Context, guildId, memberId, moderatorId string, action ModerationAction, duration *time.Duration, reason *string, appealable *bool) (*InfractionEntry, error) {
	return i.InfractMemberWithCallback(ctx,
		guildId, memberId, moderatorId,
		action, duration, reason, appealable,
		func(e InfractionEntry) error {
			return nil
		},
	)
}

// GetInfractionCaseDetails will get the details of an infraction case based on a provided id.
func (i *infractions) GetInfractionCaseDetails(ctx context.Context, guildId string, caseId int32) (*InfractionEntry, error) {
	infraction, err := i.db.queries.GetInfractionByCaseId(ctx, i.db.dbtx, gen.GetInfractionByCaseIdParams{
		GuildID: guildId,
		CaseID:  caseId,
	})
	if err != nil {
		return nil, err
	}

	return newInfractionEntryFromDetails(infraction), nil
}

// GetMemberInfractionCasePage will get a page of member infractions and its pagination details.
// Each page returns five infractions maximum.
func (i *infractions) GetMemberInfractionCasePage(ctx context.Context, guildId, memberId string, page int) (*PaginatedQuery[InfractionEntry], error) {
	page = max(page, 1)

	pageDetails, err := i.db.queries.GetMemberInfractionsPageDetails(ctx, i.db.dbtx, gen.GetMemberInfractionsPageDetailsParams{
		GuildID:  guildId,
		MemberID: memberId,
	})
	if err != nil {
		return nil, err
	}
	if pageDetails.TotalEntries <= 0 {
		return nil, ErrNoInfractions
	}

	infractionsPage, err := i.db.queries.GetMemberInfractionsPage(ctx, i.db.dbtx, gen.GetMemberInfractionsPageParams{
		GuildID:  guildId,
		MemberID: memberId,
		Page:     int16(page),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrInfractionPageNotFound
		}

		return nil, err
	}

	infractions := make([]InfractionEntry, len(infractionsPage))
	for i, inf := range infractionsPage {
		infractions[i] = *newInfractionEntryFromDetails(inf)
	}

	return &PaginatedQuery[InfractionEntry]{
		CurrentPage:  page,
		TotalPages:   int(pageDetails.TotalPages),
		TotalEntries: int(pageDetails.TotalEntries),
		Data:         infractions,
	}, nil
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

// UpdateInfractionCaseDuration will update the duration of an existing infraction case.
//
// The duration will be set based on the original infraction time and not from `time.Now()`. If there is no original infraction time,
// `time.Now()` will be used instead (this can happen with a mute or ban with no duration set).
func (i *infractions) UpdateInfractionCaseDuration(ctx context.Context, guildId string, caseId int32, duration time.Duration) (*InfractionEntry, error) {
	var infraction *InfractionEntry
	if err := i.db.withTx(ctx, func(ctx context.Context, txDb *Database) error {
		originalEntry, err := txDb.queries.GetInfractionByCaseId(ctx, txDb.dbtx, gen.GetInfractionByCaseIdParams{
			GuildID: guildId,
			CaseID:  caseId,
		})
		if err != nil {
			return err
		}

		switch originalEntry.Action {
		case ModerationActionBan, ModerationActionMute:
			expiry := time.Now().Add(duration)
			if originalEntry.CreatedAt.Valid {
				expiry = originalEntry.CreatedAt.Time.Add(duration)
			}

			if _, err := txDb.queries.ModifyScheduledInfraction(ctx, txDb.dbtx, gen.ModifyScheduledInfractionParams{
				GuildID:          guildId,
				CaseID:           caseId,
				ModifiedDuration: NullableTimeToTimestamptz(&expiry),
			}); err != nil {
				if err == pgx.ErrNoRows {
					return ErrInactiveInfraction
				}

				return err
			}

			originalEntry.ExpiresAt = NullableTimeToTimestamptz(&expiry)
		default:
			return fmt.Errorf("unsupported type") // todo: errors.
		}

		infraction = newInfractionEntryFromDetails(originalEntry)

		return nil
	}); err != nil {
		return nil, err
	}

	return infraction, nil
}

// UpdateInfractionCaseExpiry will update the raw expiration time for an existing infraction case.
//
// If the expiry time IsZero, it will be set as a permanent duration instead.
func (i *infractions) UpdateInfractionCaseExpiry(ctx context.Context, guildId string, caseId int32, expiry time.Time) (*InfractionEntry, error) {
	var infraction *InfractionEntry
	if err := i.db.withTx(ctx, func(ctx context.Context, txDb *Database) error {
		originalEntry, err := txDb.queries.GetInfractionByCaseId(ctx, txDb.dbtx, gen.GetInfractionByCaseIdParams{
			GuildID: guildId,
			CaseID:  caseId,
		})
		if err != nil {
			return err
		}

		switch originalEntry.Action {
		case ModerationActionBan, ModerationActionMute:

			if _, err := txDb.queries.ModifyScheduledInfraction(ctx, txDb.dbtx, gen.ModifyScheduledInfractionParams{
				GuildID:          guildId,
				CaseID:           caseId,
				ModifiedDuration: NullableTimeToTimestamptz(&expiry),
			}); err != nil {
				return err
			}

			originalEntry.ExpiresAt = NullableTimeToTimestamptz(&expiry)
			infraction = newInfractionEntryFromDetails(originalEntry)
		default:
			return fmt.Errorf("unsupported type") // todo: errors.
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return infraction, nil
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
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
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
