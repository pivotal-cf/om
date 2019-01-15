package api

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type certResponse struct {
	Cert string `json:"root_ca_certificate_pem"`
}

func (a Api) GetSecurityRootCACertificate() (string, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/security/root_ca_certificate", nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to submit request")
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return "", err
	}

	var certResponse certResponse
	if err := json.NewDecoder(resp.Body).Decode(&certResponse); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal response")
	}

	return certResponse.Cert, nil
}
