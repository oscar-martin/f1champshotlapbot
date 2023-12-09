package servers

import (
	"context"
	"encoding/json"
	"f1champshotlapsbot/pkg/model"
	"f1champshotlapsbot/pkg/resources"
	"fmt"
	"log"
	"net/http"
	"time"
)

func getSelectedSessionData(serverUrl string) (model.SelectedSessionData, error) {
	var selectedSessionData model.SelectedSessionData
	url := fmt.Sprintf("%s/rest/race/selection", serverUrl)
	response, err := http.Get(url)
	if err != nil {
		log.Printf("Error http-getting selected session: %s\n", err)
		return selectedSessionData, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return selectedSessionData, fmt.Errorf("error getting selected session: %s", response.Status)
	}

	err = json.NewDecoder(response.Body).Decode(&selectedSessionData)
	if err != nil {
		log.Printf("Error decoding selected session: %s\n", err)
		return selectedSessionData, err
	}
	return selectedSessionData, nil
}

func buildTrackThumbnail(serverUrl string, selectedSessionData model.SelectedSessionData) (resources.Resource, error) {
	t := resources.Resource{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	thChan := make(chan resources.Resource)
	errChan := make(chan error)
	go func() {
		th, err := resources.BuildTrackThumbnail(ctx, serverUrl, selectedSessionData.Track.ID)
		if err != nil {
			errChan <- err
			return
		}
		thChan <- th
	}()
	select {
	case <-ctx.Done():
		return t, ctx.Err()
	case err := <-errChan:
		return t, err
	case th := <-thChan:
		return th, nil
	}
}

func buildTrackSvg(serverUrl string, selectedSessionData model.SelectedSessionData) (resources.Resource, error) {
	t := resources.Resource{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	thChan := make(chan resources.Resource)
	errChan := make(chan error)
	go func() {
		th, err := resources.BuildTrackSvg(ctx, serverUrl, selectedSessionData.Track.ID)
		if err != nil {
			errChan <- err
			return
		}
		thChan <- th
	}()
	select {
	case <-ctx.Done():
		return t, ctx.Err()
	case err := <-errChan:
		return t, err
	case th := <-thChan:
		return th, nil
	}
}

func retryWithCancel(f func() error, cancel chan bool) {
	err := f()
	if err == nil {
		return
	}
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := f()
			if err != nil {
				continue
			}
			return
		case <-cancel:
			return
		}
	}
}
