package database

import (
	"context"
	"time"

	"github.com/nagomi-project/database/internal/config"
	"github.com/nagomi-project/database/internal/gen"
)

type oAuth struct {
	db *Database
}

func newOAuth(db *Database) *oAuth {
	return &oAuth{db}
}

type Session struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}

// CreateSession will create a new session.
func (o *oAuth) CreateSession(ctx context.Context, session string, clientId string, expiresAt time.Time, accessToken, refreshToken string) (*Session, error) {
	s, err := o.db.queries.CreateSession(ctx, o.db.dbtx, gen.CreateSessionParams{
		EncryptionKey: config.C.OAuthTokenEncryptionKey,
		Session:       session,
		ClientID:      clientId,
		Expiry:        NullableTimeToTimestamptz(&expiresAt),
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
	})
	if err != nil {
		return nil, err
	}

	details := &Session{
		UserID:       s.ClientID,
		AccessToken:  s.AccessToken,
		RefreshToken: s.RefreshToken,
	}

	return details, nil
}

// ValidateSession will validate a session.
func (o *oAuth) ValidateSession(ctx context.Context, session string) (*Session, error) {
	s, err := o.db.queries.ValidateSession(ctx, o.db.dbtx, gen.ValidateSessionParams{
		EncryptionKey: config.C.OAuthTokenEncryptionKey,
		Session:       session,
	})
	if err != nil {
		return nil, err
	}

	details := &Session{
		UserID:       s.ClientID,
		AccessToken:  s.AccessToken,
		RefreshToken: s.RefreshToken,
	}

	return details, nil
}

// DeleteExpiredSessions will delete all expired sessions.
func (o *oAuth) DeleteExpiredSessions(ctx context.Context) error {
	return o.db.queries.DeleteExpiredSessions(ctx, o.db.dbtx)
}
