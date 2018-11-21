package api

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

type InstallationsServiceOutput struct {
	ID         int
	Status     string
	Logs       string
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	UserName   string     `json:"user_name"`
}

func (a Api) ListInstallations() ([]InstallationsServiceOutput, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/installations", nil)
	if err != nil {
		return []InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installations endpoint: %s", err)
	}
	defer resp.Body.Close()

	var responseStruct struct {
		Installations []InstallationsServiceOutput
	}
	err = json.NewDecoder(resp.Body).Decode(&responseStruct)
	if err != nil {
		return []InstallationsServiceOutput{}, fmt.Errorf("failed to decode response: %s", err)
	}

	return responseStruct.Installations, nil
}

func (a Api) CreateInstallation(ignoreWarnings bool, deployProducts bool, productNames []string) (InstallationsServiceOutput, error) {
	var deployProductsVal interface{} = "all"
	if !deployProducts {
		deployProductsVal = "none"
	} else if len(productNames) > 0 {
		sp, err := a.ListStagedProducts()
		if err != nil {
			return InstallationsServiceOutput{}, fmt.Errorf("failed to list staged products: %v", err)
		}
		// convert list of product names to product GUIDs
		var productGUIDs []string
		for _, productName := range productNames {
			var guid string
			for _, stagedProduct := range sp.Products {
				if productName == stagedProduct.GUID {
					guid = stagedProduct.GUID
					break
				} else if productName == stagedProduct.Type {
					guid = stagedProduct.GUID
					break
				}
			}
			if guid != "" {
				productGUIDs = append(productGUIDs, guid)
			}
		}
		deployProductsVal = productGUIDs
	}

	data, err := json.Marshal(&struct {
		IgnoreWarnings string      `json:"ignore_warnings"`
		DeployProducts interface{} `json:"deploy_products"`
	}{
		IgnoreWarnings: fmt.Sprintf("%t", ignoreWarnings),
		DeployProducts: deployProductsVal,
	})
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	resp, err := a.sendAPIRequest("POST", "/api/v0/installations", data)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installations endpoint: %s", err)
	}
	defer resp.Body.Close()

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
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/installations/%d", id), nil)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installations status endpoint: %s", err)
	}
	defer resp.Body.Close()

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
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/installations/%d/logs", id), nil)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("could not make api request to installations logs endpoint: %s", err)
	}
	defer resp.Body.Close()

	var output struct {
		Logs string
	}
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return InstallationsServiceOutput{}, fmt.Errorf("failed to decode response: %s", err)
	}

	return InstallationsServiceOutput{Logs: output.Logs}, nil
}
