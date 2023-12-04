package servers

import (
	"encoding/json"
	"f1champshotlapsbot/pkg/thumbnails"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Series struct represents the "series" part of the JSON.
type Series struct {
	ShortName   string `json:"shortName"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Disabled    bool   `json:"disabled"`
	Signature   string `json:"signature"`
	Version     string `json:"version"`
}

// Track struct represents the "track" part of the JSON.
type Track struct {
	ID                    string                 `json:"id"`
	ShortName             string                 `json:"shortName"`
	Name                  string                 `json:"name"`
	SceneDesc             string                 `json:"sceneDesc"`
	Year                  string                 `json:"year"`
	Layout                string                 `json:"layout"`
	Description           string                 `json:"description"`
	Length                string                 `json:"length"`
	Type                  string                 `json:"type"`
	Localizations         map[string]interface{} `json:"localizations"`
	CategoryLocalizations map[string]interface{} `json:"categoryLocalizations"`
	PremID                int                    `json:"premId"`
	Owned                 bool                   `json:"owned"`
	Image                 string                 `json:"image"`
	Thumbnail             string                 `json:"thumbnail"`
}

// Car struct represents the "car" part of the JSON.
type Car struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	BHP                   string                 `json:"bhp"`
	UsedIn                string                 `json:"usedIn"`
	Configuration         string                 `json:"configuration"`
	FullPathTree          string                 `json:"fullPathTree"`
	VehFile               string                 `json:"vehFile"`
	Engine                string                 `json:"engine"`
	Manufacturer          string                 `json:"manufacturer"`
	Localizations         map[string]interface{} `json:"localizations"`
	CategoryLocalizations map[string]interface{} `json:"categoryLocalizations"`
	PremID                int                    `json:"premId"`
	Owned                 bool                   `json:"owned"`
	Image                 string                 `json:"image"`
	Thumbnail             string                 `json:"thumbnail"`
}

// Data struct represents the entire JSON structure.
type SelectedSessionData struct {
	Series Series `json:"series"`
	Track  Track  `json:"track"`
	Car    Car    `json:"car"`
}

func buildCurrentSessionTrackThumbnail(serverUrl string) thumbnails.Thumbnail {
	for {
		t, err := buildTrackThumbnail(serverUrl)
		if err != nil {
			delay := 15 * time.Second
			log.Printf("Error updating selected session data: %s. It will be retried in %s\n", err, delay)
			time.Sleep(delay)
			continue
		}
		return t
	}
}

func buildTrackThumbnail(serverUrl string) (thumbnails.Thumbnail, error) {
	t := thumbnails.Thumbnail{}
	url := fmt.Sprintf("%s/rest/race/selection", serverUrl)
	// fmt.Println(url)
	response, err := http.Get(url)
	if err != nil {
		log.Printf("Error http-getting selected session: %s\n", err)
		return t, err
	}
	defer response.Body.Close()

	var selectedSessionData SelectedSessionData
	err = json.NewDecoder(response.Body).Decode(&selectedSessionData)
	if err != nil {
		log.Printf("Error decoding selected session: %s\n", err)
		return t, err
	}

	th := thumbnails.NewTrackThumbnail(serverUrl, selectedSessionData.Track.ID)
	err = th.Prefetch()
	if err != nil {
		log.Printf("Error prefetching track thumbnail: %s\n", err)
		return t, err
	}
	return *th, nil
}
