package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	menuTracks             = "/circuitos"
	inlineKeyboardTimes    = "Tiempos"
	inlineKeyboardSectors  = "Sectores"
	inlineKeyboardCompound = "Gomas"
	inlineKeyboardLaps     = "Vueltas"
	inlineKeyboardTeam     = "Coches"
	inlineKeyboardDriver   = "Pilotos"
	inlineKeyboardDate     = "Fecha"

	symbolTimes    = "⏱"
	symbolSectors  = "🔂"
	symbolCompound = "🛞"
	symbolLaps     = "🏁"
	symbolTeam     = "🏎️"
	symbolDriver   = "👐"
	symbolDate     = "⌚️"

	tableDriver = "PIL"
)

var (
	trackMutex      sync.Mutex
	tracks          Tracks
	trackSessionsMu sync.Mutex
	trackSessions   = map[string]Sessions{}

	bot *tgbotapi.BotAPI

	tracksPerPage = 10
)

func main() {
	var err error
	// get token from env
	token := os.Getenv("TELEGRAM_TOKEN")
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
	// fmt.Printf("%+v\n", update.Message)
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
		// fmt.Printf("%+v\n", update.CallbackQuery)
		CallbackQueryHandler(update.CallbackQuery)
	}
}

func CallbackQueryHandler(query *tgbotapi.CallbackQuery) {
	split := strings.Split(query.Data, ":")
	if split[0] == "pager" {
		maxPages := len(tracks) / tracksPerPage
		HandleNavigationCallbackQuery(query.Message.Chat.ID, query.Message.MessageID, maxPages, tracks, split[1:]...)
		return
	} else if split[0] == inlineKeyboardTimes ||
		split[0] == inlineKeyboardSectors ||
		split[0] == inlineKeyboardCompound ||
		split[0] == inlineKeyboardLaps ||
		split[0] == inlineKeyboardTeam ||
		split[0] == inlineKeyboardDriver ||
		split[0] == inlineKeyboardDate {
		trackId := split[1]
		categoryId := split[2]

		track, found := tracks.GetTrackByID(fmt.Sprint(trackId))
		if !found {
			message := fmt.Sprintf("No hay sesiones disponibles. Vuelve a probar %s", menuTracks)
			msg := tgbotapi.NewMessage(query.Message.Chat.ID, message)
			_, err := bot.Send(msg)
			if err != nil {
				log.Println(err)
				return
			}
			return
		}
		if sessions, ok := trackSessions[track.ID]; ok {
			sessionsForCategory := sessions.GetSessionsByCategoryID(categoryId)

			if len(sessionsForCategory) > 0 {
				sort.Slice(sessionsForCategory, func(i, j int) bool {
					return sessionsForCategory[i].Time < sessionsForCategory[j].Time
				})

				// read the category name from the first session
				_, category := extractCategory(sessionsForCategory[0].Category)

				var b bytes.Buffer
				t := table.NewWriter()
				t.SetOutputMirror(&b)
				t.SetStyle(table.StyleRounded)
				t.AppendSeparator()

				t.AppendHeader(table.Row{tableDriver, split[0]})
				for _, session := range sessionsForCategory {
					if split[0] == inlineKeyboardTimes {
						t.AppendRow([]interface{}{
							getDriverCodeName(session.Driver),
							secondsToMinutes(session.Time),
						})
					} else if split[0] == inlineKeyboardSectors {
						t.AppendRow([]interface{}{
							getDriverCodeName(session.Driver),
							fmt.Sprintf("%s %s %s", toSectorTime(session.S1), toSectorTime(session.S2), toSectorTime(session.S3)),
						})
					} else if split[0] == inlineKeyboardCompound {
						tyreSlice := strings.Split(session.Fcompound, ",")
						tyre := "(desconocido)"
						if len(tyreSlice) > 0 {
							tyre = tyreSlice[len(tyreSlice)-1]
						}
						t.AppendRow([]interface{}{
							getDriverCodeName(session.Driver),
							tyre,
						})
					} else if split[0] == inlineKeyboardLaps {
						t.AppendRow([]interface{}{
							getDriverCodeName(session.Driver),
							fmt.Sprintf("%d/%d", session.Lapcountcomplete, session.Lapcount),
						})
					} else if split[0] == inlineKeyboardTeam {
						t.AppendRow([]interface{}{
							getDriverCodeName(session.Driver),
							session.CarClass,
						})
					} else if split[0] == inlineKeyboardDriver {
						t.AppendRow([]interface{}{
							getDriverCodeName(session.Driver),
							session.Driver,
						})
					} else if split[0] == inlineKeyboardDate {
						t.AppendRow([]interface{}{
							getDriverCodeName(session.Driver),
							session.DateTime,
						})
					}
				}
				t.Render()

				keyboard := getInlineKeyboardForCategory(track.ID, categoryId)

				var cfg tgbotapi.Chattable
				msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, fmt.Sprintf("```\nResultados en %q para %q\n\n%s```", track.Name, category, b.String()))
				msg.ParseMode = tgbotapi.ModeMarkdownV2
				msg.ReplyMarkup = &keyboard
				cfg = msg
				_, err := bot.Send(cfg)
				if err != nil {
					return
				}
			}
		}
	}
}

func getInlineKeyboardForCategory(trackId, categoryId string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardTimes+" "+symbolTimes, fmt.Sprintf("%s:%s:%s", inlineKeyboardTimes, trackId, categoryId)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardSectors+" "+symbolSectors, fmt.Sprintf("%s:%s:%s", inlineKeyboardSectors, trackId, categoryId)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardCompound+" "+symbolTimes, fmt.Sprintf("%s:%s:%s", inlineKeyboardCompound, trackId, categoryId)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardLaps+" "+symbolLaps, fmt.Sprintf("%s:%s:%s", inlineKeyboardLaps, trackId, categoryId)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardTeam+" "+symbolTeam, fmt.Sprintf("%s:%s:%s", inlineKeyboardTeam, trackId, categoryId)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardDriver+" "+symbolDriver, fmt.Sprintf("%s:%s:%s", inlineKeyboardDriver, trackId, categoryId)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardDate+" "+symbolDate, fmt.Sprintf("%s:%s:%s", inlineKeyboardDate, trackId, categoryId)),
		),
	)
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
			message := fmt.Sprintf("No hay circuitos disponibles. Vuelve a probar %s", menuTracks)
			msg := tgbotapi.NewMessage(chatId, message)
			_, err = bot.Send(msg)
			return err
		}
		return processCurrentTrackTimes(ctx, chatId, track)

	// Fetch all sessions for a track and a category
	case commandTrackSessionId.MatchString(command):
		trackId, _ := strconv.Atoi(commandTrackSessionId.FindStringSubmatch(command)[1])
		categoryId := commandTrackSessionId.FindStringSubmatch(command)[2]
		track, found := tracks.GetTrackByID(fmt.Sprint(trackId))
		if !found {
			message := fmt.Sprintf("No hay sesiones disponibles. Vuelve a probar %s", menuTracks)
			msg := tgbotapi.NewMessage(chatId, message)
			_, err = bot.Send(msg)
			return err
		}
		if sessions, ok := trackSessions[track.ID]; ok {
			sessionsForCategory := sessions.GetSessionsByCategoryID(categoryId)

			if len(sessionsForCategory) > 0 {
				sort.Slice(sessionsForCategory, func(i, j int) bool {
					return sessionsForCategory[i].Time < sessionsForCategory[j].Time
				})
				// read the category name from the first session
				_, category := extractCategory(sessionsForCategory[0].Category)

				var b bytes.Buffer
				t := table.NewWriter()
				t.SetOutputMirror(&b)
				t.SetStyle(table.StyleRounded)
				t.AppendSeparator()
				t.AppendHeader(table.Row{tableDriver, inlineKeyboardTimes})
				for _, session := range sessionsForCategory {
					t.AppendRow([]interface{}{
						getDriverCodeName(session.Driver),
						secondsToMinutes(session.Time),
					})
				}
				t.Render()

				keyboard := getInlineKeyboardForCategory(track.ID, categoryId)
				var cfg tgbotapi.Chattable
				msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("```\nResultados en %q para %q\n\n%s```", track.Name, category, b.String()))
				msg.ParseMode = tgbotapi.ModeMarkdownV2
				msg.ReplyMarkup = keyboard
				cfg = msg
				_, err = bot.Send(cfg)
				if err != nil {
					return err
				}

			} else {
				message := fmt.Sprintf("No hay sesiones disponibles. Vuelve a probar %s", menuTracks)
				msg := tgbotapi.NewMessage(chatId, message)
				_, err = bot.Send(msg)
				return err
			}
		} else {
			message := fmt.Sprintf("No hay sesiones disponibles. Vuelve a probar %s", menuTracks)
			msg := tgbotapi.NewMessage(chatId, message)
			_, err = bot.Send(msg)
			return err
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
			categoriesStrings[i] = " ▸ " + cat.Name + fmt.Sprintf(" ➡ /%s_%s", track.ID, cat.ID)
		}

		message += strings.Join(categoriesStrings, "\n")
	} else {
		message = "No hay sesiones disponibles"
	}
	msg := tgbotapi.NewMessage(chatId, message)
	_, err := bot.Send(msg)

	return err
}
