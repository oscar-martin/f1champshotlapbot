package tracks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
)

type Session struct {
	Driver           string  `json:"driver"`
	TrackCourse      string  `json:"TrackCourse"`
	S1               float64 `json:"s1"`
	S2               float64 `json:"s2"`
	S3               float64 `json:"s3"`
	Time             float64 `json:"time"`
	Fuel             float64 `json:"fuel"`
	Fl               float64 `json:"fl"`
	Fr               float64 `json:"fr"`
	Rl               float64 `json:"rl"`
	Rr               float64 `json:"rr"`
	Fcompound        string  `json:"fcompound"`
	Rcompound        string  `json:"rcompound"`
	DateTime         string  `json:"DateTime"`
	Category         string  `json:"category"`
	CarType          string  `json:"carType"`
	CarClass         string  `json:"carClass"`
	Team             string  `json:"team"`
	Lapcount         int     `json:"lapcount"`
	Lapcountcomplete int     `json:"lapcountcomplete"`
}

func getSessions(ctx context.Context, track string, domain string) ([]Session, error) {
	// Make a get request
	url := fmt.Sprintf("%s/v3/laps?track=%s", domain, url.QueryEscape(track))
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
	var trackSessions []Session
	err = json.Unmarshal(body, &trackSessions)
	if err != nil {
		return nil, err
	}

	sort.Slice(trackSessions, func(i, j int) bool {
		return trackSessions[i].Time < trackSessions[j].Time
	})

	return trackSessions, nil
}
