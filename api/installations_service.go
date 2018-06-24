package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/donovanhide/eventsource"
	"io"
	"net/http"
	"time"
)

const (
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

var InstallFailed = errors.New("installation was unsuccessful")

type InstallationsServiceOutput struct {
	ID         int
	Status     string
	Logs       string
	LogChan    chan string
	ErrorChan  chan error
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	UserName   string     `json:"user_name"`
}

func (a Api) ListInstallations() ([]InstallationsServiceOutput, error) {
	req, err := http.NewRequest("GET", "/api/v0/installations", nil)
	if err != nil {
		return []InstallationsServiceOutput{}, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return []InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installations endpoint: %s", err)
	}

	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return []InstallationsServiceOutput{}, err
	}

	var responseStruct struct {
		Installations []InstallationsServiceOutput
	}
	err = json.NewDecoder(resp.Body).Decode(&responseStruct)
	if err != nil {
		return []InstallationsServiceOutput{}, fmt.Errorf("failed to decode response: %s", err)
	}

	return responseStruct.Installations, nil
}

func (a Api) CreateInstallation(ignoreWarnings bool, deployProducts bool) (InstallationsServiceOutput, error) {
	deployProductsVal := "none"
	if deployProducts {
		deployProductsVal = "all"
	}

	data, err := json.Marshal(&struct {
		IgnoreWarnings string `json:"ignore_warnings"`
		DeployProducts string `json:"deploy_products"`
	}{
		IgnoreWarnings: fmt.Sprintf("%t", ignoreWarnings),
		DeployProducts: deployProductsVal,
	})
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	req, err := http.NewRequest("POST", "/api/v0/installations", bytes.NewReader(data))
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installations endpoint: %s", err)
	}

	defer resp.Body.Close()

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
		return InstallationsServiceOutput{}, fmt.Errorf("failed to decode response: %s", err)
	}

	return InstallationsServiceOutput{ID: installation.Install.ID}, nil
}

func (a Api) GetInstallation(id int) (InstallationsServiceOutput, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v0/installations/%d", id), nil)
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installations status endpoint: %s", err)
	}

	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return InstallationsServiceOutput{}, err
	}

	var output struct {
		Status string
	}
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("failed to decode response: %s", err)
	}

	return InstallationsServiceOutput{Status: output.Status}, nil
}

func (a Api) GetInstallationLogs(id int) (InstallationsServiceOutput, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v0/installations/%d/logs", id), nil)
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installations logs endpoint: %s", err)
	}

	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return InstallationsServiceOutput{}, err
	}

	var output struct {
		Logs string
	}
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("failed to decode response: %s", err)
	}

	return InstallationsServiceOutput{Logs: output.Logs}, nil
}

func (a Api) GetCurrentInstallationLogs() (InstallationsServiceOutput, error) {
	req, err := http.NewRequest("GET", "/api/v0/installations/current_log", nil)
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := a.client.Do(req)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to current installation logs endpoint: %s", err)
	}

	if err = validateStatusOK(resp); err != nil {
		return InstallationsServiceOutput{}, err
	}

	output := InstallationsServiceOutput{}
	output.LogChan = make(chan string)
	output.ErrorChan = make(chan error)

	go writeLog(resp.Body, output.LogChan, output.ErrorChan)

	return output, nil
}

func receiveEvents(r io.ReadCloser, events chan eventsource.Event, errors chan error) {
	defer r.Close()
	defer close(events)
	defer close(errors)

	decoder := eventsource.NewDecoder(r)
	for {
		ev, err := decoder.Decode()
		if err != nil {
			errors <- err
			return
		}
		events <- ev
	}
}

func writeLog(r io.ReadCloser, logChan chan string, errChan chan error) {
	type exitStatus struct {
		Code int `json:"code"`
	}

	eventChan := make(chan eventsource.Event)
	eventErrChan := make(chan error)

	go receiveEvents(r, eventChan, eventErrChan)

	for {
		select {
		case event := <-eventChan:
			// maybe handle more event types
			// https://pcf.pcf-installer.pcf-installer.norm.cf-app.com/docs#streaming-current-installation-log
			switch event.Event() {
			case "":
				logChan <- event.Data()
			case "exit":
				close(logChan)
				status := exitStatus{}
				if err := json.Unmarshal([]byte(event.Data()), &status); err != nil || status.Code != 0 {
					errChan <- InstallFailed
				}
				close(errChan)
				return
			}
		case err := <-eventErrChan:
			close(logChan)
			errChan <- fmt.Errorf("installation failed to stream logs: %s", err)
			close(errChan)
			return
		}

	}
}
