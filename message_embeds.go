package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/nagomi-project/database/internal/gen"
)

type messageEmbeds struct {
	db *Database
}

func newMessageEmbeds(db *Database) *messageEmbeds {
	return &messageEmbeds{db}
}

type MessageEmbedSettings struct {
	IgnoreChannels        []string
	IgnoreRoles           []string
	AllowCrossGuildEmbeds bool
}

func (m *messageEmbeds) getConfiguration(ctx context.Context, guildId string) (*MessageEmbedSettings, error) {
	config, err := m.db.queries.GetMessageEmbedSettings(ctx, m.db.dbtx, guildId)
	if err != nil {
		return nil, err
	}

	return &MessageEmbedSettings{
		IgnoreChannels:        config.IgnoredChannels,
		IgnoreRoles:           config.IgnoredRoles,
		AllowCrossGuildEmbeds: config.AllowCrossGuildEmbeds,
	}, nil
}

func (m *messageEmbeds) GetOrCreateConfiguration(ctx context.Context, guildId string) (*MessageEmbedSettings, error) {
	config, err := m.getConfiguration(ctx, guildId)
	if err != nil {
		if err == pgx.ErrNoRows {
			if err := m.db.queries.RegisterMessageEmbedSettingsIfMissing(ctx, m.db.dbtx, guildId); err != nil {
				return nil, err
			}

			return m.getConfiguration(ctx, guildId)
		}

		return nil, err
	}

	return config, nil
}

func (m *messageEmbeds) UpdateConfiguration(ctx context.Context, guildId string, enabled, allowCrossServer bool, ignoredChannels, ignoredRoles []string) error {
	return m.db.withTx(ctx, func(ctx context.Context, txDb *Database) error {
		if err := txDb.queries.ToggleModule(ctx, txDb.dbtx, gen.ToggleModuleParams{
			GuildID:    guildId,
			ModuleType: gen.GuildModuleTypeMessageEmbeds,
			Enabled:    enabled,
		}); err != nil {
			return err
		}

		if _, err := txDb.queries.UpsertMessageEmbedSettings(ctx, txDb.dbtx, gen.UpsertMessageEmbedSettingsParams{
			GuildID:               guildId,
			IgnoredChannels:       ignoredChannels,
			IgnoredRoles:          ignoredRoles,
			AllowCrossGuildEmbeds: allowCrossServer,
		}); err != nil {
			return err
		}

		return nil
	})
}
