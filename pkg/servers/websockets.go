package servers

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	mtStandingHistory = "standingsHistory"
	mtStandings       = "standings"
	mtSessionInfo     = "sessionInfo"
)

type Message struct {
	MessageType string `json:"type"`
	Body        any    `json:"body,omitempty"`
}

type LiveStandingData struct {
	ServerName string               `json:"serverName"`
	ServerID   string               `json:"serverId"`
	Drivers    []StandingDriverData `json:"drivers"`
}

type LiveStandingHistoryData struct {
	ServerName  string                                 `json:"serverName"`
	ServerID    string                                 `json:"serverId"`
	DriverNames []string                               `json:"driverNames"`
	DriversData map[string][]StandingHistoryDriverData `json:"driversData"`
}

type LiveSessionInfoData struct {
	ServerName  string      `json:"serverName"`
	ServerID    string      `json:"serverId"`
	SessionInfo SessionInfo `json:"sessionInfo"`
}

type StandingHistoryDriverData struct {
	Position     int     `json:"position"`
	DriverName   string  `json:"driverName"`
	SlotID       int     `json:"slotID"`
	LapTime      float64 `json:"lapTime"`
	SectorTime1  float64 `json:"sectorTime1"`
	SectorTime2  float64 `json:"sectorTime2"`
	TotalLaps    float64 `json:"totalLaps"`
	VehicleName  string  `json:"vehicleName"`
	FinishStatus string  `json:"finishStatus"`
	Pitting      bool    `json:"pitting"`
	CarClass     string  `json:"carClass"`
}

type StandingDriverData struct {
	SlotID             int             `json:"slotID"`
	DriverName         string          `json:"driverName"`
	VehicleName        string          `json:"vehicleName"`
	LapsCompleted      int             `json:"lapsCompleted"`
	Sector             string          `json:"sector"`
	FinishStatus       string          `json:"finishStatus"`
	LapDistance        float64         `json:"lapDistance"`
	PathLateral        float64         `json:"pathLateral"`
	TrackEdge          float64         `json:"trackEdge"`
	BestSectorTime1    float64         `json:"bestSectorTime1"`
	BestSectorTime2    float64         `json:"bestSectorTime2"`
	BestSectorTime3    float64         `json:"bestSectorTime3"` // synthetic field
	BestLapTime        float64         `json:"bestLapTime"`
	LastSectorTime1    float64         `json:"lastSectorTime1"`
	LastSectorTime2    float64         `json:"lastSectorTime2"`
	LastLapTime        float64         `json:"lastLapTime"`
	CurrentSectorTime1 float64         `json:"currentSectorTime1"`
	CurrentSectorTime2 float64         `json:"currentSectorTime2"`
	Pitstops           int             `json:"pitstops"`
	Penalties          int             `json:"penalties"`
	Player             bool            `json:"player"`
	InControl          int             `json:"inControl"`
	Pitting            bool            `json:"pitting"`
	Position           int             `json:"position"`
	CarClass           string          `json:"carClass"`
	TimeBehindNext     float64         `json:"timeBehindNext"`
	LapsBehindNext     float64         `json:"lapsBehindNext"`
	TimeBehindLeader   float64         `json:"timeBehindLeader"`
	LapsBehindLeader   float64         `json:"lapsBehindLeader"`
	LapStartET         float64         `json:"lapStartET"`
	CarPosition        CarPosition     `json:"carPosition"`
	CarVelocity        CarVelocity     `json:"carVelocity"`
	CarAcceleration    CarAcceleration `json:"carAcceleration"`
	Headlights         bool            `json:"headlights"`
	PitState           string          `json:"pitState"`
	ServerScored       bool            `json:"serverScored"`
	GamePhase          string          `json:"gamePhase"`
	Qualification      int             `json:"qualification"`
	TimeIntoLap        float64         `json:"timeIntoLap"`
	EstimatedLapTime   float64         `json:"estimatedLapTime"`
	PitGroup           string          `json:"pitGroup"`
	Flag               string          `json:"flag"`
	UnderYellow        bool            `json:"underYellow"`
	CountLapFlag       string          `json:"countLapFlag"`
	InGarageStall      bool            `json:"inGarageStall"`
	UpgradePack        string          `json:"upgradePack"`
	PitLapDistance     float64         `json:"pitLapDistance"`
	BestLapSectorTime1 float64         `json:"bestLapSectorTime1"`
	BestLapSectorTime2 float64         `json:"bestLapSectorTime2"`
	SteamID            int             `json:"steamID"`
	VehicleFilename    string          `json:"vehicleFilename"`
	CarID              string          `json:"carId"`
	CarNumber          string          `json:"carNumber"`
	FullTeamName       string          `json:"fullTeamName"`
	HasFocus           bool            `json:"hasFocus"`
	FuelFraction       float64         `json:"fuelFraction"`
	AttackMode         AttackMode      `json:"attackMode"`
	DrsActive          bool            `json:"drsActive"`
	Focus              bool            `json:"focus"`
}

type CarPosition struct {
	Type int     `json:"type"`
	Z    float64 `json:"z"`
	Y    float64 `json:"y"`
	X    float64 `json:"x"`
}

type CarVelocity struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
	Velocity float64 `json:"velocity"`
}

type CarAcceleration struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
	Velocity float64 `json:"velocity"`
}

type AttackMode struct {
	TotalCount     int     `json:"totalCount"`
	RemainingCount int     `json:"remainingCount"`
	TimeRemaining  float64 `json:"timeRemaining"`
}

type WindSpeed struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
	Velocity float64 `json:"velocity"`
}

type RaceCompletion struct {
	LapsCompletion float64 `json:"lapsCompletion"`
}

type SessionInfo struct {
	WebSocketRunning   bool           `json:"wsRunning,omitempty"`
	RecevingData       bool           `json:"recevingData,omitempty"`
	TrackName          string         `json:"trackName"`
	Session            string         `json:"session"`
	CurrentEventTime   float64        `json:"currentEventTime"`
	EndEventTime       float64        `json:"endEventTime"`
	MaximumLaps        int            `json:"maximumLaps"`
	LapDistance        float64        `json:"lapDistance"`
	NumberOfVehicles   int            `json:"numberOfVehicles"`
	GamePhase          int            `json:"gamePhase"`
	YellowFlagState    string         `json:"yellowFlagState"`
	SectorFlag         []string       `json:"sectorFlag"`
	StartLightFrame    int            `json:"startLightFrame"`
	NumRedLights       int            `json:"numRedLights"`
	InRealtime         bool           `json:"inRealtime"`
	PlayerName         string         `json:"playerName"`
	PlayerFileName     string         `json:"playerFileName"`
	DarkCloud          float64        `json:"darkCloud"`
	Raining            float64        `json:"raining"`
	AmbientTemp        float64        `json:"ambientTemp"`
	TrackTemp          float64        `json:"trackTemp"`
	WindSpeed          WindSpeed      `json:"windSpeed"`
	MinPathWetness     float64        `json:"minPathWetness"`
	AveragePathWetness float64        `json:"averagePathWetness"`
	MaxPathWetness     float64        `json:"maxPathWetness"`
	GameMode           string         `json:"gameMode"`
	PasswordProtected  bool           `json:"passwordProtected"`
	ServerPort         int            `json:"serverPort"`
	MaxPlayers         int            `json:"maxPlayers"`
	ServerName         string         `json:"serverName"`
	StartEventTime     float64        `json:"startEventTime"`
	RaceCompletion     RaceCompletion `json:"raceCompletion"`
}

func (s *Server) sendZeroData() {
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

func (s *Server) WebSocketReader(ctx context.Context) error {
	if s.WebSocketRunning {
		return nil
	}

	s.RecevingData = false

	defer func() {
		// init channels
		s.WebSocketRunning = false
		s.RecevingData = false
		s.sendZeroData()
	}()

	// init channels
	s.sendZeroData()

	urlString := strings.TrimPrefix(strings.TrimPrefix(s.URL, "https://"), "http://")
	u := url.URL{Scheme: "ws", Host: urlString, Path: "/websocket/controlpanel"}

	// log.Printf("trying to connect to %s", u.String())
	dealer := &websocket.Dialer{
		HandshakeTimeout:  10 * time.Second,
		EnableCompression: true,
	}
	c, _, err := dealer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("Error connecting to %s: %s", u.String(), err.Error())
		return err
	}

	s.WebSocketRunning = true
	log.Printf("connected to %s", u.String())
	s.LiveSessionInfoDataChan <- s.fromMessageToLiveSessionInfoData(s.Name, s.ID, &SessionInfo{})

	defer c.Close()
	doneErr := make(chan error)

	messageChan := make(chan Message)
	go s.dispatchMessage(ctx, messageChan, doneErr)

	go func() {
		defer close(doneErr)
		for {
			var m Message
			err = c.ReadJSON(&m)
			if err != nil {
				log.Println("read error:", err)
				doneErr <- err
				return
			}
			messageChan <- m
		}
	}()
	return <-doneErr
}

func (s *Server) dispatchMessage(ctx context.Context, messageChan <-chan Message, doneChan <-chan error) {
	timeoutTime := 5 * time.Second
	timeout := time.After(timeoutTime)

	for {
		select {
		case <-doneChan:
			return
		case <-timeout:
			// fmt.Printf("timeout waiting for message: %s\n", s.Name)
			s.RecevingData = false
			s.sendZeroData()
			timeout = time.After(timeoutTime)
		case m := <-messageChan:
			timeout = time.After(timeoutTime)
			s.RecevingData = true
			if m.MessageType == mtStandingHistory {
				body := map[string][]StandingHistoryDriverData{}
				jsonData, err := json.Marshal(m.Body)
				if err != nil {
					log.Printf("Error marshalling standingsHistory: %s\n", err.Error())
					continue
				}
				err = json.Unmarshal(jsonData, &body)
				if err != nil {
					log.Printf("Error unmarshalling standingsHistory: %s\n", err.Error())
					continue
				}
				// fmt.Printf("!!!!!!updating live standing history timing!!!!!! %d\n", len(body))
				s.LiveStandingHistoryChan <- s.fromMessageToLiveStandingHistoryData(s.Name, s.ID, &body)
			} else if m.MessageType == mtStandings {
				body := []StandingDriverData{}
				jsonData, err := json.Marshal(m.Body)
				if err != nil {
					log.Printf("Error marshalling standings: %s\n", err.Error())
					continue
				}
				err = json.Unmarshal(jsonData, &body)
				if err != nil {
					log.Printf("Error unmarshalling standings: %s\n", err.Error())
					continue
				}
				// fmt.Printf("!!!!!!updating live timing!!!!!! %d\n", len(body))
				s.LiveStandingChan <- s.fromMessageToLiveStandingData(s.Name, s.ID, body)
			} else if m.MessageType == mtSessionInfo {
				body := SessionInfo{}
				jsonData, err := json.Marshal(m.Body)
				if err != nil {
					log.Printf("Error marshalling sessionInfo: %s\n", err.Error())
					continue
				}
				err = json.Unmarshal(jsonData, &body)
				if err != nil {
					log.Printf("Error unmarshalling sessionInfo: %s\n", err.Error())
					continue
				}
				// fmt.Print("!!!!!!updating sessionInfo!!!!!!\n")
				s.LiveSessionInfoDataChan <- s.fromMessageToLiveSessionInfoData(s.Name, s.ID, &body)
			}
		}
	}
}

func (s *Server) fromMessageToLiveStandingHistoryData(serverName, serverID string, m *map[string][]StandingHistoryDriverData) LiveStandingHistoryData {
	driversDataMap := map[string][]StandingHistoryDriverData{}
	// get driver names from map
	driverNames := []string{}
	for _, driversData := range *m {
		if len(driversData) > 0 {
			driverNames = append(driverNames, driversData[0].DriverName)
			driversDataMap[driversData[0].DriverName] = driversData

			bestS1 := 0.0
			bestS2 := 0.0
			bestS3 := 0.0
			for i := range driversData {
				s1 := driversData[i].SectorTime1
				s2 := -1.0
				if s1 > 0.0 && driversData[i].SectorTime2 > 0.0 {
					s2 = driversData[i].SectorTime2 - s1
				}
				s3 := -1.0
				if s2 > 0.0 && driversData[i].LapTime > 0.0 {
					s3 = driversData[i].LapTime - s2 - s1
				}
				if bestS1 <= 0.0 {
					bestS1 = s1
				} else if s1 > 0.0 && s1 < bestS1 {
					bestS1 = s1
				}
				if bestS2 <= 0.0 {
					bestS2 = s2
				} else if s2 > 0.0 && s2 < bestS2 {
					bestS2 = s2
				}
				if bestS3 <= 0.0 {
					bestS3 = s3
				} else if s3 > 0.0 && s3 < bestS3 {
					bestS3 = s3
				}
			}

			s.BestSectorsForDriver[driversData[0].DriverName] = Sectors{
				Sector1: bestS1,
				Sector2: bestS2,
				Sector3: bestS3,
			}
		}
	}

	sort.SliceStable(driverNames, func(i, j int) bool {
		s1 := driversDataMap[driverNames[i]]
		s2 := driversDataMap[driverNames[j]]
		pos1 := int(math.Inf(1))
		pos2 := int(math.Inf(1))
		if len(s1) > 0 {
			pos1 = s1[len(s1)-1].Position
		}
		if len(s2) > 0 {
			pos2 = s2[len(s2)-1].Position
		}
		return pos1 < pos2
	})

	return LiveStandingHistoryData{
		ServerName:  serverName,
		ServerID:    serverID,
		DriverNames: driverNames,
		DriversData: driversDataMap,
	}
}

func (s *Server) fromMessageToLiveStandingData(serverName, serverID string, data []StandingDriverData) LiveStandingData {
	sort.Slice(data, func(i, j int) bool {
		return data[i].Position < data[j].Position
	})

	for i := range data {
		{
			bestSectors, found := s.BestSectorsForDriver[data[i].DriverName]
			if found {
				data[i].BestSectorTime1 = bestSectors.Sector1
				data[i].BestSectorTime2 = bestSectors.Sector2
				data[i].BestSectorTime3 = bestSectors.Sector3
			}
		}
	}

	return LiveStandingData{
		ServerName: serverName,
		ServerID:   serverID,
		Drivers:    data,
	}
}

func (s *Server) fromMessageToLiveSessionInfoData(serverName, serverID string, data *SessionInfo) LiveSessionInfoData {
	data.WebSocketRunning = s.WebSocketRunning
	data.RecevingData = s.RecevingData
	if data.ServerName == "-none-" {
		data.ServerName = serverID
	}

	return LiveSessionInfoData{
		ServerName:  serverName,
		ServerID:    serverID,
		SessionInfo: *data,
	}
}
