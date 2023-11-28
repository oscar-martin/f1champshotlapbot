package servers

import (
	"fmt"
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

type Server struct {
	ID                              string `json:"id"`
	URL                             string `json:"url"`
	Name                            string
	WebSocketRunning                bool
	RecevingData                    bool
	StartSessionPendingNotification bool
	BestSectorsForDriver            map[string]Sectors
	BestLapForDriver                map[string]int
	TopSpeedForDriver               map[string]map[int]float64
	LiveSessionInfoDataChan         chan LiveSessionInfoData     `json:"-"`
	LiveStandingHistoryChan         chan LiveStandingHistoryData `json:"-"`
	LiveStandingChan                chan LiveStandingData        `json:"-"`
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
	s.RecevingData = false
	s.StartSessionPendingNotification = false
	s.BestSectorsForDriver = make(map[string]Sectors)
	s.BestLapForDriver = make(map[string]int)
	s.TopSpeedForDriver = make(map[string]map[int]float64)
	{
		body := map[string][]StandingHistoryDriverData{}
		s.LiveStandingHistoryChan <- s.fromMessageToLiveStandingHistoryData(s.Name, s.ID, &body)
	}
	{
		body := []StandingDriverData{}
		s.LiveStandingChan <- s.fromMessageToLiveStandingData(s.Name, s.ID, body)
	}
	{
		body := SessionInfo{}
		s.LiveSessionInfoDataChan <- s.fromMessageToLiveSessionInfoData(s.Name, s.ID, &body)
	}
}
