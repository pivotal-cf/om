package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type AZConfigurationService struct {
	client httpClient
}

type AZConfigurationInput struct {
	AZNames []string
}

type AZConfigurationOutput struct {
	AZs map[string]string
}

func NewAZConfigurationService(client httpClient) AZConfigurationService {
	return AZConfigurationService{
		client: client,
	}
}

func (s AZConfigurationService) Configure(input AZConfigurationInput) (AZConfigurationOutput, error) {
	if token, err := s.getPage(); err == nil {
		return s.updatePage(token, input)
	} else {
		return AZConfigurationOutput{}, err
	}
}

func (s AZConfigurationService) getPage() (string, error) {
	var err error
	var req *http.Request
	var resp *http.Response
	var token string
	if req, err = http.NewRequest("GET", "/infrastructure/availability_zones/edit", nil); err == nil {
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

func (s AZConfigurationService) updatePage(token string, input AZConfigurationInput) (AZConfigurationOutput, error) {
	var err error
	var req *http.Request
	var resp *http.Response
	data := url.Values{}
	data.Set("_method", "put")
	data.Set("authenticity_token", token)
	for _, az := range input.AZNames {
		data.Add("availability_zones[availability_zones][][iaas_identifier]", az)
	}
	if req, err = http.NewRequest("POST", "/infrastructure/availability_zones", strings.NewReader(data.Encode())); err == nil {
		if resp, err = s.client.Do(req); err == nil {
			if err = HandleResponse(resp, http.StatusOK); err == nil {
				return s.parseResponse(input, resp)
			}
		}
	}
	return AZConfigurationOutput{}, err
}

func (s AZConfigurationService) parseResponse(input AZConfigurationInput, resp *http.Response) (AZConfigurationOutput, error) {
	output := AZConfigurationOutput{
		AZs: make(map[string]string),
	}
	r, err := regexp.Compile("name=.availability_zones\\[availability_zones\\]\\[\\]\\[guid\\].(\\s*)type=.hidden.(\\s*)value=.(.*)\\b")
	if err != nil {
		return AZConfigurationOutput{}, err
	}
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return AZConfigurationOutput{}, err
	} else {
		//Evaluate regex and get results
		matches := r.FindAllStringSubmatch(string(body), len(input.AZNames))
		if len(matches) == 0 {
			return AZConfigurationOutput{}, fmt.Errorf("Unable to find token")
		} else {
			for i, az := range input.AZNames {
				output.AZs[az] = matches[i][3]
			}
		}
	}

	return output, nil
}
