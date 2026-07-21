package database

import (
	"context"
)

type guild struct {
	db *Database
}

func newGuildSettings(db *Database) *guild {
	return &guild{db}
}

func (g *guild) EnsureGuildRegistered(ctx context.Context, guildId string) error {
	return g.db.queries.RegisterGuildIfMissing(ctx, g.db.dbtx, guildId)
}
