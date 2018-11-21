package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	yamlConverter "github.com/ghodss/yaml"
	"gopkg.in/yaml.v2"
)

type AvailabilityZoneInput struct {
	AvailabilityZones json.RawMessage `json:"availability_zones"`
}

type NetworkInput struct {
	Networks json.RawMessage `json:"networks"`
}

type Networks struct {
	Fields   map[string]interface{} `yaml:",inline"`
	Networks []*Network             `yaml:"networks"`
}

type Network struct {
	GUID   string                 `yaml:"guid,omitempty"`
	Name   string                 `yaml:"name"`
	Fields map[string]interface{} `yaml:",inline"`
}

type Cluster struct {
	GUID   string                 `yaml:"guid,omitempty"`
	Name   string                 `yaml:"cluster"`
	Fields map[string]interface{} `yaml:",inline"`
}

type AZ struct {
	GUID     string                 `yaml:"guid,omitempty"`
	Name     string                 `yaml:"name"`
	Clusters []*Cluster             `yaml:"clusters,omitempty"`
	Fields   map[string]interface{} `yaml:",inline"`
}

type AvailabilityZones struct {
	AvailabilityZones []*AZ `yaml:"availability_zones"`
}

type NetworkAndAZConfiguration struct {
	NetworkAZ json.RawMessage `json:"network_and_az,omitempty"`
}

type DirectorProperties struct {
	IAASConfiguration     json.RawMessage `json:"iaas_configuration,omitempty"`
	DirectorConfiguration json.RawMessage `json:"director_configuration,omitempty"`
	SecurityConfiguration json.RawMessage `json:"security_configuration,omitempty"`
	SyslogConfiguration   json.RawMessage `json:"syslog_configuration,omitempty"`
}

func (a Api) UpdateStagedDirectorAvailabilityZones(input AvailabilityZoneInput) error {
	azs := AvailabilityZones{}
	err := yaml.Unmarshal(input.AvailabilityZones, &azs.AvailabilityZones)
	if err != nil {
		return fmt.Errorf("provided AZ config is not well-formed JSON: %s", err)
	}

	for i, az := range azs.AvailabilityZones {
		if az.Name == "" {
			return fmt.Errorf("provided AZ config [%d] does not specify the AZ 'name'", i)
		}
	}

	azs, err = a.addGUIDToExistingAZs(azs)
	if err != nil {
		return err
	}

	decoratedConfig, err := yaml.Marshal(azs)
	if err != nil {
		return fmt.Errorf("problem marshalling request: %s", err) // un-tested
	}

	jsonData, err := yamlConverter.YAMLToJSON(decoratedConfig)
	if err != nil {
		return fmt.Errorf("problem converting request to JSON: %s", err) // un-tested
	}

	_, err = a.sendAPIRequest("PUT", "/api/v0/staged/director/availability_zones", jsonData)
	return err
}

func (a Api) UpdateStagedDirectorNetworks(input NetworkInput) error {
	networks := Networks{}
	err := yaml.Unmarshal(input.Networks, &networks)
	if err != nil {
		return fmt.Errorf("provided networks config is not well-formed JSON: %s", err)
	}

	for i, network := range networks.Networks {
		if network.Name == "" {
			return fmt.Errorf("provided networks config [%d] does not specify the network 'name'", i)
		}
	}

	networks, err = a.addGUIDToExistingNetworks(networks)
	if err != nil {
		return err
	}

	decoratedConfig, err := yaml.Marshal(networks)
	if err != nil {
		return fmt.Errorf("problem marshalling request: %s", err) // un-tested
	}

	jsonData, err := yamlConverter.YAMLToJSON(decoratedConfig)
	if err != nil {
		return fmt.Errorf("problem converting request to JSON: %s", err) // un-tested
	}

	_, err = a.sendAPIRequest("PUT", "/api/v0/staged/director/networks", jsonData)
	return err
}

func (a Api) UpdateStagedDirectorNetworkAndAZ(input NetworkAndAZConfiguration) error {
	_, err := a.sendAPIRequest("GET", "/api/v0/deployed/director/credentials", nil)
	if err == nil {
		a.logger.Println("unable to set network assignment for director as it has already been deployed")
		return err
	}
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	_, err = a.sendAPIRequest("PUT", "/api/v0/staged/director/network_and_az", jsonData)
	return err
}

func (a Api) UpdateStagedDirectorProperties(input DirectorProperties) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	_, err = a.sendAPIRequest("PUT", "/api/v0/staged/director/properties", jsonData)
	return err
}

func (a Api) addGUIDToExistingNetworks(networks Networks) (Networks, error) {
	existingNetworksResponse, err := a.sendAPIRequest("GET", "/api/v0/staged/director/networks", nil)
	if err != nil {
		if existingNetworksResponse.StatusCode != http.StatusNotFound {
			return Networks{}, fmt.Errorf("unable to fetch existing network configuration: %s", err)
		}
	}

	if existingNetworksResponse.StatusCode == http.StatusNotFound {
		a.logger.Println("unable to retrieve existing network configuration, attempting to configure anyway")
		return networks, nil
	}

	existingNetworksJSON, err := ioutil.ReadAll(existingNetworksResponse.Body)
	if err != nil {
		return Networks{}, fmt.Errorf("unable to read existing network configuration: %s", err) // un-tested
	}

	var existingNetworks Networks
	err = yaml.Unmarshal(existingNetworksJSON, &existingNetworks)
	if err != nil {
		return Networks{}, fmt.Errorf("problem retrieving existing networks: response is not well-formed: %s", err)
	}

	for _, network := range networks.Networks {
		for _, existingNetwork := range existingNetworks.Networks {
			if network.Name == existingNetwork.Name {
				network.GUID = existingNetwork.GUID
				break
			}
		}
	}
	return networks, nil
}

func (a Api) addGUIDToExistingAZs(azs AvailabilityZones) (AvailabilityZones, error) {
	existingAzsResponse, err := a.sendAPIRequest("GET", "/api/v0/staged/director/availability_zones", nil)
	if err != nil {
		if existingAzsResponse.StatusCode != http.StatusNotFound {
			return AvailabilityZones{}, fmt.Errorf("unable to fetch existing AZ configuration: %s", err)
		}
	}

	if existingAzsResponse.StatusCode == http.StatusNotFound {
		a.logger.Println("unable to retrieve existing AZ configuration, attempting to configure anyway")
		return azs, nil
	}

	existingAzsJSON, err := ioutil.ReadAll(existingAzsResponse.Body)
	if err != nil {
		return AvailabilityZones{}, fmt.Errorf("unable to read existing AZ configuration: %s", err) // un-tested
	}

	var existingAZs AvailabilityZones
	err = yaml.Unmarshal(existingAzsJSON, &existingAZs)
	if err != nil {
		return AvailabilityZones{}, fmt.Errorf("problem retrieving existing AZs: response is not well-formed: %s", err)
	}

	for _, az := range azs.AvailabilityZones {
		for _, existingAZ := range existingAZs.AvailabilityZones {
			if az.Name == existingAZ.Name {
				az.GUID = existingAZ.GUID

				for _, cluster := range az.Clusters {
					for _, existingCluster := range existingAZ.Clusters {
						if cluster.Name == existingCluster.Name {
							cluster.GUID = existingCluster.GUID
							break
						}
					}
				}

				break
			}
		}
	}
	return azs, nil
}

