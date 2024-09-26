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

	if f.OnWebAppCallback != nil {
		f.bot.Handle(tb.OnWebApp, f.OnWebAppCallback)
	}

	f.handleTextEvents()

	go f.bot.Start()
	return nil
}

func (f *Funnel) SetupOnWebAppLaunchCallback(cb func(ctx tb.Context) error) {
	f.OnWebAppCallback = cb
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
	eventMessageID := strings.ToLower(f.sanitizer.Sanitize(c.Text()))

	if _, isEventExists := f.Script[eventMessageID]; isEventExists {
		q, err := f.GetEventQueryHandler(eventMessageID)
		if err != nil {
			return fmt.Errorf("get query handler: %w", err)
		}

		return q.handleMessage(c)
	}

	// handle commands
	if f.features.Users == nil {
		return nil
	}

	if c.Chat().ID == f.features.Users.AdminChatID {
		return f.handleAdminMessage(c)
	}
	return nil
}

func (f *Funnel) GetEventQueryHandler(
	eventMessageID string,
) (*QueryHandler, error) {
	if _, isEventExists := f.Script[eventMessageID]; !isEventExists {
		return nil, fmt.Errorf("event %q not exists in funnel", eventMessageID)
	}

	menu := tb.ReplyMarkup{}

	return &QueryHandler{
		Script:         f.Script,
		EventMessageID: eventMessageID,
		EventData:      f.Script[eventMessageID],
		Menu:           &menu,
		ParseMode:      parseMode,
		Bot:            f.bot,
		FilesRoot:      f.Data.ImageRoot,
		Features:       &f.features,
		sanitizer:      f.sanitizer,
	}, nil
}

func (q *QueryHandler) createChildHandler(messageID string) (*QueryHandler, error) {
	if _, isEventExists := q.Script[messageID]; !isEventExists {
		return nil, fmt.Errorf("event %q not exists in funnel", messageID)
	}

	return &QueryHandler{
		Script:         q.Script,
		EventMessageID: messageID,
		EventData:      q.Script[messageID],
		Menu:           &tb.ReplyMarkup{},
		ParseMode:      q.ParseMode,
		Bot:            q.Bot,
		FilesRoot:      q.FilesRoot,
		Features:       q.Features,
		sanitizer:      q.sanitizer,
	}, nil
}

func (f *Funnel) handleEvent(
	eventMessageID string,
	_ FunnelEvent,
	_ tb.ParseMode,
) error {
	menu := tb.ReplyMarkup{}

	// create message handler
	q, err := f.GetEventQueryHandler(eventMessageID)
	if err != nil {
		return fmt.Errorf("get query handler: %w", err)
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

	// text query
	f.bot.Handle(strings.ToLower(eventMessageID), q.handleMessage)
	return nil
}

func (q *QueryHandler) actionNotify(telegramUserID int64, action tb.ChatAction) {
	if err := q.Bot.Notify(tb.ChatID(telegramUserID), action); err != nil {
		log.Println("notify:", err)
	}
}

func (q *QueryHandler) makeConversion(
	telegramUserID int64,
	conversion string,
	payload UserPayload,
) {
	err := q.EventData.Message.OnConversion(telegramUserID, conversion, payload)
	if err != nil {
		log.Printf(
			"handle conversion %q in tgfun: %s\n",
			conversion,
			err.Error(),
		)
	}
}

func (q *QueryHandler) handleConversions(telegramUserID int64, payload UserPayload) {
	if q.EventData.Message.Conversion != "" {
		q.makeConversion(
			telegramUserID,
			q.EventData.Message.Conversion,
			payload,
		)
		return
	}

	if len(q.EventData.Message.Conversions) > 0 {
		for _, conversion := range q.EventData.Message.Conversions {
			q.makeConversion(
				telegramUserID,
				conversion,
				payload,
			)
		}
	}
}

func (q *QueryHandler) buildMessage(
	telegramUserID int64,
	payload UserPayload,
) interface{} {
	if q.EventData.Message.Callback != nil {
		return q.EventData.Message.Callback(telegramUserID)
	}

	if q.EventData.Message.OnConversion != nil {
		q.handleConversions(telegramUserID, payload)
	}

	// get message by type
	switch getMessageType(q.EventData.Message) {
	default:
		return getTextMessage(q.EventData.Message)
	case MessageTypePhoto:
		q.actionNotify(telegramUserID, tb.UploadingPhoto)
		return getPhotoMessage(q.EventData.Message, q.FilesRoot)
	case MessageTypeDocument:
		q.actionNotify(telegramUserID, tb.UploadingDocument)
		return getDocumentMessage(q.EventData.Message, q.FilesRoot)
	case MessageTypeVideo:
		q.actionNotify(telegramUserID, tb.UploadingVideo)
		return getVideoMessage(q.EventData.Message, q.FilesRoot)
	case MessageTypeAudio:
		q.actionNotify(telegramUserID, tb.UploadingAudio)
		return getAudioMessage(q.EventData.Message, q.FilesRoot)
	}
}

func (q *QueryHandler) CustomHandle(telegramUserID int64) error {
	msg := q.buildMessage(telegramUserID, UserPayload{})
	q.buildButtons(telegramUserID)

	var format = parseMode
	if q.EventData.Message.Format != "" {
		format = string(q.EventData.Message.Format)
	}

	return q.send(telegramUserID, msg, format)
}

func (q *QueryHandler) handleMessage(ctx tb.Context) error {
	if strings.HasPrefix(ctx.Text(), startMessageCode) && ctx.Message().Payload != "" {
		sanitizedPayload := q.sanitizer.Sanitize(ctx.Message().Payload)

		payload, err := FilterUserPayload(sanitizedPayload)
		if err != nil {
			log.Println("filter user payload:", err)
			return q.buildAndSend(ctx, payload)
		}

		// обработка случая, когда пользователь вернулся в бота по backling-ссылке
		if payload.BackLinkEventID != "" {
			// найдем, есть ли эвент с таким ID
			// попробуем найти в нижнем регистре
			eventID := strings.ToLower(payload.BackLinkEventID)
			if _, isEventExists := q.Script[eventID]; isEventExists {
				// такой эвент есть
				return q.handleChildQuery(ctx, eventID, payload)
			}

			// попробуем найти в верхнем регистре
			eventID = payload.BackLinkEventID
			if _, isEventExists := q.Script[eventID]; isEventExists {
				// такой эвент есть
				return q.handleChildQuery(ctx, eventID, payload)
			}

			// эвент не найден, продолжим обработку стартового сообщения
			log.Printf("event %q not found\n", eventID)
		}
	}

	return q.buildAndSend(ctx, UserPayload{})
}

func (q *QueryHandler) handleChildQuery(
	ctx tb.Context,
	eventID string,
	payload UserPayload,
) error {
	childHandler, err := q.createChildHandler(eventID)
	if err != nil {
		return fmt.Errorf("create child handler: %w", err)
	}

	return childHandler.buildAndSend(ctx, payload)
}

func (q *QueryHandler) buildAndSend(ctx tb.Context, payload UserPayload) error {
	msg := q.buildMessage(ctx.Sender().ID, payload)
	q.buildButtons(ctx.Sender().ID)

	if q.EventData.Message.OnEvent != nil {
		if err := q.EventData.Message.OnEvent(ctx); err != nil {
			return fmt.Errorf("handle event custom callback: %w", err)
		}
	}

	if q.Features.Users != nil {
		_, err := q.Features.Users.getUserData(ctx.Sender())
		if err != nil {
			return fmt.Errorf("get user data: %w", err)
		}
	}

	return q.sendWithCheck(ctx, msg, payload)
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
			tb.ModeMarkdown,
		)
		return fmt.Errorf("send message: %w", err)
	}

	for _, telegramUserID := range telegramUserIDs {
		_, err := f.bot.Send(tb.ChatID(telegramUserID), adminPostText, tb.ModeMarkdown)
		if err != nil {
			return fmt.Errorf("send message: %w", err)
		}
	}
	return nil
}

func (q *QueryHandler) handleButton(c tb.Context) error {
	defer c.Respond()

	// button events doesn't have payload
	msg := q.buildMessage(c.Sender().ID, UserPayload{})
	q.buildButtons(c.Sender().ID)

	return q.sendWithCheck(c, msg, UserPayload{})
}

func (q *QueryHandler) sendWithCheck(
	c tb.Context,
	msg interface{},
	payload UserPayload,
) error {
	var format = parseMode
	if q.EventData.Message.Format != "" {
		format = string(q.EventData.Message.Format)
	}

	lockerPassed, err := q.checkLocker(c)
	if err != nil {
		log.Println(err)
	}
	if !lockerPassed {
		lockerMessageHandler, err := q.createChildHandler(q.EventData.SubscriptionLocker.
			LockerMessageID)
		if err != nil {
			log.Println(err)
		} else {
			msg := lockerMessageHandler.buildMessage(
				c.Sender().ID,
				payload,
			)

			lockerMessageHandler.buildButtons(c.Sender().ID)
			return lockerMessageHandler.send(
				c.Sender().ID,
				msg,
				string(lockerMessageHandler.EventData.Message.Format),
			)
		}
	}

	var args = []interface{}{tb.ParseMode(format)}
	if q.EventData.Message.DisablePreview {
		args = append(args, tb.NoPreview)
	}

	return q.send(c.Sender().ID, msg, args...)
}

// проверим, можем ли отправить сообщение или есть какие-то блокирующие штуки
// провде необходимости подписки на канал.
// возвращает true, если можно отправлять сообщение.
func (q *QueryHandler) checkLocker(c tb.Context) (bool, error) {
	if !q.EventData.SubscriptionLocker.Enabled {
		return true, nil
	}

	isJoined, err := q.isUserJoined(q.EventData.SubscriptionLocker.ChatID, c.Sender())
	if err != nil {
		return false, fmt.Errorf(
			"check user joined %v: %w",
			q.EventData.SubscriptionLocker.ChatID, err,
		)
	}
	return isJoined, nil
}

func (q *QueryHandler) isUserJoined(chatID int64, user tb.Recipient) (bool, error) {
	member, err := q.Bot.ChatMemberOf(tb.ChatID(chatID), user)
	if err != nil {
		return false, fmt.Errorf("check subscription: %w", err)
	}

	switch member.Role {
	default:
		return false, fmt.Errorf("unknown member role: %q", member.Role)
	case tb.Creator, tb.Member, tb.Administrator:
		return true, nil
	case tb.Restricted, tb.Kicked:
		return false, fmt.Errorf("sorry, you were banned")
	case tb.Left:
		return false, nil
	}
}

func (q *QueryHandler) send(
	chatID int64,
	message interface{},
	args ...interface{},
) error {
	args = append(args, q.Menu)

	messageResponse, err := q.Bot.Send(tb.ChatID(chatID), message, args...)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	if q.EventData.Message.PinThisMessage {
		if err := q.Bot.Pin(messageResponse); err != nil {
			log.Printf("pin message: %s\n", err.Error())
		}
	}
	return nil
}

func (q *QueryHandler) buildButtons(telegramUserID int64) {
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
				// next event button
				btn = q.Menu.Data(btnData.Text, btnData.NextMessageID)
			} else {
				// URL button
				btnURL := btnData.URL
				if q.Features.IsUTMTagsFeatureActive() && btnData.UseUTMTags {
					utmTags := q.Features.UTM.GetUserUTMTags(telegramUserID)
					newURL, err := addUtmTags(btnURL, utmTags)
					if err != nil {
						log.Println("failed to add utm tags to url:", err)
					}

					btnURL = newURL
				}

				btn = q.Menu.URL(btnData.Text, btnURL)
			}

			rows = append(rows, q.Menu.Row(btn))
			btns = append(btns, btn)
		}

		// handle layout style
		if q.EventData.Message.ButtonsIsColumns {
			q.Menu.Inline(rows...)
		} else {
			q.Menu.Inline(q.Menu.Row(btns...))
		}
	}
}
