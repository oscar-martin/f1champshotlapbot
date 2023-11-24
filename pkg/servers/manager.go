package servers

import (
	"context"
	"f1champshotlapsbot/pkg/caster"
	"f1champshotlapsbot/pkg/pubsub"
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
	ctx                           context.Context
	servers                       []Server
	bot                           *tgbotapi.BotAPI
	pubsubMgr                     *pubsub.PubSub
	liveSessionInfoDataCaster     caster.ChannelCaster[LiveSessionInfoData]
	liveStandingDataCaster        caster.ChannelCaster[LiveStandingData]
	liveStandingHistoryDataCaster caster.ChannelCaster[LiveStandingHistoryData]
}

func NewManager(ctx context.Context, bot *tgbotapi.BotAPI, servers []Server, pubsubMgr *pubsub.PubSub) (*Manager, error) {
	m := &Manager{
		ctx:                           ctx,
		bot:                           bot,
		servers:                       servers,
		liveSessionInfoDataCaster:     caster.JSONChannelCaster[LiveSessionInfoData]{},
		liveStandingDataCaster:        caster.JSONChannelCaster[LiveStandingData]{},
		liveStandingHistoryDataCaster: caster.JSONChannelCaster[LiveStandingHistoryData]{},
		pubsubMgr:                     pubsubMgr,
	}

	err := m.initializeServers()
	return m, err
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
	sm.checkServersOnline()
}

func (sm *Manager) initializeServers() error {
	// set up the goroutine to publish live data
	for i := range sm.servers {
		sm.servers[i].Name = sm.servers[i].ID
		sm.servers[i].BestSector3ForDriver = make(map[string]float64)
		sm.servers[i].LiveSessionInfoDataChan = make(chan LiveSessionInfoData)
		sm.servers[i].LiveStandingChan = make(chan LiveStandingData)
		sm.servers[i].LiveStandingHistoryChan = make(chan LiveStandingHistoryData)

		go func(idx int) {
			for liveSessionInfo := range sm.servers[idx].LiveSessionInfoDataChan {
				payload, err := sm.liveSessionInfoDataCaster.To(liveSessionInfo)
				if err != nil {
					log.Printf("Error casting session info to json: %s", err.Error())
				} else {
					sm.pubsubMgr.Publish(PubSubSessionInfoPreffix+sm.servers[idx].ID, payload)
				}
			}
		}(i)
		go func(idx int) {
			for liveTiming := range sm.servers[idx].LiveStandingChan {
				payload, err := sm.liveStandingDataCaster.To(liveTiming)
				if err != nil {
					log.Printf("Error casting live standing to json: %s", err.Error())
				} else {
					sm.pubsubMgr.Publish(PubSubDriversSessionPreffix+sm.servers[idx].ID, payload)
				}
			}
		}(i)
		go func(idx int) {
			for liveStanding := range sm.servers[idx].LiveStandingHistoryChan {
				payload, err := sm.liveStandingHistoryDataCaster.To(liveStanding)
				if err != nil {
					log.Printf("Error casting live standing history to json: %s", err.Error())
				} else {
					sm.pubsubMgr.Publish(PubSubStintDataPreffix+sm.servers[idx].ID, payload)
				}
			}
		}(i)
	}

	return nil
}

func (sm *Manager) checkServersOnline() {
	wg := sync.WaitGroup{}
	for i := range sm.servers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if !sm.servers[idx].WebSocketRunning {
				// set up the ws client
				go func() { _ = sm.servers[idx].WebSocketReader(sm.ctx) }()
			}
		}(i)
	}
	wg.Wait()
}
