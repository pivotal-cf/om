package api

import (
	"encoding/json"
	"fmt"
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
	resp, err := a.sendAPIRequest("GET", "/api/v0/diagnostic_report", nil)
	if err != nil {
		if resp.StatusCode == http.StatusInternalServerError {
			return DiagnosticReport{}, DiagnosticReportUnavailable{}
		}
		return DiagnosticReport{}, fmt.Errorf("could not make api request to diagnostic_report endpoint: %s", err)
	}
	defer resp.Body.Close()

	var apiResponse *struct {
		DiagnosticReport
		AddedProducts struct {
			StagedProducts   []DiagnosticProduct `json:"staged"`
			DeployedProducts []DiagnosticProduct `json:"deployed"`
		} `json:"added_products"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return DiagnosticReport{}, fmt.Errorf("invalid json received from server: %s", err)
	}

	return DiagnosticReport{
		InfrastructureType: apiResponse.DiagnosticReport.InfrastructureType,
		Stemcells:          apiResponse.Stemcells,
		StagedProducts:     apiResponse.AddedProducts.StagedProducts,
		DeployedProducts:   apiResponse.AddedProducts.DeployedProducts,
	}, nil
}
