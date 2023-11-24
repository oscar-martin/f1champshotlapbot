package servers

import (
	"fmt"
)

const (
	serverCheckHttpPath = "/"
	ServerStatusOffline = "🔴"
	ServerStatusOnline  = "🟢"
	ServerPrefixCommand = "Server"

	PubSubSessionInfoPreffix    = "sessionInfo-"
	PubSubDriversSessionPreffix = "driversSession-"
	PubSubStintDataPreffix      = "stintData-"
)

type Server struct {
	ID                      string `json:"id"`
	URL                     string `json:"url"`
	Name                    string
	WebSocketRunning        bool
	RecevingData            bool
	BestSector3ForDriver    map[string]float64
	LiveSessionInfoDataChan chan LiveSessionInfoData     `json:"-"`
	LiveStandingHistoryChan chan LiveStandingHistoryData `json:"-"`
	LiveStandingChan        chan LiveStandingData        `json:"-"`
}

func (s Server) StatusAndName() string {
	status := ServerStatusOffline
	if s.WebSocketRunning {
		status = ServerStatusOnline
	}
	return fmt.Sprintf("%s %s", status, s.Name)
}

func (s Server) CommandString(commandPrefix string) string {
	status := ServerStatusOffline
	if s.WebSocketRunning {
		status = ServerStatusOnline
	}
	return fmt.Sprintf(" ▸ %s %s ➡ %s_%s", status, s.Name, commandPrefix, s.ID)
}
