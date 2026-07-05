package database

import (
	"context"
	"database/sql"
	"errors"
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

func (g *guildSettings) CreateOrUpdateLogChannel(ctx context.Context, guildId, channelId, userId string, channelType LogChannelType, source ActionLogSource) (*gen.LogChannel, error) {
	var channel *gen.LogChannel
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

type LogChannelSettings struct {
	Type      LogChannelType
	ChannelID string
}

type GuildSettings struct {
	LastModified time.Time

	Infractions InfractionSettings
	LogChannels []LogChannelSettings
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

		logChannels, err := txDb.queries.GetGuildLogChannels(ctx, txDb.dbtx, guildId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		settings = &GuildSettings{
			LastModified: registry.UpdatedAt.Time,
			Infractions: InfractionSettings{
				AppealDuration:               infSettings.AppealDuration,
				ShouldRequestInfractionProof: infSettings.RequestInfractionProof,
			},
			LogChannels: make([]LogChannelSettings, len(logChannels)),
		}
		if infSettings.MutedRoleID.Valid {
			settings.Infractions.MutedRoleID = &infSettings.MutedRoleID.String
		}
		if infSettings.InfractionProofID.Valid {
			settings.Infractions.InfractionProofChannelID = &infSettings.InfractionProofID.String
		}

		for idx, logChannel := range logChannels {
			settings.LogChannels[idx] = LogChannelSettings{
				Type:      logChannel.Type,
				ChannelID: logChannel.ChannelID,
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return settings, nil
}
