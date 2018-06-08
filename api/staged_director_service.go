package api

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type NetworksConfigurationOutput struct {
	ICMP     bool `yaml:"icmp_checks_enabled,omitempty"`
	Networks []NetworkConfigurationOutput
}

type NetworkConfigurationOutput struct {
	Name           string         `yaml:"name"`
	ServiceNetwork *bool          `yaml:"service_network,omitempty"`
	Subnets        []SubnetOutput `yaml:"subnets,omitempty"`
}

type SubnetOutput struct {
	IAASIdentifier    string   `yaml:"iaas_identifier"`
	CIDR              string   `yaml:"cidr"`
	DNS               string   `yaml:"dns"`
	Gateway           string   `yaml:"gateway"`
	ReservedIPRanges  string   `yaml:"reserved_ip_ranges"`
	AvailabilityZones []string `yaml:"availability_zone_names,omitempty"`
}

func (a Api) GetStagedDirectorProperties() (map[string]map[string]interface{}, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/staged/director/properties", nil)
	if err != nil {
		return nil, err // un-tested
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var properties map[string]map[string]interface{}
	if err = yaml.Unmarshal(body, &properties); err != nil {
		return nil, fmt.Errorf("could not parse json: %s", err)
	}

	return properties, nil
}

func (a Api) GetStagedDirectorAvailabilityZones() (map[string][]map[string]interface{}, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/staged/director/availability_zones", nil)
	if err != nil {
		return nil, err // un-tested
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var properties map[string][]map[string]interface{}
	if err = yaml.Unmarshal(body, &properties); err != nil {
		return nil, fmt.Errorf("could not parse json: %s", err)
	}

	return properties, nil
}

func (a Api) GetStagedDirectorNetworks() (NetworksConfigurationOutput, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/staged/director/networks", nil)
	var properties NetworksConfigurationOutput
	if err != nil {
		return properties, err // un-tested
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return properties, err
	}

	if err = yaml.Unmarshal(body, &properties); err != nil {
		return properties, fmt.Errorf("could not parse json: %s", err)
	}

	return properties, nil
}
