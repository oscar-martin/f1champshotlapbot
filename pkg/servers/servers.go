package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

const (
	serverCheckHttpPath = "/"
	ServerStatusOffline = "🔴"
	ServerStatusOnline  = "🟢"
	ServerPrefixCommand = "Server"

	PubSubServersTopic          = "servers"
	PubSubDriversSessionPreffix = "driversSession-"
	PubSubStintDataPreffix      = "stintData-"
)

type Server struct {
	ID     string
	URL    string
	Name   string
	Online bool
}

func getServers(ctx context.Context, domain string) ([]Server, error) {
	servers := []Server{
		{
			ID:  "Server1",
			URL: "http://localhost:10001",
		},
		{
			ID:  "Server2",
			URL: "http://localhost:10002",
		},
		{
			ID:  "Server3",
			URL: "http://localhost:10003",
		},
		{
			ID:  "Server4",
			URL: "http://localhost:10004",
		},
	}

	for i := range servers {
		servers[i].Name = servers[i].ID
	}
	return servers, nil
}

func (s Server) StatusAndName() string {
	status := ServerStatusOffline
	if s.Online {
		status = ServerStatusOnline
	}
	return fmt.Sprintf("%s %s", status, s.Name)
}

func (s Server) CommandString(commandPrefix string) string {
	status := ServerStatusOffline
	if s.Online {
		status = ServerStatusOnline
	}
	return fmt.Sprintf(" ▸ %s %s ➡ %s_%s", status, s.Name, commandPrefix, s.ID)
}

func (s Server) GetSessionInfo(ctx context.Context) (SessionInfo, error) {
	// Make a get request
	url := fmt.Sprintf("%s/rest/watch/sessionInfo", s.URL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return SessionInfo{}, err
	}

	// Do the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SessionInfo{}, err
	}

	// Close the response body on function return
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SessionInfo{}, err
	}

	// Unmarshal the response body into a TrackResponse struct
	var sessionInfo SessionInfo
	err = json.Unmarshal(body, &sessionInfo)
	if err != nil {
		return SessionInfo{}, err
	}

	return sessionInfo, nil
}

func (s Server) GetDriverSessions(ctx context.Context) (DriversSession, error) {
	dss := []DriverSession{}
	drivers := rand.Intn(10) + 5
	for i := 0; i < drivers; i++ {
		dss = append(dss, DriverSession{
			Driver:           "Driver Player" + fmt.Sprintf("%d", i+1),
			Diff:             4.444,
			S1:               11.111,
			S2:               22.222,
			S3:               33.333,
			Time:             81.111,
			CarType:          "Car1",
			CarClass:         "Class1",
			Team:             "Team1",
			Compound:         "Soft",
			Lapcount:         1,
			Lapcountcomplete: 1,
			S1InBestLap:      11.111,
			S2InBestLap:      22.222,
			S3InBestLap:      33.333,
			BestLap:          81.111,
			BestS1:           11.111,
			BestS2:           22.222,
			BestS3:           33.333,
			OptimumLap:       81.111,
			MaxSpeed:         111.1,
		})
	}
	return DriversSession{
		ServerName: s.Name,
		ServerID:   s.ID,
		Drivers:    dss,
	}, nil
}

func (s Server) GetStintData(ctx context.Context) (StintData, error) {
	dss := map[string]DriverStint{}
	drivers := rand.Intn(10) + 5
	for i := 0; i < drivers; i++ {
		lapsCount := rand.Intn(10) + 5
		laps := []LapTime{}
		for j := 0; j < lapsCount; j++ {
			laps = append(laps, LapTime{
				LapTime:  81.111,
				S1:       11.111,
				S2:       22.222,
				S3:       33.333,
				MaxSpeed: 111.1,
				Diff:     4.444,
			})
		}
		ds := DriverStint{
			Driver:     "Driver Player" + fmt.Sprintf("%d", i+1),
			Laps:       laps,
			OptimumLap: laps[2],
			BestLap:    laps[1],
			CarType:    "Car1",
			CarClass:   "Class1",
			Team:       "Team1",
		}
		dss[ds.Driver] = ds
	}
	return StintData{
		ServerName: s.Name,
		ServerID:   s.ID,
		Drivers:    dss,
	}, nil
}
