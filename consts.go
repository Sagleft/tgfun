package tgfun

import tb "gopkg.in/telebot.v3"

const (
	startMessageCode     = "/start"
	parseMode            = tb.ModeMarkdown
	adminPostToAllPrefix = "!all"
)

type ParseFormat string

const (
	ParseFormatMarkdown ParseFormat = ParseFormat(tb.ModeMarkdown)
	ParseFormatHTML     ParseFormat = ParseFormat(tb.ModeHTML)
)

type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypePhoto    MessageType = "photo"
	MessageTypeDocument MessageType = "document"
	MessageTypeVideo    MessageType = "video"
)
