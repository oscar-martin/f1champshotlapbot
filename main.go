package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	menuTracks = "/tracks"
)

// struct to process the next json payload
//
//	{
//		"driver": "Sanchez",
//		"TrackCourse": "Autodromo do Interlagos",
//		"s1": 17.7163,
//		"s2": 38.2925,
//		"s3": 17.0742,
//		"time": 73.083,
//		"fuel": 0.627,
//		"fl": 0.89,
//		"fr": 0.871,
//		"rl": 0.941,
//		"rr": 0.941,
//		"fcompound": "0,Soft",
//		"rcompound": "0,Soft",
//		"DateTime": "2022-11-11 18:38:36",
//		"category": "F1Champs 2022,Ferrari",
//		"carType": "Ferrari",
//		"carClass": "Ferrari",
//		"team": "Scuderia Ferrari",
//		"lapcount": 51,
//		"lapcountcomplete": 16
//	},

var (
	trackMutex      sync.Mutex
	tracks          Tracks
	trackSessionsMu sync.Mutex
	trackSessions   = map[string]Sessions{}
	// Menu texts
	// firstMenu  = "<b>Menu 1</b>\n\nA beautiful menu with a shiny inline button."
	// secondMenu = "<b>Menu 2</b>\n\nA better menu with even more shiny inline buttons."

	// Button texts
	// nextButton     = "Next"
	// backButton     = "Back"
	// tutorialButton = "Tutorial"

	// Store bot screaming status
	// screaming = false
	bot *tgbotapi.BotAPI

	// // Keyboard layout for the first menu. One button, one row
	// firstMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
	// 	tgbotapi.NewInlineKeyboardRow(
	// 		tgbotapi.NewInlineKeyboardButtonData(nextButton, nextButton),
	// 	),
	// )

	// // Keyboard layout for the second menu. Two buttons, one per row
	// secondMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
	// 	tgbotapi.NewInlineKeyboardRow(
	// 		tgbotapi.NewInlineKeyboardButtonData(backButton, backButton),
	// 	),
	// 	tgbotapi.NewInlineKeyboardRow(
	// 		tgbotapi.NewInlineKeyboardButtonURL(tutorialButton, "https://core.telegram.org/bots/api"),
	// 	),
	// )
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
	log.Println("Start listening for updates. Press enter to stop")

	// Wait for a newline symbol, then cancel handling updates
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
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
		handleMessage(ctx, update.Message)
	// Handle button clicks
	case update.CallbackQuery != nil:
		handleButton(ctx, update.CallbackQuery)
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

	// if strings.HasPrefix(text, "/") {
	// 	err = handleCommand(message.Chat.ID, text)
	// } else if screaming && len(text) > 0 {
	// 	msg := tgbotapi.NewMessage(message.Chat.ID, strings.ToUpper(text))
	// 	// To preserve markdown, we attach entities (bold, italic..)
	// 	msg.Entities = message.Entities
	// 	_, err = bot.Send(msg)
	// } else {
	// 	// This is equivalent to forwarding, without the sender's name
	// 	copyMsg := tgbotapi.NewCopyMessage(message.Chat.ID, message.Chat.ID, message.MessageID)
	// 	_, err = bot.CopyMessage(copyMsg)
	// }

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

func handleButton(ctx context.Context, query *tgbotapi.CallbackQuery) {
	var text string

	markup := tgbotapi.NewInlineKeyboardMarkup()
	message := query.Message

	// if query.Data == nextButton {
	// 	text = secondMenu
	// 	markup = secondMenuMarkup
	// } else if query.Data == backButton {
	// 	text = firstMenu
	// 	markup = firstMenuMarkup
	// }

	callbackCfg := tgbotapi.NewCallback(query.ID, "")
	bot.Send(callbackCfg)

	// Replace menu text and keyboard
	msg := tgbotapi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, text, markup)
	msg.ParseMode = tgbotapi.ModeHTML
	bot.Send(msg)
}

// When we get a command, we react accordingly
func handleCommand(ctx context.Context, chatId int64, command string) error {
	var err error

	commandTrackId := regexp.MustCompile(`^\/(\d+)$`)
	commandTrackSessionId := regexp.MustCompile(`^\/(\d+)_(.+)$`)
	switch {

	// Fetch all tracks
	case command == menuTracks:
		trackMutex.Lock()
		defer trackMutex.Unlock()
		tracks, err = getTracks(ctx)
		if err != nil {
			return err
		}

		var message string
		if len(tracks) > 0 {
			tracksStrings := make([]string, len(tracks))
			for i, track := range tracks {
				tracksStrings[i] = track.Name + " --> " + track.Command
			}
			message = strings.Join(tracksStrings, "\n")
		} else {
			message = "No hay circuitos disponibles"
		}
		msg := tgbotapi.NewMessage(chatId, message)
		_, err = bot.Send(msg)
		if err != nil {
			return err
		}

	// Fetch all sessions for a track
	case commandTrackId.MatchString(command):
		trackId, _ := strconv.Atoi(commandTrackId.FindStringSubmatch(command)[1])
		track, found := tracks.GetTrackByID(fmt.Sprint(trackId))
		if !found {
			message := "No hay circuitos disponibles"
			msg := tgbotapi.NewMessage(chatId, message)
			_, err = bot.Send(msg)
			if err != nil {
				return err
			}
		}
		return processCurrentTrackTimes(ctx, chatId, track)

	case commandTrackSessionId.MatchString(command):
		trackId, _ := strconv.Atoi(commandTrackSessionId.FindStringSubmatch(command)[1])
		categoryId := commandTrackSessionId.FindStringSubmatch(command)[2]

		track, found := tracks.GetTrackByID(fmt.Sprint(trackId))
		if !found {
			message := "No hay sesiones disponibles"
			msg := tgbotapi.NewMessage(chatId, message)
			_, err = bot.Send(msg)
			if err != nil {
				return err
			}
			return nil
		}
		if sessions, ok := trackSessions[track.ID]; ok {
			sessionsForCategory := sessions.GetSessionsByCategoryID(categoryId)

			if len(sessionsForCategory) > 0 {
				sort.Slice(sessionsForCategory, func(i, j int) bool {
					return sessionsForCategory[i].Time < sessionsForCategory[j].Time
				})

				var b bytes.Buffer
				t := table.NewWriter()
				t.SetOutputMirror(&b)
				t.SetStyle(table.StyleRounded)
				t.AppendSeparator()
				// t.AppendHeader(table.Row{"Dri", "Times", "S1", "S2", "S3"})
				t.AppendHeader(table.Row{"Dri", "Times", "Sectors"})
				// var message string
				for _, session := range sessionsForCategory {
					t.AppendRow([]interface{}{
						getDriverCodeName(session.Driver),
						secondsToMinutes(session.Time),
						fmt.Sprintf("%s %s %s", toSectorTime(session.S1), toSectorTime(session.S2), toSectorTime(session.S3)),
						// toSectorTime(session.S1),
						// toSectorTime(session.S2),
						// toSectorTime(session.S3),
					})
				}
				t.Render()

				msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("```%s```", b.String()))
				msg.ParseMode = tgbotapi.ModeMarkdownV2

				// msg := tgbotapi.NewMessage(chatId, message)
				_, err = bot.Send(msg)
				if err != nil {
					return err
				}

			} else {
				message := "No hay sesiones disponibles"
				msg := tgbotapi.NewMessage(chatId, message)
				_, err = bot.Send(msg)
				if err != nil {
					return err
				}
			}
		} else {
			message := "No hay sesiones disponibles"
			msg := tgbotapi.NewMessage(chatId, message)
			_, err = bot.Send(msg)
			if err != nil {
				return err
			}
		}

	// case command == menuCurrentTrack:
	// 	if currentTrack == "" {
	// 		message := "No hay circuitos disponibles"
	// 		msg := tgbotapi.NewMessage(chatId, message)
	// 		_, err = bot.Send(msg)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}

	// 	// find track index in tracks
	// 	for i, track := range tracks {
	// 		if track == currentTrack {
	// 			return processCurrentTrackTimes(chatId, i, currentTrack)
	// 		}
	// 	}
	// 	message := "No hay circuitos disponibles"
	// 	msg := tgbotapi.NewMessage(chatId, message)
	// 	_, err = bot.Send(msg)
	// 	if err != nil {
	// 		return err
	// 	}

	case command == "/whisper":
		// var b bytes.Buffer
		// t := table.NewWriter()
		// t.SetOutputMirror(&b)
		// t.AppendHeader(table.Row{"#", "First Name", "Last Name", "Salary"})
		// t.AppendRows([]table.Row{
		// 	{1, "Arya", "Stark", 3000},
		// 	{20, "Jon", "Snow", 2000, "You know nothing, Jon Snow!"},
		// })
		// t.AppendRow([]interface{}{300, "Tyrion", "Lannister", 5000})
		// t.AppendFooter(table.Row{"", "", "Total", 10000})
		// t.RenderMarkdown()
		// message := "```\n" +
		// 	"| one   | two |" + "\n" +
		// 	"| ----- | --- |" + "\n" +
		// 	"| two   |   2 |" + "\n" +
		// 	"| three |   3 |" + "\n" +
		// 	"```"
		// msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("```%s```", b.String()))
		// msg.ParseMode = tgbotapi.ModeMarkdownV2
		// _, err = bot.Send(msg)
		// if err != nil {
		// 	return err
		// }

	case command == "/menu":
		// err = sendMenu(chatId)
	}

	return err
}

func processCurrentTrackTimes(ctx context.Context, chatId int64, track Track) error {
	trackSessionsMu.Lock()
	defer trackSessionsMu.Unlock()
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

	var message string
	if len(cats) > 0 {
		categoriesStrings := make([]string, len(cats))
		for i, cat := range cats {
			categoriesStrings[i] = cat.Name + fmt.Sprintf(" ---> /%s_%s", track.ID, cat.ID)
		}

		message = strings.Join(categoriesStrings, "\n")
	} else {
		message = "No hay sesiones disponibles"
	}
	msg := tgbotapi.NewMessage(chatId, message)
	_, err := bot.Send(msg)

	return err
}

// func sendMenu(chatId int64) error {
// 	msg := tgbotapi.NewMessage(chatId, firstMenu)
// 	msg.ParseMode = tgbotapi.ModeHTML
// 	msg.ReplyMarkup = firstMenuMarkup
// 	_, err := bot.Send(msg)
// 	return err
// }
