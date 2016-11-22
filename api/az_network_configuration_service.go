package api

import (
	"net/http"
	"net/url"
	"strings"
)

type AZNetworkConfigurationService struct {
	client httpClient
}

type AZNetworkConfigurationInput struct {
	SingletonAZ      string
	SingletonNetwork string
	AZs              map[string]string
	Networks         map[string]string
}

type AZNetworkConfigurationOutput struct {
}

func NewAZNetworkConfigurationService(client httpClient) AZNetworkConfigurationService {
	return AZNetworkConfigurationService{
		client: client,
	}
}

func (s AZNetworkConfigurationService) Configure(input AZNetworkConfigurationInput) (AZNetworkConfigurationOutput, error) {
	if token, err := s.getPage(); err == nil {
		return s.updatePage(token, input)
	} else {
		return AZNetworkConfigurationOutput{}, err
	}
}

func (s AZNetworkConfigurationService) getPage() (string, error) {
	var err error
	var req *http.Request
	var resp *http.Response
	var token string
	if req, err = http.NewRequest("GET", "/infrastructure/director/az_and_network_assignment/edit", nil); err == nil {
		if resp, err = s.client.Do(req); err == nil {
			defer resp.Body.Close()
			if err = HandleResponse(resp, http.StatusOK); err == nil {
				token, err = GetCSRFToken(resp)
				return token, err
			}
		}
	}
	return "", err
}

func (s AZNetworkConfigurationService) updatePage(token string, input AZNetworkConfigurationInput) (AZNetworkConfigurationOutput, error) {
	var err error
	var req *http.Request
	var resp *http.Response
	data := url.Values{}
	data.Set("_method", "put")
	data.Set("authenticity_token", token)
	data.Set("bosh_product[singleton_availability_zone_reference]", input.AZs[input.SingletonAZ])
	data.Set("bosh_product[network_reference]", input.Networks[input.SingletonNetwork])

	if req, err = http.NewRequest("POST", "/infrastructure/director/az_and_network_assignment", strings.NewReader(data.Encode())); err == nil {
		if resp, err = s.client.Do(req); err == nil {
			if err = HandleResponse(resp, http.StatusOK); err == nil {
				return AZNetworkConfigurationOutput{}, nil
			}
		}
	}
	return AZNetworkConfigurationOutput{}, err
}
