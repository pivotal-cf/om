package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ImportInstallationInput struct {
	ContentLength   int64
	Installation    io.Reader
	ContentType     string
	PollingInterval int
}

func (a Api) DownloadInstallationAssetCollection(outputFile string, pollingInterval int) error {
	req, err := http.NewRequest("GET", "/api/v0/installation_asset_collection", nil)
	if err != nil {
		return err
	}

	resp, err := a.progressClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not make api request to installation_asset_collection endpoint: %s", err)
	}

	if err = validateStatusOK(resp); err != nil {
		return fmt.Errorf("request failed: unexpected response")
	}

	defer resp.Body.Close()

	outputFileHandle, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("cannot create output file: %s", err)
	}

	bytesWritten, err := io.Copy(outputFileHandle, resp.Body)
	if err != nil {
		return fmt.Errorf("cannot write output file: %s", err)
	}

	if bytesWritten != resp.ContentLength {
		return fmt.Errorf("invalid response length (expected %d, got %d)", resp.ContentLength, bytesWritten)
	}

	return nil
}

func (a Api) UploadInstallationAssetCollection(input ImportInstallationInput) error {
	req, err := http.NewRequest("POST", "/api/v0/installation_asset_collection", input.Installation)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", input.ContentType)
	req.ContentLength = input.ContentLength

	resp, err := a.unauthedProgressClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not make api request to installation_asset_collection endpoint: %s", err)
	}

	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) DeleteInstallationAssetCollection() (InstallationsServiceOutput, error) {
	req, err := http.NewRequest("DELETE", "/api/v0/installation_asset_collection", bytes.NewBuffer([]byte(`{"errands": {}}`)))
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installation_asset_collection endpoint: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusGone {
		return InstallationsServiceOutput{}, nil
	}

	if err = validateStatusOK(resp); err != nil {
		return InstallationsServiceOutput{}, err
	}

	var installation struct {
		Install struct {
			ID int
		}
	}

	err = json.NewDecoder(resp.Body).Decode(&installation)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not read response from installation_asset_collection endpoint: %s", err)
	}

	return InstallationsServiceOutput{ID: installation.Install.ID}, nil
}
