package live

import (
	"context"
	"f1champshotlapsbot/pkg/menus"
	"f1champshotlapsbot/pkg/settings"
	"fmt"
	"log"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ContextUser string
type ContextChatID string

const (
	UserContextKey         ContextUser   = "user"
	ChatContextKey         ContextChatID = "chat"
	inlineKeyboardTestday                = settings.TestDay
	inlineKeyboardPractice               = settings.Practice
	inlineKeyboardQual                   = settings.Qual
	inlineKeyboardWarmup                 = settings.Warmup
	inlineKeyboardRace                   = settings.Race

	symbolNotifications     = "🔔"
	subcommandNotifications = "notifications"
)

type SettingsApp struct {
	bot          *tgbotapi.BotAPI
	appMenu      menus.ApplicationMenu
	menuKeyboard tgbotapi.ReplyKeyboardMarkup
	sm           *settings.Manager
	mu           sync.Mutex
}

func NewSettingsApp(bot *tgbotapi.BotAPI, appMenu menus.ApplicationMenu, sm *settings.Manager) *SettingsApp {
	sa := &SettingsApp{
		bot:     bot,
		sm:      sm,
		appMenu: appMenu,
	}

	return sa
}

func (sa *SettingsApp) Menu() tgbotapi.ReplyKeyboardMarkup {
	return sa.menuKeyboard
}

func (sa *SettingsApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	return false, nil
}

func (sa *SettingsApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery) error) {
	data := strings.Split(query.Data, ":")
	if data[0] == subcommandNotifications {
		sa.mu.Lock()
		defer sa.mu.Unlock()
		return true, func(ctx context.Context, query *tgbotapi.CallbackQuery) error {
			userID := data[1]
			sessionType := data[2]

			chatCtxValue := ctx.Value(ChatContextKey)
			if chatCtxValue == nil {
				msg := tgbotapi.NewMessage(query.Message.Chat.ID, "No se pudo leer información del chat")
				msg.ReplyMarkup = sa.appMenu.PrevMenu()
				_, err := sa.bot.Send(msg)
				return err
			}
			chat := chatCtxValue.(*tgbotapi.Chat)
			chatID := fmt.Sprintf("%d", chat.ID)

			err := sa.sm.ToggleNotificationForSessionStarted(userID, chatID, sessionType)
			if err != nil {
				msg := tgbotapi.NewMessage(query.Message.Chat.ID, "No se pudo cambiar el estado de la notificación")
				msg.ReplyMarkup = sa.appMenu.PrevMenu()
				_, err := sa.bot.Send(msg)
				return err
			}
			return sa.renderNotifications(&query.Message.MessageID)(ctx, query.Message.Chat.ID)
		}
	}
	return false, nil
}

func (sa *SettingsApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	// fmt.Printf("SETTINGS: button: %s. appName: %s\n", button, buttonSettings)
	if button == buttonSettings {
		return true, sa.renderNotifications(nil)
	} else if button == sa.appMenu.ButtonBackTo() {
		return true, func(ctx context.Context, chatId int64) error {
			msg := tgbotapi.NewMessage(chatId, "OK")
			msg.ReplyMarkup = sa.appMenu.PrevMenu()
			_, err := sa.bot.Send(msg)
			return err
		}
	}
	return false, nil
}

func (sa *SettingsApp) renderNotifications(messageID *int) func(ctx context.Context, chatId int64) error {
	return func(ctx context.Context, chatId int64) error {
		userCtxValue := ctx.Value(UserContextKey)
		if userCtxValue == nil {
			msg := tgbotapi.NewMessage(chatId, "No se pudo leer el usuario")
			msg.ReplyMarkup = sa.appMenu.PrevMenu()
			_, err := sa.bot.Send(msg)
			return err
		}
		user := userCtxValue.(*tgbotapi.User)
		userID := fmt.Sprintf("%d", user.ID)
		notificationStatus, err := sa.sm.ListNotifications(userID)
		if err != nil {
			log.Println(err)
			msg := tgbotapi.NewMessage(chatId, "No se pudieron leer los de notificaciones para el usuario")
			msg.ReplyMarkup = sa.appMenu.PrevMenu()
			_, err := sa.bot.Send(msg)
			return err
		}
		keyboard := getSettingsInlineKeyboard(userID, notificationStatus)
		var cfg tgbotapi.Chattable
		if messageID == nil {
			msg := tgbotapi.NewMessage(chatId, "Estado de notificaciones\n(Solo notifica la primera sesión)")
			msg.ReplyMarkup = keyboard
			cfg = msg
		} else {
			msg := tgbotapi.NewEditMessageText(chatId, *messageID, "Estado de notificaciones\n(Solo notifica la primera sesión)")
			msg.ReplyMarkup = &keyboard
			cfg = msg
		}
		_, err = sa.bot.Send(cfg)
		return err
	}
}

func getSettingsInlineKeyboard(userID string, n settings.Notifications) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardTestday+" "+n.TestDaySymbol(), fmt.Sprintf("%s:%s:%s", subcommandNotifications, userID, inlineKeyboardTestday)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardPractice+" "+n.PracticeSymbol(), fmt.Sprintf("%s:%s:%s", subcommandNotifications, userID, inlineKeyboardPractice)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardQual+" "+n.QualSymbol(), fmt.Sprintf("%s:%s:%s", subcommandNotifications, userID, inlineKeyboardQual)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardWarmup+" "+n.WarmupSymbol(), fmt.Sprintf("%s:%s:%s", subcommandNotifications, userID, inlineKeyboardWarmup)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardRace+" "+n.RaceSymbol(), fmt.Sprintf("%s:%s:%s", subcommandNotifications, userID, inlineKeyboardRace)),
		),
	)
}
