package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
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
	ID     int
	Status string
	Logs   string
}

func NewInstallationsService(client httpClient) InstallationsService {
	return InstallationsService{
		client: client,
	}
}

func (is InstallationsService) Trigger() (InstallationsServiceOutput, error) {
	req, err := http.NewRequest("POST", "/api/v0/installations", strings.NewReader(`{}`))
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := is.client.Do(req)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installations endpoint: %s", err)
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

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return InstallationsServiceOutput{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return InstallationsServiceOutput{}, fmt.Errorf("request failed: unexpected response:\n%s", out)
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

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return InstallationsServiceOutput{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return InstallationsServiceOutput{}, fmt.Errorf("request failed: unexpected response:\n%s", out)
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
