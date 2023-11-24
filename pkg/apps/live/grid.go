package live

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
	subcommandShowLiveTiming = "show_live_timing"
	tableDriver              = "PIL"
)

type GridApp struct {
	bot                        *tgbotapi.BotAPI
	appMenu                    menus.ApplicationMenu
	serverID                   string
	liveStandingData           servers.LiveStandingData
	liveStandingDataUpdateChan <-chan string
	caster                     caster.ChannelCaster[servers.LiveStandingData]
	mu                         sync.Mutex
	menuKeyboard               tgbotapi.ReplyKeyboardMarkup
}

func NewGridApp(bot *tgbotapi.BotAPI, appMenu menus.ApplicationMenu, pubsubMgr *pubsub.PubSub, serverID string) *GridApp {
	menuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(appMenu.ButtonBackTo()),
		),
	)

	ga := &GridApp{
		bot:                        bot,
		appMenu:                    appMenu,
		serverID:                   serverID,
		caster:                     caster.JSONChannelCaster[servers.LiveStandingData]{},
		liveStandingDataUpdateChan: pubsubMgr.Subscribe(servers.PubSubDriversSessionPreffix + serverID),
		menuKeyboard:               menuKeyboard,
	}

	go ga.updater()

	return ga
}

func (ga *GridApp) updater() {
	for payload := range ga.liveStandingDataUpdateChan {
		// fmt.Println("Updating DriverSessions")
		dss, err := ga.caster.From(payload)
		if err != nil {
			fmt.Printf("Error casting DriverSessions: %s\n", err.Error())
			continue
		}
		ga.update(dss)
	}
}

func (ga *GridApp) update(lsd servers.LiveStandingData) {
	ga.mu.Lock()
	defer ga.mu.Unlock()
	ga.liveStandingData = lsd
}

func (ga *GridApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	return false, nil
}

func (ga *GridApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery) error) {
	data := strings.Split(query.Data, ":")
	if data[0] == subcommandShowLiveTiming && data[1] == ga.serverID {
		ga.mu.Lock()
		defer ga.mu.Unlock()
		return true, func(ctx context.Context, query *tgbotapi.CallbackQuery) error {
			return ga.handleSessionDataCallbackQuery(query.Message.Chat.ID, &query.Message.MessageID, data[2:]...)
		}
	}
	return false, nil
}

func (ga *GridApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	ga.mu.Lock()
	defer ga.mu.Unlock()

	// fmt.Printf("GRID: button: %s. appName: %s\n", button, buttonGrid+" "+ga.driversSession.ServerName)
	if button == buttonGrid+" "+ga.liveStandingData.ServerName {
		return true, ga.renderGrid()
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

func (ga *GridApp) renderGrid() func(ctx context.Context, chatId int64) error {
	return func(ctx context.Context, chatId int64) error {
		err := ga.sendSessionData(chatId, nil, ga.liveStandingData, inlineKeyboardBestLap)
		if err != nil {
			log.Printf("An error occured: %s", err.Error())
		}
		return nil
	}
}

func (ga *GridApp) handleSessionDataCallbackQuery(chatId int64, messageId *int, data ...string) error {
	infoType := data[0]
	return ga.sendSessionData(chatId, messageId, ga.liveStandingData, infoType)
}

func (ga *GridApp) sendSessionData(chatId int64, messageId *int, driversSession servers.LiveStandingData, infoType string) error {
	if len(driversSession.Drivers) > 0 {
		var b bytes.Buffer
		t := table.NewWriter()
		t.SetOutputMirror(&b)
		t.SetStyle(table.StyleRounded)
		t.AppendSeparator()

		switch infoType {
		case inlineKeyboardStatus:
			t.AppendHeader(table.Row{tableDriver, "Sectores", "S" /*, "FUEL"*/})
		case inlineKeyboardInfo:
			t.AppendHeader(table.Row{tableDriver, "Nombre" /*, "Núm"*/, "Lap"})
		case inlineKeyboardLastLap:
			t.AppendHeader(table.Row{tableDriver, "Última", "Mejor"})
		case inlineKeyboardOptimumLap:
			t.AppendHeader(table.Row{tableDriver, "Óptimo", "Mejor"})
		default:
			t.AppendHeader(table.Row{tableDriver, infoType})
		}
		for idx, driverStat := range driversSession.Drivers {
			switch infoType {
			case inlineKeyboardStatus:
				state := ""
				if driverStat.InGarageStall {
					state = "P"
				} else if driverStat.Pitting {
					state = "P"
				}
				var s1 float64
				s2 := -1.0
				s3 := -1.0
				if driverStat.CurrentSectorTime1 > 0.0 {
					// s1 is done in current lap
					s1 = driverStat.CurrentSectorTime1
					if s1 > 0.0 && driverStat.CurrentSectorTime2 > 0.0 {
						// s2 is done in current lap
						s2 = driverStat.CurrentSectorTime2 - s1
					}
				} else {
					s1 = driverStat.LastSectorTime1
					if s1 > 0.0 && driverStat.LastSectorTime2 > 0.0 {
						s2 = driverStat.LastSectorTime2 - s1
					}
					if s2 > 0.0 && driverStat.LastLapTime > 0.0 {
						s3 = driverStat.LastLapTime - s2 - s1
					}
				}
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.DriverName),
					fmt.Sprintf("%s %s %s", helper.ToSectorTime(s1), helper.ToSectorTime(s2), helper.ToSectorTime(s3)),
					// fmt.Sprintf("%.0f%%", driverStat.FuelFraction*100),
					state,
				})
			case inlineKeyboardInfo:
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.DriverName),
					driverStat.DriverName,
					// driverStat.CarNumber,
					driverStat.LapsCompleted,
				})
			case inlineKeyboardDiff:
				diff := ""
				if idx == 0 {
					diff = helper.SecondsToMinutes(driverStat.BestLapTime)
				} else {
					diff = helper.SecondsToDiff(driverStat.BestLapTime - driversSession.Drivers[0].BestLapTime)
				}
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.DriverName),
					diff,
				})
			case inlineKeyboardBestLap:
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.DriverName),
					helper.SecondsToMinutes(driverStat.BestLapTime),
				})
			case inlineKeyboardLastLap:
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.DriverName),
					helper.SecondsToMinutes(driverStat.LastLapTime),
					helper.SecondsToMinutes(driverStat.BestLapTime),
				})
			case inlineKeyboardOptimumLap:
				optimumLap := -1.0
				if driverStat.BestSectorTime2 > 0.0 && driverStat.BestSectorTime3 > 0.0 {
					optimumLap = driverStat.BestSectorTime2 + driverStat.BestSectorTime3
				}
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.DriverName),
					helper.SecondsToMinutes(optimumLap),
					helper.SecondsToMinutes(driverStat.BestLapTime),
				})
			case inlineKeyboardBestLapSectors:
				bs1 := driverStat.BestLapSectorTime1
				bs2 := -1.0
				if bs1 > 0.0 {
					bs2 = driverStat.BestLapSectorTime2 - bs1
				}
				bs3 := -1.0
				if bs2 > 0.0 && driverStat.BestLapTime > 0.0 {
					bs3 = driverStat.BestLapTime - bs2 - bs1
				}
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.DriverName),
					fmt.Sprintf("%s %s %s", helper.ToSectorTime(bs1), helper.ToSectorTime(bs2), helper.ToSectorTime(bs3)),
				})
			case inlineKeyboardLastLapSectors:
				ls1 := driverStat.LastSectorTime1
				ls2 := -1.0
				if ls1 > 0.0 && driverStat.LastSectorTime2 > 0.0 {
					ls2 = driverStat.LastSectorTime2 - ls1
				}
				ls3 := -1.0
				if ls2 > 0.0 && driverStat.LastLapTime > 0.0 {
					ls3 = driverStat.LastLapTime - ls2 - ls1
				}
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.DriverName),
					fmt.Sprintf("%s %s %s", helper.ToSectorTime(ls1), helper.ToSectorTime(ls2), helper.ToSectorTime(ls3)),
				})
			case inlineKeyboardOptimumLapSectors:
				ls1 := driverStat.BestSectorTime1
				ls2 := -1.0
				if ls1 > 0.0 && driverStat.BestSectorTime2 > 0.0 {
					ls2 = driverStat.BestSectorTime2 - ls1
				}
				ls3 := -1.0
				if driverStat.BestSectorTime3 > 0.0 {
					ls3 = driverStat.BestSectorTime3
				}
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.DriverName),
					fmt.Sprintf("%s %s %s", helper.ToSectorTime(ls1), helper.ToSectorTime(ls2), helper.ToSectorTime(ls3)),
				})
			case inlineKeyboardLaps:
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.DriverName),
					fmt.Sprintf("%d", driverStat.LapsCompleted),
				})
			case inlineKeyboardTeam:
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.DriverName),
					driverStat.CarClass,
				})
			case inlineKeyboardDriver:
				t.AppendRow([]interface{}{
					helper.GetDriverCodeName(driverStat.DriverName),
					driverStat.DriverName,
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
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardBestLap+" "+symbolTimes, fmt.Sprintf("%s:%s:%s", subcommandShowLiveTiming, serverID, inlineKeyboardBestLap)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardBestLapSectors, fmt.Sprintf("%s:%s:%s", subcommandShowLiveTiming, serverID, inlineKeyboardBestLapSectors)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardLastLap+" "+symbolTimes, fmt.Sprintf("%s:%s:%s", subcommandShowLiveTiming, serverID, inlineKeyboardLastLap)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardLastLapSectors, fmt.Sprintf("%s:%s:%s", subcommandShowLiveTiming, serverID, inlineKeyboardLastLapSectors)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardOptimumLap+" "+symbolTimes, fmt.Sprintf("%s:%s:%s", subcommandShowLiveTiming, serverID, inlineKeyboardOptimumLap)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardOptimumLapSectors, fmt.Sprintf("%s:%s:%s", subcommandShowLiveTiming, serverID, inlineKeyboardOptimumLapSectors)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardStatus, fmt.Sprintf("%s:%s:%s", subcommandShowLiveTiming, serverID, inlineKeyboardStatus)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardInfo, fmt.Sprintf("%s:%s:%s", subcommandShowLiveTiming, serverID, inlineKeyboardInfo)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardDiff, fmt.Sprintf("%s:%s:%s", subcommandShowLiveTiming, serverID, inlineKeyboardDiff)),
		),
	)
}
