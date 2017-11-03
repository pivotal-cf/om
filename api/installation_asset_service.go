package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type ImportInstallationInput struct {
	ContentLength   int64
	Installation    io.Reader
	ContentType     string
	PollingInterval int
}

type InstallationAssetService struct {
	client     httpClient
	progress   progress
	liveWriter liveWriter
}

type logger interface {
	Printf(format string, v ...interface{})
}

//go:generate counterfeiter -o ./fakes/livewriter.go --fake-name LiveWriter . liveWriter
type liveWriter interface {
	io.Writer
	Start()
	Stop()
}

func NewInstallationAssetService(client httpClient, progress progress, liveWriter liveWriter) InstallationAssetService {
	return InstallationAssetService{
		client:     client,
		progress:   progress,
		liveWriter: liveWriter,
	}
}

func (ia InstallationAssetService) Export(outputFile string, pollingInterval int) error {
	req, err := http.NewRequest("GET", "/api/v0/installation_asset_collection", nil)
	if err != nil {
		return err
	}

	requestComplete := make(chan bool)
	progressComplete := make(chan bool)

	go func() {
		ia.liveWriter.Start()

		liveLog := log.New(ia.liveWriter, "", 0)
		startTime := time.Now().Round(time.Second)
		ticker := time.NewTicker(time.Duration(pollingInterval) * time.Second)

		for {
			select {
			case <-requestComplete:
				ticker.Stop()
				ia.liveWriter.Stop()
				progressComplete <- true
			case now := <-ticker.C:
				liveLog.Printf("%s elapsed, waiting for response from Ops Manager...\r", now.Round(time.Second).Sub(startTime).String())
			}
		}
	}()

	resp, err := ia.client.Do(req)
	requestComplete <- true
	<-progressComplete

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

	outputFileHandle, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("cannot create output file: %s", err)
	}

	bytesWritten, err := io.Copy(outputFileHandle, progressReader)
	if err != nil {
		return fmt.Errorf("cannot write output file: %s", err)
	}

	if bytesWritten != resp.ContentLength {
		return fmt.Errorf("invalid response length (expected %d, got %d)", resp.ContentLength, bytesWritten)
	}

	ia.progress.End()

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

	requestComplete := make(chan bool)
	progressComplete := make(chan bool)

	go func() {
		ia.progress.Kickoff()
		ia.liveWriter.Start()

		for {
			if ia.progress.GetCurrent() == ia.progress.GetTotal() {
				ia.progress.End()
				break
			}

			time.Sleep(time.Second)
		}

		liveLog := log.New(ia.liveWriter, "", 0)
		startTime := time.Now().Round(time.Second)
		ticker := time.NewTicker(time.Duration(input.PollingInterval) * time.Second)

		for {
			select {
			case <-requestComplete:
				ticker.Stop()
				ia.liveWriter.Stop()
				progressComplete <- true
			case now := <-ticker.C:
				liveLog.Printf("%s elapsed, waiting for response from Ops Manager...\r", now.Round(time.Second).Sub(startTime).String())
			}
		}
	}()

	resp, err := ia.client.Do(req)
	requestComplete <- true
	<-progressComplete

	if err != nil {
		return fmt.Errorf("could not make api request to installation_asset_collection endpoint: %s", err)
	}

	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (ia InstallationAssetService) Delete() (InstallationsServiceOutput, error) {
	req, err := http.NewRequest("DELETE", "/api/v0/installation_asset_collection", bytes.NewBuffer([]byte(`{"errands": {}}`)))
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := ia.client.Do(req)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installation_asset_collection endpoint: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusGone {
		return InstallationsServiceOutput{}, nil
	}

	if err = ValidateStatusOK(resp); err != nil {
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
