package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type DiagnosticProduct struct {
	Name      string     `json:"name"`
	Version   string     `json:"version"`
	Stemcell  string     `json:"stemcell,omitempty"`
	Stemcells []Stemcell `json:"stemcells,omitempty"`
}
type DiagnosticReport struct {
	InfrastructureType string   `json:"infrastructure_type"`
	Stemcells          []string `json:"stemcells,omitempty"`
	StagedProducts     []DiagnosticProduct
	DeployedProducts   []DiagnosticProduct
	AvailableStemcells []Stemcell `json:"available_stemcells,omitempty"`
	FullReport         string
}

type Stemcell struct {
	Filename string
	OS       string
	Version  string
}

type DiagnosticReportUnavailable struct{}

func (du DiagnosticReportUnavailable) Error() string {
	return "diagnostic report is currently unavailable"
}

func (a Api) GetDiagnosticReport() (DiagnosticReport, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/diagnostic_report", nil)
	if err != nil {
		return DiagnosticReport{}, fmt.Errorf("could not make api request to diagnostic_report endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		return DiagnosticReport{}, DiagnosticReportUnavailable{}
	}

	if err = validateStatusOK(resp); err != nil {
		return DiagnosticReport{}, err
	}

	var apiResponse *struct {
		DiagnosticReport
		AddedProducts struct {
			StagedProducts   []DiagnosticProduct `json:"staged"`
			DeployedProducts []DiagnosticProduct `json:"deployed"`
		} `json:"added_products"`
	}

	reportBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(reportBytes, &apiResponse); err != nil {
		return DiagnosticReport{}, fmt.Errorf("invalid json received from server: %w", err)
	}

	return DiagnosticReport{
		InfrastructureType: apiResponse.DiagnosticReport.InfrastructureType,
		Stemcells:          apiResponse.Stemcells,
		StagedProducts:     apiResponse.AddedProducts.StagedProducts,
		DeployedProducts:   apiResponse.AddedProducts.DeployedProducts,
		AvailableStemcells: apiResponse.AvailableStemcells,
		FullReport:         string(reportBytes),
	}, nil
}
