package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/pivotal-cf/om/common"
)

type ImportInstallationInput struct {
	ContentLength int64
	Installation  io.Reader
	ContentType   string
}

type InstallationAssetService struct {
	client   httpClient
	progress progress
	logger   common.Logger
}

func NewInstallationAssetService(client httpClient, progress progress, logger common.Logger) InstallationAssetService {
	return InstallationAssetService{
		client:   client,
		progress: progress,
		logger:   logger,
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
		for {
			select {
			case _ = <-respChan:
				return
			default:
				time.Sleep(1 * time.Second)
				elapsedTime++
				ia.logger.Printf("%ds elapsed, waiting for response from Ops Manager...\r", elapsedTime)
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
	go func() {
		var elapsedTime int
		for {
			select {
			case _ = <-respChan:
				return
			default:
				time.Sleep(1 * time.Second)
				if ia.progress.GetCurrent() == ia.progress.GetTotal() { // so that we only start logging time elapsed after the progress bar is done
					elapsedTime++
					ia.logger.Printf("%ds elapsed, waiting for response from Ops Manager...\r", elapsedTime)
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
