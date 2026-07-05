package database

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// NullableTimeToTimestamptz will return a pgtype.Timestamptz
func NullableTimeToTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t != nil {
		return pgtype.Timestamptz{Time: *t, Valid: true}
	}

	return pgtype.Timestamptz{Valid: false}
}

// NullableStringToText will return a pgtype.Text
func NullableStringToText(s *string) pgtype.Text {
	if s != nil {
		return pgtype.Text{String: *s, Valid: true}
	}

	return pgtype.Text{Valid: false}
}

// NullableBoolToBool will return a pgtype.Bool
func NullableBoolToBool(b *bool) pgtype.Bool {
	if b != nil {
		return pgtype.Bool{Bool: *b, Valid: true}
	}

	return pgtype.Bool{Valid: false}
}
