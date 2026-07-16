package database

import "errors"

var (
	ErrUserNotMuted      = errors.New("user is not muted")
	ErrUserAlreadyMuted  = errors.New("user is already muted")
	ErrUserNotBanned     = errors.New("user is not banned")
	ErrUserAlreadyBanned = errors.New("user is already banned")

	ErrNoInfractions          = errors.New("user has no infractions")
	ErrInfractionPageNotFound = errors.New("the page for infactions does not exist")
	ErrInactiveInfraction     = errors.New("the infraction is no longer active")
)
