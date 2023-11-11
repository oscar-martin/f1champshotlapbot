package servers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendServersData(chatId int64, commandPrefix string, servers []Server, bot *tgbotapi.BotAPI) error {
	message := "Servidores disponibles:\n\n"
	for _, server := range servers {
		message += server.CommandString(commandPrefix) + "\n"
	}
	msg := tgbotapi.NewMessage(chatId, message)

	_, err := bot.Send(msg)
	return err
}
