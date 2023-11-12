package apps

import (
	"context"
	"f1champshotlapsbot/pkg/menus"
	"f1champshotlapsbot/pkg/servers"
	"fmt"
	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	buttonStint = "Piloto"
	buttonGrid  = "Parrilla"
)

type ServerApp struct {
	bot          *tgbotapi.BotAPI
	apiDomain    string
	appMenu      menus.ApplicationMenu
	sm           *servers.Manager
	menuKeyboard tgbotapi.ReplyKeyboardMarkup
}

func NewServerApp(ctx context.Context, bot *tgbotapi.BotAPI, domain string, appMenu menus.ApplicationMenu, sm *servers.Manager) *ServerApp {
	menuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(buttonStint),
			tgbotapi.NewKeyboardButton(buttonGrid),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(appMenu.ButtonBackTo()),
		),
	)

	return &ServerApp{
		apiDomain:    domain,
		bot:          bot,
		appMenu:      appMenu,
		sm:           sm,
		menuKeyboard: menuKeyboard,
	}
}

func (la *ServerApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	return false, nil
}

func (la *ServerApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery)) {
	return false, nil
}

func (la *ServerApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
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
		commandServerId := regexp.MustCompile(fmt.Sprintf(`^%s(\d+)$`, serverPrefixCommand))
		if commandServerId.MatchString(button) {
			serverId := commandServerId.FindStringSubmatch(button)[1]
			return true, func(ctx context.Context, chatId int64) error {
				msg, err := la.sm.RenderServerId(serverId)(ctx, chatId)
				if err != nil {
					return err
				}
				msg.ReplyMarkup = la.menuKeyboard
				_, err = la.bot.Send(msg)
				return err
			}
		}
	}
	return false, nil
}
