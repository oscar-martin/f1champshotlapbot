package main

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jedib0t/go-pretty/v6/table"
)

func HandleSessionDataCallbackQuery(chatId int64, messageId *int, data ...string) {
	infoType := data[0]
	trackId := data[1]
	categoryId := data[2]
	err := SendSessionData(chatId, messageId, trackId, categoryId, infoType)
	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

func SendSessionData(chatId int64, messageId *int, trackId, categoryId, infoType string) error {
	track, found := tracks.GetTrackByID(trackId)
	if !found {
		message := fmt.Sprintf("El circuito seleccionado no se ha encontrado. Vuelve a probar %s", menuTracks)
		msg := tgbotapi.NewMessage(chatId, message)
		_, err := bot.Send(msg)
		return err
	}
	sessions, ok := trackSessions[track.ID]
	if !ok {
		message := fmt.Sprintf("No se han encontrado la sesiones para el circuito. Vuelve a probar %s", menuTracks)
		msg := tgbotapi.NewMessage(chatId, message)
		_, err := bot.Send(msg)
		return err
	}
	sessionsForCategory := sessions.GetSessionsByCategoryID(categoryId)

	if len(sessionsForCategory) > 0 {
		sort.Slice(sessionsForCategory, func(i, j int) bool {
			return sessionsForCategory[i].Time < sessionsForCategory[j].Time
		})

		// read the category name from the first session
		_, category := extractCategory(sessionsForCategory[0].Category)

		var b bytes.Buffer
		t := table.NewWriter()
		t.SetOutputMirror(&b)
		t.SetStyle(table.StyleRounded)
		t.AppendSeparator()

		t.AppendHeader(table.Row{tableDriver, infoType})
		for _, session := range sessionsForCategory {
			switch infoType {
			case inlineKeyboardTimes:
				t.AppendRow([]interface{}{
					getDriverCodeName(session.Driver),
					secondsToMinutes(session.Time),
				})
			case inlineKeyboardSectors:
				t.AppendRow([]interface{}{
					getDriverCodeName(session.Driver),
					fmt.Sprintf("%s %s %s", toSectorTime(session.S1), toSectorTime(session.S2), toSectorTime(session.S3)),
				})
			case inlineKeyboardCompound:
				tyreSlice := strings.Split(session.Fcompound, ",")
				tyre := "(desconocido)"
				if len(tyreSlice) > 0 {
					tyre = tyreSlice[len(tyreSlice)-1]
				}
				t.AppendRow([]interface{}{
					getDriverCodeName(session.Driver),
					tyre,
				})
			case inlineKeyboardLaps:
				t.AppendRow([]interface{}{
					getDriverCodeName(session.Driver),
					fmt.Sprintf("%d/%d", session.Lapcountcomplete, session.Lapcount),
				})
			case inlineKeyboardTeam:
				t.AppendRow([]interface{}{
					getDriverCodeName(session.Driver),
					session.CarClass,
				})
			case inlineKeyboardDriver:
				t.AppendRow([]interface{}{
					getDriverCodeName(session.Driver),
					session.Driver,
				})
			case inlineKeyboardDate:
				t.AppendRow([]interface{}{
					getDriverCodeName(session.Driver),
					session.DateTime,
				})
			}
		}
		t.Render()

		keyboard := getInlineKeyboardForCategory(track.ID, categoryId)
		var cfg tgbotapi.Chattable
		if messageId == nil {
			msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("```\nResultados en %q para %q\n\n%s```", track.Name, category, b.String()))
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = keyboard
			cfg = msg
		} else {
			msg := tgbotapi.NewEditMessageText(chatId, *messageId, fmt.Sprintf("```\nResultados en %q para %q\n\n%s```", track.Name, category, b.String()))
			msg.ParseMode = tgbotapi.ModeMarkdownV2
			msg.ReplyMarkup = &keyboard
			cfg = msg
		}
		_, err := bot.Send(cfg)
		return err
	} else {
		message := "No hay sesiones registradas"
		msg := tgbotapi.NewMessage(chatId, message)
		_, err := bot.Send(msg)
		return err
	}
}
func getInlineKeyboardForCategory(trackId, categoryId string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardTimes+" "+symbolTimes, fmt.Sprintf("%s:%s:%s:%s", subcommandShowSessionData, inlineKeyboardTimes, trackId, categoryId)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardSectors+" "+symbolSectors, fmt.Sprintf("%s:%s:%s:%s", subcommandShowSessionData, inlineKeyboardSectors, trackId, categoryId)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardCompound+" "+symbolTimes, fmt.Sprintf("%s:%s:%s:%s", subcommandShowSessionData, inlineKeyboardCompound, trackId, categoryId)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardLaps+" "+symbolLaps, fmt.Sprintf("%s:%s:%s:%s", subcommandShowSessionData, inlineKeyboardLaps, trackId, categoryId)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardTeam+" "+symbolTeam, fmt.Sprintf("%s:%s:%s:%s", subcommandShowSessionData, inlineKeyboardTeam, trackId, categoryId)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardDriver+" "+symbolDriver, fmt.Sprintf("%s:%s:%s:%s", subcommandShowSessionData, inlineKeyboardDriver, trackId, categoryId)),
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardDate+" "+symbolDate, fmt.Sprintf("%s:%s:%s:%s", subcommandShowSessionData, inlineKeyboardDate, trackId, categoryId)),
		),
	)
}
