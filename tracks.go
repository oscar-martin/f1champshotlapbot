package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type Track struct {
	Command string
	ID      string
	Name    string
}

type Tracks []Track

func (t Tracks) Len() int {
	return len(t)
}

func (t Tracks) GetTrackByID(id string) (Track, bool) {
	for _, track := range t {
		if track.ID == id {
			return track, true
		}
	}
	return Track{}, false
}

func (t Tracks) GetRange(from, to int) []Track {
	return t[from:to]
}

func (t Track) String() string {
	return " ▸ " + t.Name + " ➡ " + t.Command
}

func getTracks(ctx context.Context) (Tracks, error) {
	// Make a get request
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.f1champs.es/v3/laps?tracklist=tracklist", nil)
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
	var tracks []Track
	for _, trackName := range trackNames {
		tracks = append(tracks, Track{
			Command: "/" + toID(trackName),
			ID:      toID(trackName),
			Name:    trackName,
		})
	}

	return tracks, nil
}
