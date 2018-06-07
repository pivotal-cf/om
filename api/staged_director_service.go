package api

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

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

func (a Api) GetStagedDirectorNetworks() (map[string]interface{}, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/staged/director/networks", nil)
	if err != nil {
		return nil, err // un-tested
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var properties map[string]interface{}
	if err = yaml.Unmarshal(body, &properties); err != nil {
		return nil, fmt.Errorf("could not parse json: %s", err)
	}

	return properties, nil
}
