package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	yamlConverter "github.com/ghodss/yaml"
	"github.com/pkg/errors"
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

type DirectorProperties json.RawMessage

func (a Api) UpdateStagedDirectorAvailabilityZones(input AvailabilityZoneInput, ignoreVerifierWarnings bool) error {
	azs := AvailabilityZones{}
	err := yaml.Unmarshal(input.AvailabilityZones, &azs.AvailabilityZones)
	if err != nil {
		return errors.Wrap(err, "provided AZ config is not well-formed JSON")
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
		return errors.Wrap(err, "problem marshalling request") // un-tested
	}

	jsonData, err := yamlConverter.YAMLToJSON(decoratedConfig)
	if err != nil {
		return errors.Wrap(err, "problem converting request to JSON") // un-tested
	}

	resp, err := a.sendAPIRequest("PUT", "/api/v0/staged/director/availability_zones", jsonData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = validateStatusOKOrVerificationWarning(resp, ignoreVerifierWarnings); err != nil {
		return err
	}

	return nil
}

func (a Api) UpdateStagedDirectorNetworks(input NetworkInput) error {
	networks := Networks{}
	err := yaml.Unmarshal(input.Networks, &networks)
	if err != nil {
		return errors.Wrap(err, "provided networks config is not well-formed JSON")
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
		return errors.Wrap(err, "problem marshalling request") // un-tested
	}

	jsonData, err := yamlConverter.YAMLToJSON(decoratedConfig)
	if err != nil {
		return errors.Wrap(err, "problem converting request to JSON") // un-tested
	}

	resp, err := a.sendAPIRequest("PUT", "/api/v0/staged/director/networks", jsonData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) UpdateStagedDirectorNetworkAndAZ(input NetworkAndAZConfiguration) error {
	resp, err := a.sendAPIRequest("GET", "/api/v0/deployed/director/credentials", nil)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		a.logger.Println("unable to set network assignment for director as it has already been deployed")
		return nil
	case http.StatusNotFound:
		jsonData, err := json.Marshal(&input)
		if err != nil {
			return errors.Wrap(err, "could not marshal json")
		}

		resp, err = a.sendAPIRequest("PUT", "/api/v0/staged/director/network_and_az", jsonData)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if err = validateStatusOK(resp); err != nil {
			return err
		}

		return nil
	default:
		return fmt.Errorf("unexpected request status code: %d", resp.StatusCode)
	}
}

func (a Api) UpdateStagedDirectorProperties(input DirectorProperties) error {
	resp, err := a.sendAPIRequest("PUT", "/api/v0/staged/director/properties", input)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) addGUIDToExistingNetworks(networks Networks) (Networks, error) {
	existingNetworksResponse, err := a.sendAPIRequest("GET", "/api/v0/staged/director/networks", nil)
	if err != nil {
		return Networks{}, err
	}
	defer existingNetworksResponse.Body.Close()

	if existingNetworksResponse.StatusCode == http.StatusNotFound {
		a.logger.Println("unable to retrieve existing network configuration, attempting to configure anyway")
		return Networks{}, nil
	}

	if err = validateStatusOK(existingNetworksResponse); err != nil {
		return Networks{}, err
	}

	existingNetworksJSON, err := ioutil.ReadAll(existingNetworksResponse.Body)
	if err != nil {
		return Networks{}, errors.Wrap(err, "unable to read existing network configuration") // un-tested
	}

	var existingNetworks Networks
	err = yaml.Unmarshal(existingNetworksJSON, &existingNetworks)
	if err != nil {
		return Networks{}, errors.Wrap(err, "problem retrieving existing networks: response is not well-formed")
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
		return AvailabilityZones{}, err
	}

	switch {
	case existingAzsResponse.StatusCode == http.StatusOK:
		a.logger.Println("successfully fetched AZs, continuing")
	case existingAzsResponse.StatusCode == http.StatusNotFound:
		a.logger.Println("unable to retrieve existing AZ configuration, attempting to configure anyway")
		return azs, nil
	default:
		return AvailabilityZones{}, fmt.Errorf("received unexpected status while fetching AZ configuration: %d", existingAzsResponse.StatusCode)
	}

	existingAzsJSON, err := ioutil.ReadAll(existingAzsResponse.Body)
	if err != nil {
		return AvailabilityZones{}, errors.Wrap(err, "unable to read existing AZ configuration") // un-tested
	}

	var existingAZs AvailabilityZones
	err = yaml.Unmarshal(existingAzsJSON, &existingAZs)
	if err != nil {
		return AvailabilityZones{}, errors.Wrap(err, "problem retrieving existing AZs: response is not well-formed")
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
