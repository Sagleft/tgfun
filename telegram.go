package tgfun

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	tb "gopkg.in/telebot.v3"
)

func (f *Funnel) formatMessages() {
	for key, value := range f.Script {
		value.Message.Text = formatMessage(f.Script[key].Message.Text)
		f.Script[key] = value
	}
}

func formatMessage(message string) string {
	var result []string
	lines := strings.Split(message, "\n\n")
	for _, val := range lines {
		val = strings.Trim(val, " ")
		val = strings.ReplaceAll(val, "	", "")

		result = append(result, val)
	}
	return strings.Join(result, "\n\n")
}

// Run funnel. This is a non-blocking operation
func (f *Funnel) Run() error {
	if f.Data.Token == "" {
		return errors.New("bot token is not set")
	}

	f.formatMessages()

	var err error
	f.bot, err = tb.NewBot(tb.Settings{
		Token:  f.Data.Token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return errors.New("failed to setup telegram bot: " + err.Error())
	}

	if err := f.handleStartMessage(); err != nil {
		return fmt.Errorf("handle start message: %w", err)
	}

	if err := f.handleScriptEvents(); err != nil {
		return fmt.Errorf("handle script events: %w", err)
	}

	f.handleTextEvents()

	go f.bot.Start()
	return nil
}

func (f *Funnel) handleStartMessage() error {
	parseMode := tb.ParseMode(parseMode)

	startEvent, isStartMessageFound := f.Script[startMessageCode]
	if !isStartMessageFound {
		return errors.New("start message not found in script")
	}

	return f.handleEvent(startMessageCode, startEvent, parseMode)
}

func (f *Funnel) handleScriptEvents() error {
	parseMode := tb.ParseMode(parseMode)

	for eventMessageID, eventData := range f.Script {
		if eventMessageID == startMessageCode {
			continue
		}

		err := f.handleEvent(eventMessageID, eventData, parseMode)
		if err != nil {
			return fmt.Errorf("handle event: %w", err)
		}
	}

	return nil
}

func (f *Funnel) handleTextEvents() {
	f.bot.Handle(tb.OnText, f.handleTextMessage)
}

func (f *Funnel) handleTextMessage(c tb.Context) error {
	if f.features.Users == nil {
		return nil
	}

	if c.Chat().ID == f.features.Users.AdminChatID {
		return f.handleAdminMessage(c)
	}
	return nil
}

func (f *Funnel) handleEvent(
	eventMessageID string,
	eventData FunnelEvent,
	parseMode tb.ParseMode,
) error {
	menu := tb.ReplyMarkup{}

	// create message handler
	q := queryHandler{
		EventMessageID: eventMessageID,
		EventData:      eventData,
		Menu:           &menu,
		ParseMode:      parseMode,
		Bot:            f.bot,
		ImageRoot:      f.Data.ImageRoot,
		Features:       &f.features,
	}

	// build message
	if strings.Contains(eventMessageID, "/") {
		// command or text message
		f.bot.Handle(eventMessageID, q.handleMessage)
		return nil
	}

	// button query
	btnListener := menu.Data("listener", eventMessageID)
	f.bot.Handle(&btnListener, q.handleButton)
	return nil
}

func (q *queryHandler) buildMessage() interface{} {
	if q.EventData.Message.Image == "" {
		if q.EventData.Message.Text == "" {
			return "unknown message type"
		}
		return q.EventData.Message.Text
	}
	photo := &tb.Photo{}
	if strings.Contains(q.EventData.Message.Image, "http") {
		photo.File = tb.FromURL(q.EventData.Message.Image)
	} else {

		imgPath := q.ImageRoot
		if !strings.HasSuffix(imgPath, "/") {
			imgPath += "/"
		}
		imgPath += q.EventData.Message.Image
		photo.File = tb.FromDisk(imgPath)
	}
	if q.EventData.Message.Text != "" {
		photo.Caption = q.EventData.Message.Text
	}
	return photo
}

func (q *queryHandler) handleMessage(c tb.Context) error {
	msg := q.buildMessage()
	q.buildButtons()

	if q.Features.Users != nil {
		_, err := q.Features.Users.getUserData(c.Sender())
		if err != nil {
			return fmt.Errorf("get user data: %w", err)
		}
	}

	if _, err := q.Bot.Send(c.Sender(), msg, q.Menu, parseMode); err != nil {
		log.Println("failed to send query handler (from message) message: " + err.Error())
	}
	return nil
}

func (f *Funnel) handleAdminMessage(c tb.Context) error {
	if !strings.HasPrefix(c.Text(), adminPostToAllPrefix) {
		if _, err := f.bot.Send(c.Sender(), "Не могу разобрать сообщение"); err != nil {
			return fmt.Errorf("send message: %w", err)
		}
		return nil
	}

	adminPostText := strings.Replace(c.Text(), adminPostToAllPrefix, "", 1)
	telegramUserIDs, err := f.features.Users.getUsersTelegramIDs()
	if err != nil {
		_, err := f.bot.Send(
			tb.ChatID(f.features.Users.AdminChatID),
			err.Error(),
			parseMode,
		)
		return fmt.Errorf("send message: %w", err)
	}

	for _, telegramUserID := range telegramUserIDs {
		_, err := f.bot.Send(tb.ChatID(telegramUserID), adminPostText, tb.ModeHTML)
		if err != nil {
			return fmt.Errorf("send message: %w", err)
		}
	}
	return nil
}

func (q *queryHandler) handleButton(c tb.Context) error {
	defer c.Respond()

	msg := q.buildMessage()
	q.buildButtons()

	_, err := q.Bot.Send(c.Sender(), msg, q.Menu, parseMode)
	if err != nil {
		return fmt.Errorf("failed to send query handler (from btn) message: %w", err)
	}
	return nil
}

func (q *queryHandler) buildButtons() {
	if q.EventData.Message.Buttons == nil {
		return
	}

	if len(q.EventData.Message.Buttons) == 0 {
		return
	}

	if len(q.EventData.Message.Buttons) > 0 {
		var rows []tb.Row
		var btns []tb.Btn

		for _, btnData := range q.EventData.Message.Buttons {
			var btn tb.Btn
			if btnData.URL == "" {
				btn = q.Menu.Data(btnData.Text, btnData.NextMessageID)
			} else {
				btn = q.Menu.URL(btnData.Text, btnData.URL)
			}

			rows = append(rows, q.Menu.Row(btn))
			btns = append(btns, btn)
		}

		if q.EventData.Message.ButtonsIsColumns {
			q.Menu.Inline(rows...)
		} else {
			q.Menu.Inline(q.Menu.Row(btns...))
		}
	}
}
