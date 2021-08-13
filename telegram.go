package tgfun

import (
	"errors"
	"log"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (f *Funnel) setupBot() error {
	if f.Data.Token == "" {
		return errors.New("bot token is not set")
	}

	var err error
	f.bot, err = tb.NewBot(tb.Settings{
		Token:  f.Data.Token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return errors.New("failed to setup telegram bot: " + err.Error())
	}
	return nil
}

// Run funnel. This is a non-blocking operation
func (f *Funnel) Run() error {
	err := f.createBindings()
	if err != nil {
		return err
	}

	// check telegram bot connection in cron
	checkBotCron := newCronHandler(f.checkBot, time.Second*botCheckCronTimeSeconds)
	go checkBotCron.run()

	// listen requests
	go f.bot.Start()
	return nil
}

func (f *Funnel) createBindings() error {
	return checkErrors(
		f.handleStartMessage,
		f.handleScriptEvents,
	)
}

func (f *Funnel) checkBot() {
	if f.bot == nil {
		// bot is not initiated
		err := f.setupBot()
		if err != nil {
			log.Fatalln(err)
		}
		return
	}

	// bot already initiated. send test query
	_, err := f.bot.GetCommands()
	if err == nil {
		return
	}
	log.Println("failed to get bot updates: " + err.Error())
	log.Println("trying to reconnect")

	// re-setup bot
	err = checkErrors(
		f.setupBot,
		f.createBindings,
	)
	if err != nil {
		log.Println(err)
	}
}

func (f *Funnel) handleStartMessage() error {
	menu := tb.ReplyMarkup{}
	parseMode := tb.ParseMode(parseMode)

	startEvent, isStartMessageFound := f.Script[startMessageID]
	if !isStartMessageFound {
		return errors.New("start message not found in script")
	}

	return f.handleEvent(startMessageID, startEvent, &menu, parseMode)
}

func (f *Funnel) handleScriptEvents() error {
	menu := tb.ReplyMarkup{}
	parseMode := tb.ParseMode(parseMode)

	for eventMessageID, eventData := range f.Script {
		if eventMessageID == startMessageID {
			continue
		}

		err := f.handleEvent(eventMessageID, eventData, &menu, parseMode)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Funnel) handleEvent(
	eventMessageID string,
	eventData FunnelEvent,
	menu *tb.ReplyMarkup,
	parseMode tb.ParseMode,
) error {
	// create message handler
	q := queryHandler{
		EventMessageID: eventMessageID,
		EventData:      eventData,
		Menu:           menu,
		ParseMode:      parseMode,
		Bot:            f.bot,
		ImageRoot:      f.ImageRoot,
	}

	// build message
	if strings.Contains(eventMessageID, "/") {
		// command or text message
		f.bot.Handle(eventMessageID, q.handleMessage)
	} else {
		// button query
		btnListener := menu.Data("listener", eventMessageID)
		f.bot.Handle(&btnListener, q.handleButton)
	}
	return nil
}

func (q *queryHandler) buildMessage() interface{} {
	if q.EventData.Message.Image == "" {
		if q.EventData.Message.Text == "" {
			return "unknown message type"
		}
		return q.EventData.Message.Text
	}
	photo := &tb.Photo{
		ParseMode: parseMode,
	}
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

func (q *queryHandler) handleMessage(m *tb.Message) {
	msg := q.buildMessage()
	q.buildButtons()

	_, err := q.Bot.Send(m.Sender, msg, q.Menu, parseMode)
	if err != nil {
		log.Println("failed to send query handler (from message) message: " + err.Error())
	}
}

func (q *queryHandler) handleButton(c *tb.Callback) {
	msg := q.buildMessage()
	q.buildButtons()

	_, err := q.Bot.Send(c.Sender, msg, q.Menu, parseMode)
	if err != nil {
		log.Println("failed to send query handler (from btn) message: " + err.Error())
	}
}

func (q *queryHandler) buildButtons() {
	if q.EventData.Message.Buttons == nil {
		return
	}
	if len(q.EventData.Message.Buttons) == 0 {
		return
	}
	buttons := []tb.Btn{}
	var btn tb.Btn
	for _, btnData := range q.EventData.Message.Buttons {
		if btnData.URL == "" {
			btn = q.Menu.Data(btnData.Text, btnData.NextMessageID)
		} else {
			btn = q.Menu.URL(btnData.Text, btnData.URL)
		}

		buttons = append(buttons, btn)
	}
	if len(buttons) > 0 {
		q.Menu.Inline(
			q.Menu.Row(buttons...),
		)
	}
}
