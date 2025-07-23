package api

import (
	"encoding/json"
	"fmt"
	"time"
)

const expiringCertificatesEndpoint = "/api/v0/deployed/certificates?expires_within=%s"

type ExpiringCertificatesResponse struct {
	Certificates []ExpiringCertificate `json:"certificates"`
}

type ExpiringCertificate struct {
	Issuer                string    `json:"issuer"`
	ValidFrom             time.Time `json:"valid_from"`
	ValidUntil            time.Time `json:"valid_until"`
	Configurable          bool      `json:"configurable"`
	PropertyReference     string    `json:"property_reference"`
	PropertyType          string    `json:"property_type"`
	ProductGUID           string    `json:"product_guid"`
	Location              string    `json:"location"`
	VariablePath          string    `json:"variable_path"`
	RotationProcedureName string    `json:"rotation_procedure_name"`
	RotationProcedureUrl  string    `json:"rotation_procedure_url"`
}

func (a Api) ListExpiringCertificates(expiresWithin string) ([]ExpiringCertificate, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf(expiringCertificatesEndpoint, expiresWithin), nil)
	if err != nil {
		return nil, fmt.Errorf("could not make api request to certificates endpoint: %w", err)
	}

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	var expiringCertificatesResponse ExpiringCertificatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&expiringCertificatesResponse); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return expiringCertificatesResponse.Certificates, nil
}

type DeployedCertificatesResponse struct {
	Certificates []ExpiringCertificate `json:"certificates"`
}

func (a Api) ListDeployedCertificates() ([]ExpiringCertificate, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/deployed/certificates", nil)
	if err != nil {
		return nil, fmt.Errorf("could not make api request to deployed certificates endpoint: %w", err)
	}

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	var deployedCertificatesResponse DeployedCertificatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&deployedCertificatesResponse); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return deployedCertificatesResponse.Certificates, nil
}
