package tgfun

import (
	"database/sql"

	tb "gopkg.in/telebot.v3"
)

// Funnel - telegram bot funnel
type Funnel struct {
	// public
	Data   FunnelData
	Script FunnelScript

	OnWebAppCallback func(ctx tb.Context) error

	// protected
	bot      *tb.Bot
	features funnelFeatures
}

type funnelFeatures struct {
	Users *UsersFeature
}

// UsersFeature - feature to enable users db
type UsersFeature struct {
	// required
	DBConn    *sql.DB
	TableName string

	// optional
	AdminChatID int64
}

type userData struct {
	ID         int64
	TelegramID int64
	Name       string
	TgName     string
}

// FunnelData - data container for Funnel struct
type FunnelData struct {
	Token     string `json:"token"`
	ImageRoot string `json:"imageRoot"`
}

// FunnelEvent - user interaction event
type FunnelEvent struct {
	Message EventMessage `json:"message"`
}

// EventMessage - funnel event message data
type EventMessage struct {
	// main data
	Text string `json:"text"`

	// instead of main data
	Callback BuildMessageCallback `json:"-"` // use it to redefine message

	// additional events
	OnEvent OnEventCallback `json:"-"`

	// optional
	Image            string               `json:"image"` // local filename or URL
	File             FileData             `json:"file"`
	Video            VideoData            `json:"video"`
	Buttons          []MessageButton      `json:"buttons"`
	Format           ParseFormat          `json:"format"`
	ButtonsIsColumns bool                 `json:"buttonsIsColumns"`
	Conversion       string               `json:"conversion"`
	OnConversion     OnConversionCallback `json:"-"`
	PinThisMessage   bool                 `json:"pin"`
}

type VideoData struct {
	Path             string `json:"path"`
	PreviewImagePath string `json:"preview"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
}

type FileData struct {
	Path             string `json:"path"`
	Name             string `json:"name"`
	PreviewImagePath string `json:"preview"`
}

type OnConversionCallback func(ctx tb.Context, conversionTag string) error

type OnEventCallback func(tb.Context) error

type BuildMessageCallback func(tb.Context) interface{}

// MessageButton - funnel event message button
type MessageButton struct {
	Text          string `json:"text"`
	NextMessageID string `json:"nextID"` // optional for URL-buttons
	URL           string `json:"url"`    // optional. only for URL-buttons
}

// FunnelScript - funnel scenario
type FunnelScript map[string]FunnelEvent // message ID -> event

// telegram user query handler
type QueryHandler struct {
	EventMessageID string
	EventData      FunnelEvent
	Menu           *tb.ReplyMarkup
	ParseMode      tb.ParseMode
	Bot            *tb.Bot
	FilesRoot      string // inherit from Funnel
	Features       *funnelFeatures
}
