package tracks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	tracksPerPage = 10
)

type Manager struct {
	tracks    []*Track
	mu        sync.Mutex
	apiDomain string
	bot       *tgbotapi.BotAPI
}

func NewTrackManager(bot *tgbotapi.BotAPI, domain string) *Manager {
	return &Manager{
		apiDomain: domain,
		bot:       bot,
	}
}

func (tm *Manager) Sync(ctx context.Context, ticker *time.Ticker, exitChan chan bool) {
	go func() {
		for {
			select {
			case <-exitChan:
				return
			case t := <-ticker.C:
				fmt.Println("Resetting tracks and sessions at: ", t)
				tm.mu.Lock()
				tm.tracks = []*Track{}
				tm.mu.Unlock()
			}
		}
	}()
}

func (tm *Manager) Lock() {
	tm.mu.Lock()
}

func (tm *Manager) Unlock() {
	tm.mu.Unlock()
}

func (tm *Manager) GetTracks(ctx context.Context) ([]*Track, error) {
	if len(tm.tracks) == 0 {
		// if there is no tracks, fetch them
		ts, err := getTracks(ctx, tm.apiDomain)
		if err != nil {
			return ts, err
		}
		tm.tracks = ts
	}

	return tm.tracks, nil
}

func (tm *Manager) GetTrackByID(id string) (*Track, bool) {
	for _, track := range tm.tracks {
		if track.ID == id {
			return track, true
		}
	}
	return &Track{}, false
}

func (tm *Manager) GetRange(from, to int) []*Track {
	return tm.tracks[from:to]
}

func getTracks(ctx context.Context, domain string) ([]*Track, error) {
	// Make a get request
	url := fmt.Sprintf("%s/v3/laps?tracklist=tracklist", domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Do the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Close the response body on function return
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response body into a TrackResponse struct
	var trackNames []string
	err = json.Unmarshal(body, &trackNames)
	if err != nil {
		return nil, err
	}

	// Create a slice of Track structs
	var tracks []*Track
	for _, trackName := range trackNames {
		track := Track{
			Command: "/" + toID(trackName),
			ID:      toID(trackName),
			Name:    trackName,
			mu:      sync.Mutex{},
		}
		tracks = append(tracks, &track)
	}

	return tracks, nil
}
