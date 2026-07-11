package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	infractionsChannel = "infraction_expiry_schedule_events"
)

type DatabaseNotifier struct {
	conn *pgx.Conn

	Infractions chan InfractionsNotificationEvent
}

// NewDatabaseNotifier will create a new DatabaseNotifier object that is used for listening for certain operations made.
func NewDatabaseNotifier(conn *pgx.Conn) *DatabaseNotifier {
	return &DatabaseNotifier{
		conn:        conn,
		Infractions: make(chan InfractionsNotificationEvent, 32),
	}
}

func (n *DatabaseNotifier) dispatch(ctx context.Context, notif *pgconn.Notification) error {
	switch notif.Channel {
	case infractionsChannel:
		var event InfractionsNotificationEvent

		if err := json.Unmarshal([]byte(notif.Payload), &event); err != nil {
			return err
		}

		select {
		case n.Infractions <- event:
			return nil
		case <-ctx.Done():
			return nil
		}
	default:
		return nil
	}
}

// Run will begin the listener.
func (n *DatabaseNotifier) Run(ctx context.Context) error {
	if _, err := n.conn.Exec(ctx, fmt.Sprintf("LISTEN %s", pgx.Identifier{infractionsChannel}.Sanitize())); err != nil {
		return err
	}
	defer n.conn.Exec(ctx, fmt.Sprintf("UNSUBSCRIBE %s", pgx.Identifier{infractionsChannel}.Sanitize())) //nolint:errcheck

	defer close(n.Infractions)

	for {
		notif, err := n.conn.WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			return err
		}

		if err := n.dispatch(ctx, notif); err != nil {
			return err
		}
	}
}

type InfractionEventType string

var (
	InfractionEventTypeInserted InfractionEventType = "INSERT"
	InfractionEventTypeUpdated  InfractionEventType = "UPDATE"
	InfractionEventTypeRemoved  InfractionEventType = "DELETE"
)

type InfractionsNotificationEvent struct {
	Type       InfractionEventType `json:"type"`
	CaseNumber int32               `json:"case_number"`
	ExpiresAt  *time.Time          `json:"expires_at"`
	GuildID    string              `json:"guild_id"`
	MemberID   string              `json:"member_id"`
	Action     InfractionAction    `json:"action"`
}
