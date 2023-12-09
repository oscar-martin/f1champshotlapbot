package servers

import (
	"context"
	"encoding/json"
	"f1champshotlapsbot/pkg/helper"
	"f1champshotlapsbot/pkg/model"
	"log"
	"math"
	"net/url"
	"slices"
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

func (s *Server) WebSocketReader(ctx context.Context) error {
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
	s.LiveSessionInfoDataChan <- s.fromMessageToLiveSessionInfoData(s.Name, s.ID, &model.SessionInfo{})

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
			s.reset()
			timeout = time.After(timeoutTime)
		case m := <-messageChan:
			timeout = time.After(timeoutTime)
			if !s.ReceivingData {
				s.StartSessionPendingNotification = true
			}
			s.ReceivingData = true
			if m.MessageType == mtStandingHistory {
				shdd := map[string][]model.StandingHistoryDriverData{}
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
				sdd := []model.StandingDriverData{}
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
					log.Printf("Signaling First Driver entered in server %s. Session: %s\n", s.Name, s.SessionStarted.SessionType)
					s.FirstDriverEnteredChan <- s.SessionStarted
				}

				// fmt.Printf("!!!!!!updating live timing!!!!!! %d\n", len(body))
				lsd, cp := s.fromMessageToLiveStandingData(s.Name, s.ID, sdd)
				s.LiveStandingChan <- lsd
				s.CarsPositionChan <- cp

			} else if m.MessageType == mtSessionInfo {
				si := model.SessionInfo{}
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
					ss := model.ServerStarted{
						ServerName:  s.Name,
						ServerID:    s.ID,
						SessionType: si.Session,
						TrackName:   si.TrackName,
						EventTime:   si.CurrentEventTime,
					}
					s.ServerStartedChan <- ss
				}

				// fmt.Print("!!!!!!updating sessionInfo!!!!!!\n")
				s.LiveSessionInfoDataChan <- s.fromMessageToLiveSessionInfoData(s.Name, s.ID, &si)
			}
		}
	}
}

func (s *Server) fromMessageToLiveStandingHistoryData(serverName, serverID string, idToDriverDataSlice *map[string][]model.StandingHistoryDriverData) model.LiveStandingHistoryData {
	driversDataMap := map[string][]model.StandingHistoryDriverData{}
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

			carId := s.DriverToCarId[driverName]

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
				driversData[i].CarId = carId
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

	return model.LiveStandingHistoryData{
		ServerName:  serverName,
		ServerID:    serverID,
		DriverNames: driverNames,
		DriversData: driversDataMap,
	}
}

func (s *Server) fromMessageToLiveStandingData(serverName, serverID string, data []model.StandingDriverData) (model.LiveStandingData, []model.CarPosition) {
	carsPosition := []model.CarPosition{}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Position < data[j].Position
	})

	for i := range data {
		// update car position
		{
			cp := data[i].CarPosition
			cp.DriverShortName = helper.GetDriverCodeName(data[i].DriverName)
			carsPosition = append(carsPosition, cp)
		}
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

		// update driver car id
		{
			s.DriverToCarId[data[i].DriverName] = data[i].CarID
		}
	}

	slices.Reverse(carsPosition)
	return model.LiveStandingData{
		ServerName: serverName,
		ServerID:   serverID,
		Drivers:    data,
	}, carsPosition
}

func (s *Server) fromMessageToLiveSessionInfoData(serverName, serverID string, data *model.SessionInfo) model.LiveSessionInfoData {
	data.WebSocketRunning = s.WebSocketRunning
	data.ReceivingData = s.ReceivingData
	data.LiveMapPath = s.LiveMapPath
	data.LiveMapDomain = s.LiveMapDomain
	if data.ServerName == "-none-" {
		data.ServerName = serverID
	}

	return model.LiveSessionInfoData{
		ServerName:  serverName,
		ServerID:    serverID,
		SessionInfo: *data,
	}
}
