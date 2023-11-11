package tracks

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	symbolInit = "⏮"
	symbolPrev = "◀️"
	symbolNext = "▶️"
	symbolEnd  = "⏭"
)

func SendTracksData(chatId int64, currentPage, count, maxPages int, messageId *int, tm *Manager) error {
	text, keyboard := TracksTextMarkup(currentPage, count, maxPages, tm)

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

	_, err := tm.bot.Send(cfg)
	return err
}

func TracksTextMarkup(currentPage, count, maxPages int, tm *Manager) (text string, markup tgbotapi.InlineKeyboardMarkup) {
	ts := tm.GetRange(currentPage*count, currentPage*count+count)
	var trackNames []string
	for _, track := range ts {
		trackNames = append(trackNames, track.CommandString())
	}
	text = fmt.Sprintf("Elige el circuito de la lista (%d/%d):\n\n", currentPage+1, maxPages)
	text += strings.Join(trackNames, "\n")

	var rows []tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardButtonData(symbolInit, fmt.Sprintf("%s:init:%d:%d", subcommandShowTracks, currentPage, count)))
	rows = append(rows, tgbotapi.NewInlineKeyboardButtonData(symbolPrev, fmt.Sprintf("%s:prev:%d:%d", subcommandShowTracks, currentPage, count)))
	rows = append(rows, tgbotapi.NewInlineKeyboardButtonData(symbolNext, fmt.Sprintf("%s:next:%d:%d", subcommandShowTracks, currentPage, count)))
	rows = append(rows, tgbotapi.NewInlineKeyboardButtonData(symbolEnd, fmt.Sprintf("%s:end:%d:%d", subcommandShowTracks, currentPage, count)))

	markup = tgbotapi.NewInlineKeyboardMarkup(rows)
	return
}

func HandleTrackDataCallbackQuery(chatId int64, messageId, maxPages int, tm *Manager, data ...string) {
	pagerType := data[0]
	currentPage, _ := strconv.Atoi(data[1])
	itemsPerPage, _ := strconv.Atoi(data[2])

	if pagerType == "next" {
		nextPage := currentPage + 1
		if nextPage < maxPages {
			_ = SendTracksData(chatId, nextPage, itemsPerPage, maxPages, &messageId, tm)
		}
	}
	if pagerType == "prev" {
		previousPage := currentPage - 1
		if previousPage >= 0 {
			_ = SendTracksData(chatId, previousPage, itemsPerPage, maxPages, &messageId, tm)
		}
	}
	if pagerType == "init" {
		_ = SendTracksData(chatId, 0, itemsPerPage, maxPages, &messageId, tm)
	}
	if pagerType == "end" {
		_ = SendTracksData(chatId, maxPages-1, itemsPerPage, maxPages, &messageId, tm)
	}
}
