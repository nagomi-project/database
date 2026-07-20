package database

import (
	"context"
	"fmt"
	"time"

	"github.com/nagomi-project/database/internal/gen"
)

type guildSettings struct {
	db *Database
}

func newGuildSettings(db *Database) *guildSettings {
	return &guildSettings{db}
}

type EventLogChannel struct {
	IgnoreChannels []string
	IgnoreRoles    []string

	UserJoin           *string
	UserLeave          *string
	UserKick           *string
	UserBan            *string
	UserUnban          *string
	UserRolesUpdate    *string
	UserNicknameUpdate *string
	UserVoiceJoin      *string
	UserVoiceMove      *string
	UserVoiceLeave     *string

	MessageCreate      *string
	MessageEdit        *string
	MessageDelete      *string
	MessageImageRemove *string

	ChannelCreate *string
	ChannelUpdate *string
	ChannelDelete *string

	EmojiCreate *string
	EmojiUpdate *string
	EmojiDelete *string
}

func (g *guildSettings) formatLogChannels(channels []gen.EventLogChannel, options gen.EventLogSetting) *EventLogChannel {
	logChannels := &EventLogChannel{
		IgnoreChannels: options.IgnoredChannels,
		IgnoreRoles:    options.IgnoreRoles,
	}

	for _, channel := range channels {
		switch channel.Type {
		case EventLogTypeUserjoin:
			logChannels.UserJoin = &channel.ChannelID
		case EventLogTypeUserleave:
			logChannels.UserLeave = &channel.ChannelID
		case EventLogTypeUserkick:
			logChannels.UserKick = &channel.ChannelID
		case EventLogTypeUserban:
			logChannels.UserBan = &channel.ChannelID
		case EventLogTypeUserunban:
			logChannels.UserUnban = &channel.ChannelID
		case EventLogTypeUserrolesUpdate:
			logChannels.UserRolesUpdate = &channel.ChannelID
		case EventLogTypeUsernicknameUpdate:
			logChannels.UserNicknameUpdate = &channel.ChannelID
		case EventLogTypeUservoiceJoin:
			logChannels.UserVoiceJoin = &channel.ChannelID
		case EventLogTypeUservoiceMove:
			logChannels.UserVoiceMove = &channel.ChannelID
		case EventLogTypeUservoiceLeave:
			logChannels.UserVoiceLeave = &channel.ChannelID
		case EventLogTypeMessagecreate:
			logChannels.MessageCreate = &channel.ChannelID
		case EventLogTypeMessageedit:
			logChannels.MessageEdit = &channel.ChannelID
		case EventLogTypeMessagedelete:
			logChannels.MessageDelete = &channel.ChannelID
		case EventLogTypeMessageimageRemove:
			logChannels.MessageImageRemove = &channel.ChannelID
		case EventLogTypeChannelcreate:
			logChannels.ChannelCreate = &channel.ChannelID
		case EventLogTypeChannelupdate:
			logChannels.ChannelUpdate = &channel.ChannelID
		case EventLogTypeChanneldelete:
			logChannels.ChannelDelete = &channel.ChannelID
		case EventLogTypeEmojicreate:
			logChannels.EmojiCreate = &channel.ChannelID
		case EventLogTypeEmojiupdate:
			logChannels.EmojiUpdate = &channel.ChannelID
		case EventLogTypeEmojidelete:
			logChannels.EmojiDelete = &channel.ChannelID
		}
	}

	return logChannels
}

func (g *guildSettings) getLogChannels(ctx context.Context, dbtx gen.DBTX, guildId string) (*EventLogChannel, error) {
	options, err := g.db.queries.GetEventLogSettings(ctx, dbtx, guildId)
	if err != nil {
		return nil, err
	}

	channels, err := g.db.queries.GetEventLogChannels(ctx, dbtx, guildId)
	if err != nil {
		return nil, err
	}

	return g.formatLogChannels(channels, options), nil
}

func (g *guildSettings) GetLogChannels(ctx context.Context, guildId string) (*EventLogChannel, error) {
	return g.getLogChannels(ctx, g.db.dbtx, guildId)
}

func (g *guildSettings) CreateOrUpdateLogChannel(ctx context.Context, guildId, channelId, userId string, channelType LogChannelType, source ActionLogSource) (*gen.EventLogChannel, error) {
	var channel *gen.EventLogChannel
	if err := g.db.withTx(ctx, func(ctx context.Context, txDb *Database) error {
		if err := g.db.queries.UpdateGuildRegistryTime(ctx, txDb.dbtx, guildId); err != nil {
			return err
		}

		entry, err := txDb.queries.UpsertLogChannel(ctx, txDb.dbtx, gen.UpsertLogChannelParams{
			Type:      channelType,
			GuildID:   guildId,
			ChannelID: channelId,
		})
		if err != nil {
			return err
		}

		if _, err := txDb.queries.LogAction(ctx, txDb.dbtx, gen.LogActionParams{
			GuildID:      guildId,
			ActorID:      userId,
			ActionType:   ActionLogTypeGuildSettingsUpdate,
			ActionSource: source,
			ActionReason: fmt.Sprintf("updated log channel: type=%s, channel_id=%s", channelType, channelId),
		}); err != nil {
			return err
		}

		channel = &entry
		return nil
	}); err != nil {
		return nil, err
	}

	return channel, nil
}

func (g *guildSettings) RemoveLogChannel(ctx context.Context, guildId, userId string, channelType LogChannelType, source ActionLogSource) error {
	return g.db.withTx(ctx, func(ctx context.Context, txDb *Database) error {
		if err := g.db.queries.UpdateGuildRegistryTime(ctx, txDb.dbtx, guildId); err != nil {
			return err
		}

		entry, err := txDb.queries.RemoveLogChannel(ctx, txDb.dbtx, gen.RemoveLogChannelParams{
			Type:    channelType,
			GuildID: guildId,
		})
		if err != nil {
			return err
		}

		if _, err := txDb.queries.LogAction(ctx, txDb.dbtx, gen.LogActionParams{
			GuildID:      guildId,
			ActorID:      userId,
			ActionType:   ActionLogTypeGuildSettingsUpdate,
			ActionSource: source,
			ActionReason: fmt.Sprintf("removed log channel: type=%s, channel_id=%s", channelType, entry.ChannelID),
		}); err != nil {
			return err
		}

		return nil
	})
}

type InfractionSettings struct {
	MutedRoleID                  *string
	AppealDuration               int16
	ShouldRequestInfractionProof bool
	InfractionProofChannelID     *string
}

type GuildSettings struct {
	LastModified time.Time

	Infractions      InfractionSettings
	EventLogChannels EventLogChannel
}

func (g *guildSettings) GetOrCreateGuildSettings(ctx context.Context, guildId string) (*GuildSettings, error) {
	var settings *GuildSettings

	if err := g.db.withTx(ctx, func(ctx context.Context, txDb *Database) error {
		if err := txDb.queries.RegisterGuildIfMissing(ctx, txDb.dbtx, guildId); err != nil {
			return err
		}

		registry, err := txDb.queries.GetRegisteredGuild(ctx, txDb.dbtx, guildId)
		if err != nil {
			return err
		}

		if err := txDb.queries.RegisterInfractionSettingsIfMissing(ctx, txDb.dbtx, guildId); err != nil {
			return err
		}

		infSettings, err := txDb.queries.GetGuildInfractionSettings(ctx, txDb.dbtx, guildId)
		if err != nil {
			return err
		}

		logChannels, err := g.getLogChannels(ctx, txDb.dbtx, guildId)
		if err != nil {
			return err
		}

		settings = &GuildSettings{
			LastModified: registry.UpdatedAt.Time,
			Infractions: InfractionSettings{
				AppealDuration:               infSettings.AppealDuration,
				ShouldRequestInfractionProof: infSettings.RequestInfractionProof,
			},
			EventLogChannels: *logChannels,
		}
		if infSettings.MutedRoleID.Valid {
			settings.Infractions.MutedRoleID = &infSettings.MutedRoleID.String
		}
		if infSettings.InfractionProofID.Valid {
			settings.Infractions.InfractionProofChannelID = &infSettings.InfractionProofID.String
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return settings, nil
}

type ActionLog struct {
	CreatedAt time.Time
	ID        int64
	ActorID   string
	Type      ActionLogType
	Source    ActionLogSource
	Action    string
}

// GetActionLogHistory gets a page of actions performed in a guild.
// Each page returns five actions maximum.
func (g *guildSettings) GetActionLogHistory(ctx context.Context, guildId string, page int) (*PaginatedQuery[ActionLog], error) {
	page = max(page, 1)

	pageDetails, err := g.db.queries.GetActionLogPageDetails(ctx, g.db.dbtx, gen.GetActionLogPageDetailsParams{
		GuildID: guildId,
	})
	if err != nil {
		return nil, err
	}

	logPage, err := g.db.queries.GetActionLogPage(ctx, g.db.dbtx, gen.GetActionLogPageParams{
		GuildID: guildId,
		Page:    int16(page),
	})
	if err != nil {
		return nil, err
	}

	actions := make([]ActionLog, len(logPage))
	for idx, entry := range logPage {
		actions[idx] = ActionLog{
			CreatedAt: entry.CreatedAt.Time,
			ID:        entry.ID,
			ActorID:   entry.ActorID,
			Type:      entry.Type,
			Source:    entry.Source,
			Action:    entry.Action,
		}
	}

	return &PaginatedQuery[ActionLog]{
		CurrentPage:  page,
		TotalPages:   int(pageDetails.TotalPages),
		TotalEntries: int(pageDetails.TotalEntries),
		Data:         actions,
	}, nil
}
