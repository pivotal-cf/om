package api

import (
	"encoding/json"
	"fmt"
	"sort"
)

type Job struct {
	GUID string
	Name string
}

func (a Api) ListStagedProductJobs(productGUID string) (map[string]string, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/jobs", productGUID), nil)
	if err != nil {
		return nil, fmt.Errorf("could not make api request to jobs endpoint: %w", err)
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
		return nil, fmt.Errorf("failed to decode jobs json response: %w", err)
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

		jobProperties, err := a.GetStagedProductJobResourceConfig(productGUID, jobGUID)
		if err != nil {
			return fmt.Errorf("could not fetch existing job configuration for job %s: %s", name, err)
		}

		// Check if the existing config has instances set to "automatic" (string)
		// This indicates the job has instances_configurable: false
		// If user is trying to set a numeric value, we need to remove it to avoid API errors
		existingInstances, hasExistingInstances := jobProperties["instances"]
		if hasExistingInstances {
			if existingInstancesStr, ok := existingInstances.(string); ok && existingInstancesStr == "automatic" {
				// Job has instances_configurable: false, check if user is trying to set a numeric value
				// Handle both map[string]interface{} and map[interface{}]interface{} (from YAML unmarshaling)
				var userConfig map[string]interface{}
				var userConfigMapInterface map[interface{}]interface{}
				var hasUserConfig bool

				if configMap, ok := config[name].(map[string]interface{}); ok {
					userConfig = configMap
					hasUserConfig = true
				} else if configMapInterface, ok := config[name].(map[interface{}]interface{}); ok {
					userConfigMapInterface = configMapInterface
					hasUserConfig = true
				}

				if hasUserConfig {
					var userInstances interface{}
					var hasUserInstances bool

					if userConfig != nil {
						userInstances, hasUserInstances = userConfig["instances"]
					} else if userConfigMapInterface != nil {
						userInstances, hasUserInstances = userConfigMapInterface["instances"]
					}

					if hasUserInstances {
						// Check if user is trying to set a numeric value (not "automatic" string)
						switch userInstances.(type) {
						case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
							// User is trying to set a numeric value, but job doesn't allow it
							// Remove instances from user config to preserve "automatic"
							if userConfig != nil {
								delete(userConfig, "instances")
							} else if userConfigMapInterface != nil {
								delete(userConfigMapInterface, "instances")
							}
						case string:
							// If user explicitly sets "automatic", that's fine - no need to remove it
						}
					}
				}
			}
		}

		prop, err := a.getJSONProperties(config[name])
		if err != nil {
			return fmt.Errorf("could not unmarshall resource configuration for job %s: %v", name, err)
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
		return JobProperties{}, fmt.Errorf("could not make api request to resource_config endpoint: %w", err)
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
		return fmt.Errorf("could not make api request to jobs resource_config endpoint: %w", err)
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
