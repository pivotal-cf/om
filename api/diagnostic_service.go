package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type DiagnosticProduct struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Stemcell string `json:"stemcell,omitempty"`
}

type DiagnosticReport struct {
	InfrastructureType string `json:"infrastructure_type"`
	Stemcells          []string
	StagedProducts     []DiagnosticProduct
	DeployedProducts   []DiagnosticProduct
}

type DiagnosticReportUnavailable struct{}

func (du DiagnosticReportUnavailable) Error() string {
	return "diagnostic report is currently unavailable"
}

func (a Api) GetDiagnosticReport() (DiagnosticReport, error) {
	req, err := http.NewRequest("GET", "/api/v0/diagnostic_report", nil)
	if err != nil {
		return DiagnosticReport{}, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return DiagnosticReport{}, fmt.Errorf("could not make api request to diagnostic_report endpoint: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		return DiagnosticReport{}, DiagnosticReportUnavailable{}
	}

	if err = validateStatusOK(resp); err != nil {
		return DiagnosticReport{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return DiagnosticReport{}, err
	}

	var apiResponse *struct {
		DiagnosticReport
		AddedProducts struct {
			StagedProducts   []DiagnosticProduct `json:"staged"`
			DeployedProducts []DiagnosticProduct `json:"deployed"`
		} `json:"added_products"`
	}

	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		return DiagnosticReport{}, fmt.Errorf("invalid json received from server: %s", err)
	}

	return DiagnosticReport{
		InfrastructureType: apiResponse.DiagnosticReport.InfrastructureType,
		Stemcells:          apiResponse.Stemcells,
		StagedProducts:     apiResponse.AddedProducts.StagedProducts,
		DeployedProducts:   apiResponse.AddedProducts.DeployedProducts,
	}, nil
}
