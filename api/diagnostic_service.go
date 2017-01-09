package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type DiagnosticService struct {
	client httpClient
}

type DiagnosticProduct struct {
	Name     string
	Version  string
	Stemcell string
}

type DiagnosticReport struct {
	Stemcells      []string
	StagedProducts []DiagnosticProduct
}

type DiagnosticReportUnavailable struct{}

func (du DiagnosticReportUnavailable) Error() string {
	return "diagnostic report is currently unavailable"
}

func NewDiagnosticService(client httpClient) DiagnosticService {
	return DiagnosticService{
		client: client,
	}
}

func (ds DiagnosticService) Report() (DiagnosticReport, error) {
	req, err := http.NewRequest("GET", "/api/v0/diagnostic_report", nil)
	if err != nil {
		return DiagnosticReport{}, err
	}

	resp, err := ds.client.Do(req)
	if err != nil {
		return DiagnosticReport{}, fmt.Errorf("could not make api request to diagnostic_report endpoint: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		return DiagnosticReport{}, DiagnosticReportUnavailable{}
	}

	if err = ValidateStatusOK(resp); err != nil {
		return DiagnosticReport{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return DiagnosticReport{}, err
	}

	var apiResponse *struct {
		DiagnosticReport
		AddedProducts struct {
			StagedProducts []DiagnosticProduct `json:"staged"`
		} `json:"added_products"`
	}

	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		return DiagnosticReport{}, fmt.Errorf("invalid json received from server: %s", err)
	}

	return DiagnosticReport{
		Stemcells:      apiResponse.Stemcells,
		StagedProducts: apiResponse.AddedProducts.StagedProducts,
	}, nil
}
