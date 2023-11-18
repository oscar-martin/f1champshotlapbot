package apps

import (
	"context"
	"f1champshotlapsbot/pkg/menus"
	"f1champshotlapsbot/pkg/tracks"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	buttonTracks = "Circuitos"
	buttonActual = "Actual"
)

type HotlapsApp struct {
	bot          *tgbotapi.BotAPI
	apiDomain    string
	appMenu      menus.ApplicationMenu
	tm           *tracks.Manager
	menuKeyboard tgbotapi.ReplyKeyboardMarkup
}

func NewHotlapsApp(ctx context.Context, bot *tgbotapi.BotAPI, domain string, appMenu menus.ApplicationMenu, exitChan chan bool, refreshTicker *time.Ticker) *HotlapsApp {
	menuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(buttonTracks),
			tgbotapi.NewKeyboardButton(buttonActual),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(appMenu.ButtonBackTo()),
		),
	)

	tm := tracks.NewTrackManager(bot, domain)
	tm.Sync(ctx, refreshTicker, exitChan)

	return &HotlapsApp{
		apiDomain:    domain,
		bot:          bot,
		appMenu:      appMenu,
		tm:           tm,
		menuKeyboard: menuKeyboard,
	}
}

func (hl *HotlapsApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	commandTrackId := regexp.MustCompile(`^\/(\d+)$`)
	commandTrackSessionId := regexp.MustCompile(`^\/(\d+)_(.+)$`)
	if commandTrackId.MatchString(command) {
		// show categories for track id
		trackId, _ := strconv.Atoi(commandTrackId.FindStringSubmatch(command)[1])
		return true, hl.tm.RenderCategoriesForTrackId(trackId)
	} else if commandTrackSessionId.MatchString(command) {
		// show sessions for track
		trackId := commandTrackSessionId.FindStringSubmatch(command)[1]
		categoryId := commandTrackSessionId.FindStringSubmatch(command)[2]
		return true, hl.tm.RenderSessionForCategoryAndTrack(trackId, categoryId)
	}
	return false, nil
}

func (hl *HotlapsApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery) error) {
	data := strings.Split(query.Data, ":")
	if data[0] == tracks.SubcommandShowTracks {
		return true, hl.tm.RenderShowTracksCallback(data)
	} else if data[0] == tracks.SubcommandShowSessionData {
		return true, hl.tm.RenderSessionsCallback(data)
	}
	return false, nil
}

func (hl *HotlapsApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	// fmt.Printf("HOTLAP: button: %s. appName: %s\n", button, hl.appMenu.Name)
	if button == hl.appMenu.Name {
		return true, func(ctx context.Context, chatId int64) error {
			message := fmt.Sprintf("%s application\n\n", hl.appMenu.Name)
			msg := tgbotapi.NewMessage(chatId, message)
			msg.ReplyMarkup = hl.menuKeyboard
			_, err := hl.bot.Send(msg)
			return err
		}
	} else if button == hl.appMenu.ButtonBackTo() {
		return true, func(ctx context.Context, chatId int64) error {
			msg := tgbotapi.NewMessage(chatId, "OK")
			msg.ReplyMarkup = hl.appMenu.PrevMenu()
			_, err := hl.bot.Send(msg)
			return err
		}
	} else if button == buttonTracks {
		return true, hl.tm.RenderTracks()
	} else if button == buttonActual {
		return true, hl.tm.RenderCurrentSession()
	}
	// fmt.Print("HOTLAP: FALSE\n")
	return false, nil
}
