package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type InstallationService struct {
	client   httpClient
	progress progress
}

func NewInstallationService(client httpClient, progress progress) InstallationService {
	return InstallationService{
		client:   client,
		progress: progress,
	}
}

func (is InstallationService) Export(outputFile string) error {
	req, err := http.NewRequest("GET", "/api/v0/installation_asset_collection", nil)
	if err != nil {
		return err
	}

	resp, err := is.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make api request to installation_asset_collection endpoint: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: unexpected response")
	}

	is.progress.SetTotal(resp.ContentLength)
	is.progress.Kickoff()
	progressReader := is.progress.NewBarReader(resp.Body)

	respBody, err := ioutil.ReadAll(progressReader)
	if err != nil {
		return fmt.Errorf("request failed: response cannot be read") //Can't test drive
	}

	is.progress.End()

	err = ioutil.WriteFile(outputFile, respBody, 0644)
	if err != nil {
		return fmt.Errorf("request failed: cannot write to output file: %s", err)
	}

	return nil
}
