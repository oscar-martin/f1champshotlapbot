package live

import (
	"context"
	"f1champshotlapsbot/pkg/apps"
	"f1champshotlapsbot/pkg/caster"
	"f1champshotlapsbot/pkg/menus"
	"f1champshotlapsbot/pkg/pubsub"
	"f1champshotlapsbot/pkg/servers"
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	liveAppName    = "Live"
	buttonSettings = "Ajustes"
)

type LiveApp struct {
	bot               *tgbotapi.BotAPI
	appMenu           menus.ApplicationMenu
	menuKeyboard      tgbotapi.ReplyKeyboardMarkup
	accepters         []apps.Accepter
	serversUpdateChan <-chan string
	caster            caster.ChannelCaster[[]servers.Server]
	mu                sync.Mutex
}

func NewLiveApp(ctx context.Context, bot *tgbotapi.BotAPI, domain string, pubsubMgr *pubsub.PubSub, appMenu menus.ApplicationMenu, exitChan chan bool, refreshTicker *time.Ticker) *LiveApp {
	sm := servers.NewManager(ctx, bot, domain, pubsubMgr)
	sm.Sync(refreshTicker, exitChan)

	ss, err := sm.GetInitialServers()
	if err != nil {
		fmt.Printf("Error getting initial servers: %s\n", err.Error())
	}

	la := &LiveApp{
		bot:               bot,
		appMenu:           appMenu,
		caster:            caster.JSONChannelCaster[[]servers.Server]{},
		serversUpdateChan: pubsubMgr.Subscribe(servers.PubSubServersTopic),
	}

	la.accepters = []apps.Accepter{}
	for _, server := range ss {
		serverAppMenu := menus.NewApplicationMenu(server.StatusAndName(), liveAppName, la)
		serverApp := NewServerApp(la.bot, serverAppMenu, pubsubMgr, server.ID)
		la.accepters = append(la.accepters, serverApp)
	}

	la.update(ss)
	go la.updater()

	return la
}

func (la *LiveApp) update(ss []servers.Server) {
	menuKeyboard := tgbotapi.NewReplyKeyboard()
	menuKeyboard.Keyboard = make([][]tgbotapi.KeyboardButton, len(ss)+1)

	for idx, server := range ss {
		menuKeyboard.Keyboard[idx] = []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(server.StatusAndName())}
	}
	backButtonRow := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(la.appMenu.ButtonBackTo()),
		tgbotapi.NewKeyboardButton(buttonSettings),
	)
	menuKeyboard.Keyboard[len(ss)] = backButtonRow

	la.menuKeyboard = menuKeyboard
}

func (la *LiveApp) updater() {
	for payload := range la.serversUpdateChan {
		// fmt.Println("Updating servers statuses")
		ss, err := la.caster.From(payload)
		if err != nil {
			fmt.Printf("Error casting servers: %s\n", err.Error())
			continue
		}
		la.mu.Lock()
		la.update(ss)
		la.mu.Unlock()
	}
}

func (la *LiveApp) Menu() tgbotapi.ReplyKeyboardMarkup {
	la.mu.Lock()
	defer la.mu.Unlock()

	return la.menuKeyboard
}

func (la *LiveApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	for _, accepter := range la.accepters {
		accept, handler := accepter.AcceptCommand(command)
		if accept {
			return true, handler
		}
	}

	return false, nil
}

func (la *LiveApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery) error) {
	for _, accepter := range la.accepters {
		accept, handler := accepter.AcceptCallback(query)
		if accept {
			return true, handler
		}
	}

	return false, nil
}

func (la *LiveApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	la.mu.Lock()
	defer la.mu.Unlock()

	// fmt.Printf("LIVE: button: %s. appName: %s\n", button, la.appMenu.Name)
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
			msg.ReplyMarkup = la.appMenu.PrevMenu()
			_, err := la.bot.Send(msg)
			return err
		}
	}
	for _, accepter := range la.accepters {
		accept, handler := accepter.AcceptButton(button)
		if accept {
			return true, handler
		}
	}
	// fmt.Print("LIVE: FALSE\n")
	return false, nil

}
