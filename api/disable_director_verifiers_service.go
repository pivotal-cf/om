package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const listDirectorVerifiersEndpoint = "/api/v0/staged/director/verifiers/install_time"
const disableDirectorVerifiersEndpointTemplate = "/api/v0/staged/director/verifiers/install_time/%s"

type Verifier struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

type verifiersResponse struct {
	Verifiers []Verifier `json:"verifiers"`
}

func (a Api) ListDirectorVerifiers() ([]Verifier, error) {
	resp, err := a.sendAPIRequest("GET", listDirectorVerifiersEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("could not make api request to list_director_verifiers endpoint: %w", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	verifiersBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var verifierResponse verifiersResponse
	if err := json.Unmarshal(verifiersBytes, &verifierResponse); err != nil {
		return nil, fmt.Errorf("could not unmarshal list_director_verifiers response: %w", err)
	}

	return verifierResponse.Verifiers, nil
}

func (a Api) DisableDirectorVerifiers(verifiers []string) error {
	for _, verifier := range verifiers {
		resp, err := a.sendAPIRequest("PUT", fmt.Sprintf(disableDirectorVerifiersEndpointTemplate, verifier), []byte(`{ "enabled": false }`))
		if err != nil {
			return fmt.Errorf("could not make api request to disable_director_verifiers endpoint: %w", err)
		}
		resp.Body.Close()

		if err = validateStatusOK(resp); err != nil {
			return err
		}
	}

	return nil
}
