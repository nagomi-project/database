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
	EventLogTypeUserJoin           = gen.EventLogTypeUserJoin
	EventLogTypeUserLeave          = gen.EventLogTypeUserLeave
	EventLogTypeUserKick           = gen.EventLogTypeUserKick
	EventLogTypeUserBan            = gen.EventLogTypeUserBan
	EventLogTypeUserUnban          = gen.EventLogTypeUserUnban
	EventLogTypeUserTimeout        = gen.EventLogTypeUserTimeout
	EventLogTypeUserRolesUpdate    = gen.EventLogTypeUserRolesUpdate
	EventLogTypeUserNicknameUpdate = gen.EventLogTypeUserNicknameUpdate
	EventLogTypeUserVoiceJoin      = gen.EventLogTypeUserVoiceJoin
	EventLogTypeUserVoiceMove      = gen.EventLogTypeUserVoiceMove
	EventLogTypeUserVoiceLeave     = gen.EventLogTypeUserVoiceLeave
	EventLogTypeMessageEdit        = gen.EventLogTypeMessageEdit
	EventLogTypeMessageDelete      = gen.EventLogTypeMessageDelete
	EventLogTypeMessageImageRemove = gen.EventLogTypeMessageImageRemove
	EventLogTypeChannelCreate      = gen.EventLogTypeChannelCreate
	EventLogTypeChannelUpdate      = gen.EventLogTypeChannelUpdate
	EventLogTypeChannelDelete      = gen.EventLogTypeChannelDelete
	EventLogTypeRoleCreate         = gen.EventLogTypeRoleCreate
	EventLogTypeRoleUpdate         = gen.EventLogTypeRoleUpdate
	EventLogTypeRoleDelete         = gen.EventLogTypeRoleDelete
	EventLogTypeEmojiCreate        = gen.EventLogTypeEmojiCreate
	EventLogTypeEmojiUpdate        = gen.EventLogTypeEmojiUpdate
	EventLogTypeEmojiDelete        = gen.EventLogTypeEmojiDelete
)

type ModerationAction = gen.ModerationAction

const (
	ModerationActionNote    = gen.ModerationActionNote
	ModerationActionWarn    = gen.ModerationActionWarn
	ModerationActionMute    = gen.ModerationActionMute
	ModerationActionUnmute  = gen.ModerationActionUnmute
	ModerationActionKick    = gen.ModerationActionKick
	ModerationActionBan     = gen.ModerationActionBan
	ModerationActionUnban   = gen.ModerationActionUnban
	ModerationActionSoftban = gen.ModerationActionSoftban
)

type AppealStatus = gen.AppealStatus

const (
	AppealStatusSubmitted = gen.AppealStatusSubmitted
	AppealStatusApproved  = gen.AppealStatusApproved
	AppealStatusDenied    = gen.AppealStatusDenied
	AppealStatusBlocked   = gen.AppealStatusBlocked
	AppealStatusUnblocked = gen.AppealStatusUnblocked
)

type GuildModuleType = gen.GuildModuleType

const (
	GuildModuleTypeInfractions      = gen.GuildModuleTypeInfractions
	GuildModuleTypeBanAppeals       = gen.GuildModuleTypeBanAppeals
	GuildModuleTypeEventLogs        = gen.GuildModuleTypeEventLogs
	GuildModuleTypeTickets          = gen.GuildModuleTypeTickets
	GuildModuleTypeModMail          = gen.GuildModuleTypeModMail
	GuildModuleTypeVoiceRooms       = gen.GuildModuleTypeVoiceRooms
	GuildModuleTypeActivityTracking = gen.GuildModuleTypeActivityTracking
)
