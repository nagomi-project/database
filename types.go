package database

import "github.com/nagomi-project/database/internal/gen"

type PaginatedQuery[T any] struct {
	CurrentPage  int
	TotalPages   int
	TotalEntries int

	Data []T
}

type ActionLogSource = gen.ActionLogSource

const (
	ActionLogSourcePanel   = gen.ActionLogSourcePanel
	ActionLogSourceDiscord = gen.ActionLogSourceDiscord
)

type ActionLogType = gen.ActionLogType

const (
	ActionLogTypeGuildSettingsUpdate = gen.ActionLogTypeGuildSettingsUpdate
)

type LogChannelType = gen.EventLogType

const (
	EventLogTypeUserjoin           = gen.EventLogTypeUserjoin
	EventLogTypeUserleave          = gen.EventLogTypeUserleave
	EventLogTypeUserkick           = gen.EventLogTypeUserkick
	EventLogTypeUserban            = gen.EventLogTypeUserban
	EventLogTypeUserunban          = gen.EventLogTypeUserunban
	EventLogTypeUserrolesUpdate    = gen.EventLogTypeUserrolesUpdate
	EventLogTypeUsernicknameUpdate = gen.EventLogTypeUsernicknameUpdate
	EventLogTypeUservoiceJoin      = gen.EventLogTypeUservoiceJoin
	EventLogTypeUservoiceMove      = gen.EventLogTypeUservoiceMove
	EventLogTypeUservoiceLeave     = gen.EventLogTypeUservoiceLeave
	EventLogTypeMessagecreate      = gen.EventLogTypeMessagecreate
	EventLogTypeMessageedit        = gen.EventLogTypeMessageedit
	EventLogTypeMessagedelete      = gen.EventLogTypeMessagedelete
	EventLogTypeMessageimageRemove = gen.EventLogTypeMessageimageRemove
	EventLogTypeChannelcreate      = gen.EventLogTypeChannelcreate
	EventLogTypeChannelupdate      = gen.EventLogTypeChannelupdate
	EventLogTypeChanneldelete      = gen.EventLogTypeChanneldelete
	EventLogTypeRolecreate         = gen.EventLogTypeRolecreate
	EventLogTypeRoleupdate         = gen.EventLogTypeRoleupdate
	EventLogTypeRoledelete         = gen.EventLogTypeRoledelete
	EventLogTypeEmojicreate        = gen.EventLogTypeEmojicreate
	EventLogTypeEmojiupdate        = gen.EventLogTypeEmojiupdate
	EventLogTypeEmojidelete        = gen.EventLogTypeEmojidelete
)

type InfractionAction = gen.InfractionAction

const (
	InfractionActionNote    = gen.InfractionActionNote
	InfractionActionWarn    = gen.InfractionActionWarn
	InfractionActionMute    = gen.InfractionActionMute
	InfractionActionUnmute  = gen.InfractionActionUnmute
	InfractionActionKick    = gen.InfractionActionKick
	InfractionActionBan     = gen.InfractionActionBan
	InfractionActionUnban   = gen.InfractionActionUnban
	InfractionActionSoftban = gen.InfractionActionSoftban
)

type AppealStatus = gen.AppealStatus

const (
	AppealStatusSubmitted = gen.AppealStatusSubmitted
	AppealStatusApproved  = gen.AppealStatusApproved
	AppealStatusDenied    = gen.AppealStatusDenied
	AppealStatusBlocked   = gen.AppealStatusBlocked
	AppealStatusUnblocked = gen.AppealStatusUnblocked
)
