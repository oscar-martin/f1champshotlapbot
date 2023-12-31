package main

import (
	"context"
	"encoding/json"
	"f1champshotlapsbot/pkg/apps"
	"f1champshotlapsbot/pkg/apps/mainapp"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/oscar-martin/rfactor2telegrambot/pkg/apps/live"
	"github.com/oscar-martin/rfactor2telegrambot/pkg/notification"
	"github.com/oscar-martin/rfactor2telegrambot/pkg/servers"
	"github.com/oscar-martin/rfactor2telegrambot/pkg/settings"
	"github.com/oscar-martin/rfactor2telegrambot/pkg/webserver"
	"golang.org/x/text/language"

	_ "net/http/pprof"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	EnvTelegramToken = "TELEGRAM_TOKEN"
	EnvHotlapsDomain = "API_DOMAIN"
	// format: <server_id>,<server_url>;<server_id>,<server_url>;...
	// format example: "ServerID1,http://localhost:10001;ServerID2,http://localhost:10002;ServerID3,http://localhost:10003"
	EnvServers          = "RF2_SERVERS"
	EnvLiveMapDomain    = "LIVEMAP_DOMAIN"
	EnvWebServerAddress = "WEBSERVER_ADDRESS"
)

var (
	domain        = ""
	liveMapDomain = ""
	bot           *tgbotapi.BotAPI
	app           apps.Accepter
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	// flag.Parse()
	// if *cpuprofile != "" {
	// 	f, err := os.Create(*cpuprofile)
	// 	if err != nil {
	// 		log.Fatal("could not create CPU profile: ", err)
	// 	}
	// 	defer f.Close() // error handling omitted for example
	// 	if err := pprof.StartCPUProfile(f); err != nil {
	// 		log.Fatal("could not start CPU profile: ", err)
	// 	}
	// 	defer pprof.StopCPUProfile()
	// }

	// go func() {
	// 	_ = http.ListenAndServe("0.0.0.0:8081", nil)
	// }()

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

	liveMapDomain = os.Getenv(EnvLiveMapDomain)
	if liveMapDomain == "" {
		log.Fatalf("%s is not set", EnvLiveMapDomain)
	}
	liveMapDomain = strings.TrimRight(liveMapDomain, "/")

	rf2Servers := os.Getenv(EnvServers)
	if rf2Servers == "" {
		log.Fatalf("%s is not set", EnvServers)
	}

	var webServerAddr = ":8080"
	if os.Getenv(EnvWebServerAddress) != "" {
		webServerAddr = os.Getenv(EnvWebServerAddress)
	}

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

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustLoadMessageFile("active.es.json")
	loc := i18n.NewLocalizer(bundle, "es")

	exitChan := make(chan bool)
	refreshHotlapsTicker := time.NewTicker(60 * time.Minute)
	refreshServersTicker := time.NewTicker(10 * time.Second)

	settings, err := settings.NewManager()
	if err != nil {
		log.Fatalf("Error creating settings manager: %s", err.Error())
	}

	nm := notification.NewManager(ctx, bot, settings, loc)
	go nm.Start(exitChan)

	// build the main app
	ss, err := createServers(rf2Servers, liveMapDomain)
	if err != nil {
		log.Fatalf("Error creating servers: %s", err.Error())
	}
	ws := webserver.NewManager()
	sm, err := servers.NewManager(ctx, bot, ss, ws, loc)
	if err != nil {
		log.Fatalf("Error creating servers manager: %s", err.Error())
	}
	// ws.Debug()

	app, err = mainapp.NewMainApp(ctx, bot, domain, ss, exitChan, refreshHotlapsTicker, settings, loc)
	if err != nil {
		log.Fatalf("Error creating main app: %s", err.Error())
	}

	// start syncing once the apps are created
	go sm.Sync(refreshServersTicker, exitChan)
	go ws.Serve(webServerAddr)

	// Tell the user the bot is online
	log.Println("Start listening for updates. Press Ctrl-C to stop it")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// lock the main thread until we receive a signal
	<-sigs

	refreshHotlapsTicker.Stop()
	refreshServersTicker.Stop()
	exitChan <- true

	settings.Close()

	cancel()

	// if *memprofile != "" {
	// 	f, err := os.Create(*memprofile)
	// 	if err != nil {
	// 		log.Fatal("could not create memory profile: ", err)
	// 	}
	// 	defer f.Close() // error handling omitted for example
	// 	runtime.GC()    // get up-to-date statistics
	// 	if err := pprof.WriteHeapProfile(f); err != nil {
	// 		log.Fatal("could not write memory profile: ", err)
	// 	}
	// }
}

func createServers(rf2Servers, domain string) ([]servers.Server, error) {
	serversStr := strings.Split(rf2Servers, ";")
	ss := []servers.Server{}
	for _, serverStr := range serversStr {
		serverData := strings.Split(serverStr, ",")
		if len(serverData) != 2 {
			return nil, fmt.Errorf("Invalid server data: %s", serverStr)
		}
		server := servers.NewServer(serverData[0], serverData[1], domain)
		ss = append(ss, server)
	}
	return ss, nil
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
		user := update.Message.From
		if user == nil {
			return
		}
		ctx = context.WithValue(ctx, live.UserContextKey, user)
		ctx = context.WithValue(ctx, live.ChatContextKey, update.Message.Chat)
		MessageHandler(ctx, update.Message)
	// Handle button clicks
	case update.CallbackQuery != nil:
		user := update.CallbackQuery.From
		if user == nil {
			return
		}
		if update.CallbackQuery.Message == nil {
			return
		}
		ctx = context.WithValue(ctx, live.UserContextKey, user)
		ctx = context.WithValue(ctx, live.ChatContextKey, update.CallbackQuery.Message.Chat)
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
