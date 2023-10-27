package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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
	menuTracks       = "/tracks"
	menuCurrentTrack = "/current_track"
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
type Session struct {
	Driver           string  `json:"driver"`
	TrackCourse      string  `json:"TrackCourse"`
	S1               float64 `json:"s1"`
	S2               float64 `json:"s2"`
	S3               float64 `json:"s3"`
	Time             float64 `json:"time"`
	Fuel             float64 `json:"fuel"`
	Fl               float64 `json:"fl"`
	Fr               float64 `json:"fr"`
	Rl               float64 `json:"rl"`
	Rr               float64 `json:"rr"`
	Fcompound        string  `json:"fcompound"`
	Rcompound        string  `json:"rcompound"`
	DateTime         string  `json:"DateTime"`
	Category         string  `json:"category"`
	CarType          string  `json:"carType"`
	CarClass         string  `json:"carClass"`
	Team             string  `json:"team"`
	Lapcount         int     `json:"lapcount"`
	Lapcountcomplete int     `json:"lapcountcomplete"`
}

var (
	trackMutex      sync.Mutex
	tracks          = []string{}
	trackSessionsMu sync.Mutex
	trackSessions   = map[string][]Session{}
	currentTrack    = ""
	// Menu texts
	firstMenu  = "<b>Menu 1</b>\n\nA beautiful menu with a shiny inline button."
	secondMenu = "<b>Menu 2</b>\n\nA better menu with even more shiny inline buttons."

	// Button texts
	nextButton     = "Next"
	backButton     = "Back"
	tutorialButton = "Tutorial"

	// Store bot screaming status
	screaming = false
	bot       *tgbotapi.BotAPI

	// Keyboard layout for the first menu. One button, one row
	firstMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(nextButton, nextButton),
		),
	)

	// Keyboard layout for the second menu. Two buttons, one per row
	secondMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(backButton, backButton),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(tutorialButton, "https://core.telegram.org/bots/api"),
		),
	)
)

func main() {
	var err error
	bot, err = tgbotapi.NewBotAPI("6361517001:AAGMS30BNTmLkBoc9OzhygaIl4yhuduXMzM")
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
	bufio.NewReader(os.Stdin).ReadBytes('\n')
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
			handleUpdate(update)
		}
	}
}

func handleUpdate(update tgbotapi.Update) {
	switch {
	// Handle messages
	case update.Message != nil:
		handleMessage(update.Message)
		break

	// Handle button clicks
	case update.CallbackQuery != nil:
		handleButton(update.CallbackQuery)
		break
	}
}

func handleMessage(message *tgbotapi.Message) {
	user := message.From
	text := message.Text

	if user == nil {
		return
	}

	// Print to console
	log.Printf("%s wrote %s", user.FirstName, text)

	var err error
	if strings.HasPrefix(text, "/") {
		err = handleCommand(message.Chat.ID, text)
	} else if screaming && len(text) > 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, strings.ToUpper(text))
		// To preserve markdown, we attach entities (bold, italic..)
		msg.Entities = message.Entities
		_, err = bot.Send(msg)
	} else {
		// This is equivalent to forwarding, without the sender's name
		copyMsg := tgbotapi.NewCopyMessage(message.Chat.ID, message.Chat.ID, message.MessageID)
		_, err = bot.CopyMessage(copyMsg)
	}

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

// When we get a command, we react accordingly
func handleCommand(chatId int64, command string) error {
	var err error

	commandTrackId := regexp.MustCompile(`^\/(\d+)$`)
	commandTrackSessionId := regexp.MustCompile(`^\/(\d+)_(.+)$`)
	switch {
	case command == menuTracks:
		tracks, err = getTracks()
		if err != nil {
			return err
		}

		var message string
		if len(tracks) > 0 {
			currentTrack = tracks[0]
			// sort tracks by name
			sort.Slice(tracks, func(i, j int) bool {
				return tracks[i] < tracks[j]
			})

			// iterate over the tracks to append a suffix
			tracksStrings := make([]string, len(tracks))
			for i, track := range tracks {
				tracksStrings[i] = track + " /" + fmt.Sprintf("%d", i)
			}

			message = strings.Join(append(tracksStrings, fmt.Sprintf("%s ---> %s", currentTrack, menuCurrentTrack)), "\n")
		} else {
			message = "No hay circuitos disponibles"
		}
		msg := tgbotapi.NewMessage(chatId, message)
		_, err = bot.Send(msg)
		if err != nil {
			return err
		}

	case commandTrackId.MatchString(command):
		trackId, _ := strconv.Atoi(commandTrackId.FindStringSubmatch(command)[1])
		if trackId >= len(tracks) {
			message := "No hay circuitos disponibles"
			msg := tgbotapi.NewMessage(chatId, message)
			_, err = bot.Send(msg)
			if err != nil {
				return err
			}
		}
		return processCurrentTrackTimes(chatId, trackId, tracks[trackId])

	case commandTrackSessionId.MatchString(command):
		trackId, _ := strconv.Atoi(commandTrackSessionId.FindStringSubmatch(command)[1])
		category := commandTrackSessionId.FindStringSubmatch(command)[2]

		if trackId >= len(tracks) {
			message := "No hay circuitos disponibles"
			msg := tgbotapi.NewMessage(chatId, message)
			_, err = bot.Send(msg)
			if err != nil {
				return err
			}
		}

		track := tracks[trackId]
		if sessions, ok := trackSessions[track]; ok {
			sessionsForCategory := []Session{}
			for _, session := range sessions {
				if extractCategory(session.Category) == category {
					sessionsForCategory = append(sessionsForCategory, session)
				}
			}

			if len(sessionsForCategory) > 0 {
				sort.Slice(sessionsForCategory, func(i, j int) bool {
					return sessionsForCategory[i].Time < sessionsForCategory[j].Time
				})

				var b bytes.Buffer
				t := table.NewWriter()
				t.SetOutputMirror(&b)
				t.AppendSeparator()
				t.AppendHeader(table.Row{"Driver", "Times", "S1", "S2", "S3"})
				// var message string
				for _, session := range sessionsForCategory {
					t.AppendRow([]interface{}{session.Driver, secondsToMinutes(session.Time), session.S1, session.S2, session.S3})
					// message += fmt.Sprintf("%s %s (%.3f, %.3f, %.3f)\n", session.Driver, secondsToMinutes(session.Time), session.S1, session.S2, session.S3)
				}

				// t.AppendRows([]table.Row{
				// 	{1, "Arya", "Stark", 3000},
				// 	{20, "Jon", "Snow", 2000, "You know nothing, Jon Snow!"},
				// })
				// t.AppendFooter(table.Row{"", "", "Total", 10000})
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

	case command == menuCurrentTrack:
		if currentTrack == "" {
			message := "No hay circuitos disponibles"
			msg := tgbotapi.NewMessage(chatId, message)
			_, err = bot.Send(msg)
			if err != nil {
				return err
			}
		}

		// find track index in tracks
		for i, track := range tracks {
			if track == currentTrack {
				return processCurrentTrackTimes(chatId, i, currentTrack)
			}
		}
		message := "No hay circuitos disponibles"
		msg := tgbotapi.NewMessage(chatId, message)
		_, err = bot.Send(msg)
		if err != nil {
			return err
		}

	case command == "/whisper":
		var b bytes.Buffer
		t := table.NewWriter()
		t.SetOutputMirror(&b)
		t.AppendHeader(table.Row{"#", "First Name", "Last Name", "Salary"})
		t.AppendRows([]table.Row{
			{1, "Arya", "Stark", 3000},
			{20, "Jon", "Snow", 2000, "You know nothing, Jon Snow!"},
		})
		t.AppendRow([]interface{}{300, "Tyrion", "Lannister", 5000})
		t.AppendFooter(table.Row{"", "", "Total", 10000})
		t.RenderMarkdown()
		// message := "```\n" +
		// 	"| one   | two |" + "\n" +
		// 	"| ----- | --- |" + "\n" +
		// 	"| two   |   2 |" + "\n" +
		// 	"| three |   3 |" + "\n" +
		// 	"```"

		msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("```%s```", b.String()))
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		_, err = bot.Send(msg)
		if err != nil {
			return err
		}

	case command == "/menu":
		err = sendMenu(chatId)
	}

	return err
}

// method to convert from seconds to minutes:seconds:milliseconds
func secondsToMinutes(seconds float64) string {
	minutes := int(seconds / 60)
	seconds = seconds - float64(minutes*60)
	milliseconds := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d.%03d", minutes, int(seconds), milliseconds)
}

func extractCategory(category string) string {
	if len(category) > 0 {
		c := strings.ToLower(strings.Split(category, ",")[0])
		return strings.ReplaceAll(c, " ", "_")
	}
	return ""
}

func processCurrentTrackTimes(chatId int64, trackId int, track string) error {
	trackSessionsMu.Lock()
	defer trackSessionsMu.Unlock()
	var sessions []Session
	if trackSessions[track] == nil {
		var err error
		sessions, err = getTrackSessions(track)
		if err != nil {
			return err
		}
		trackSessions[track] = sessions
	} else {
		sessions = trackSessions[track]
	}

	categories := map[string]string{}
	for _, session := range sessions {
		if _, exits := categories[session.Category]; !exits {
			c := extractCategory(session.Category)
			if c != "" {
				categories[c] = ""
			}
		}
	}

	cats := make([]string, 0, len(categories))

	for k := range categories {
		cats = append(cats, k)
	}
	sort.Strings(cats)

	var message string
	if len(cats) > 0 {
		categoriesStrings := make([]string, len(cats))
		for i, cat := range cats {
			categoriesStrings[i] = cat + fmt.Sprintf(" ---> /%d_%s", trackId, cat)
		}

		message = strings.Join(categoriesStrings, "\n")
	} else {
		message = "No hay sesiones disponibles"
	}
	msg := tgbotapi.NewMessage(chatId, message)
	_, err := bot.Send(msg)

	return err
}

func getTrackSessions(track string) ([]Session, error) {
	// Make a get request
	resp, err := http.Get("https://api.f1champs.es/v3/laps?track=" + url.QueryEscape(track))
	if err != nil {
		return nil, err
	}

	// Close the response body on function return
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response body into a TrackResponse struct
	var trackSessions []Session
	err = json.Unmarshal(body, &trackSessions)
	if err != nil {
		return nil, err
	}

	return trackSessions, nil
}

// getTracks gets all tracks from URL https://api.f1champs.es/v3/laps?tracklist=tracklist
// using a http.get call
func getTracks() ([]string, error) {
	trackMutex.Lock()
	defer trackMutex.Unlock()
	// Make a get request
	resp, err := http.Get("https://api.f1champs.es/v3/laps?tracklist=tracklist")
	if err != nil {
		return nil, err
	}

	// Close the response body on function return
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response body into a TrackResponse struct
	var tracks []string
	err = json.Unmarshal(body, &tracks)
	if err != nil {
		return nil, err
	}

	return tracks, nil
}

func handleButton(query *tgbotapi.CallbackQuery) {
	var text string

	markup := tgbotapi.NewInlineKeyboardMarkup()
	message := query.Message

	if query.Data == nextButton {
		text = secondMenu
		markup = secondMenuMarkup
	} else if query.Data == backButton {
		text = firstMenu
		markup = firstMenuMarkup
	}

	callbackCfg := tgbotapi.NewCallback(query.ID, "")
	bot.Send(callbackCfg)

	// Replace menu text and keyboard
	msg := tgbotapi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, text, markup)
	msg.ParseMode = tgbotapi.ModeHTML
	bot.Send(msg)
}

func sendMenu(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, firstMenu)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = firstMenuMarkup
	_, err := bot.Send(msg)
	return err
}
