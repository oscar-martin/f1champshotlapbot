package servers

import (
	"context"
	"f1champshotlapsbot/pkg/livemap"
	"f1champshotlapsbot/pkg/model"
	"f1champshotlapsbot/pkg/pubsub"
	"f1champshotlapsbot/pkg/resources"
	"f1champshotlapsbot/pkg/webserver"
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
	ctx     context.Context
	servers []Server
	bot     *tgbotapi.BotAPI
}

func NewManager(ctx context.Context, bot *tgbotapi.BotAPI, servers []Server, ws *webserver.Manager) (*Manager, error) {
	m := &Manager{
		ctx:     ctx,
		bot:     bot,
		servers: servers,
	}

	err := m.initializeServers(ws)
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

func (sm *Manager) initializeServers(ws *webserver.Manager) error {
	// set up the goroutine to publish live data
	for i := range sm.servers {
		sm.servers[i].Name = sm.servers[i].ID
		sm.servers[i].BestSectorsForDriver = make(map[string]Sectors)
		sm.servers[i].BestLapForDriver = make(map[string]int)
		sm.servers[i].TopSpeedForDriver = make(map[string]map[int]float64)
		sm.servers[i].LiveSessionInfoDataChan = make(chan model.LiveSessionInfoData)
		sm.servers[i].LiveStandingChan = make(chan model.LiveStandingData)
		sm.servers[i].LiveStandingHistoryChan = make(chan model.LiveStandingHistoryData)
		sm.servers[i].ThumbnailChan = make(chan resources.Resource)
		sm.servers[i].ServerStartedChan = make(chan model.ServerStarted)
		sm.servers[i].ServerStoppedChan = make(chan string)
		sm.servers[i].FirstDriverEnteredChan = make(chan model.ServerStarted)
		sm.servers[i].SelectedSessionDataChan = make(chan model.SelectedSessionData)
		sm.servers[i].CarsPositionChan = make(chan []model.CarPosition)
		sm.servers[i].LiveMapPath = fmt.Sprintf("/servers/%d", i)
		sm.servers[i].LiveMap = livemap.NewLiveMap(ws.GetRouter(sm.servers[i].ID, sm.servers[i].LiveMapPath), sm.servers[i].ID, sm.servers[i].LiveMapPath)

		go func(idx int) {
			for liveSessionInfo := range sm.servers[idx].LiveSessionInfoDataChan {
				pubsub.LiveSessionInfoDataPubSub.Publish(pubsub.PubSubSessionInfoPreffix+sm.servers[idx].ID, liveSessionInfo)
			}
		}(i)

		go func(idx int) {
			for liveTiming := range sm.servers[idx].LiveStandingChan {
				pubsub.LiveStandingDataPubSub.Publish(pubsub.PubSubDriversSessionPreffix+sm.servers[idx].ID, liveTiming)
			}
		}(i)

		go func(idx int) {
			for liveStanding := range sm.servers[idx].LiveStandingHistoryChan {
				pubsub.LiveStandingHistoryPubSub.Publish(pubsub.PubSubStintDataPreffix+sm.servers[idx].ID, liveStanding)
			}
		}(i)

		go func(idx int) {
			for thumbnail := range sm.servers[idx].ThumbnailChan {
				pubsub.TrackThumbnailPubSub.Publish(pubsub.PubSubThumbnailPreffix+sm.servers[idx].ID, thumbnail)
			}
		}(i)

		go func(idx int) {
			for serverStarted := range sm.servers[idx].ServerStartedChan {
				pubsub.SessionStartedPubSub.Publish(pubsub.PubSubSessionStartedPreffix, serverStarted)
			}
		}(i)

		go func(idx int) {
			for serverStopped := range sm.servers[idx].ServerStoppedChan {
				pubsub.SessionStoppedPubSub.Publish(pubsub.PubSubSessionStoppedPreffix, serverStopped)
			}
		}(i)

		go func(idx int) {
			for firstDriverEnteredInSession := range sm.servers[idx].FirstDriverEnteredChan {
				pubsub.FirstDriverEnteredPubSub.Publish(pubsub.PubSubFirstDriverEnteredPreffix, firstDriverEnteredInSession)
			}
		}(i)

		go func(idx int) {
			for selectedSessionData := range sm.servers[idx].SelectedSessionDataChan {
				pubsub.SelectedSessionDataPubSub.Publish(pubsub.PubSubSelectedSessionDataPreffix+sm.servers[idx].ID, selectedSessionData)
			}
		}(i)

		go func(idx int) {
			for carsPosition := range sm.servers[idx].CarsPositionChan {
				pubsub.CarsPositionPubSub.Publish(pubsub.PubSubCarsPositionPreffix+sm.servers[idx].ID, carsPosition)
			}
		}(i)

		// run update goroutine
		go func(idx int) {
			sm.servers[idx].eventHandler()
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
				go func() {
					// fmt.Printf("Starting websocket reader for server %s\n", sm.servers[idx].ID)
					err := sm.servers[idx].WebSocketReader(sm.ctx)
					if err != nil {
						log.Printf("Error reading websocket: %s", err.Error())
					}
				}()
			}
		}(i)
	}
	wg.Wait()
}
