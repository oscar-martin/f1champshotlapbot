package hotlaps

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Hotlaps struct {
	bot       *tgbotapi.BotAPI
	apiDomain string
}

func NewHotlaps(bot *tgbotapi.BotAPI, domain string) *Hotlaps {
	return &Hotlaps{
		apiDomain: domain,
		bot:       bot,
	}
}

func (hl *Hotlaps) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	return false, nil
}

func (hl *Hotlaps) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery)) {
	return false, nil
}

func (hl *Hotlaps) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	return false, nil
}
