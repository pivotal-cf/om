package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type NetworkConfigurationService struct {
	client httpClient
}

type NetworkConfigurationInput struct {
	AZs              map[string]string
	Networks         []Network
	EnableICMPChecks bool
}

type NetworkConfigurationOutput struct {
	Networks map[string]string
}

type Network struct {
	Name           string
	ServiceNetwork bool
	Subnets        []Subnet
}
type Subnet struct {
	Name             string
	CIDR             string
	ReservedIPRanges string
	DNS              string
	Gateway          string
	AZName           string
}

func NewNetworkConfigurationService(client httpClient) NetworkConfigurationService {
	return NetworkConfigurationService{
		client: client,
	}
}

func (s NetworkConfigurationService) Configure(input NetworkConfigurationInput) (NetworkConfigurationOutput, error) {
	if token, err := s.getPage(); err == nil {
		return s.updatePage(token, input)
	} else {
		return NetworkConfigurationOutput{}, err
	}
}

func (s NetworkConfigurationService) getPage() (string, error) {
	var err error
	var req *http.Request
	var resp *http.Response
	var token string
	if req, err = http.NewRequest("GET", "/infrastructure/networks/edit", nil); err == nil {
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

func (s NetworkConfigurationService) updatePage(token string, input NetworkConfigurationInput) (NetworkConfigurationOutput, error) {
	var err error
	var req *http.Request
	var resp *http.Response
	data := url.Values{}
	data.Set("_method", "put")
	data.Set("authenticity_token", token)
	data.Set("infrastructure[icmp_checks_enabled]", GetBooleanAsFormValue(input.EnableICMPChecks))
	for i, network := range input.Networks {
		data.Add(fmt.Sprintf("network_collection[networks_attributes][%d][guid]", i), "0")
		data.Add(fmt.Sprintf("network_collection[networks_attributes][%d][name]", i), network.Name)
		data.Add(fmt.Sprintf("network_collection[networks_attributes][%d][service_network]", i), GetBooleanAsFormValue(network.ServiceNetwork))
		for s, subnet := range network.Subnets {
			data.Add(fmt.Sprintf("network_collection[networks_attributes][%d][subnets][%d][iaas_identifier]", i, s), subnet.Name)
			data.Add(fmt.Sprintf("network_collection[networks_attributes][%d][subnets][%d][cidr]", i, s), subnet.CIDR)
			data.Add(fmt.Sprintf("network_collection[networks_attributes][%d][subnets][%d][reserved_ip_ranges]", i, s), subnet.ReservedIPRanges)
			data.Add(fmt.Sprintf("network_collection[networks_attributes][%d][subnets][%d][dns]", i, s), subnet.DNS)
			data.Add(fmt.Sprintf("network_collection[networks_attributes][%d][subnets][%d][gateway]", i, s), subnet.Gateway)
			data.Add(fmt.Sprintf("network_collection[networks_attributes][%d][subnets][%d][availability_zone_references][]", i, s), input.AZs[subnet.AZName])
		}
	}
	if req, err = http.NewRequest("POST", "/infrastructure/networks/update", strings.NewReader(data.Encode())); err == nil {
		if resp, err = s.client.Do(req); err == nil {
			if err = HandleResponse(resp, http.StatusOK); err == nil {
				return s.parseResponse(input, resp)
			}
		}
	}
	return NetworkConfigurationOutput{}, err
}

func (s NetworkConfigurationService) parseResponse(input NetworkConfigurationInput, resp *http.Response) (NetworkConfigurationOutput, error) {
	output := NetworkConfigurationOutput{
		Networks: make(map[string]string),
	}
	r, err := regexp.Compile("div\\sclass=.content.\\sid=.(.*)\\b")
	if err != nil {
		return NetworkConfigurationOutput{}, err
	}
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return NetworkConfigurationOutput{}, err
	} else {
		//Evaluate regex and get results
		matches := r.FindAllStringSubmatch(string(body), len(input.Networks))
		if len(matches) == 0 {
			return NetworkConfigurationOutput{}, fmt.Errorf("Unable to find token")
		} else {
			for i, network := range input.Networks {
				output.Networks[network.Name] = matches[i][1]
			}
		}
	}
	return output, nil
}
