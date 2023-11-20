package main

import (
	"context"
	"f1champshotlapsbot/pkg/apps"
	"f1champshotlapsbot/pkg/apps/mainapp"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"f1champshotlapsbot/pkg/pubsub"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	EnvTelegramToken = "TELEGRAM_TOKEN"
	EnvHotlapsDomain = "API_DOMAIN"
)

var (
	domain    = ""
	bot       *tgbotapi.BotAPI
	app       apps.Accepter
	pubsubMgr *pubsub.PubSub
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

	pubsubMgr = pubsub.NewPubSub()

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

	exitChan := make(chan bool)
	refreshHotlapsTicker := time.NewTicker(60 * time.Minute)
	refreshServersTicker := time.NewTicker(10 * time.Second)

	// build the main app
	app = mainapp.NewMainApp(ctx, bot, domain, pubsubMgr, exitChan, refreshHotlapsTicker, refreshServersTicker)

	// Tell the user the bot is online
	log.Println("Start listening for updates. Press Ctrl-C to stop it")

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
	switch {
	// Handle messages
	case update.Message != nil:
		MessageHandler(ctx, update.Message)
	// Handle button clicks
	case update.CallbackQuery != nil:
		err := CallbackQueryHandler(ctx, update.CallbackQuery)
		if err != nil {
			log.Printf("An error occured: %s", err.Error())
		}
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
	if accept, handler := app.AcceptButton(button); accept {
		return handler(ctx, chatId)
	}
	return nil
}

// When we get a command, we react accordingly
func handleCommand(ctx context.Context, chatId int64, command string) error {
	if accept, handler := app.AcceptCommand(command); accept {
		return handler(ctx, chatId)
	}
	return nil
}

func CallbackQueryHandler(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	if accept, handler := app.AcceptCallback(query); accept {
		return handler(ctx, query)
	}
	return nil
}
