package main

import (
	"context"
	"f1champshotlapsbot/pkg/menus"
	"f1champshotlapsbot/pkg/servers"
	"f1champshotlapsbot/pkg/tracks"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	menuStart     = "/start"
	menuMenu      = "/menu"
	buttonHotlaps = "Hotlaps"
	buttonLive    = "Live"
	appName       = "menu"
)

var (
	menuKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(buttonHotlaps),
			tgbotapi.NewKeyboardButton(buttonLive),
		),
	)
)

type Menu struct {
	bot       *tgbotapi.BotAPI
	accepters []Accepter
}

func NewMenu(ctx context.Context, bot *tgbotapi.BotAPI, domain string, exitChan chan bool, refreshHotlapsTicker, refreshServersTicker *time.Ticker) *Menu {
	hotlapsAppMenu := menus.NewApplicationMenu(buttonHotlaps, appName, menuKeyboard)
	tm := tracks.NewTrackManager(bot, domain, hotlapsAppMenu)

	serversAppMenu := menus.NewApplicationMenu(buttonLive, appName, menuKeyboard)
	sm := servers.NewManager(bot, domain, serversAppMenu)
	accepters := []Accepter{tm, sm}

	// start the trackmanager sync process
	tm.Sync(ctx, refreshHotlapsTicker, exitChan)
	sm.Sync(ctx, refreshServersTicker, exitChan)

	return &Menu{
		bot:       bot,
		accepters: accepters,
	}
}

func (m *Menu) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
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

func (m *Menu) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery)) {
	for _, accepter := range m.accepters {
		accept, handler := accepter.AcceptCallback(query)
		if accept {
			return true, handler
		}
	}

	return false, nil
}

func (m *Menu) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	for _, accepter := range m.accepters {
		accept, handler := accepter.AcceptButton(button)
		if accept {
			return true, handler
		}
	}
	return false, nil
}

func (m *Menu) renderStart() func(ctx context.Context, chatId int64) error {
	return func(ctx context.Context, chatId int64) error {
		message := "Hola, soy el bot de F1Champs que permite ver las Hotlaps registradas y sesiones en curso.\n\n"
		message += "Puedes usar el siguiente comando:\n\n"
		message += fmt.Sprintf("%s - Muestra el menú del bot\n", menuMenu)
		msg := tgbotapi.NewMessage(chatId, message)
		msg.ReplyMarkup = menuKeyboard
		_, err := bot.Send(msg)
		return err
	}
}

func (m *Menu) renderMenu() func(ctx context.Context, chatId int64) error {
	return func(ctx context.Context, chatId int64) error {
		message := "Menú del bot.\n\n"
		msg := tgbotapi.NewMessage(chatId, message)
		msg.ReplyMarkup = menuKeyboard
		_, err := bot.Send(msg)
		return err
	}
}
