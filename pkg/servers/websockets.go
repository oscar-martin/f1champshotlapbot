package servers

import (
	"context"
	"encoding/json"
	"fmt"
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
	TopSpeed     float64 `json:"topSpeed"` // synthetic field
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
	TopSpeedPerLap     map[int]float64 `json:"TopSpeedPerLap"`  // synthetic field
	BestLap            int             `json:"BestLap"`         // synthetic field
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

type ServerStarted struct {
	ServerName  string  `json:"serverName"`
	ServerID    string  `json:"serverId"`
	SessionType string  `json:"sessionType"`
	TrackName   string  `json:"trackName"`
	EventTime   float64 `json:"eventTime"`
}

func (ss ServerStarted) String() string {
	return fmt.Sprintf("  ▸ Servidor: %s\n  ▸ Sesión: %s\n  ▸ Circuito: %s", ss.ServerName, ss.SessionType, ss.TrackName)
}

func (s *Server) WebSocketReader(ctx context.Context, newSessionChannel chan<- ServerStarted) error {
	if s.WebSocketRunning {
		return nil
	}

	defer func() {
		s.WebSocketRunning = false
		s.reset()
	}()

	// init channels
	s.reset()

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
	go s.dispatchMessage(ctx, messageChan, doneErr, newSessionChannel)

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

func (s *Server) dispatchMessage(ctx context.Context, messageChan <-chan Message, doneChan <-chan error, newSessionChannel chan<- ServerStarted) {
	timeoutTime := 5 * time.Second
	timeout := time.After(timeoutTime)

	for {
		select {
		case <-doneChan:
			return
		case <-timeout:
			// fmt.Printf("timeout waiting for message: %s\n", s.Name)
			s.reset()
			timeout = time.After(timeoutTime)
		case m := <-messageChan:
			timeout = time.After(timeoutTime)
			if !s.RecevingData {
				s.StartSessionPendingNotification = true
			}
			s.RecevingData = true
			if m.MessageType == mtStandingHistory {
				shdd := map[string][]StandingHistoryDriverData{}
				jsonData, err := json.Marshal(m.Body)
				if err != nil {
					log.Printf("Error marshalling standingsHistory: %s\n", err.Error())
					continue
				}
				err = json.Unmarshal(jsonData, &shdd)
				if err != nil {
					log.Printf("Error unmarshalling standingsHistory: %s\n", err.Error())
					continue
				}
				// fmt.Printf("!!!!!!updating live standing history timing!!!!!! %d\n", len(body))
				s.LiveStandingHistoryChan <- s.fromMessageToLiveStandingHistoryData(s.Name, s.ID, &shdd)
			} else if m.MessageType == mtStandings {
				sdd := []StandingDriverData{}
				jsonData, err := json.Marshal(m.Body)
				if err != nil {
					log.Printf("Error marshalling standings: %s\n", err.Error())
					continue
				}
				err = json.Unmarshal(jsonData, &sdd)
				if err != nil {
					log.Printf("Error unmarshalling standings: %s\n", err.Error())
					continue
				}
				if s.StartSessionPendingNotification &&
					s.SessionStarted.ServerID != "" &&
					len(sdd) > 0 /* player */ {
					s.StartSessionPendingNotification = false
					// only send the notification once at least one player comes in
					// otherwise, it's probably a session that was already running when the bot started
					log.Printf("Signaling serverStarted subscribers for Server %s started: %s\n", s.Name, s.SessionStarted.SessionType)
					newSessionChannel <- s.SessionStarted
				}

				// fmt.Printf("!!!!!!updating live timing!!!!!! %d\n", len(body))
				s.LiveStandingChan <- s.fromMessageToLiveStandingData(s.Name, s.ID, sdd)
			} else if m.MessageType == mtSessionInfo {
				si := SessionInfo{}
				jsonData, err := json.Marshal(m.Body)
				if err != nil {
					log.Printf("Error marshalling sessionInfo: %s\n", err.Error())
					continue
				}
				err = json.Unmarshal(jsonData, &si)
				if err != nil {
					log.Printf("Error unmarshalling sessionInfo: %s\n", err.Error())
					continue
				}
				if s.SessionStarted.ServerID == "" {
					log.Printf("Server %s started receiving data for session: %s\n", s.Name, si.Session)
					s.SessionStarted = ServerStarted{
						ServerName:  s.Name,
						ServerID:    s.ID,
						SessionType: si.Session,
						TrackName:   si.TrackName,
						EventTime:   si.CurrentEventTime,
					}
					go func() {
						trackThumbnail := buildCurrentSessionTrackThumbnail(s.URL)
						log.Printf("Built track thumbnail: %s\n", trackThumbnail.String())
						s.ThumbnailChan <- trackThumbnail
					}()
				}

				// fmt.Print("!!!!!!updating sessionInfo!!!!!!\n")
				s.LiveSessionInfoDataChan <- s.fromMessageToLiveSessionInfoData(s.Name, s.ID, &si)
			}
		}
	}
}

func (s *Server) fromMessageToLiveStandingHistoryData(serverName, serverID string, idToDriverDataSlice *map[string][]StandingHistoryDriverData) LiveStandingHistoryData {
	driversDataMap := map[string][]StandingHistoryDriverData{}
	// get driver names from map
	driverNames := []string{}

	// as drivers can come in and out of the server, we need to keep track of the total laps
	// by aggregating the laps of the drivers that have the same name
	// sort idToDriverDataSlice by key as their data is ordered by key id.
	sortedKeys := []string{}
	for k := range *idToDriverDataSlice {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	playersPosition := map[string]int{}

	// iterate over map
	for _, k := range sortedKeys {
		driversData := (*idToDriverDataSlice)[k]
		if len(driversData) > 0 {
			driverName := driversData[0].DriverName
			existingDriversData, found := driversDataMap[driverName]
			if !found {
				driverNames = append(driverNames, driverName)
				driversDataMap[driverName] = driversData
			} else {
				numLaps := 0.0
				if len(existingDriversData) > 0.0 {
					numLaps = existingDriversData[len(existingDriversData)-1].TotalLaps
				}
				for i := range driversData {
					driversData[i].TotalLaps += numLaps
				}

				driversData = append(existingDriversData, driversData...)
				driversDataMap[driverName] = driversData
			}

			topSpeedForDriver, topSpeedForDriverFound := s.TopSpeedForDriver[driverName]

			bestS1 := 0.0
			bestS2 := 0.0
			bestS3 := 0.0
			for i := range driversData {
				playerPosition, ok := playersPosition[driverName]
				if !ok {
					playerPosition = int(math.Inf(1))
				}
				if driversData[i].Position < playerPosition {
					playerPosition = driversData[i].Position
				}
				playersPosition[driverName] = playerPosition

				if topSpeedForDriverFound {
					driversData[i].TopSpeed = topSpeedForDriver[i]
				} else {
					driversData[i].TopSpeed = -1.0
				}
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

			s.BestSectorsForDriver[driverName] = Sectors{
				Sector1: bestS1,
				Sector2: bestS2,
				Sector3: bestS3,
			}
		}
	}

	// sort driver names by position
	sort.SliceStable(driverNames, func(i, j int) bool {
		pos1 := playersPosition[driverNames[i]]
		pos2 := playersPosition[driverNames[j]]
		bestLap1 := s.BestSectorsForDriver[driverNames[i]].TimeLap()
		bestLap2 := s.BestSectorsForDriver[driverNames[j]].TimeLap()
		if pos1 == pos2 {
			if bestLap1 > 0.0 && bestLap2 > 0.0 {
				return bestLap1 < bestLap2
			}
			return bestLap1 > 0.0
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
		// update topSpeed
		{
			topSpeedPerLaps, found := s.TopSpeedForDriver[data[i].DriverName]
			if !found {
				s.TopSpeedForDriver[data[i].DriverName] = map[int]float64{}
				topSpeedPerLaps = s.TopSpeedForDriver[data[i].DriverName]
			}
			topSpeed, found := topSpeedPerLaps[data[i].LapsCompleted]
			speed := data[i].CarVelocity.Velocity * 3.6
			if !found {
				topSpeedPerLaps[data[i].LapsCompleted] = speed
			} else if topSpeed < speed {
				topSpeedPerLaps[data[i].LapsCompleted] = speed
			}
			data[i].TopSpeedPerLap = topSpeedPerLaps
		}

		// update best sectors
		{
			bestSectors, found := s.BestSectorsForDriver[data[i].DriverName]
			if found {
				data[i].BestSectorTime1 = bestSectors.Sector1
				data[i].BestSectorTime2 = bestSectors.Sector2
				data[i].BestSectorTime3 = bestSectors.Sector3
			}
		}

		// update best lap
		{
			if data[i].BestLapTime == data[i].LastLapTime && data[i].BestLapTime > 0.0 {
				lap := data[i].LapsCompleted - 1
				data[i].BestLap = lap
				s.BestLapForDriver[data[i].DriverName] = lap
			} else {
				bestLap, found := s.BestLapForDriver[data[i].DriverName]
				if !found {
					bestLap = -1
				}
				data[i].BestLap = bestLap
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
