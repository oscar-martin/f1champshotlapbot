package apps

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Accepter interface {
	AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error)
	AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error)
	AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery) error)
}
