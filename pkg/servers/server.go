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
	ID                      string `json:"id"`
	URL                     string `json:"url"`
	Name                    string
	WebSocketRunning        bool
	RecevingData            bool
	BestSectorsForDriver    map[string]Sectors
	LiveSessionInfoDataChan chan LiveSessionInfoData     `json:"-"`
	LiveStandingHistoryChan chan LiveStandingHistoryData `json:"-"`
	LiveStandingChan        chan LiveStandingData        `json:"-"`
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
