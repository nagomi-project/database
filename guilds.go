package database

import (
	"context"
	"fmt"

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
	if err := g.db.WithTx(ctx, func(ctx context.Context, txDb *Database) error {
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
	return g.db.WithTx(ctx, func(ctx context.Context, txDb *Database) error {
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
