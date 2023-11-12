package apps

import (
	"context"
	"f1champshotlapsbot/pkg/menus"
	"f1champshotlapsbot/pkg/servers"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	liveAppName         = "Live"
	serverPrefixCommand = "Server"
	buttonServer1       = "Server1"
	buttonServer2       = "Server2"
	buttonServer3       = "Server3"
	buttonServer4       = "Server4"
)

type LiveApp struct {
	bot          *tgbotapi.BotAPI
	apiDomain    string
	appMenu      menus.ApplicationMenu
	sm           *servers.Manager
	menuKeyboard tgbotapi.ReplyKeyboardMarkup
	serverApp    *ServerApp
}

func NewLiveApp(ctx context.Context, bot *tgbotapi.BotAPI, domain string, appMenu menus.ApplicationMenu, exitChan chan bool, refreshTicker *time.Ticker) *LiveApp {
	menuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(buttonServer1),
			tgbotapi.NewKeyboardButton(buttonServer2),
			tgbotapi.NewKeyboardButton(buttonServer3),
			tgbotapi.NewKeyboardButton(buttonServer4),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(appMenu.ButtonBackTo()),
		),
	)

	sm := servers.NewManager(bot, domain)
	sm.Sync(ctx, refreshTicker, exitChan)

	serverAppMenu := menus.NewApplicationMenu(buttonLive, liveAppName, menuKeyboard)
	serverApp := NewServerApp(ctx, bot, domain, serverAppMenu, sm)

	return &LiveApp{
		apiDomain:    domain,
		bot:          bot,
		appMenu:      appMenu,
		sm:           sm,
		menuKeyboard: menuKeyboard,
		serverApp:    serverApp,
	}
}

func (la *LiveApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	return false, nil
}

func (la *LiveApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery)) {
	return false, nil
}

func (la *LiveApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	if button == la.appMenu.Name {
		return true, func(ctx context.Context, chatId int64) error {
			message := fmt.Sprintf("%s application\n\n", la.appMenu.Name)
			msg := tgbotapi.NewMessage(chatId, message)
			msg.ReplyMarkup = la.menuKeyboard
			_, err := la.bot.Send(msg)
			return err
		}
	} else if button == la.appMenu.ButtonBackTo() {
		return true, func(ctx context.Context, chatId int64) error {
			msg := tgbotapi.NewMessage(chatId, "OK")
			msg.ReplyMarkup = la.appMenu.PrevMenu
			_, err := la.bot.Send(msg)
			return err
		}
	} else {
		return la.serverApp.AcceptButton(button)
	}
}
