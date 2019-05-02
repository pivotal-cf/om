package api

import (
	"net/http"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type NetworksConfigurationOutput struct {
	ICMP     bool `yaml:"icmp_checks_enabled"`
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

type AvailabilityZonesOutput struct {
	AvailabilityZones []AvailabilityZoneOutput `yaml:"availability_zones"`
}

type AvailabilityZoneOutput struct {
	Name                  string          `yaml:"name"`
	Clusters              []ClusterOutput `yaml:"clusters,omitempty"`
	IAASConfigurationGUID string          `yaml:"iaas_configuration_guid,omitempty"`
}

type ClusterOutput struct {
	Cluster      string `yaml:"cluster"`
	ResourcePool string `yaml:"resource_pool"`
}

func (a Api) GetStagedDirectorProperties(redact bool) (map[string]interface{}, error) {
	var queryString string

	if redact {
		queryString = "/api/v0/staged/director/properties?redact=true"
	} else {
		queryString = "/api/v0/staged/director/properties?redact=false"
	}

	resp, err := a.sendAPIRequest("GET", queryString, nil)
	if err != nil {
		return nil, err // un-tested
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	var properties map[string]interface{}
	if err = yaml.NewDecoder(resp.Body).Decode(&properties); err != nil {
		return nil, errors.Wrap(err, "could not parse json")
	}

	return properties, nil
}

func (a Api) GetStagedDirectorIaasConfigurations(redact bool) (map[string][]map[string]interface{}, error) {
	var queryString string

	if redact {
		queryString = "/api/v0/staged/director/iaas_configurations?redact=true"
	} else {
		queryString = "/api/v0/staged/director/iaas_configurations?redact=false"
	}

	resp, err := a.sendAPIRequest("GET", queryString, nil)
	if err != nil {
		return nil, err // un-tested
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	var properties map[string][]map[string]interface{}
	if err = yaml.NewDecoder(resp.Body).Decode(&properties); err != nil {
		return nil, errors.Wrap(err, "could not parse json")
	}

	return properties, nil
}

func (a Api) GetStagedDirectorAvailabilityZones() (AvailabilityZonesOutput, error) {
	var properties AvailabilityZonesOutput

	resp, err := a.sendAPIRequest("GET", "/api/v0/staged/director/availability_zones", nil)
	if err != nil {
		return properties, err // un-tested
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusMethodNotAllowed {
		return properties, nil
	}

	if err = validateStatusOK(resp); err != nil {
		return AvailabilityZonesOutput{}, err
	}

	if err = yaml.NewDecoder(resp.Body).Decode(&properties); err != nil {
		return properties, errors.Wrap(err, "could not parse json")
	}

	return properties, nil
}

func (a Api) GetStagedDirectorNetworks() (NetworksConfigurationOutput, error) {
	var properties NetworksConfigurationOutput

	resp, err := a.sendAPIRequest("GET", "/api/v0/staged/director/networks", nil)
	if err != nil {
		return properties, err // un-tested
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return NetworksConfigurationOutput{}, err
	}

	if err = yaml.NewDecoder(resp.Body).Decode(&properties); err != nil {
		return properties, errors.Wrap(err, "could not parse json")
	}

	return properties, nil
}
