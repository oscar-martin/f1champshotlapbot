package menus

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	buttonBackTo = "Volver a"
)

type Menuer interface {
	Menu() tgbotapi.ReplyKeyboardMarkup
}

type ApplicationMenu struct {
	Name       string
	From       string
	prevMenuer Menuer
}

func NewApplicationMenu(name, from string, prevMenuer Menuer) ApplicationMenu {
	return ApplicationMenu{
		Name:       name,
		From:       from,
		prevMenuer: prevMenuer,
	}
}

func (am *ApplicationMenu) ButtonBackTo() string {
	return buttonBackTo + " " + am.From
}

func (am *ApplicationMenu) PrevMenu() tgbotapi.ReplyKeyboardMarkup {
	return am.prevMenuer.Menu()
}
