package servers

import (
	"f1champshotlapsbot/pkg/livemap"
	"f1champshotlapsbot/pkg/model"
	"f1champshotlapsbot/pkg/pubsub"
	"f1champshotlapsbot/pkg/resources"
	"fmt"
	"log"
	"sync"
)

const (
	serverCheckHttpPath          = "/"
	ServerStatusOffline          = "🔴"
	ServerStatusOnline           = "🟢"
	ServerStatusOnlineButNotData = "🟡"
	ServerPrefixCommand          = "Server"
)

type Sectors struct {
	Sector1 float64 `json:"sector1"`
	Sector2 float64 `json:"sector2"`
	Sector3 float64 `json:"sector3"`
}

func (s Sectors) TimeLap() float64 {
	if s.Sector1 > 0.0 && s.Sector2 > 0.0 && s.Sector3 > 0.0 {
		return s.Sector1 + s.Sector2 + s.Sector3
	}
	return -1.0
}

type Server struct {
	mu                              *sync.Mutex
	ID                              string `json:"id"`
	URL                             string `json:"url"`
	Name                            string
	WebSocketRunning                bool
	ReceivingData                   bool
	StartSessionPendingNotification bool
	BestSectorsForDriver            map[string]Sectors
	BestLapForDriver                map[string]int
	TopSpeedForDriver               map[string]map[int]float64
	DriverToCarId                   map[string]string
	SessionStarted                  model.ServerStarted
	LiveSessionInfoDataChan         chan model.LiveSessionInfoData     `json:"-"`
	LiveStandingHistoryChan         chan model.LiveStandingHistoryData `json:"-"`
	LiveStandingChan                chan model.LiveStandingData        `json:"-"`
	ThumbnailChan                   chan resources.Resource            `json:"-"`
	ServerStartedChan               chan model.ServerStarted           `json:"-"`
	ServerStoppedChan               chan string                        `json:"-"`
	FirstDriverEnteredChan          chan model.ServerStarted           `json:"-"`
	SelectedSessionDataChan         chan model.SelectedSessionData     `json:"-"`
	CarsPositionChan                chan []model.CarPosition           `json:"-"`
	cancelDownloadingChan           chan bool                          `json:"-"`
	LiveMap                         *livemap.LiveMap                   `json:"-"`
	LiveMapPath                     string                             `json:"liveMapPath"`
	LiveMapDomain                   string                             `json:"liveMapDomain"`
}

func NewServer(id, url, domain string) Server {
	return Server{
		mu:                   &sync.Mutex{},
		ID:                   id,
		URL:                  url,
		LiveMapDomain:        domain,
		WebSocketRunning:     false,
		ReceivingData:        false,
		BestSectorsForDriver: make(map[string]Sectors),
		DriverToCarId:        make(map[string]string),
		BestLapForDriver:     make(map[string]int),
		TopSpeedForDriver:    make(map[string]map[int]float64),
	}
}

func (s *Server) eventHandler() {
	startedChan := pubsub.SessionStartedPubSub.Subscribe(pubsub.PubSubSessionStartedPreffix)
	stoppedChan := pubsub.SessionStoppedPubSub.Subscribe(pubsub.PubSubSessionStoppedPreffix)
	selectedSessionData := pubsub.SelectedSessionDataPubSub.Subscribe(pubsub.PubSubSelectedSessionDataPreffix + s.ID)
	for {
		select {
		case ss := <-startedChan:
			if ss.ServerID == s.ID {
				s.SessionStarted = ss

				s.cancelDownloadingChan = make(chan bool)
				// fetch session data and send it to the channel
				go retryWithCancel(func() error {
					ssd, err := getSelectedSessionData(s.URL)
					if err != nil {
						log.Printf("Error getting selected session data: %s. It will be retried soon\n", err)
						return err
					}
					log.Printf("Selected session data received for Server %s\n", s.ID)
					s.SelectedSessionDataChan <- ssd
					return nil
				}, s.cancelDownloadingChan)
			}
		case id := <-stoppedChan:
			if id == s.ID {
				// force to stop any pending download (due to errors)
				if s.cancelDownloadingChan != nil {
					close(s.cancelDownloadingChan)
					s.cancelDownloadingChan = nil
				}
				s.LiveMap.StopSession()
			}
		case ssd := <-selectedSessionData:
			s.cancelDownloadingChan = make(chan bool)
			// fetch track thumbnail and send it to the channel
			go retryWithCancel(func() error {
				t, err := buildTrackThumbnail(s.URL, ssd)
				if err != nil {
					log.Printf("Error getting track thumbnail data: %s. It will be retried soon\n", err)
					return err
				}
				log.Printf("Track thumbnail received for Server %s\n", s.ID)
				s.ThumbnailChan <- t
				return nil
			}, s.cancelDownloadingChan)

			// fetch track svg
			go retryWithCancel(func() error {
				svgTrackResource, err := buildTrackSvg(s.URL, ssd)
				if err != nil {
					log.Printf("Error getting track svg data: %s. It will be retried soon\n", err)
					return err
				}
				log.Printf("SVG Track received for Server %s\n", s.ID)
				s.LiveMap.StartSession(ssd, svgTrackResource)
				return nil
			}, s.cancelDownloadingChan)
		}
	}
}

func (s Server) Status() string {
	status := ServerStatusOffline
	if s.WebSocketRunning {
		if s.ReceivingData {
			status = ServerStatusOnline
		} else {
			status = ServerStatusOnlineButNotData
		}
	}
	return status
}

func (s Server) StatusAndName() string {
	return fmt.Sprintf("%s %s", s.Status(), s.Name)
}

func (s *Server) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ReceivingData = false
	s.StartSessionPendingNotification = false
	s.BestSectorsForDriver = make(map[string]Sectors)
	s.DriverToCarId = make(map[string]string)
	s.BestLapForDriver = make(map[string]int)
	s.TopSpeedForDriver = make(map[string]map[int]float64)
	s.SessionStarted = model.ServerStarted{}
	{
		body := map[string][]model.StandingHistoryDriverData{}
		s.LiveStandingHistoryChan <- s.fromMessageToLiveStandingHistoryData(s.Name, s.ID, &body)
	}
	{
		body := []model.StandingDriverData{}
		lsd, cp := s.fromMessageToLiveStandingData(s.Name, s.ID, body)
		s.LiveStandingChan <- lsd
		s.CarsPositionChan <- cp
	}
	{
		body := model.SessionInfo{}
		s.LiveSessionInfoDataChan <- s.fromMessageToLiveSessionInfoData(s.Name, s.ID, &body)
	}
	{
		s.ServerStoppedChan <- s.ID
	}
}
