package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

type DiagnosticService struct {
	client httpClient
}

type DiagnosticReport struct {
	Stemcells []string
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

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return DiagnosticReport{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return DiagnosticReport{}, fmt.Errorf("request failed: unexpected response:\n%s", out)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return DiagnosticReport{}, err
	}

	var report *DiagnosticReport
	err = json.Unmarshal(body, &report)
	if err != nil {
		return DiagnosticReport{}, fmt.Errorf("invalid json received from server: %s", err)
	}

	return *report, nil
}
