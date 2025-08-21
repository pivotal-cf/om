package api

import (
	"encoding/json"
	"fmt"
	"time"
)

const baseCertificatesEndpoint = "/api/v0/deployed/certificates"

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

// ListCertificates retrieves certificates from the deployed certificates endpoint.
// If expiresWithin is provided (non-empty), it filters for expiring certificates.
// If expiresWithin is empty, it returns all deployed certificates.
func (a Api) ListCertificates(expiresWithin string) ([]ExpiringCertificate, error) {
	endpoint := baseCertificatesEndpoint
	if expiresWithin != "" {
		endpoint = fmt.Sprintf("%s?expires_within=%s", baseCertificatesEndpoint, expiresWithin)
	}

	resp, err := a.sendAPIRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("could not make api request to certificates endpoint: %w", err)
	}

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	var certificatesResponse ExpiringCertificatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&certificatesResponse); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return certificatesResponse.Certificates, nil
}
