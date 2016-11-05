package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

type ImportInstallationInput struct {
	ContentLength int64
	Installation  io.Reader
	ContentType   string
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

func (ia InstallationAssetService) Export(outputFile string) error {
	req, err := http.NewRequest("GET", "/api/v0/installation_asset_collection", nil)
	if err != nil {
		return err
	}

	respChan := make(chan error)
	go func() {
		var elapsedTime int
		var liveLog logger
		for {
			select {
			case _ = <-respChan:
				ia.liveWriter.Stop()
				return
			default:
				if elapsedTime == 0 {
					ia.liveWriter.Start()
					liveLog = log.New(ia.liveWriter, "", 0)
				}
				time.Sleep(1 * time.Second)
				elapsedTime++
				liveLog.Printf("%ds elapsed, waiting for response from Ops Manager...\r", elapsedTime)
			}
		}
	}()

	resp, err := ia.client.Do(req)
	respChan <- err
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
	respChan := make(chan error)
	progressDone := ia.trackProgress()
	go func() {
		var elapsedTime int
		var liveLog logger
		for {
			select {
			case _ = <-respChan:
				ia.liveWriter.Stop()
				return
			case _ = <-progressDone:
				ia.progress.End()
				ia.liveWriter.Start()
				liveLog = log.New(ia.liveWriter, "", 0)
			default:
				time.Sleep(1 * time.Second)
				if liveLog != nil {
					elapsedTime++
					liveLog.Printf("%ds elapsed, waiting for response from Ops Manager...\r", elapsedTime)
				}
			}
		}
	}()

	resp, err := ia.client.Do(req)
	respChan <- err
	if err != nil {
		return fmt.Errorf("could not make api request to installation_asset_collection endpoint: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return fmt.Errorf("request failed: unexpected response:\n%s", out)
	}

	return nil
}

func (ia InstallationAssetService) Delete() (InstallationsServiceOutput, error) {
	req, err := http.NewRequest("DELETE", "/api/v0/installation_asset_collection", nil)
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := ia.client.Do(req)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installation_asset_collection endpoint: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return InstallationsServiceOutput{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return InstallationsServiceOutput{}, fmt.Errorf("request failed: unexpected response:\n%s", out)
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

func (ia InstallationAssetService) trackProgress() chan string {
	progressChan := make(chan string)
	go func() {
		for {
			if ia.progress.GetCurrent() == ia.progress.GetTotal() {
				progressChan <- "done"
				break
			}
		}
	}()
	return progressChan
}
