package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

type ImportInstallationInput struct {
	ContentLength int64
	Installation  io.Reader
	ContentType   string
}

type InstallationAssetService struct {
	client   httpClient
	progress progress
}

func NewInstallationAssetService(client httpClient, progress progress) InstallationAssetService {
	return InstallationAssetService{
		client:   client,
		progress: progress,
	}
}

func (ia InstallationAssetService) Export(outputFile string) error {
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

func (ia InstallationAssetService) Import(input ImportInstallationInput) error {
	ia.progress.SetTotal(input.ContentLength)
	body := ia.progress.NewBarReader(input.Installation)

	req, err := http.NewRequest("POST", "/api/v0/installation_asset_collection", body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", input.ContentType)
	req.ContentLength = input.ContentLength

	ia.progress.Kickoff()

	resp, err := ia.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make api request to installation_asset_collection endpoint: %s", err)
	}

	defer resp.Body.Close()

	ia.progress.End()

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return fmt.Errorf("request failed: unexpected response:\n%s", out)
	}

	return nil
}
