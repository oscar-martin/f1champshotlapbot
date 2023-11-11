package main

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	menuStart = "/start"
	menuMenu  = "/menu"
)

type Menu struct {
	bot *tgbotapi.BotAPI
}

func NewMenu(bot *tgbotapi.BotAPI) *Menu {
	return &Menu{
		bot: bot,
	}
}

func (m *Menu) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	if command == menuStart {
		return true, m.renderStart()
	} else if command == menuMenu {
		return true, m.renderMenu()
	}
	return false, nil
}

func (m *Menu) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery)) {
	return false, nil
}

func (s *Menu) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
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
