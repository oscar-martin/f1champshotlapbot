package thumbnails

import (
	"encoding/json"
	"f1champshotlapsbot/pkg/layout"
	"fmt"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	thumbnailsDir = "./thumbnails"
)

func init() {
	// create thumbnail dir if not exists
	if _, err := os.Stat(thumbnailsDir); os.IsNotExist(err) {
		os.Mkdir(thumbnailsDir, 0755)
	}
}

type Thumbnail struct {
	Id        string
	FilePath  string
	ServerUrl string
	Endpoint  string
}

func NewTrackThumbnail(url, id string) *Thumbnail {
	return &Thumbnail{
		Id:        id,
		Endpoint:  "rest/race/track/%s/trackmap",
		ServerUrl: url,
	}
}

func (t Thumbnail) IsZero() bool {
	return t.Id == "" || t.FilePath == ""
}

func (t Thumbnail) String() string {
	return fmt.Sprintf("TrackID: %s, FilePath: %s, ServerURL: %s", t.Id, t.FilePath, t.ServerUrl)
}

func (t Thumbnail) requestUrl() string {
	return fmt.Sprintf("%s/%s", t.ServerUrl, fmt.Sprintf(t.Endpoint, t.Id))
}

func (t *Thumbnail) Prefetch() error {
	_, err := t.FileData()
	return err
}

func (t *Thumbnail) FileData() (tgbotapi.RequestFileData, error) {
	if t.Id == "" {
		return nil, fmt.Errorf("thumbnail is not initialized")
	}
	if t.FilePath != "" {
		return tgbotapi.FilePath(t.FilePath), nil
	}
	aiwData, err := t.fetchAIWData()
	if err != nil {
		log.Printf("Error fetching track aiw data: %s\n", err)
		return nil, err
	}

	filePath := t.buildFilePath()
	err = layout.BuildLayoutPNG(filePath, aiwData)
	if err != nil {
		log.Printf("Error building layout png: %s\n", err)
		return nil, err
	}

	t.FilePath = filePath
	return tgbotapi.FilePath(t.FilePath), nil
}

func (t Thumbnail) buildFilePath() string {
	return fmt.Sprintf("%s/%s.png", thumbnailsDir, t.Id)
}

func (t Thumbnail) fetchAIWData() (layout.AIW, error) {
	url := t.requestUrl()
	response, err := http.Get(url)
	if err != nil {
		log.Printf("Error http-getting aiw data: %s\n", err)
		return nil, err
	}
	defer response.Body.Close()

	var layoutData layout.AIW
	err = json.NewDecoder(response.Body).Decode(&layoutData)
	return layoutData, err
}
