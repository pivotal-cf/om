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
	Name                  string                 `yaml:"name"`
	IAASConfigurationGUID string                 `yaml:"iaas_configuration_guid,omitempty"`
	IAASConfigurationName string                 `yaml:"iaas_configuration_name"`
	Fields                map[string]interface{} `yaml:",inline"`
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

	azResp, err := a.sendAPIRequest("GET", "/api/v0/staged/director/availability_zones", nil)
	if err != nil {
		return AvailabilityZonesOutput{}, err
	}
	defer azResp.Body.Close()

	if azResp.StatusCode == http.StatusMethodNotAllowed {
		return AvailabilityZonesOutput{}, nil
	}

	if err = validateStatusOK(azResp); err != nil {
		return AvailabilityZonesOutput{}, err
	}

	if err = yaml.NewDecoder(azResp.Body).Decode(&properties); err != nil {
		return AvailabilityZonesOutput{}, errors.Wrap(err, "could not parse json")
	}

	iaasResp, err := a.sendAPIRequest("GET", "/api/v0/staged/director/iaas_configurations", nil)
	if err != nil {
		return AvailabilityZonesOutput{}, err // un-tested
	}
	defer iaasResp.Body.Close()

	if err = validateStatusOK(iaasResp); err != nil {
		return AvailabilityZonesOutput{}, err
	}

	var iaasConfigs struct {
		Configs []struct {
			Name string
			GUID string
		} `yaml:"iaas_configurations"`
	}

	if err = yaml.NewDecoder(iaasResp.Body).Decode(&iaasConfigs); err != nil {
		return AvailabilityZonesOutput{}, errors.Wrap(err, "could not parse json")
	}

	for _, iaas := range iaasConfigs.Configs {
		for index, property := range properties.AvailabilityZones {
			if property.IAASConfigurationGUID == iaas.GUID {
				property.IAASConfigurationName = iaas.Name
				property.IAASConfigurationGUID = ""

				properties.AvailabilityZones[index] = property
			}
		}
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
