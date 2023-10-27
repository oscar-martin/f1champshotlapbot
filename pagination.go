package main

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendTracksData(chatId int64, currentPage, count, maxPages int, messageId *int, tracks Tracks) error {
	text, keyboard := TracksTextMarkup(currentPage, count, maxPages, tracks)

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

	_, err := bot.Send(cfg)
	return err
}

func TracksTextMarkup(currentPage, count, maxPages int, tracks Tracks) (text string, markup tgbotapi.InlineKeyboardMarkup) {
	ts := tracks.GetRange(currentPage*count, currentPage*count+count)
	var trackNames []string
	for _, track := range ts {
		trackNames = append(trackNames, track.String())
	}
	text = strings.Join(trackNames, "\n")

	var rows []tgbotapi.InlineKeyboardButton
	if currentPage > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardButtonData("Anterior", fmt.Sprintf("pager:prev:%d:%d", currentPage, count)))
	}
	if currentPage < maxPages-1 {
		rows = append(rows, tgbotapi.NewInlineKeyboardButtonData("Siguiente", fmt.Sprintf("pager:next:%d:%d", currentPage, count)))
	}

	markup = tgbotapi.NewInlineKeyboardMarkup(rows)
	return
}

func HandleNavigationCallbackQuery(chatId int64, messageId, maxPages int, tracks Tracks, data ...string) {
	pagerType := data[0]
	currentPage, _ := strconv.Atoi(data[1])
	itemsPerPage, _ := strconv.Atoi(data[2])

	if pagerType == "next" {
		nextPage := currentPage + 1
		if nextPage < maxPages {
			_ = SendTracksData(chatId, nextPage, itemsPerPage, maxPages, &messageId, tracks)
		}
	}
	if pagerType == "prev" {
		previousPage := currentPage - 1
		if previousPage >= 0 {
			_ = SendTracksData(chatId, previousPage, itemsPerPage, maxPages, &messageId, tracks)
		}
	}
}
