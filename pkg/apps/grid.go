package apps

import (
	"context"
	"f1champshotlapsbot/pkg/caster"
	"f1champshotlapsbot/pkg/menus"
	"f1champshotlapsbot/pkg/pubsub"
	"f1champshotlapsbot/pkg/servers"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const ()

type GridApp struct {
	bot                      *tgbotapi.BotAPI
	appMenu                  menus.ApplicationMenu
	serverID                 string
	driversSession           servers.DriversSession
	driversSessionUpdateChan <-chan string
	caster                   caster.ChannelCaster[servers.DriversSession]
	mu                       sync.Mutex
	menuKeyboard             tgbotapi.ReplyKeyboardMarkup
}

func NewGridApp(bot *tgbotapi.BotAPI, appMenu menus.ApplicationMenu, pubsubMgr *pubsub.PubSub, serverID string) *GridApp {
	menuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(appMenu.ButtonBackTo()),
		),
	)

	ga := &GridApp{
		bot:                      bot,
		appMenu:                  appMenu,
		serverID:                 serverID,
		caster:                   caster.JSONChannelCaster[servers.DriversSession]{},
		driversSessionUpdateChan: pubsubMgr.Subscribe(servers.PubSubDriversSessionPreffix + serverID),
		menuKeyboard:             menuKeyboard,
	}

	go ga.updater()

	return ga
}

func (ga *GridApp) updater() {
	for payload := range ga.driversSessionUpdateChan {
		fmt.Println("Updating DriverSessions")
		dss, err := ga.caster.From(payload)
		if err != nil {
			fmt.Printf("Error casting DriverSessions: %s\n", err.Error())
			continue
		}
		ga.mu.Lock()
		ga.update(dss)
		ga.mu.Unlock()
	}
}

func (ga *GridApp) update(dss servers.DriversSession) {
	ga.driversSession = dss
}

func (ga *GridApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	return false, nil
}

func (ga *GridApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery) error) {
	return false, nil
}

func (ga *GridApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	ga.mu.Lock()
	defer ga.mu.Unlock()

	// fmt.Printf("GRID: button: %s. appName: %s\n", button, buttonGrid+" "+ga.driversSession.ServerName)
	if button == buttonGrid+" "+ga.driversSession.ServerName {
		return true, func(ctx context.Context, chatId int64) error {
			message := fmt.Sprintf("%s application\n\nDrivers: %d", ga.appMenu.Name, len(ga.driversSession.Drivers))
			msg := tgbotapi.NewMessage(chatId, message)
			msg.ReplyMarkup = ga.menuKeyboard
			_, err := ga.bot.Send(msg)
			return err
		}
	} else if button == ga.appMenu.ButtonBackTo() {
		return true, func(ctx context.Context, chatId int64) error {
			msg := tgbotapi.NewMessage(chatId, "OK1")
			msg.ReplyMarkup = ga.appMenu.PrevMenu()
			_, err := ga.bot.Send(msg)
			return err
		}
	}
	// fmt.Print("GRID: FALSE\n")
	return false, nil
}
