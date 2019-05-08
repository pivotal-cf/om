package api

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

const preDeployDirectorEndpoint = "/api/v0/staged/director/pre_deploy_check"
const stagedProductsEndpoint = "/api/v0/staged/products"
const preDeployProductEndpointTemplate = "/api/v0/staged/products/%s/pre_deploy_check"

type PreDeployNetwork struct {
	Assigned bool `json:"assigned"`
}

type PreDeployAvailabilityZone struct {
	Assigned bool `json:"assigned"`
}

type PreDeployStemcells struct {
	Assigned                bool   `json:"assigned"`
	RequiredStemcellVersion string `json:"required_stemcell_version"`
	RequiredStemcellOS      string `json:"required_stemcell_os"`
}

type PreDeployProperty struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Errors  []string          `json:"errors"`
	Records []PreDeployRecord `json:"records"`
}

type PreDeployRecord struct {
	Index  int                 `json:"index"`
	Errors []PreDeployProperty `json:"errors"`
}

type PreDeployResources struct {
	Jobs []PreDeployJob `json:"jobs"`
}

type PreDeployJob struct {
	Identifier string   `json:"identifier"`
	GUID       string   `json:"guid"`
	Errors     []string `json:"error"`
}

type PreDeployVerifier struct {
	Type      string   `json:"type"`
	Errors    []string `json:"errors"`
	Ignorable bool     `json:"ignorable"`
}

type PreDeployCheck struct {
	Identifier       string                    `json:"identifier"`
	Complete         bool                      `json:"complete"`
	Network          PreDeployNetwork          `json:"network"`
	AvailabilityZone PreDeployAvailabilityZone `json:"availability_zone"`
	Stemcells        []PreDeployStemcells      `json:"stemcells"`
	Properties       []PreDeployProperty       `json:"properties"`
	Resources        PreDeployResources        `json:"resources"`
	Verifiers        []PreDeployVerifier       `json:"verifiers"`
}

type PendingDirectorChangesOutput struct {
	EndpointResults PreDeployCheck `json:"pre_deploy_check"`
}

type PendingProductChangesOutput struct {
	EndpointResults PreDeployCheck `json:"pre_deploy_check"`
}

func (a Api) ListPendingDirectorChanges() (PendingDirectorChangesOutput, error) {
	resp, err := a.sendAPIRequest("GET", preDeployDirectorEndpoint, nil)
	if err != nil {
		return PendingDirectorChangesOutput{}, errors.Wrap(err, "could not make api request to pre_deploy_check endpoint")
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return PendingDirectorChangesOutput{}, err
	}

	reportBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return PendingDirectorChangesOutput{}, err
	}

	var pendingDirectorChanges PendingDirectorChangesOutput
	if err := json.Unmarshal(reportBytes, &pendingDirectorChanges); err != nil {
		return PendingDirectorChangesOutput{}, errors.Wrap(err, "could not unmarshal pre_deploy_check response")
	}

	return pendingDirectorChanges, nil
}

func (a Api) ListAllPendingProductChanges() ([]PendingProductChangesOutput, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf(stagedProductsEndpoint), nil)
	if err != nil {
		return []PendingProductChangesOutput{}, errors.Wrap(err, "could not make api request to pre_deploy_check endpoint")
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return []PendingProductChangesOutput{}, err
	}

	var stagedProducts []struct {
		GUID string `json:"guid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stagedProducts); err != nil {
		return []PendingProductChangesOutput{}, errors.Wrap(err, "could not unmarshal pre_deploy_check response")
	}

	var allPendingProductChanges []PendingProductChangesOutput
	for _, sp := range stagedProducts {
		var resp *http.Response
		resp, err = a.sendAPIRequest("GET", fmt.Sprintf(preDeployProductEndpointTemplate, sp.GUID), nil)

		if err != nil {
			return []PendingProductChangesOutput{}, errors.Wrap(err, "could not make api request to pre_deploy_check endpoint")
		}

		if err = validateStatusOK(resp); err != nil {
			return []PendingProductChangesOutput{}, err
		}

		var pendingProductChanges PendingProductChangesOutput
		if err := json.NewDecoder(resp.Body).Decode(&pendingProductChanges); err != nil {
			return []PendingProductChangesOutput{}, errors.Wrap(err, "could not unmarshal pre_deploy_check response")
		}
		allPendingProductChanges = append(allPendingProductChanges, pendingProductChanges)
		_ = resp.Body.Close()
	}

	return allPendingProductChanges, nil
}
