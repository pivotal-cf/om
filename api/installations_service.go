package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

type InstallationsService struct {
	client httpClient
}

type InstallationsServiceOutput struct {
	ID         int
	Status     string
	Logs       string
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	UserName   string     `json:"user_name"`
}

func NewInstallationsService(client httpClient) InstallationsService {
	return InstallationsService{
		client: client,
	}
}

func installationList(state string, is InstallationsService) ([]InstallationsServiceOutput, error) {
	req, err := http.NewRequest("GET", "/api/v0/installations", nil)
	if err != nil {
		return []InstallationsServiceOutput{}, err
	}

	resp, err := is.client.Do(req)
	if err != nil {
		return []InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installations endpoint: %s", err)
	}

	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return []InstallationsServiceOutput{}, err
	}

	var responseStruct struct {
		Installations []InstallationsServiceOutput
	}
	err = json.NewDecoder(resp.Body).Decode(&responseStruct)
	if err != nil {
		return []InstallationsServiceOutput{}, fmt.Errorf("failed to decode response: %s", err)
	}

	var returnStruct struct {
		Installations []InstallationsServiceOutput
	}

	if state == "" {
		return responseStruct.Installations, nil
	} else {
		for _, installation := range responseStruct.Installations {
			if installation.Status == state {
				returnStruct.Installations = append(returnStruct.Installations, installation)
			}
		}
	}

	return returnStruct.Installations, nil
}

func (is InstallationsService) RunningInstallation() (InstallationsServiceOutput, error) {
	installationOutput, err := installationList(StatusRunning, is)
	if len(installationOutput) > 0 {
		return installationOutput[0], err
	} else {
		return InstallationsServiceOutput{}, err
	}
}

func (is InstallationsService) ListInstallations() ([]InstallationsServiceOutput, error) {
	return installationList("", is)
}

func (is InstallationsService) Trigger(ignoreWarnings bool) (InstallationsServiceOutput, error) {
	ignore := fmt.Sprintf(`{"ignore_warnings": "%t"}`, ignoreWarnings)

	req, err := http.NewRequest("POST", "/api/v0/installations", strings.NewReader(ignore))
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := is.client.Do(req)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installations endpoint: %s", err)
	}

	defer resp.Body.Close()

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
		return InstallationsServiceOutput{}, fmt.Errorf("failed to decode response: %s", err)
	}

	return InstallationsServiceOutput{ID: installation.Install.ID}, nil
}

func (is InstallationsService) Status(id int) (InstallationsServiceOutput, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v0/installations/%d", id), nil)
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := is.client.Do(req)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installations status endpoint: %s", err)
	}

	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
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

func (is InstallationsService) Logs(id int) (InstallationsServiceOutput, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v0/installations/%d/logs", id), nil)
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := is.client.Do(req)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installations logs endpoint: %s", err)
	}

	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
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
