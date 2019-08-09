package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

const expiringCertificatesEndpoint = "/api/v0/deployed/certificates?expires_within=%s"

type ExpiringCertificatesResponse struct {
	Certificates []ExpiringCertificate `json:"certificates"`
}

type ExpiringCertificate struct {
	Issuer            string    `json:"issuer"`
	ValidFrom         time.Time `json:"valid_from"`
	ValidUntil        time.Time `json:"valid_until"`
	Configurable      bool      `json:"configurable"`
	PropertyReference string    `json:"property_reference"`
	PropertyType      string    `json:"property_type"`
	ProductGUID       string    `json:"product_guid"`
	Location          string    `json:"location"`
	VariablePath      string    `json:"variable_path"`
}

func (a Api) ListExpiringCertificates(expiresWithin string) ([]ExpiringCertificate, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf(expiringCertificatesEndpoint, expiresWithin), nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not make api request to certificates endpoint")
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
