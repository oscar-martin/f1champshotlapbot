package servers

import (
	"f1champshotlapsbot/pkg/model"
	"f1champshotlapsbot/pkg/thumbnails"
	"fmt"
	"sync"
)

const (
	serverCheckHttpPath          = "/"
	ServerStatusOffline          = "🔴"
	ServerStatusOnline           = "🟢"
	ServerStatusOnlineButNotData = "🟡"
	ServerPrefixCommand          = "Server"

	PubSubSessionInfoPreffix    = "sessionInfo-"
	PubSubDriversSessionPreffix = "driversSession-"
	PubSubStintDataPreffix      = "stintData-"
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
	RecevingData                    bool
	StartSessionPendingNotification bool
	BestSectorsForDriver            map[string]Sectors
	BestLapForDriver                map[string]int
	TopSpeedForDriver               map[string]map[int]float64
	DriverToCarId                   map[string]string
	SessionStarted                  ServerStarted
	LiveSessionInfoDataChan         chan model.LiveSessionInfoData     `json:"-"`
	LiveStandingHistoryChan         chan model.LiveStandingHistoryData `json:"-"`
	LiveStandingChan                chan model.LiveStandingData        `json:"-"`
	ThumbnailChan                   chan thumbnails.Thumbnail          `json:"-"`
}

func NewServer(id, url string) Server {
	return Server{
		mu:                   &sync.Mutex{},
		ID:                   id,
		URL:                  url,
		WebSocketRunning:     false,
		RecevingData:         false,
		BestSectorsForDriver: make(map[string]Sectors),
		DriverToCarId:        make(map[string]string),
		BestLapForDriver:     make(map[string]int),
		TopSpeedForDriver:    make(map[string]map[int]float64),
	}
}

func (s Server) Status() string {
	status := ServerStatusOffline
	if s.WebSocketRunning {
		if s.RecevingData {
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
	s.RecevingData = false
	s.StartSessionPendingNotification = false
	s.BestSectorsForDriver = make(map[string]Sectors)
	s.DriverToCarId = make(map[string]string)
	s.BestLapForDriver = make(map[string]int)
	s.TopSpeedForDriver = make(map[string]map[int]float64)
	s.SessionStarted = ServerStarted{}
	{
		body := map[string][]model.StandingHistoryDriverData{}
		s.LiveStandingHistoryChan <- s.fromMessageToLiveStandingHistoryData(s.Name, s.ID, &body)
	}
	{
		body := []model.StandingDriverData{}
		s.LiveStandingChan <- s.fromMessageToLiveStandingData(s.Name, s.ID, body)
	}
	{
		body := model.SessionInfo{}
		s.LiveSessionInfoDataChan <- s.fromMessageToLiveSessionInfoData(s.Name, s.ID, &body)
	}
}
