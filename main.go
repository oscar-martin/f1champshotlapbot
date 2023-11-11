package main

import (
	"context"
	"f1champshotlapsbot/pkg/servers"
	"f1champshotlapsbot/pkg/tracks"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	EnvTelegramToken = "TELEGRAM_TOKEN"
	EnvHotlapsDomain = "API_DOMAIN"
)

type Accepter interface {
	AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error)
	AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error)
	AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery))
}

var (
	domain = ""
	m      *Menu
	tm     *tracks.Manager
	sm     *servers.Manager
	bot    *tgbotapi.BotAPI

	accepters = []Accepter{}

	menuKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(tracks.ButtonHotlaps),
			tgbotapi.NewKeyboardButton(servers.ButtonLive),
		),
	)
)

func main() {
	var err error
	// get token from env
	token := os.Getenv(EnvTelegramToken)
	if token == "" {
		log.Fatalf("%s is not set", EnvTelegramToken)
	}
	domain = os.Getenv(EnvHotlapsDomain)
	if domain == "" {
		log.Fatalf("%s is not set", EnvHotlapsDomain)
	}

	domain = strings.TrimRight(domain, "/")

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		// Abort if something is wrong
		log.Panic(err)
	}

	// Set this to true to log all interactions with telegram servers
	bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Create a new cancellable background context. Calling `cancel()` leads to the cancellation of the context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// `updates` is a golang channel which receives telegram updates
	updates := bot.GetUpdatesChan(u)

	// Pass cancellable context to goroutine
	go receiveUpdates(ctx, updates)

	// TODO: remove this
	// ---------------------
	CreateServers([]int{10001, 10002, 10004})
	// ---------------------

	m = NewMenu(bot)
	tm = tracks.NewTrackManager(bot, domain)
	sm = servers.NewManager(bot, domain)

	// Register the accepters
	accepters = append(accepters, m)
	accepters = append(accepters, tm)
	accepters = append(accepters, sm)

	// Tell the user the bot is online
	log.Println("Start listening for updates. Press Ctrl-C to stop it")

	refreshHotlapsTicker := time.NewTicker(60 * time.Minute)
	refreshServersTicker := time.NewTicker(5 * time.Minute)
	exitChan := make(chan bool)

	// start the trackmanager sync process
	tm.Sync(ctx, refreshHotlapsTicker, exitChan)
	sm.Sync(ctx, refreshServersTicker, exitChan)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// lock the main thread until we receive a signal
	<-sigs

	refreshHotlapsTicker.Stop()
	refreshServersTicker.Stop()
	exitChan <- true

	cancel()
}

func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	// `for {` means the loop is infinite until we manually stop it
	for {
		select {
		// stop looping if ctx is cancelled
		case <-ctx.Done():
			return
		// receive update from channel and then handle it
		case update := <-updates:
			handleUpdate(ctx, update)
		}
	}
}

func handleUpdate(ctx context.Context, update tgbotapi.Update) {
	tm.Lock()
	defer tm.Unlock()
	switch {
	// Handle messages
	case update.Message != nil:
		MessageHandler(ctx, update.Message)
	// Handle button clicks
	case update.CallbackQuery != nil:
		CallbackQueryHandler(ctx, update.CallbackQuery)
	}
}

func MessageHandler(ctx context.Context, message *tgbotapi.Message) {
	user := message.From
	text := message.Text

	if user == nil {
		return
	}

	// Print to console
	log.Printf("%s wrote %s", user.FirstName, text)

	var err error
	if message.IsCommand() {
		// text is `/command-name`
		err = handleCommand(ctx, message.Chat.ID, text)
	} else {
		// text is `button-text`
		err = handleButton(ctx, message.Chat.ID, text)
	}

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

// When we get a button clicked, we react accordingly
func handleButton(ctx context.Context, chatId int64, button string) error {
	for _, accepter := range accepters {
		if accept, handler := accepter.AcceptButton(button); accept {
			return handler(ctx, chatId)
		}
	}
	return nil
}

// When we get a command, we react accordingly
func handleCommand(ctx context.Context, chatId int64, command string) error {
	for _, accepter := range accepters {
		if accept, handler := accepter.AcceptCommand(command); accept {
			return handler(ctx, chatId)
		}
	}
	return nil
}

func CallbackQueryHandler(ctx context.Context, query *tgbotapi.CallbackQuery) {
	for _, accepter := range accepters {
		if accept, handler := accepter.AcceptCallback(query); accept {
			handler(ctx, query)
			return
		}
	}
}
