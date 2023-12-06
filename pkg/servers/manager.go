package servers

import (
	"context"
	"f1champshotlapsbot/pkg/model"
	"f1champshotlapsbot/pkg/pubsub"
	"f1champshotlapsbot/pkg/thumbnails"
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
	ctx               context.Context
	servers           []Server
	bot               *tgbotapi.BotAPI
	newSessionChannel chan<- ServerStarted
}

func NewManager(ctx context.Context, bot *tgbotapi.BotAPI, servers []Server, newSessionChannel chan<- ServerStarted) (*Manager, error) {
	m := &Manager{
		ctx:               ctx,
		bot:               bot,
		servers:           servers,
		newSessionChannel: newSessionChannel,
	}

	err := m.initializeServers()
	return m, err
}

func (sm *Manager) Sync(ticker *time.Ticker, exitChan chan bool) {
	sm.doSync(time.Now())
	for {
		select {
		case <-exitChan:
			return
		case t := <-ticker.C:
			sm.doSync(t)
		}
	}
}

func (sm *Manager) doSync(t time.Time) {
	sm.checkServersOnline()
}

func (sm *Manager) initializeServers() error {
	// set up the goroutine to publish live data
	for i := range sm.servers {
		sm.servers[i].Name = sm.servers[i].ID
		sm.servers[i].BestSectorsForDriver = make(map[string]Sectors)
		sm.servers[i].BestLapForDriver = make(map[string]int)
		sm.servers[i].TopSpeedForDriver = make(map[string]map[int]float64)
		sm.servers[i].LiveSessionInfoDataChan = make(chan model.LiveSessionInfoData)
		sm.servers[i].LiveStandingChan = make(chan model.LiveStandingData)
		sm.servers[i].LiveStandingHistoryChan = make(chan model.LiveStandingHistoryData)
		sm.servers[i].ThumbnailChan = make(chan thumbnails.Thumbnail)

		go func(idx int) {
			for {
				select {
				case liveSessionInfo := <-sm.servers[idx].LiveSessionInfoDataChan:
					pubsub.LiveSessionInfoDataPubSub.Publish(PubSubSessionInfoPreffix+sm.servers[idx].ID, liveSessionInfo)
				case liveTiming := <-sm.servers[idx].LiveStandingChan:
					pubsub.LiveStandingDataPubSub.Publish(PubSubDriversSessionPreffix+sm.servers[idx].ID, liveTiming)
				case liveStanding := <-sm.servers[idx].LiveStandingHistoryChan:
					pubsub.LiveStandingHistoryPubSub.Publish(PubSubStintDataPreffix+sm.servers[idx].ID, liveStanding)
				case thumbnail := <-sm.servers[idx].ThumbnailChan:
					pubsub.TrackThumbnailPubSub.Publish(thumbnails.PubSubThumbnailPreffix+sm.servers[idx].ID, thumbnail)
				}
			}
		}(i)
		// go func(idx int) {
		// 	for liveSessionInfo := range sm.servers[idx].LiveSessionInfoDataChan {
		// 		pubsub.LiveSessionInfoDataPubSub.Publish(PubSubSessionInfoPreffix+sm.servers[idx].ID, liveSessionInfo)
		// 	}
		// }(i)
		// go func(idx int) {
		// 	for liveTiming := range sm.servers[idx].LiveStandingChan {
		// 		pubsub.LiveStandingDataPubSub.Publish(PubSubDriversSessionPreffix+sm.servers[idx].ID, liveTiming)
		// 	}
		// }(i)
		// go func(idx int) {
		// 	for liveStanding := range sm.servers[idx].LiveStandingHistoryChan {
		// 		pubsub.LiveStandingHistoryPubSub.Publish(PubSubStintDataPreffix+sm.servers[idx].ID, liveStanding)
		// 	}
		// }(i)
		// go func(idx int) {
		// 	for thumbnail := range sm.servers[idx].ThumbnailChan {
		// 		pubsub.TrackThumbnailPubSub.Publish(thumbnails.PubSubThumbnailPreffix+sm.servers[idx].ID, thumbnail)
		// 	}
		// }(i)
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
				go func() {
					// fmt.Printf("Starting websocket reader for server %s\n", sm.servers[idx].ID)
					err := sm.servers[idx].WebSocketReader(sm.ctx, sm.newSessionChannel)
					if err != nil {
						log.Printf("Error reading websocket: %s", err.Error())
					}
				}()
			}
		}(i)
	}
	wg.Wait()
}
