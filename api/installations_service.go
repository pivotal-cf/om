package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
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

type ApplyErrandChanges struct {
	Errands map[string]ProductErrand `yaml:"errands" json:"errands,omitempty"`
}

type ProductErrand struct {
	RunPostDeploy map[string]interface{} `yaml:"run_post_deploy" json:"run_post_deploy,omitempty"`
	RunPreDelete  map[string]interface{} `yaml:"run_pre_delete" json:"run_pre_delete,omitempty"`
}

func (a Api) fetchProductGUID() (map[string]string, error) {
	productGuidMapping := map[string]string{}
	sp, err := a.ListStagedProducts()
	if err != nil {
		return productGuidMapping, err
	}
	dp, err := a.ListDeployedProducts()
	if err != nil {
		return productGuidMapping, err
	}

	for _, stagedProduct := range sp.Products {
		productGuidMapping[stagedProduct.Type] = stagedProduct.GUID
	}
	for _, deployedProduct := range dp {
		productGuidMapping[deployedProduct.Type] = deployedProduct.GUID
	}

	return productGuidMapping, nil
}

func (a Api) ListInstallations() ([]InstallationsServiceOutput, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/installations", nil)
	if err != nil {
		return []InstallationsServiceOutput{}, errors.Wrap(err, "could not make api request to installations endpoint")
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
		return []InstallationsServiceOutput{}, errors.Wrap(err, "failed to decode response")
	}

	return responseStruct.Installations, nil
}

func (a Api) CreateInstallation(ignoreWarnings bool, deployProducts bool, productNames []string, errands ApplyErrandChanges) (InstallationsServiceOutput, error) {
	productGuidMapping, err := a.fetchProductGUID()
	if err != nil {
		return InstallationsServiceOutput{}, errors.Wrap(err, "failed to list staged and/or deployed products")
	}

	var deployProductsVal interface{} = "all"
	if !deployProducts {
		deployProductsVal = "none"
	} else if len(productNames) > 0 {
		var productGUIDs []string
		for _, productName := range productNames {
			if productGUID, ok := productGuidMapping[productName]; ok {
				productGUIDs = append(productGUIDs, productGUID)
			} else {
				return InstallationsServiceOutput{}, fmt.Errorf("failed to fetch product GUID for product: %s", productName)
			}
		}
		deployProductsVal = productGUIDs
	}

	errandsPayload := map[string]ProductErrand{}

	for productName, errandConfig := range errands.Errands {
		if productGUID, ok := productGuidMapping[productName]; ok {
			errandsPayload[productGUID] = errandConfig
		} else {
			return InstallationsServiceOutput{}, fmt.Errorf("failed to fetch product GUID for product: %s", productName)
		}
	}

	data, err := json.Marshal(&struct {
		IgnoreWarnings string                   `json:"ignore_warnings"`
		DeployProducts interface{}              `json:"deploy_products"`
		Errands        map[string]ProductErrand `json:"errands,omitempty"`
	}{
		IgnoreWarnings: fmt.Sprintf("%t", ignoreWarnings),
		DeployProducts: deployProductsVal,
		Errands:        errandsPayload,
	})
	if err != nil {
		return InstallationsServiceOutput{}, err
	}

	resp, err := a.sendAPIRequest("POST", "/api/v0/installations", data)
	if err != nil {
		return InstallationsServiceOutput{}, errors.Wrap(err, "could not make api request to installations endpoint")
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
		return InstallationsServiceOutput{}, errors.Wrap(err, "failed to decode response")
	}

	return InstallationsServiceOutput{ID: installation.Install.ID}, nil
}

func (a Api) GetInstallation(id int) (InstallationsServiceOutput, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/installations/%d", id), nil)
	if err != nil {
		return InstallationsServiceOutput{}, errors.Wrap(err, "could not make api request to installations status endpoint")
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
		return InstallationsServiceOutput{}, errors.Wrap(err, "failed to decode response")
	}

	return InstallationsServiceOutput{Status: output.Status}, nil
}

func (a Api) GetInstallationLogs(id int) (InstallationsServiceOutput, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/installations/%d/logs", id), nil)
	if err != nil {
		return InstallationsServiceOutput{}, errors.Wrap(err, "could not make api request to installations logs endpoint")
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
		return InstallationsServiceOutput{}, errors.Wrap(err, "failed to decode response")
	}

	return InstallationsServiceOutput{Logs: output.Logs}, nil
}
