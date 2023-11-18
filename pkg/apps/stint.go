package apps

import (
	"context"
	"f1champshotlapsbot/pkg/menus"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const ()

type StintApp struct {
	bot          *tgbotapi.BotAPI
	appMenu      menus.ApplicationMenu
	serverName   string
	menuKeyboard tgbotapi.ReplyKeyboardMarkup
}

func NewStintApp(bot *tgbotapi.BotAPI, appMenu menus.ApplicationMenu, serverName string) *StintApp {
	menuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(appMenu.ButtonBackTo()),
		),
	)

	return &StintApp{
		bot:          bot,
		appMenu:      appMenu,
		serverName:   serverName,
		menuKeyboard: menuKeyboard,
	}
}

func (sa *StintApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	return false, nil
}

func (sa *StintApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery) error) {
	return false, nil
}

func (sa *StintApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	// fmt.Printf("STINT: button: %s. appName: %s\n", button, sa.appMenu.Name)
	if button == buttonStint+" "+sa.serverName {
		return true, func(ctx context.Context, chatId int64) error {
			message := fmt.Sprintf("%s application\n\n", sa.appMenu.Name)
			msg := tgbotapi.NewMessage(chatId, message)
			msg.ReplyMarkup = sa.menuKeyboard
			_, err := sa.bot.Send(msg)
			return err
		}
	} else if button == sa.appMenu.ButtonBackTo() {
		return true, func(ctx context.Context, chatId int64) error {
			msg := tgbotapi.NewMessage(chatId, "OK")
			msg.ReplyMarkup = sa.appMenu.PrevMenu()
			_, err := sa.bot.Send(msg)
			return err
		}
	}
	// fmt.Print("STINT: FALSE\n")
	return false, nil
}
