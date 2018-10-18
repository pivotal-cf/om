package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type FederationTokenService struct {
	client Client
}

type FederationToken struct {
	AccessKeyID     string `json:"access_key_id,omitempty" yaml:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty" yaml:"secret_access_key,omitempty"`
	SessionToken    string `json:"session_token,omitempty" yaml:"session_token,omitempty"`
	Bucket          string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Region          string `json:"region,omitempty" yaml:"region,omitempty"`
}

type createFederationTokenBody struct {
	ProductID string `json:"product_id"`
}

func (f FederationTokenService) GenerateFederationToken(productSlug string) (FederationToken, error) {
	url := fmt.Sprintf("/federation_token")

	body := createFederationTokenBody{
		ProductID: productSlug,
	}

	b, err := json.Marshal(body)
	if err != nil {
		// Untested as we cannot force an error because we are marshalling
		// a known-good body
		return FederationToken{}, err
	}

	resp, err := f.client.MakeRequest(
		"POST",
		url,
		http.StatusOK,
		bytes.NewReader(b),
	)

	if err != nil {
		return FederationToken{}, err
	}
	defer resp.Body.Close()

	var token FederationToken
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return FederationToken{}, err
	}

	return token, nil
}
