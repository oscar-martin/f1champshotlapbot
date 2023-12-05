package live

import (
	"context"
	"f1champshotlapsbot/pkg/apps"
	"f1champshotlapsbot/pkg/caster"
	"f1champshotlapsbot/pkg/menus"
	"f1champshotlapsbot/pkg/pubsub"
	"f1champshotlapsbot/pkg/servers"
	"f1champshotlapsbot/pkg/settings"
	"fmt"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	liveAppName    = "LiveTiming"
	buttonSettings = "Ajustes"
)

type LiveApp struct {
	bot                        *tgbotapi.BotAPI
	appMenu                    menus.ApplicationMenu
	menuKeyboard               tgbotapi.ReplyKeyboardMarkup
	accepters                  []apps.Accepter
	servers                    []servers.Server
	liveSessionInfoUpdateChans []<-chan string
	liveSessionInfoDataCaster  caster.ChannelCaster[servers.LiveSessionInfoData]
	mu                         sync.Mutex
}

func NewLiveApp(ctx context.Context, bot *tgbotapi.BotAPI, pubsubMgr *pubsub.PubSub, ss []servers.Server, appMenu menus.ApplicationMenu, sm *settings.Manager) (*LiveApp, error) {
	liveSessionInfoUpdateChans := []<-chan string{}
	for _, server := range ss {
		liveSessionInfoUpdateChans = append(liveSessionInfoUpdateChans, pubsubMgr.Subscribe(servers.PubSubSessionInfoPreffix+server.ID))
	}
	la := &LiveApp{
		bot:                        bot,
		appMenu:                    appMenu,
		liveSessionInfoDataCaster:  caster.JSONChannelCaster[servers.LiveSessionInfoData]{},
		liveSessionInfoUpdateChans: liveSessionInfoUpdateChans,
		servers:                    ss,
	}

	la.accepters = []apps.Accepter{}
	for _, server := range ss {
		serverAppMenu := menus.NewApplicationMenu(server.StatusAndName(), liveAppName, la)
		serverApp := NewServerApp(la.bot, serverAppMenu, pubsubMgr, server.ID, server.URL)
		la.accepters = append(la.accepters, serverApp)
	}

	settingsApp := NewSettingsApp(la.bot, appMenu, sm)
	la.accepters = append(la.accepters, settingsApp)

	la.updateKeyboard()

	for _, liveSessionInfoUpdateChan := range la.liveSessionInfoUpdateChans {
		go la.updater(liveSessionInfoUpdateChan)
	}

	return la, nil
}

func (la *LiveApp) updateKeyboard() {
	buttons := [][]tgbotapi.KeyboardButton{}
	for idx := range la.servers {
		if idx%2 == 0 {
			buttons = append(buttons, []tgbotapi.KeyboardButton{})
		}
		buttons[len(buttons)-1] = append(buttons[len(buttons)-1], tgbotapi.NewKeyboardButton(la.servers[idx].StatusAndName()))
	}
	backButtonRow := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(la.appMenu.ButtonBackTo()),
		tgbotapi.NewKeyboardButton(buttonSettings),
	)

	buttons = append(buttons, backButtonRow)

	menuKeyboard := tgbotapi.NewReplyKeyboard()
	menuKeyboard.Keyboard = buttons
	la.menuKeyboard = menuKeyboard
}

func (la *LiveApp) update(lsid servers.LiveSessionInfoData) {
	for idx := range la.servers {
		if la.servers[idx].ID == lsid.ServerID {
			if lsid.SessionInfo.ServerName != "" {
				la.servers[idx].Name = lsid.SessionInfo.ServerName
			}
			la.servers[idx].WebSocketRunning = lsid.SessionInfo.WebSocketRunning
			la.servers[idx].RecevingData = lsid.SessionInfo.RecevingData
		}
	}
	la.updateKeyboard()
}

func (la *LiveApp) updater(c <-chan string) {
	for payload := range c {
		lsid, err := la.liveSessionInfoDataCaster.From(payload)
		if err != nil {
			log.Printf("Error casting session info: %s\n", err.Error())
			continue
		}
		la.mu.Lock()
		la.update(lsid)
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
			message := fmt.Sprintf("%s\n", la.appMenu.Name)
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
