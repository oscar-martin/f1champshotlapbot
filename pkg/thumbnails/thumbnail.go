package thumbnails

import (
	"context"
	"encoding/json"
	"f1champshotlapsbot/pkg/layout"
	"fmt"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	thumbnailsDir          = "./thumbnails"
	PubSubThumbnailPreffix = "thumbnail_"
)

func init() {
	// create thumbnail dir if not exists
	if _, err := os.Stat(thumbnailsDir); os.IsNotExist(err) {
		os.Mkdir(thumbnailsDir, 0755)
	}
}

type pngBuilder func(ctx context.Context, url, filePath string) error

type Thumbnail struct {
	Id         string
	FilePath   string
	ServerUrl  string
	Endpoint   string
	pngBuilder pngBuilder
	prefix     string
}

func BuildTrackThumbnail(ctx context.Context, url, id string) (Thumbnail, error) {
	th := Thumbnail{
		Id:         id,
		Endpoint:   "rest/race/track/%s/trackmap",
		ServerUrl:  url,
		pngBuilder: pngBuilderForTrack,
		prefix:     "track_",
	}

	return th, th.build(ctx)
}

func BuildSmallCarThumbnail(ctx context.Context, url, id string) (Thumbnail, error) {
	th := Thumbnail{
		Id:         id,
		Endpoint:   "rest/race/car/%s/image?type=IMAGE_SMALL",
		ServerUrl:  url,
		pngBuilder: pngBuilderForCar,
		prefix:     "car_",
	}

	return th, th.build(ctx)
}

func BuildCarThumbnail(ctx context.Context, url, id string) (Thumbnail, error) {
	th := Thumbnail{
		Id:         id,
		Endpoint:   "rest/race/car/%s/image",
		ServerUrl:  url,
		pngBuilder: pngBuilderForCar,
		prefix:     "car_",
	}

	return th, th.build(ctx)
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

func (t *Thumbnail) build(ctx context.Context) error {
	if t.Id == "" {
		return fmt.Errorf("thumbnail is not initialized")
	}
	filePath := t.buildFilePath()
	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("thumbnail for %q already exists\n", t.Id)
		t.FilePath = filePath
	} else if os.IsNotExist(err) {
		err := t.pngBuilder(ctx, t.requestUrl(), filePath)
		if err != nil {
			log.Printf("Error building png: %s\n", err)
			return err
		}
	} else {
		return err
	}

	t.FilePath = filePath
	return nil
}

func (t *Thumbnail) FileData() (tgbotapi.RequestFileData, error) {
	if t.Id == "" || t.FilePath == "" {
		return nil, fmt.Errorf("thumbnail is not initialized")
	}

	return tgbotapi.FilePath(t.FilePath), nil
}

func (t Thumbnail) buildFilePath() string {
	return fmt.Sprintf("%s/%s%s.png", thumbnailsDir, t.prefix, t.Id)
}

func pngBuilderForTrack(ctx context.Context, url, filePath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error getting car image: %s (%s)", response.Status, url)
	}

	var layoutData layout.AIW
	err = json.NewDecoder(response.Body).Decode(&layoutData)
	if err != nil {
		return err
	}
	err = layout.BuildLayoutPNG(filePath, layoutData)
	if err != nil {
		return err
	}
	return nil
}

// create a pngBuilder from a http call
func pngBuilderForCar(ctx context.Context, url, filePath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error getting car image: %s (%s)", response.Status, url)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.ReadFrom(response.Body)
	if err != nil {
		return err
	}
	return nil
}
