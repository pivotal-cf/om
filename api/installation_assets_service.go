package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type InstallationAssetsService struct {
	client   httpClient
	progress progress
}

func NewInstallationAssetsService(client httpClient, progress progress) InstallationAssetsService {
	return InstallationAssetsService{
		client:   client,
		progress: progress,
	}
}

func (ia InstallationAssetsService) Export(outputFile string) error {
	req, err := http.NewRequest("GET", "/api/v0/installation_asset_collection", nil)
	if err != nil {
		return err
	}

	resp, err := ia.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make api request to installation_asset_collection endpoint: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: unexpected response")
	}

	ia.progress.SetTotal(resp.ContentLength)
	ia.progress.Kickoff()
	progressReader := ia.progress.NewBarReader(resp.Body)
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(progressReader)
	if err != nil {
		return fmt.Errorf("request failed: response cannot be read")
	}

	ia.progress.End()

	err = ioutil.WriteFile(outputFile, respBody, 0644)
	if err != nil {
		return fmt.Errorf("request failed: cannot write to output file: %s", err)
	}

	return nil
}
