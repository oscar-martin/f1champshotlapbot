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
	"sort"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	SubcommandShowDrivers = "show_drivers"
	tableLap              = "LAP"
)

type StintApp struct {
	bot                 *tgbotapi.BotAPI
	appMenu             menus.ApplicationMenu
	serverID            string
	stintData           servers.StintData
	stintDataUpdateChan <-chan string
	caster              caster.ChannelCaster[servers.StintData]
	mu                  sync.Mutex

	menuKeyboard tgbotapi.ReplyKeyboardMarkup
}

func NewStintApp(bot *tgbotapi.BotAPI, appMenu menus.ApplicationMenu, pubsubMgr *pubsub.PubSub, serverID string) *StintApp {
	menuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(appMenu.ButtonBackTo()),
		),
	)

	sa := &StintApp{
		bot:                 bot,
		appMenu:             appMenu,
		serverID:            serverID,
		caster:              caster.JSONChannelCaster[servers.StintData]{},
		stintDataUpdateChan: pubsubMgr.Subscribe(servers.PubSubStintDataPreffix + serverID),
		menuKeyboard:        menuKeyboard,
	}

	go sa.updater()

	return sa
}

func (sa *StintApp) updater() {
	for payload := range sa.stintDataUpdateChan {
		// fmt.Println("Updating StintData")
		dss, err := sa.caster.From(payload)
		if err != nil {
			fmt.Printf("Error casting StintData: %s\n", err.Error())
			continue
		}
		sa.mu.Lock()
		sa.update(dss)
		sa.mu.Unlock()
	}
}

func (sa *StintApp) update(sd servers.StintData) {
	sa.stintData = sd
}

func (sa *StintApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	return false, nil
}

func (sa *StintApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery) error) {
	data := strings.Split(query.Data, ":")
	// fmt.Printf("STINT: callback: %s. appName: %s\n", query.Data, SubcommandShowDrivers)
	if data[0] == SubcommandShowDrivers && data[2] == sa.serverID {
		sa.mu.Lock()
		defer sa.mu.Unlock()
		return true, func(ctx context.Context, query *tgbotapi.CallbackQuery) error {
			return sa.handleStintDataCallbackQuery(query.Message.Chat.ID, &query.Message.MessageID, data[1:]...)
		}
	}
	return false, nil
}

func (sa *StintApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	// fmt.Printf("STINT: button: %s. appName: %s\n", button, buttonStint+" "+sa.stintData.ServerName)
	if button == buttonStint+" "+sa.stintData.ServerName {
		return true, func(ctx context.Context, chatId int64) error {
			if len(sa.stintData.Drivers) > 0 {
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
	} else if button == sa.appMenu.ButtonBackTo() {
		return true, func(ctx context.Context, chatId int64) error {
			msg := tgbotapi.NewMessage(chatId, "OK")
			msg.ReplyMarkup = sa.appMenu.PrevMenu()
			_, err := sa.bot.Send(msg)
			return err
		}
	} else if strings.HasPrefix(button, fmt.Sprintf("%s:%s", SubcommandShowDrivers, sa.stintData.ServerID)) {
		driver := button[strings.Index(button, ":")+1:]
		return true, sa.renderStint(driver)
	}
	// fmt.Print("STINT: FALSE\n")
	return false, nil
}

func (sa *StintApp) renderStint(driver string) func(ctx context.Context, chatId int64) error {
	return func(ctx context.Context, chatId int64) error {
		driverStint, found := sa.stintData.Drivers[driver]
		if found {
			err := sa.sendStintData(chatId, nil, driverStint, sa.stintData.ServerName, sa.stintData.ServerID, inlineKeyboardTimes)
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
}

func (sa *StintApp) handleStintDataCallbackQuery(chatId int64, messageId *int, data ...string) error {
	infoType := data[0]
	driver := data[2]
	driverStint, found := sa.stintData.Drivers[driver]
	if found {
		err := sa.sendStintData(chatId, messageId, driverStint, sa.stintData.ServerName, sa.stintData.ServerID, infoType)
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

func (sa *StintApp) sendStintData(chatId int64, messageId *int, driverStint servers.DriverStint, serverName, serverId, infoType string) error {
	if len(driverStint.Laps) > 0 {
		var b bytes.Buffer
		t := table.NewWriter()
		t.SetOutputMirror(&b)
		t.SetStyle(table.StyleRounded)
		t.AppendSeparator()

		t.AppendHeader(table.Row{tableLap, infoType})
		for idx, lapData := range driverStint.Laps {
			switch infoType {
			case inlineKeyboardTimes:
				t.AppendRow([]interface{}{
					fmt.Sprintf("%d", idx+1),
					helper.SecondsToMinutes(lapData.LapTime),
				})
			case inlineKeyboardSectors:
				t.AppendRow([]interface{}{
					fmt.Sprintf("%d", idx+1),
					fmt.Sprintf("%s %s %s", helper.ToSectorTime(lapData.S1), helper.ToSectorTime(lapData.S2), helper.ToSectorTime(lapData.S3)),
				})
			case inlineKeyboardMaxSpeed:
				t.AppendRow([]interface{}{
					fmt.Sprintf("%d", idx+1),
					fmt.Sprintf("%.1f", lapData.MaxSpeed),
				})
			case inlineKeyboardDiff:
				t.AppendRow([]interface{}{
					fmt.Sprintf("%d", idx+1),
					fmt.Sprintf("%.1fs", lapData.Diff),
				})
			}
		}
		t.Render()

		keyboard := getStintInlineKeyboard(driverStint.Driver, serverId)
		var cfg tgbotapi.Chattable
		if messageId == nil {
			msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("```\nDatos de %s en %q\n\n%s```", driverStint.Driver, serverName, b.String()))
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = keyboard
			cfg = msg
		} else {
			msg := tgbotapi.NewEditMessageText(chatId, *messageId, fmt.Sprintf("```\nDatos de %s en %q\n\n%s```", driverStint.Driver, serverName, b.String()))
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
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardTimes+" "+symbolTimes, fmt.Sprintf("%s:%s:%s:%s", SubcommandShowDrivers, inlineKeyboardTimes, serverID, driver)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardDiff+" "+symbolDiff, fmt.Sprintf("%s:%s:%s:%s", SubcommandShowDrivers, inlineKeyboardDiff, serverID, driver)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardSectors+" "+symbolSectors, fmt.Sprintf("%s:%s:%s:%s", SubcommandShowDrivers, inlineKeyboardSectors, serverID, driver)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardMaxSpeed+" "+symbolMaxSpeed, fmt.Sprintf("%s:%s:%s:%s", SubcommandShowDrivers, inlineKeyboardMaxSpeed, serverID, driver)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardUpdate+" "+symbolUpdate, fmt.Sprintf("%s:%s:%s:%s", SubcommandShowDrivers, inlineKeyboardSectors, serverID, driver)),
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
	drivers := make([]string, 0, len(sa.stintData.Drivers))

	for d := range sa.stintData.Drivers {
		drivers = append(drivers, d)
	}
	sort.Strings(drivers)
	buttons := [][]tgbotapi.InlineKeyboardButton{}

	for idx, driver := range drivers {
		if idx%2 == 0 {
			buttons = append(buttons, []tgbotapi.InlineKeyboardButton{})
		}
		buttons[len(buttons)-1] = append(buttons[len(buttons)-1], tgbotapi.NewInlineKeyboardButtonData(driver, fmt.Sprintf("%s:%s:%s:%s", SubcommandShowDrivers, inlineKeyboardTimes, sa.stintData.ServerID, driver)))
	}
	text = "Elige el piloto de la lista:\n\n"
	markup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
	return
}
