package apps

import (
	"bytes"
	"context"
	"f1champshotlapsbot/pkg/caster"
	"f1champshotlapsbot/pkg/helper"
	"f1champshotlapsbot/pkg/menus"
	"f1champshotlapsbot/pkg/pubsub"
	"f1champshotlapsbot/pkg/servers"
	"fmt"
	"log"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	inlineKeyboardTimes    = "Tiempos"
	inlineKeyboardSectors  = "Sectores"
	inlineKeyboardCompound = "Gomas"
	inlineKeyboardLaps     = "Vueltas"
	inlineKeyboardTeam     = "Coches"
	inlineKeyboardDriver   = "Pilotos"
	inlineKeyboardUpdate   = "Actualizar"
	inlineKeyboardDiff     = "Diferencia"
	inlineKeyboardMaxSpeed = "Máx. Vel"

	symbolTimes    = "⏱"
	symbolSectors  = "🔂"
	symbolCompound = "🛞"
	symbolLaps     = "🏁"
	symbolTeam     = "🏎️"
	symbolDriver   = "👐"
	symbolUpdate   = "🔄"
	symbolDiff     = "⏲️"
	symbolMaxSpeed = "🚀"

	SubcommandShowLiveTiming = "show_live_timing"

	tableDriver = "PIL"
)

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
		// fmt.Println("Updating DriverSessions")
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
	data := strings.Split(query.Data, ":")
	if data[0] == SubcommandShowLiveTiming && data[2] == ga.serverID {
		ga.mu.Lock()
		defer ga.mu.Unlock()
		return true, func(ctx context.Context, query *tgbotapi.CallbackQuery) error {
			return ga.handleSessionDataCallbackQuery(query.Message.Chat.ID, &query.Message.MessageID, data[1:]...)
		}
	}
	return false, nil
}

func (ga *GridApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	ga.mu.Lock()
	defer ga.mu.Unlock()

	// fmt.Printf("GRID: button: %s. appName: %s\n", button, buttonGrid+" "+ga.driversSession.ServerName)
	if button == buttonGrid+" "+ga.driversSession.ServerName {
		return true, ga.RenderGrid()
	} else if button == ga.appMenu.ButtonBackTo() {
		return true, func(ctx context.Context, chatId int64) error {
			msg := tgbotapi.NewMessage(chatId, "OK")
			msg.ReplyMarkup = ga.appMenu.PrevMenu()
			_, err := ga.bot.Send(msg)
			return err
		}
	}
	// fmt.Print("GRID: FALSE\n")
	return false, nil
}

func (ga *GridApp) RenderGrid() func(ctx context.Context, chatId int64) error {
	return func(ctx context.Context, chatId int64) error {
		err := ga.SendSessionData(chatId, nil, ga.driversSession, inlineKeyboardTimes)
		if err != nil {
			log.Printf("An error occured: %s", err.Error())
		}
		return nil
	}
}

func (ga *GridApp) handleSessionDataCallbackQuery(chatId int64, messageId *int, data ...string) error {
	infoType := data[0]
	return ga.SendSessionData(chatId, messageId, ga.driversSession, infoType)
}

func (ga *GridApp) SendSessionData(chatId int64, messageId *int, driversSession servers.DriversSession, infoType string) error {
	if len(driversSession.Drivers) > 0 {
		var b bytes.Buffer
		t := table.NewWriter()
		t.SetOutputMirror(&b)
		t.SetStyle(table.StyleRounded)
		t.AppendSeparator()

		t.AppendHeader(table.Row{tableDriver, infoType})
		for _, driverStat := range driversSession.Drivers {
			switch infoType {
			case inlineKeyboardTimes:
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.Driver),
					helper.SecondsToMinutes(driverStat.Time),
				})
			case inlineKeyboardSectors:
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.Driver),
					fmt.Sprintf("%s %s %s", helper.ToSectorTime(driverStat.S1), helper.ToSectorTime(driverStat.S2), helper.ToSectorTime(driverStat.S3)),
				})
			case inlineKeyboardCompound:
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.Driver),
					driverStat.Compound,
				})
			case inlineKeyboardLaps:
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.Driver),
					fmt.Sprintf("%d/%d", driverStat.Lapcountcomplete, driverStat.Lapcount),
				})
			case inlineKeyboardTeam:
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.Driver),
					driverStat.CarClass,
				})
			case inlineKeyboardDriver:
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.Driver),
					driverStat.Driver,
				})
			case inlineKeyboardDiff:
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.Driver),
					fmt.Sprintf("%.1fs", driverStat.Diff),
				})
			}
		}
		t.Render()

		keyboard := getInlineKeyboard(driversSession.ServerID)
		var cfg tgbotapi.Chattable
		if messageId == nil {
			msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("```\nDatos de la sesión actual en %q\n\n%s```", driversSession.ServerName, b.String()))
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = keyboard
			cfg = msg
		} else {
			msg := tgbotapi.NewEditMessageText(chatId, *messageId, fmt.Sprintf("```\nDatos de la sesión actual en %q\n\n%s```", driversSession.ServerName, b.String()))
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = &keyboard
			cfg = msg
		}
		_, err := ga.bot.Send(cfg)
		return err
	} else {
		message := "No hay pilotos en la sesión"
		msg := tgbotapi.NewMessage(chatId, message)
		_, err := ga.bot.Send(msg)
		return err
	}
}

func getInlineKeyboard(serverID string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardTimes+" "+symbolTimes, fmt.Sprintf("%s:%s:%s", SubcommandShowLiveTiming, inlineKeyboardTimes, serverID)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardDiff+" "+symbolDiff, fmt.Sprintf("%s:%s:%s", SubcommandShowLiveTiming, inlineKeyboardDiff, serverID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardSectors+" "+symbolSectors, fmt.Sprintf("%s:%s:%s", SubcommandShowLiveTiming, inlineKeyboardSectors, serverID)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardCompound+" "+symbolTimes, fmt.Sprintf("%s:%s:%s", SubcommandShowLiveTiming, inlineKeyboardCompound, serverID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardLaps+" "+symbolLaps, fmt.Sprintf("%s:%s:%s", SubcommandShowLiveTiming, inlineKeyboardLaps, serverID)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardTeam+" "+symbolTeam, fmt.Sprintf("%s:%s:%s", SubcommandShowLiveTiming, inlineKeyboardTeam, serverID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardDriver+" "+symbolDriver, fmt.Sprintf("%s:%s:%s", SubcommandShowLiveTiming, inlineKeyboardDriver, serverID)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardUpdate+" "+symbolUpdate, fmt.Sprintf("%s:%s:%s", SubcommandShowLiveTiming, inlineKeyboardSectors, serverID)),
		),
	)
}
