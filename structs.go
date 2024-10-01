package tgfun

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/microcosm-cc/bluemonday"
	tb "gopkg.in/telebot.v3"
)

type UserPayload struct {
	UTMSource       string `json:"s"`
	UTMCampaign     string `json:"c"`
	BackLinkEventID string `json:"b"`
	Yclid           string `json:"y"`
}

func (p UserPayload) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		log.Println("encode payload:", err)
		return "{}"
	}

	return string(data)
}

func (p UserPayload) IsEmpty() bool {
	return p.UTMSource == "" &&
		p.UTMCampaign == "" &&
		p.BackLinkEventID == ""
}

// Funnel - telegram bot funnel
type Funnel struct {
	// public
	Data   FunnelData
	Script FunnelScript

	OnWebAppCallback func(ctx tb.Context) error

	// protected
	bot       *tb.Bot
	features  funnelFeatures
	sanitizer *bluemonday.Policy
}

type funnelFeatures struct {
	Users     *UsersFeature
	UTM       *UTMTagsFeature
	UserInput *UserInputFeature
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
	Message            EventMessage `json:"message"`
	SubscriptionLocker EventLocker  `json:"locker"`
}

type EventLocker struct {
	Enabled         bool   `json:"enabled"`
	ChatID          int64  `json:"chatID"`
	LockerMessageID string `json:"lockerMessageID"`
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
	Audio            AudioData            `json:"audio"`
	Video            VideoData            `json:"video"`
	Buttons          []MessageButton      `json:"buttons"`
	Format           ParseFormat          `json:"format"`
	ButtonsIsColumns bool                 `json:"buttonsIsColumns"`
	Conversion       string               `json:"conversion"`  // optional
	Conversions      []string             `json:"conversions"` // optional
	OnConversion     OnConversionCallback `json:"-"`
	PinThisMessage   bool                 `json:"pin"`
	DisablePreview   bool                 `json:"disablePreview"`
}

type AudioData struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Duration int    `json:"duration"`
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

type OnConversionCallback func(
	telegramUserID int64,
	conversionTag string,
	payload UserPayload,
) error

type OnEventCallback func(tb.Context) error

type BuildMessageCallback func(telegramUserID int64) interface{}

// MessageButton - funnel event message button
type MessageButton struct {
	Text          string `json:"text"`
	NextMessageID string `json:"nextID"`     // optional for URL-buttons
	URL           string `json:"url"`        // optional. only for URL-buttons
	UseUTMTags    bool   `json:"useUtmTags"` // optional
}

// FunnelScript - funnel scenario
type FunnelScript map[string]FunnelEvent // message ID -> event

// telegram user query handler
type QueryHandler struct {
	Script FunnelScript

	EventMessageID string
	EventData      FunnelEvent
	Menu           *tb.ReplyMarkup
	ParseMode      tb.ParseMode
	Bot            *tb.Bot
	FilesRoot      string // inherit from Funnel
	Features       *funnelFeatures
	sanitizer      *bluemonday.Policy
}
