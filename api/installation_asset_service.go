package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

type ImportInstallationInput struct {
	ContentLength   int64
	Installation    io.Reader
	ContentType     string
	PollingInterval int
}

func (a Api) DownloadInstallationAssetCollection(outputFile string) error {
	resp, err := a.sendProgressAPIRequest("GET", "/api/v0/installation_asset_collection", nil)
	if err != nil {
		return errors.Wrap(err, "could not make api request to installation_asset_collection endpoint")
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	outputFileHandle, err := os.Create(outputFile)
	if err != nil {
		return errors.Wrap(err, "cannot create output file")
	}

	bytesWritten, err := io.Copy(outputFileHandle, resp.Body)
	if err != nil {
		return errors.Wrap(err, "cannot write output file")
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
		return errors.Wrap(err, "could not make api request to installation_asset_collection endpoint")
	}

	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) DeleteInstallationAssetCollection() (InstallationsServiceOutput, error) {
	resp, err := a.sendAPIRequest("DELETE", "/api/v0/installation_asset_collection", []byte(`{"errands": {}}`))
	if err != nil {
		return InstallationsServiceOutput{}, errors.Wrap(err, "could not make api request to installation_asset_collection endpoint")
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
		return InstallationsServiceOutput{}, errors.Wrap(err, "could not read response from installation_asset_collection endpoint")
	}

	return InstallationsServiceOutput{ID: installation.Install.ID}, nil
}
