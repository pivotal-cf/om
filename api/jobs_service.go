package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type JobsService struct {
	client httpClient
}

type JobProperties struct {
	Instances              interface{}  `json:"instances"`
	PersistentDisk         *Disk        `json:"persistent_disk,omitempty"`
	InstanceType           InstanceType `json:"instance_type"`
	InternetConnected      *bool        `json:"internet_connected,omitempty"`
	LBNames                []string     `json:"elb_names"`
	NSXSecurityGroups      []string     `json:"nsx_security_groups,omitempty"`
	NSXLBS                 []NSXLB      `json:"nsx_lbs,omitempty"`
	FloatingIPs            string       `json:"floating_ips,omitempty"`
	AdditionalVMExtensions []string     `json:"additional_vm_extensions,omitempty"`
}

type NSXLB struct {
	EdgeName      string `json:"edge_name"`
	PoolName      string `json:"pool_name"`
	SecurityGroup string `json:"security_group"`
	Port          string `json:"port"`
}

type Disk struct {
	Size string `json:"size_mb"`
}

type InstanceType struct {
	ID string `json:"id"`
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

func (j JobsService) Jobs(productGUID string) (map[string]string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/jobs", productGUID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not make api request to jobs endpoint: %s", err)
	}

	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return nil, err
	}

	var jobsOutput struct {
		Jobs []Job
	}

	err = json.NewDecoder(resp.Body).Decode(&jobsOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to decode jobs json response: %s", err)
	}

	jobGUIDMap := make(map[string]string)
	for _, job := range jobsOutput.Jobs {
		jobGUIDMap[job.Name] = job.GUID
	}

	return jobGUIDMap, nil
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

	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return JobProperties{}, err
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

func (j JobsService) ConfigureJob(productGUID, jobGUID string, jobProperties JobProperties) error {
	bodyBytes := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(bodyBytes).Encode(jobProperties)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v0/staged/products/%s/jobs/%s/resource_config", productGUID, jobGUID), bodyBytes)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := j.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make api request to jobs resource_config endpoint: %s", err)
	}

	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
