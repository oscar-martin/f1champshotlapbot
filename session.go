package main

import (
	"context"
	"encoding/json"
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

type Category struct {
	ID   string
	Name string
}

type Sessions []Session

func (s Sessions) GetCategories() []Category {
	cats := map[string]string{}
	for _, session := range s {
		if _, exits := cats[session.Category]; !exits {
			id, name := extractCategory(session.Category)
			if id != "" {
				cats[id] = name
			}
		}
	}

	categories := make([]Category, 0, len(cats))

	for id, name := range cats {
		categories = append(categories, Category{ID: id, Name: name})
	}

	// sort categories by name
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	return categories
}

func (s Sessions) GetSessionsByCategoryID(catId string) []Session {
	sessionsForCategory := []Session{}
	for _, session := range s {
		id, _ := extractCategory(session.Category)
		if id == catId {
			sessionsForCategory = append(sessionsForCategory, session)
		}
	}
	return sessionsForCategory
}

func GetSessions(ctx context.Context, track string) (Sessions, error) {
	// Make a get request
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.f1champs.es/v3/laps?track="+url.QueryEscape(track), nil)
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

	return trackSessions, nil
}
