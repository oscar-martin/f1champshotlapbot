package servers

import (
	"context"
	"f1champshotlapsbot/pkg/menus"
	"fmt"
	"regexp"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	serverPrefixCommand = "/server"
	MenuLive            = "/directo"
	ButtonLive          = "Live"
	menuKeyboard        tgbotapi.ReplyKeyboardMarkup
)

type Manager struct {
	mu        sync.Mutex
	apiDomain string
	servers   []Server
	bot       *tgbotapi.BotAPI
	appMenu   menus.ApplicationMenu
}

func NewManager(bot *tgbotapi.BotAPI, domain string, appMenu menus.ApplicationMenu) *Manager {
	menuKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Server1"),
			tgbotapi.NewKeyboardButton("Server2"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(appMenu.ButtonBackTo()),
		),
	)
	return &Manager{
		apiDomain: domain,
		bot:       bot,
		appMenu:   appMenu,
	}
}

func (sm *Manager) Lock() {
	sm.mu.Lock()
}

func (sm *Manager) Unlock() {
	sm.mu.Unlock()
}

func (sm *Manager) Sync(ctx context.Context, ticker *time.Ticker, exitChan chan bool) {
	go func() {
		for {
			select {
			case <-exitChan:
				return
			case t := <-ticker.C:
				fmt.Println("Refreshing servers statuses: ", t)
				sm.mu.Lock()
				sm.servers = []Server{}
				sm.mu.Unlock()
			}
		}
	}()
}

func (sm *Manager) GetServers(ctx context.Context) ([]Server, error) {
	if len(sm.servers) == 0 {
		// if there is no servers, fetch them
		ss, err := getServers(ctx, sm.apiDomain)
		if err != nil {
			return ss, err
		}
		sm.servers = ss
	}

	return sm.servers, nil
}

func (sm *Manager) GetServerById(id string) (Server, bool) {
	for _, s := range sm.servers {
		if s.ID == id {
			return s, true
		}
	}
	return Server{}, false
}

func (sm *Manager) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	commandServerId := regexp.MustCompile(fmt.Sprintf(`^\%s_(\d+)$`, serverPrefixCommand))
	if commandServerId.MatchString(command) {
		serverId := commandServerId.FindStringSubmatch(command)[1]
		return true, sm.renderServerId(serverId)
	} else if command == MenuLive {
		return true, sm.renderServers()
	}
	return false, nil
}

func (sm *Manager) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery)) {
	return false, nil
}

func (sm *Manager) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	if button == sm.appMenu.Name {
		// return sm.AcceptCommand(MenuLive)
		return true, func(ctx context.Context, chatId int64) error {
			message := fmt.Sprintf("%s application\n\n", sm.appMenu.Name)
			msg := tgbotapi.NewMessage(chatId, message)
			msg.ReplyMarkup = menuKeyboard
			_, err := sm.bot.Send(msg)
			return err
		}
	} else if button == sm.appMenu.ButtonBackTo() {
		return true, func(ctx context.Context, chatId int64) error {
			msg := tgbotapi.NewMessage(chatId, "OK")
			msg.ReplyMarkup = sm.appMenu.PrevMenu
			_, err := sm.bot.Send(msg)
			return err
		}
	}
	return false, nil
}
