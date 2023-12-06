package mainapp

import (
	"context"
	"f1champshotlapsbot/pkg/apps"
	"f1champshotlapsbot/pkg/apps/hotlaps"
	"f1champshotlapsbot/pkg/apps/live"
	"f1champshotlapsbot/pkg/apps/sessions"
	"f1champshotlapsbot/pkg/menus"
	"f1champshotlapsbot/pkg/servers"
	"f1champshotlapsbot/pkg/settings"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	menuStart      = "/start"
	menuMenu       = "/menu"
	buttonHotlaps  = "Hotlaps"
	buttonLive     = "LiveTiming"
	buttonSessions = "Sessions"
	appName        = "menu"
)

var (
	menuKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(buttonHotlaps),
			// tgbotapi.NewKeyboardButton(buttonSessions),
			tgbotapi.NewKeyboardButton(buttonLive),
		),
	)
)

type menuer struct{}

func (m menuer) Menu() tgbotapi.ReplyKeyboardMarkup {
	return menuKeyboard
}

type MainApp struct {
	bot       *tgbotapi.BotAPI
	accepters []apps.Accepter
}

func NewMainApp(ctx context.Context, bot *tgbotapi.BotAPI, domain string, ss []servers.Server, exitChan chan bool, refreshHotlapsTicker *time.Ticker, sm *settings.Manager) (*MainApp, error) {
	hotlapsAppMenu := menus.NewApplicationMenu(buttonHotlaps, appName, menuer{})
	hotlapApp := hotlaps.NewHotlapsApp(ctx, bot, domain, hotlapsAppMenu, exitChan, refreshHotlapsTicker)

	sessionsAppMenu := menus.NewApplicationMenu(buttonSessions, appName, menuer{})
	sessionsApp := sessions.NewSessionsApp(ctx, bot, domain, sessionsAppMenu)

	liveAppMenu := menus.NewApplicationMenu(buttonLive, appName, menuer{})
	liveApp, err := live.NewLiveApp(ctx, bot, ss, liveAppMenu, sm)
	if err != nil {
		return nil, err
	}

	accepters := []apps.Accepter{hotlapApp, sessionsApp, liveApp}

	return &MainApp{
		bot:       bot,
		accepters: accepters,
	}, nil
}

func (m *MainApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	if command == menuStart {
		return true, m.renderStart()
	} else if command == menuMenu {
		return true, m.renderMenu()
	}
	for _, accepter := range m.accepters {
		accept, handler := accepter.AcceptCommand(command)
		if accept {
			return true, handler
		}
	}

	return false, nil
}

func (m *MainApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery) error) {
	for _, accepter := range m.accepters {
		accept, handler := accepter.AcceptCallback(query)
		if accept {
			return true, handler
		}
	}

	return false, nil
}

func (m *MainApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	for _, accepter := range m.accepters {
		accept, handler := accepter.AcceptButton(button)
		if accept {
			return true, handler
		}
	}
	return false, nil
}

func (m *MainApp) renderStart() func(ctx context.Context, chatId int64) error {
	return func(ctx context.Context, chatId int64) error {
		message := "Hola, soy el bot de F1Champs que permite ver las Hotlaps registradas y sesiones en curso.\n\n"
		message += "Puedes usar el siguiente comando:\n\n"
		message += fmt.Sprintf("%s - Muestra el menú del bot\n", menuMenu)
		msg := tgbotapi.NewMessage(chatId, message)
		msg.ReplyMarkup = menuKeyboard
		_, err := m.bot.Send(msg)
		return err
	}
}

func (m *MainApp) renderMenu() func(ctx context.Context, chatId int64) error {
	return func(ctx context.Context, chatId int64) error {
		message := "Menú del bot.\n\n"
		msg := tgbotapi.NewMessage(chatId, message)
		msg.ReplyMarkup = menuKeyboard
		_, err := m.bot.Send(msg)
		return err
	}
}
