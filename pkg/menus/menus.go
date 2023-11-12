package menus

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	buttonBackTo = "Back to"
)

type ApplicationMenu struct {
	Name     string
	From     string
	PrevMenu tgbotapi.ReplyKeyboardMarkup
}

func NewApplicationMenu(name, from string, prevMenu tgbotapi.ReplyKeyboardMarkup) ApplicationMenu {
	return ApplicationMenu{
		Name:     name,
		From:     from,
		PrevMenu: prevMenu,
	}
}

func (am *ApplicationMenu) ButtonBackTo() string {
	return buttonBackTo + " " + am.From
}
