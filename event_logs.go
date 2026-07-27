package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/nagomi-project/database/internal/gen"
)

type eventLog struct {
	db *Database
}

func newEventLog(db *Database) *eventLog {
	return &eventLog{db}
}

type EventLogChannel struct {
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

	MessageEdit        *string
	MessageDelete      *string
	MessageImageRemove *string

	ChannelCreate *string
	ChannelUpdate *string
	ChannelDelete *string

	RoleCreate *string
	RoleUpdate *string
	RoleDelete *string

	EmojiCreate *string
	EmojiUpdate *string
	EmojiDelete *string
}

func (e *eventLog) formatLogChannels(channels []gen.EventLogChannel) *EventLogChannel {
	logChannels := &EventLogChannel{}

	for _, channel := range channels {
		switch channel.Type {
		case EventLogTypeUserJoin:
			logChannels.UserJoin = &channel.ChannelID
		case EventLogTypeUserLeave:
			logChannels.UserLeave = &channel.ChannelID
		case EventLogTypeUserKick:
			logChannels.UserKick = &channel.ChannelID
		case EventLogTypeUserBan:
			logChannels.UserBan = &channel.ChannelID
		case EventLogTypeUserUnban:
			logChannels.UserUnban = &channel.ChannelID
		case EventLogTypeUserRolesUpdate:
			logChannels.UserRolesUpdate = &channel.ChannelID
		case EventLogTypeUserNicknameUpdate:
			logChannels.UserNicknameUpdate = &channel.ChannelID
		case EventLogTypeUserVoiceJoin:
			logChannels.UserVoiceJoin = &channel.ChannelID
		case EventLogTypeUserVoiceMove:
			logChannels.UserVoiceMove = &channel.ChannelID
		case EventLogTypeUserVoiceLeave:
			logChannels.UserVoiceLeave = &channel.ChannelID
		case EventLogTypeMessageEdit:
			logChannels.MessageEdit = &channel.ChannelID
		case EventLogTypeMessageDelete:
			logChannels.MessageDelete = &channel.ChannelID
		case EventLogTypeMessageImageRemove:
			logChannels.MessageImageRemove = &channel.ChannelID
		case EventLogTypeChannelCreate:
			logChannels.ChannelCreate = &channel.ChannelID
		case EventLogTypeChannelUpdate:
			logChannels.ChannelUpdate = &channel.ChannelID
		case EventLogTypeChannelDelete:
			logChannels.ChannelDelete = &channel.ChannelID
		case EventLogTypeRoleCreate:
			logChannels.RoleCreate = &channel.ChannelID
		case EventLogTypeRoleUpdate:
			logChannels.RoleUpdate = &channel.ChannelID
		case EventLogTypeRoleDelete:
			logChannels.RoleDelete = &channel.ChannelID
		case EventLogTypeEmojiCreate:
			logChannels.EmojiCreate = &channel.ChannelID
		case EventLogTypeEmojiUpdate:
			logChannels.EmojiUpdate = &channel.ChannelID
		case EventLogTypeEmojiDelete:
			logChannels.EmojiDelete = &channel.ChannelID
		}
	}

	return logChannels
}

type EventLogSettings struct {
	IgnoreChannels []string
	IgnoreRoles    []string
	IgnoreBots     bool
	Channels       EventLogChannel
}

func (e *eventLog) getConfiguration(ctx context.Context, guildId string) (*EventLogSettings, error) {
	config, err := e.db.queries.GetEventLogSettings(ctx, e.db.dbtx, guildId)
	if err != nil {
		return nil, err
	}

	channels, err := e.GetLogChannels(ctx, guildId)
	if err != nil {
		return nil, err
	}

	return &EventLogSettings{
		IgnoreChannels: config.IgnoredChannels,
		IgnoreRoles:    config.IgnoredRoles,
		IgnoreBots:     config.IgnoreBots,
		Channels:       *channels,
	}, nil
}

func (e *eventLog) GetOrCreateConfiguration(ctx context.Context, guildId string) (*EventLogSettings, error) {
	config, err := e.getConfiguration(ctx, guildId)
	if err != nil {
		if err == pgx.ErrNoRows {
			if err := e.db.queries.RegisterEventLogSettingsIfMissing(ctx, e.db.dbtx, guildId); err != nil {
				return nil, err
			}

			return e.getConfiguration(ctx, guildId)
		}

		return nil, err
	}

	return config, nil
}

func (e *eventLog) GetLogChannels(ctx context.Context, guildId string) (*EventLogChannel, error) {
	channels, err := e.db.queries.GetEventLogChannels(ctx, e.db.dbtx, guildId)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	return e.formatLogChannels(channels), nil
}

func (e *eventLog) CreateOrUpdateLogChannel(ctx context.Context, guildId, channelId, userId string, channelType LogChannelType, source ActionLogSource) (*gen.EventLogChannel, error) {
	var channel *gen.EventLogChannel
	if err := e.db.withTx(ctx, func(ctx context.Context, txDb *Database) error {
		if err := e.db.queries.UpdateGuildRegistryTime(ctx, txDb.dbtx, guildId); err != nil {
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

func (e *eventLog) CreateOrUpdateManyLogChannels(ctx context.Context, guildId string, channelIds map[LogChannelType]*string, source ActionLogSource) ([]gen.EventLogChannel, error) {
	var channels []gen.EventLogChannel
	if err := e.db.withTx(ctx, func(ctx context.Context, txDb *Database) error {
		var err error
		channels, err = txDb.EventLog.createOrUpdateManyLogChannels(ctx, guildId, channelIds, source)
		return err
	}); err != nil {
		return nil, err
	}

	return channels, nil
}

func (e *eventLog) createOrUpdateManyLogChannels(ctx context.Context, guildId string, channelIds map[LogChannelType]*string, source ActionLogSource) ([]gen.EventLogChannel, error) {
	var (
		ids         []string
		types       []string
		removeTypes []string
	)

	for t, id := range channelIds {
		if !t.Valid() {
			return nil, fmt.Errorf("invalid event log type: %q", t)
		}
		if id == nil || *id == "" {
			removeTypes = append(removeTypes, string(t))
			continue
		}

		ids = append(ids, *id)
		types = append(types, string(t))
	}

	if len(removeTypes) > 0 {
		if _, err := e.db.queries.RemoveManyLogChannels(ctx, e.db.dbtx, gen.RemoveManyLogChannelsParams{
			GuildID: guildId,
			Types:   removeTypes,
		}); err != nil {
			return nil, err
		}
	}

	if len(ids) == 0 {
		return nil, nil
	}

	return e.db.queries.UpsertManyLogChannels(ctx, e.db.dbtx, gen.UpsertManyLogChannelsParams{
		GuildID:    guildId,
		ChannelIds: ids,
		Types:      types,
	})
}

func (e *eventLog) UpdateConfiguration(ctx context.Context, guildId string, enabled, ignoreBots bool, ignoredChannels, ignoredRoles []string, channelIds map[LogChannelType]*string) error {
	return e.db.withTx(ctx, func(ctx context.Context, txDb *Database) error {
		if err := txDb.queries.ToggleModule(ctx, txDb.dbtx, gen.ToggleModuleParams{
			GuildID:    guildId,
			ModuleType: gen.GuildModuleTypeEventLogs,
			Enabled:    enabled,
		}); err != nil {
			return err
		}

		if _, err := txDb.queries.UpsertEventLogSettings(ctx, txDb.dbtx, gen.UpsertEventLogSettingsParams{
			GuildID:         guildId,
			IgnoredChannels: ignoredChannels,
			IgnoredRoles:    ignoredRoles,
			IgnoreBots:      ignoreBots,
		}); err != nil {
			return err
		}

		if _, err := txDb.EventLog.createOrUpdateManyLogChannels(ctx, guildId, channelIds, ActionLogSourcePanel); err != nil {
			return err
		}

		return nil
	})
}

func (e *eventLog) RemoveLogChannel(ctx context.Context, guildId, userId string, channelType LogChannelType, source ActionLogSource) error {
	return e.db.withTx(ctx, func(ctx context.Context, txDb *Database) error {
		if err := e.db.queries.UpdateGuildRegistryTime(ctx, txDb.dbtx, guildId); err != nil {
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
