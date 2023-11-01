package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	menuStart              = "/start"
	menuTracks             = "/circuitos"
	inlineKeyboardTimes    = "Tiempos"
	inlineKeyboardSectors  = "Sectores"
	inlineKeyboardCompound = "Gomas"
	inlineKeyboardLaps     = "Vueltas"
	inlineKeyboardTeam     = "Coches"
	inlineKeyboardDriver   = "Pilotos"
	inlineKeyboardDate     = "Fecha"

	EnvTelegramToken = "TELEGRAM_TOKEN"
	EnvHotlapsDomain = "API_DOMAIN"

	symbolTimes    = "⏱"
	symbolSectors  = "🔂"
	symbolCompound = "🛞"
	symbolLaps     = "🏁"
	symbolTeam     = "🏎️"
	symbolDriver   = "👐"
	symbolDate     = "⌚️"

	symbolInit = "⏮"
	symbolPrev = "◀️"
	symbolNext = "▶️"
	symbolEnd  = "⏭"

	tableDriver = "PIL"

	subcommandShowTracks      = "show_tracks"
	subcommandShowSessionData = "show_session_data"
)

var (
	trackMutex      sync.Mutex
	tracks          Tracks
	trackSessionsMu sync.Mutex
	trackSessions   = map[string]Sessions{}
	domain          = ""

	bot *tgbotapi.BotAPI

	tracksPerPage = 10
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

	// Tell the user the bot is online
	log.Println("Start listening for updates. Press Ctrl-C to stop it")

	ticker := time.NewTicker(60 * time.Minute)
	tickerDone := make(chan bool)

	go func() {
		for {
			select {
			case <-tickerDone:
				return
			case t := <-ticker.C:
				fmt.Println("Resetting tracks and sessions at: ", t)
				trackMutex.Lock()
				tracks = Tracks{}
				trackMutex.Unlock()
				trackSessionsMu.Lock()
				trackSessions = map[string]Sessions{}
				trackSessionsMu.Unlock()
			}
		}
	}()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// lock the main thread until we receive a signal
	<-sigs

	ticker.Stop()
	tickerDone <- true

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
	trackMutex.Lock()
	defer trackMutex.Unlock()
	trackSessionsMu.Lock()
	defer trackSessionsMu.Unlock()
	switch {
	// Handle messages
	case update.Message != nil:
		handleMessage(ctx, update.Message)
	// Handle button clicks
	case update.CallbackQuery != nil:
		CallbackQueryHandler(update.CallbackQuery)
	}
}

func CallbackQueryHandler(query *tgbotapi.CallbackQuery) {
	split := strings.Split(query.Data, ":")
	if split[0] == subcommandShowTracks {
		maxPages := len(tracks) / tracksPerPage
		HandleTrackDataCallbackQuery(query.Message.Chat.ID, query.Message.MessageID, maxPages, tracks, split[1:]...)
		return
	} else if split[0] == subcommandShowSessionData {
		HandleSessionDataCallbackQuery(query.Message.Chat.ID, &query.Message.MessageID, split[1:]...)
	}
}

func handleMessage(ctx context.Context, message *tgbotapi.Message) {
	user := message.From
	text := message.Text

	if user == nil {
		return
	}

	// Print to console
	log.Printf("%s wrote %s", user.FirstName, text)

	var err error
	if message.IsCommand() {
		err = handleCommand(ctx, message.Chat.ID, text)
	}

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

// When we get a command, we react accordingly
func handleCommand(ctx context.Context, chatId int64, command string) error {
	var err error

	commandTrackId := regexp.MustCompile(`^\/(\d+)$`)
	commandTrackSessionId := regexp.MustCompile(`^\/(\d+)_(.+)$`)
	switch {
	case command == menuStart:
		message := "Hola, soy el bot de F1Champs que permite ver las Hotlaps registradas. Puedes usar los siguientes comandos:\n\n"
		message += menuTracks + " - Muestra la lista de circuitos"
		msg := tgbotapi.NewMessage(chatId, message)
		_, err = bot.Send(msg)
		return err

	// Fetch all tracks
	case command == menuTracks:
		if len(tracks) == 0 {
			// if there is no tracks, fetch them
			tracks, err = getTracks(ctx)
			if err != nil {
				return err
			}
		}

		if len(tracks) > 0 {
			err := SendTracksData(chatId, 0, tracksPerPage, len(tracks)/tracksPerPage, nil, tracks)
			if err != nil {
				return err
			}
		} else {
			message := "No hay circuitos disponibles"
			msg := tgbotapi.NewMessage(chatId, message)
			_, err = bot.Send(msg)
			return err
		}

	// Fetch all sessions for a track
	case commandTrackId.MatchString(command):
		trackId, _ := strconv.Atoi(commandTrackId.FindStringSubmatch(command)[1])
		track, found := tracks.GetTrackByID(fmt.Sprint(trackId))
		if !found {
			message := fmt.Sprintf("El circuito seleccionado no se ha encontrado. Vuelve a listarlos con %s", menuTracks)
			msg := tgbotapi.NewMessage(chatId, message)
			_, err = bot.Send(msg)
			return err
		}
		return processCurrentTrackTimes(ctx, chatId, track)

	// Fetch all sessions for a track and a category
	case commandTrackSessionId.MatchString(command):
		trackId := commandTrackSessionId.FindStringSubmatch(command)[1]
		categoryId := commandTrackSessionId.FindStringSubmatch(command)[2]
		err := SendSessionData(chatId, nil, trackId, categoryId, inlineKeyboardTimes)
		if err != nil {
			log.Printf("An error occured: %s", err.Error())
		}
	}

	return err
}

func processCurrentTrackTimes(ctx context.Context, chatId int64, track Track) error {
	var sessions Sessions
	if trackSessions[track.ID] == nil {
		var err error
		sessions, err = GetSessions(ctx, track.Name)
		if err != nil {
			return err
		}
		trackSessions[track.ID] = sessions
	} else {
		sessions = trackSessions[track.ID]
	}

	cats := sessions.GetCategories()

	message := fmt.Sprintf("Elige categoría para %s:\n\n", track.Name)
	if len(cats) > 0 {
		categoriesStrings := make([]string, len(cats))
		for i, cat := range cats {
			categoriesStrings[i] = cat.CommandString(track.ID)
		}

		message += strings.Join(categoriesStrings, "\n")
	} else {
		message = "No hay categorías para este circuito"
	}
	msg := tgbotapi.NewMessage(chatId, message)
	_, err := bot.Send(msg)

	return err
}
