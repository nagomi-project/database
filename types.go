package database

import "github.com/nagomi-project/database/internal/gen"

type ActionLogSource = gen.ActionLogSource

const (
	ActionLogSourcePanel   = gen.ActionLogSourcePanel
	ActionLogSourceDiscord = gen.ActionLogSourceDiscord
)

type ActionLogType = gen.ActionLogType

const (
	ActionLogTypeGuildSettingsUpdate = gen.ActionLogTypeGuildSettingsUpdate
)

type LogChannelType = gen.LogChannelType

const (
	LogChannelTypeAll           = gen.LogChannelTypeAll
	LogChannelTypeMessage       = gen.LogChannelTypeMessage
	LogChannelTypeUser          = gen.LogChannelTypeUser
	LogChannelTypeInfractionLog = gen.LogChannelTypeInfractionLog
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
