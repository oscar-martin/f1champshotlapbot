package sessions

import (
	"context"
	"fmt"

	"github.com/oscar-martin/rfactor2telegrambot/pkg/menus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SessionsApp struct {
	bot          *tgbotapi.BotAPI
	apiDomain    string
	appMenu      menus.ApplicationMenu
	menuKeyboard tgbotapi.ReplyKeyboardMarkup
}

func NewSessionsApp(ctx context.Context, bot *tgbotapi.BotAPI, domain string, appMenu menus.ApplicationMenu) *SessionsApp {
	menuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(appMenu.ButtonBackTo()),
		),
	)

	return &SessionsApp{
		apiDomain:    domain,
		bot:          bot,
		appMenu:      appMenu,
		menuKeyboard: menuKeyboard,
	}
}

func (sa *SessionsApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	return false, nil
}

func (sa *SessionsApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery) error) {
	return false, nil
}

func (sa *SessionsApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	// fmt.Printf("SESSIONS: button: %s. appName: %s\n", button, sa.appMenu.Name)
	if button == sa.appMenu.Name {
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
	// fmt.Print("SESSIONS: FALSE\n")
	return false, nil
}
