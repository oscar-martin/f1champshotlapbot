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
	subcommandShowDrivers = "show_drivers"
	tableLap              = "LAP"
)

type StintApp struct {
	bot                        *tgbotapi.BotAPI
	appMenu                    menus.ApplicationMenu
	serverID                   string
	liveStandingData           servers.LiveStandingHistoryData
	liveStandingDataUpdateChan <-chan string
	caster                     caster.ChannelCaster[servers.LiveStandingHistoryData]
	mu                         sync.Mutex
	menuKeyboard               tgbotapi.ReplyKeyboardMarkup
}

func NewStintApp(bot *tgbotapi.BotAPI, appMenu menus.ApplicationMenu, pubsubMgr *pubsub.PubSub, serverID string) *StintApp {
	menuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(appMenu.ButtonBackTo()),
		),
	)

	sa := &StintApp{
		bot:                        bot,
		appMenu:                    appMenu,
		serverID:                   serverID,
		caster:                     caster.JSONChannelCaster[servers.LiveStandingHistoryData]{},
		liveStandingDataUpdateChan: pubsubMgr.Subscribe(servers.PubSubStintDataPreffix + serverID),
		menuKeyboard:               menuKeyboard,
	}

	go sa.updater()

	return sa
}

func (sa *StintApp) updater() {
	for payload := range sa.liveStandingDataUpdateChan {
		// fmt.Println("Updating StintData")
		lsd, err := sa.caster.From(payload)
		if err != nil {
			fmt.Printf("Error casting StintData: %s\n", err.Error())
			continue
		}
		sa.update(lsd)
	}
}

func (sa *StintApp) update(lsd servers.LiveStandingHistoryData) {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.liveStandingData = lsd
}

func (sa *StintApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	return false, nil
}

func (sa *StintApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery) error) {
	data := strings.Split(query.Data, ":")
	if data[0] == subcommandShowDrivers && data[1] == sa.serverID {
		sa.mu.Lock()
		defer sa.mu.Unlock()
		return true, func(ctx context.Context, query *tgbotapi.CallbackQuery) error {
			return sa.handleStintDataCallbackQuery(query.Message.Chat.ID, &query.Message.MessageID, data[2:]...)
		}
	}
	return false, nil
}

func (sa *StintApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	// fmt.Printf("STINT: button: %s. appName: %s\n", button, buttonStint+" "+sa.stintData.ServerName)
	if button == buttonStint+" "+sa.liveStandingData.ServerName {
		return true, sa.renderDrivers()
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

func (sa *StintApp) renderDrivers() func(ctx context.Context, chatId int64) error {
	return func(ctx context.Context, chatId int64) error {
		if len(sa.liveStandingData.DriverNames) > 0 {
			err := sa.sendDriversData(chatId, nil)
			if err != nil {
				return err
			}
		} else {
			message := "No hay pilotos en la sesión"
			msg := tgbotapi.NewMessage(chatId, message)
			_, err := sa.bot.Send(msg)
			return err
		}
		return nil
	}
}

func (sa *StintApp) handleStintDataCallbackQuery(chatId int64, messageId *int, data ...string) error {
	infoType := data[0]
	driver := data[1]
	driverData, found := sa.liveStandingData.DriversData[driver]
	if found {
		err := sa.sendStintData(chatId, messageId, driverData, driver, sa.liveStandingData.ServerName, sa.liveStandingData.ServerID, infoType)
		if err != nil {
			log.Printf("An error occured: %s", err.Error())
		}
	} else {
		message := fmt.Sprintf("No hay datos para el piloto %s", driver)
		msg := tgbotapi.NewMessage(chatId, message)
		_, err := sa.bot.Send(msg)
		return err
	}
	return nil
}

func (sa *StintApp) sendStintData(chatId int64, messageId *int, driverData []servers.StandingHistoryDriverData, driverName, serverName, serverId, infoType string) error {
	if len(driverData) > 0 {
		var b bytes.Buffer
		t := table.NewWriter()
		t.SetOutputMirror(&b)
		style := table.StyleRounded
		style.Options.DrawBorder = false
		t.SetStyle(style)
		t.AppendSeparator()

		t.AppendHeader(table.Row{tableLap, infoType})
		for idx, lapData := range driverData {
			switch infoType {
			case inlineKeyboardTimes:
				t.AppendRow([]interface{}{
					fmt.Sprintf("%d", idx+1),
					helper.SecondsToMinutes(lapData.LapTime),
				})
			case inlineKeyboardSectors:
				ls1 := lapData.SectorTime1
				ls2 := -1.0
				if ls1 > 0.0 && lapData.SectorTime2 > 0.0 {
					ls2 = lapData.SectorTime2 - ls1
				}
				ls3 := -1.0
				if ls2 > 0.0 && lapData.LapTime > 0.0 {
					ls3 = lapData.LapTime - ls2 - ls1
				}

				t.AppendRow([]interface{}{
					fmt.Sprintf("%d", idx+1),
					fmt.Sprintf("%s %s %s", helper.ToSectorTime(ls1), helper.ToSectorTime(ls2), helper.ToSectorTime(ls3)),
				})
			}
		}
		t.Render()

		keyboard := getStintInlineKeyboard(driverName, serverId)
		var cfg tgbotapi.Chattable
		if messageId == nil {
			msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("```\nDatos de %s en %q\n\n%s```", driverName, serverName, b.String()))
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = keyboard
			cfg = msg
		} else {
			msg := tgbotapi.NewEditMessageText(chatId, *messageId, fmt.Sprintf("```\nDatos de %s en %q\n\n%s```", driverName, serverName, b.String()))
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = &keyboard
			cfg = msg
		}
		_, err := sa.bot.Send(cfg)
		return err
	} else {
		message := "No hay vueltas en la sesión"
		msg := tgbotapi.NewMessage(chatId, message)
		_, err := sa.bot.Send(msg)
		return err
	}
}

func getStintInlineKeyboard(driver, serverID string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardTimes+" "+symbolTimes, fmt.Sprintf("%s:%s:%s:%s", subcommandShowDrivers, serverID, inlineKeyboardTimes, driver)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardSectors+" "+symbolSectors, fmt.Sprintf("%s:%s:%s:%s", subcommandShowDrivers, serverID, inlineKeyboardSectors, driver)),
		),
	)
}

func (sa *StintApp) sendDriversData(chatId int64, messageId *int) error {
	text, keyboard := sa.driversTextMarkup()

	var cfg tgbotapi.Chattable
	if messageId == nil {
		msg := tgbotapi.NewMessage(chatId, text)
		msg.ReplyMarkup = keyboard
		cfg = msg
	} else {
		msg := tgbotapi.NewEditMessageText(chatId, *messageId, text)
		msg.ReplyMarkup = &keyboard
		cfg = msg
	}

	_, err := sa.bot.Send(cfg)
	return err
}

func (sa *StintApp) driversTextMarkup() (text string, markup tgbotapi.InlineKeyboardMarkup) {
	buttons := [][]tgbotapi.InlineKeyboardButton{}

	for idx, driver := range sa.liveStandingData.DriverNames {
		if idx%2 == 0 {
			buttons = append(buttons, []tgbotapi.InlineKeyboardButton{})
		}
		buttons[len(buttons)-1] = append(buttons[len(buttons)-1], tgbotapi.NewInlineKeyboardButtonData(driver, fmt.Sprintf("%s:%s:%s:%s", subcommandShowDrivers, sa.liveStandingData.ServerID, inlineKeyboardTimes, driver)))
	}
	text = "Elige el piloto de la lista:\n\n"
	markup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
	return
}
