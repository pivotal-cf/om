package api

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pkg/errors"
)

type Job struct {
	GUID string
	Name string
}

func (a Api) ListStagedProductJobs(productGUID string) (map[string]string, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/jobs", productGUID), nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not make api request to jobs endpoint")
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	var jobsOutput struct {
		Jobs []Job
	}

	err = json.NewDecoder(resp.Body).Decode(&jobsOutput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode jobs json response")
	}

	jobGUIDMap := make(map[string]string)
	for _, job := range jobsOutput.Jobs {
		jobGUIDMap[job.Name] = job.GUID
	}

	return jobGUIDMap, nil
}

func (a Api) ConfigureJobResourceConfig(productGUID string, config map[string]interface{}) error {
	jobs, err := a.ListStagedProductJobs(productGUID)
	if err != nil {
		return fmt.Errorf("failed to fetch jobs: %s", err)
	}

	var names []string
	for name := range config {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		jobGUID, ok := jobs[name]
		if !ok {
			return fmt.Errorf("unable to find job guid for job %s", name)
		}

		prop, err := a.getJSONProperties(config[name])
		if err != nil {
			return fmt.Errorf("could not unmarshall resource configuration for job %s: %v", name, err)
		}

		jobProperties, err := a.GetStagedProductJobResourceConfig(productGUID, jobGUID)
		if err != nil {
			return fmt.Errorf("could not fetch existing job configuration for job %s: %s", name, err)
		}

		err = json.Unmarshal([]byte(prop), &jobProperties)
		if err != nil {
			return fmt.Errorf("failed to unmarshal jobProperties for job %s: %s", name, err)
		}

		err = a.updateStagedProductJobResourceConfig(productGUID, jobGUID, jobProperties)
		if err != nil {
			return fmt.Errorf("failed to configure resources for %s: %s", name, err)
		}
	}

	return nil
}

type JobProperties map[string]interface{}

func (a Api) GetStagedProductJobResourceConfig(productGUID, jobGUID string) (JobProperties, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/jobs/%s/resource_config", productGUID, jobGUID), nil)
	if err != nil {
		return JobProperties{}, errors.Wrap(err, "could not make api request to resource_config endpoint")
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return JobProperties{}, err
	}

	var existingConfig JobProperties
	if err := json.NewDecoder(resp.Body).Decode(&existingConfig); err != nil {
		return JobProperties{}, err
	}

	return existingConfig, nil
}

func (a Api) updateStagedProductJobResourceConfig(productGUID, jobGUID string, jobProperties JobProperties) error {
	jsonPayload, err := json.Marshal(jobProperties)
	if err != nil {
		return err
	}

	resp, err := a.sendAPIRequest("PUT", fmt.Sprintf("/api/v0/staged/products/%s/jobs/%s/resource_config", productGUID, jobGUID), jsonPayload)
	if err != nil {
		return errors.Wrap(err, "could not make api request to jobs resource_config endpoint")
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
