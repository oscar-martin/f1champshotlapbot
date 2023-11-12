package servers

import (
	"context"
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	MenuLive   = "/directo"
	ButtonLive = "Live"
)

type Manager struct {
	mu        sync.Mutex
	apiDomain string
	servers   []Server
	bot       *tgbotapi.BotAPI
}

func NewManager(bot *tgbotapi.BotAPI, domain string) *Manager {
	return &Manager{
		apiDomain: domain,
		bot:       bot,
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
