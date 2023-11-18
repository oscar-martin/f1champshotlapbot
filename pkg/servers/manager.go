package servers

import (
	"context"
	"f1champshotlapsbot/pkg/caster"
	"f1champshotlapsbot/pkg/pubsub"
	"fmt"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	MenuLive   = "/directo"
	ButtonLive = "Live"
)

type Manager struct {
	ctx       context.Context
	mu        sync.Mutex
	apiDomain string
	// servers   []Server
	bot                  *tgbotapi.BotAPI
	pubsubMgr            *pubsub.PubSub
	serversCaster        caster.ChannelCaster[[]Server]
	sessionInfoCaster    caster.ChannelCaster[SessionInfo]
	driversSessionCaster caster.ChannelCaster[DriversSession]
}

func NewManager(ctx context.Context, bot *tgbotapi.BotAPI, domain string, pubsubMgr *pubsub.PubSub) *Manager {
	return &Manager{
		ctx:                  ctx,
		apiDomain:            domain,
		bot:                  bot,
		serversCaster:        caster.JSONChannelCaster[[]Server]{},
		sessionInfoCaster:    caster.JSONChannelCaster[SessionInfo]{},
		driversSessionCaster: caster.JSONChannelCaster[DriversSession]{},
		pubsubMgr:            pubsubMgr,
	}
}

func (sm *Manager) Lock() {
	sm.mu.Lock()
}

func (sm *Manager) Unlock() {
	sm.mu.Unlock()
}

func (sm *Manager) Sync(ticker *time.Ticker, exitChan chan bool) {
	sm.doSync(time.Now())
	go func() {
		for {
			select {
			case <-exitChan:
				return
			case t := <-ticker.C:
				sm.doSync(t)
			}
		}
	}()
}

func (sm *Manager) doSync(t time.Time) {
	fmt.Println("Refreshing servers statuses: ", t)
	ss := sm.checkServersOnline()
	payload, err := sm.serversCaster.To(ss)
	if err != nil {
		log.Printf("Error casting servers to json: %s", err.Error())
	} else {
		sm.pubsubMgr.Publish(PubSubServersTopic, payload)
	}
}

func (sm *Manager) GetInitialServers() ([]Server, error) {
	return getServers(sm.ctx, sm.apiDomain)
}

// func (sm *Manager) GetServers() ([]Server, error) {
// 	if len(sm.servers) == 0 {
// 		// if there is no servers, fetch them
// 		ss, err := getServers(sm.ctx, sm.apiDomain)
// 		if err != nil {
// 			return ss, err
// 		}
// 		sm.servers = ss
// 	}

// 	return sm.servers, nil
// }

// func (sm *Manager) GetServerById(id string) (Server, bool) {
// 	for _, s := range sm.servers {
// 		if s.ID == id {
// 			return s, true
// 		}
// 	}
// 	return Server{}, false
// }

func (sm *Manager) checkServersOnline() []Server {
	wg := sync.WaitGroup{}
	ss, _ := sm.GetInitialServers()
	for i := range ss {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// get session info
			sessionInfo, err := ss[idx].GetSessionInfo(sm.ctx)
			sessionInfo.ServerID = ss[idx].ID
			if err != nil {
				sessionInfo.Online = false
				log.Printf("Error checking server %s: %s", ss[idx].Name, err.Error())
			} else {
				ss[idx].Online = true
				ss[idx].Name = sessionInfo.ServerName
			}
			sessionInfo.Online = ss[idx].Online
			payload, err := sm.sessionInfoCaster.To(sessionInfo)
			if err != nil {
				log.Printf("Error casting servers to json: %s", err.Error())
			} else {
				sm.pubsubMgr.Publish(ss[idx].ID, payload)
			}
			// get driver sessions
			if sessionInfo.Online {
				dss, err := ss[idx].GetDriverSessions(sm.ctx)
				if err != nil {
					log.Printf("Error getting driver sessions: %s", err.Error())
				}
				payload, err := sm.driversSessionCaster.To(dss)
				if err != nil {
					log.Printf("Error casting servers to json: %s", err.Error())
				} else {
					sm.pubsubMgr.Publish(PubSubDriversSessionPreffix+ss[idx].ID, payload)
				}
			}
		}(i)
	}
	wg.Wait()
	return ss
}
