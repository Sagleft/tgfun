package tgfun

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

// Funnel - telegram bot funnel
type Funnel struct {
	// public
	ImageRoot string
	Data      FunnelData
	Script    FunnelScript

	// protected
	bot *tb.Bot
}

// FunnelData - data container for Funnel struct
type FunnelData struct {
	Token string `json:"token"`
}

// FunnelEvent - user interaction event
type FunnelEvent struct {
	Message EventMessage `json:"message"`
}

// EventMessage - funnel event message data
type EventMessage struct {
	//ID      string          `json:"id"`
	Text    string          `json:"text"`
	Image   string          `json:"image"`   // local filename or URL. optional
	Buttons []MessageButton `json:"buttons"` // optional
}

// MessageButton - funnel event message button
type MessageButton struct {
	Text          string `json:"text"`
	NextMessageID string `json:"nextID"` // optional for URL-buttons
	URL           string `json:"url"`    // optional. only for URL-buttons
}

// FunnelScript - funnel scenario
type FunnelScript map[string]FunnelEvent // message ID -> event

// telegram user query handler
type queryHandler struct {
	EventMessageID string
	EventData      FunnelEvent
	Menu           *tb.ReplyMarkup
	ParseMode      tb.ParseMode
	Bot            *tb.Bot
	ImageRoot      string // inherit from Funnel
}
