package live

import (
	"context"
	"f1champshotlapsbot/pkg/apps"
	"f1champshotlapsbot/pkg/caster"
	"f1champshotlapsbot/pkg/menus"
	"f1champshotlapsbot/pkg/pubsub"
	"f1champshotlapsbot/pkg/servers"
	"fmt"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	buttonStint = "Tanda"
	buttonGrid  = "Parrilla"
)

type ServerApp struct {
	bot                   *tgbotapi.BotAPI
	appMenu               menus.ApplicationMenu
	menuKeyboard          tgbotapi.ReplyKeyboardMarkup
	gridApp               *GridApp
	stintApp              *StintApp
	accepters             []apps.Accepter
	serverID              string
	sessionInfo           servers.SessionInfo
	sessionInfoUpdateChan <-chan string
	caster                caster.ChannelCaster[servers.SessionInfo]
	mu                    sync.Mutex
}

func sanitizeServerName(name string) string {
	fixed := strings.TrimPrefix(name, servers.ServerStatusOnline)
	fixed = strings.TrimPrefix(fixed, servers.ServerStatusOffline)
	return strings.TrimSpace(fixed)
}

func NewServerApp(bot *tgbotapi.BotAPI, appMenu menus.ApplicationMenu, pubsubMgr *pubsub.PubSub, serverID string) *ServerApp {
	sa := &ServerApp{
		bot:                   bot,
		appMenu:               appMenu,
		serverID:              serverID,
		caster:                caster.JSONChannelCaster[servers.SessionInfo]{},
		sessionInfoUpdateChan: pubsubMgr.Subscribe(serverID),
	}

	go sa.updater()

	gridAppMenu := menus.NewApplicationMenu("", serverID, sa)
	gridApp := NewGridApp(bot, gridAppMenu, pubsubMgr, serverID)

	stintAppMenu := menus.NewApplicationMenu("", serverID, sa)
	stintApp := NewStintApp(bot, stintAppMenu, pubsubMgr, serverID)

	accepters := []apps.Accepter{gridApp, stintApp}

	sa.accepters = accepters
	sa.gridApp = gridApp
	sa.stintApp = stintApp
	return sa
}

func (sa *ServerApp) update(si servers.SessionInfo) {
	stint := buttonStint + " " + si.ServerName
	grid := buttonGrid + " " + si.ServerName
	menuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(stint),
			tgbotapi.NewKeyboardButton(grid),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(sa.appMenu.ButtonBackTo()),
		),
	)

	sa.menuKeyboard = menuKeyboard
	sa.sessionInfo = si
}

func (sa *ServerApp) updater() {
	for payload := range sa.sessionInfoUpdateChan {
		// fmt.Println("Updating SessionInfo")
		s, err := sa.caster.From(payload)
		if err != nil {
			fmt.Printf("Error casting SessionInfo: %s\n", err.Error())
			continue
		}
		sa.mu.Lock()
		sa.update(s)
		sa.mu.Unlock()
	}
}

func (sa *ServerApp) Menu() tgbotapi.ReplyKeyboardMarkup {
	return sa.menuKeyboard
}

func (sa *ServerApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	for _, accepter := range sa.accepters {
		accept, handler := accepter.AcceptCommand(command)
		if accept {
			return true, handler
		}
	}

	return false, nil
}

func (sa *ServerApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery) error) {
	for _, accepter := range sa.accepters {
		accept, handler := accepter.AcceptCallback(query)
		if accept {
			return true, handler
		}
	}

	return false, nil
}

func (sa *ServerApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	// fmt.Printf("SERVER: button: %s. appName: %s\n", button, sa.sessionInfo.ServerName)
	if sanitizeServerName(button) == sa.sessionInfo.ServerName {
		return true, func(ctx context.Context, chatId int64) error {
			if !sa.sessionInfo.Online {
				msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("El servidor %s está apagado", sa.sessionInfo.ServerName))
				msg.ReplyMarkup = sa.appMenu.PrevMenu()
				_, err := sa.bot.Send(msg)
				return err
			}
			si := sa.sessionInfo
			msg := tgbotapi.NewMessage(chatId,
				fmt.Sprintf(`%s:

	‣ Circuito: %s (%0.fm)
	‣ Tiempo restante: %s
	‣ Sesión: %s (Vueltas: %d)
	‣ Coches: %d
	‣ Lluvia: %d%% (min: %d%%. max: %d%%)
	‣ Temperatura (Pista/Ambiente): %0.fºC/%0.fºC
	`, sa.sessionInfo.ServerName, si.TrackName, si.LapDistance, si.TimeToEnd, si.Session, si.MaximumLaps, si.NumberOfVehicles, si.Raining, si.MinPathWetness, si.MaxPathWetness, si.TrackTemp, si.AmbientTemp))
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
	} else {
		for _, accepter := range sa.accepters {
			accept, handler := accepter.AcceptButton(button)
			if accept {
				return true, handler
			}
		}
		// fmt.Print("SERVER: FALSE\n")
		return false, nil
	}
}
