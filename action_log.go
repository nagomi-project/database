package database

import (
	"context"
	"time"

	"github.com/nagomi-project/database/internal/gen"
)

type actionLog struct {
	db *Database
}

func newActionLog(db *Database) *actionLog {
	return &actionLog{db}
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
func (a *actionLog) GetActionLogHistory(ctx context.Context, guildId string, page int) (*PaginatedQuery[ActionLog], error) {
	page = max(page, 1)

	pageDetails, err := a.db.queries.GetActionLogPageDetails(ctx, a.db.dbtx, gen.GetActionLogPageDetailsParams{
		GuildID: guildId,
	})
	if err != nil {
		return nil, err
	}

	logPage, err := a.db.queries.GetActionLogPage(ctx, a.db.dbtx, gen.GetActionLogPageParams{
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
