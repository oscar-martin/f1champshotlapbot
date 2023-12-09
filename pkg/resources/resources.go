package resources

import (
	"context"
	"encoding/json"
	"f1champshotlapsbot/pkg/layout"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	ResourcesDir = "./resources"
)

func init() {
	// create thumbnail dir if not exists
	if _, err := os.Stat(ResourcesDir); os.IsNotExist(err) {
		os.Mkdir(ResourcesDir, 0755)
	}
}

type builder func(ctx context.Context, url, filePath string) error

type Resource struct {
	id        string
	serverUrl string
	endpoint  string
	builder   builder
	prefix    string
	suffix    string
	_type     string
}

func BuildTrackThumbnail(ctx context.Context, url, id string) (Resource, error) {
	r := Resource{
		id:        id,
		endpoint:  "rest/race/track/%s/trackmap",
		serverUrl: url,
		builder:   pngBuilderForTrack,
		prefix:    "track_",
		suffix:    ".png",
		_type:     "track",
	}

	return r.build(ctx, id)
}

func BuildTrackSvg(ctx context.Context, url, id string) (Resource, error) {
	r := Resource{
		endpoint:  "rest/race/track/%s/trackmap",
		serverUrl: url,
		builder:   svgBuilderForTrack,
		prefix:    "track_",
		suffix:    ".svg",
		_type:     "svg-track",
	}

	return r.build(ctx, id)
}

func BuildSmallCarThumbnail(ctx context.Context, url, id string) (Resource, error) {
	r := Resource{
		endpoint:  "rest/race/car/%s/image?type=IMAGE_SMALL",
		serverUrl: url,
		builder:   pngBuilderForCar,
		prefix:    "car_",
		suffix:    ".png",
		_type:     "small_car",
	}

	return r.build(ctx, id)
}

func BuildCarThumbnail(ctx context.Context, url, id string) (Resource, error) {
	r := Resource{
		endpoint:  "rest/race/car/%s/image",
		serverUrl: url,
		builder:   pngBuilderForCar,
		prefix:    "car_",
		suffix:    ".png",
		_type:     "car",
	}

	return r.build(ctx, id)
}

func (r Resource) buildFilePath(id string) string {
	return fmt.Sprintf("%s/%s%s%s", ResourcesDir, r.prefix, id, r.suffix)
}

func (r Resource) IsZero() bool {
	return r.id == ""
}

func (r Resource) String() string {
	return fmt.Sprintf("ID: %s, Type: %s, ServerURL: %s", r.id, r._type, r.serverUrl)
}

func (r Resource) FilePath() string {
	return r.buildFilePath(r.id)
}

func (r Resource) FileName() string {
	return fmt.Sprintf("%s%s%s", r.prefix, r.id, r.suffix)
}

func (r Resource) requestUrl(id string) string {
	return fmt.Sprintf("%s/%s", r.serverUrl, fmt.Sprintf(r.endpoint, id))
}

func (r *Resource) build(ctx context.Context, id string) (Resource, error) {
	if id == "" {
		return *r, fmt.Errorf("id cannot be empty")
	}
	filePath := r.buildFilePath(id)
	if _, err := os.Stat(filePath); err == nil {
		log.Printf("resource for %q already exists\n", id)
	} else if os.IsNotExist(err) {
		err := r.builder(ctx, r.requestUrl(id), filePath)
		if err != nil {
			log.Printf("Error building resource: %s\n", err)
			return *r, err
		}
	} else {
		return *r, err
	}

	r.id = id
	return *r, nil
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
		return fmt.Errorf("error getting track image: %s (%s)", response.Status, url)
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

func svgBuilderForTrack(ctx context.Context, url, filePath string) error {
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
		return fmt.Errorf("error getting track image: %s (%s)", response.Status, url)
	}

	var layoutData layout.AIW
	err = json.NewDecoder(response.Body).Decode(&layoutData)
	if err != nil {
		return err
	}
	err = layout.BuildLayoutSVG(filePath, layoutData)
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
