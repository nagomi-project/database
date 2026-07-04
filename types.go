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
