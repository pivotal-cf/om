package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

type JobsService struct {
	client httpClient
}

type JobConfigurationInput struct {
	ProductGUID string
	Jobs        JobConfig
}

type JobConfig map[string]JobProperties

type JobProperties struct {
	Instances         int          `json:"instances"`
	PersistentDisk    *Disk        `json:"persistent_disk,omitempty"`
	InstanceType      InstanceType `json:"instance_type"`
	InternetConnected bool         `json:"internet_connected"`
	LBNames           []string     `json:"elb_names"`
}

type Disk struct {
	Size string `json:"size_mb"`
}

type InstanceType struct {
	ID string `json:"id"`
}

type JobsOutput struct {
	Jobs []Job
}

type Job struct {
	GUID string
	Name string
}

func NewJobsService(client httpClient) JobsService {
	return JobsService{
		client: client,
	}
}

func (j JobsService) Jobs(productGUID string) (JobsOutput, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/jobs", productGUID), nil)
	if err != nil {
		return JobsOutput{}, err
	}

	resp, err := j.client.Do(req)
	if err != nil {
		return JobsOutput{}, fmt.Errorf("could not make api request to jobs endpoint: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return JobsOutput{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return JobsOutput{}, fmt.Errorf("request failed: unexpected response:\n%s", out)
	}

	var output JobsOutput
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return JobsOutput{}, fmt.Errorf("failed to decode jobs json response: %s", err)
	}

	return output, nil
}

func (j JobsService) GetExistingJobConfig(productGUID, jobGUID string) (JobProperties, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/jobs/%s/resource_config", productGUID, jobGUID), nil)
	if err != nil {
		return JobProperties{}, err
	}

	resp, err := j.client.Do(req)
	if err != nil {
		return JobProperties{}, fmt.Errorf("could not make api request to resource_config endpoint: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return JobProperties{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return JobProperties{}, fmt.Errorf("request failed: unexpected response:\n%s", out)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return JobProperties{}, err
	}

	var existingConfig JobProperties
	err = json.Unmarshal(content, &existingConfig)
	if err != nil {
		return JobProperties{}, err
	}

	return existingConfig, nil
}

func (j JobsService) Configure(input JobConfigurationInput) error {
	var requests []*http.Request
	for jobGUID, resourceConfig := range input.Jobs {
		bodyBytes := bytes.NewBuffer([]byte{})
		err := json.NewEncoder(bodyBytes).Encode(resourceConfig)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v0/staged/products/%s/jobs/%s/resource_config", input.ProductGUID, jobGUID), bodyBytes)
		if err != nil {
			return err
		}

		req.Header.Add("Content-Type", "application/json")

		requests = append(requests, req)
	}

	for _, req := range requests {
		resp, err := j.client.Do(req)
		if err != nil {
			return fmt.Errorf("could not make api request to jobs resource_config endpoint: %s", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			out, err := httputil.DumpResponse(resp, true)
			if err != nil {
				return fmt.Errorf("request failed: unexpected response: %s", err)
			}

			return fmt.Errorf("request failed: unexpected response:\n%s", out)
		}
	}

	return nil
}
