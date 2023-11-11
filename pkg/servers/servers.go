package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"log"
)

const (
	serverCheckHttpPath = "/"
)

type Server struct {
	ID     string
	URL    string
	Name   string
	Online bool
}

func checkServerOnline(ctx context.Context, server Server) (bool, error) {
	// try to reach the http server
	client := &http.Client{Timeout: 5 * time.Second}
	_, err := client.Get(fmt.Sprintf("%s%s", server.URL, serverCheckHttpPath))
	if err != nil {
		// A timeout error occurred
		// time.Sleep(3 * time.Second)
		if os.IsTimeout(err) {
			return false, nil
		}

		// This was an error, but not a timeout
		log.Printf("Error checking server %s: %s", server.Name, err.Error())
		return false, err
	}
	return true, nil
}

func getServers(ctx context.Context, domain string) ([]Server, error) {
	servers := []Server{
		{
			ID:   "1",
			URL:  "http://localhost:10001",
			Name: "Server 1",
		},
		{
			ID:   "2",
			URL:  "http://localhost:10002",
			Name: "Server 2",
		},
		{
			ID:   "3",
			URL:  "http://localhost:10003",
			Name: "Server 3",
		},
		{
			ID:   "4",
			URL:  "http://localhost:10004",
			Name: "Server 4",
		},
	}

	wg := sync.WaitGroup{}
	for i := range servers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			online, err := checkServerOnline(ctx, servers[idx])
			if err != nil {
				online = false
				log.Printf("Error checking server %s: %s", servers[idx].Name, err.Error())
			}
			servers[idx].Online = online
		}(i)
	}
	wg.Wait()
	return servers, nil
}

func (s Server) CommandString(commandPrefix string) string {
	status := "🔴"
	if s.Online {
		status = "🟢"
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
