package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/nagomi-project/database/internal/gen"
)

type modules struct {
	db *Database
}

func newModules(db *Database) *modules {
	return &modules{db}
}

func (m *modules) IsEnabled(ctx context.Context, guildId string, module GuildModuleType) (bool, error) {
	status, err := m.db.queries.IsModuleEnabled(ctx, m.db.dbtx, gen.IsModuleEnabledParams{
		GuildID:    guildId,
		ModuleType: module,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}

		return false, err
	}

	return status, err
}

func (m *modules) SetEnabled(ctx context.Context, guildId string, module GuildModuleType, enabled bool) error {
	return m.db.queries.ToggleModule(ctx, m.db.dbtx, gen.ToggleModuleParams{
		GuildID:    guildId,
		ModuleType: module,
		Enabled:    enabled,
	})
}
