package api

import (
	"encoding/json"
	"fmt"
)

// TODO: add omitempty everywhere
type JobProperties struct {
	Instances              interface{}  `json:"instances" yaml:"instances"`
	PersistentDisk         *Disk        `json:"persistent_disk,omitempty" yaml:"persistent_disk,omitempty"`
	InstanceType           InstanceType `json:"instance_type" yaml:"instance_type"`
	InternetConnected      *bool        `json:"internet_connected,omitempty" yaml:"internet_connected,omitempty"`
	LBNames                []string     `json:"elb_names" yaml:"elb_names,omitempty"`
	NSXSecurityGroups      []string     `json:"nsx_security_groups,omitempty" yaml:"nsx_security_groups,omitempty"`
	NSXLBS                 []NSXLB      `json:"nsx_lbs,omitempty" yaml:"nsx_lbs,omitempty"`
	FloatingIPs            string       `json:"floating_ips,omitempty" yaml:"floating_ips,omitempty"`
	AdditionalVMExtensions []string     `json:"additional_vm_extensions,omitempty" yaml:"additional_vm_extensions,omitempty"`
}

type NSXLB struct {
	EdgeName      string `json:"edge_name" yaml:"edge_name"`
	PoolName      string `json:"pool_name" yaml:"pool_name"`
	SecurityGroup string `json:"security_group" yaml:"security_group"`
	Port          string `json:"port" yaml:"port"`
}

type Disk struct {
	Size string `json:"size_mb" yaml:"size_mb"`
}

type InstanceType struct {
	ID string `json:"id" yaml:"id"`
}

type Job struct {
	GUID string
	Name string
}

func (a Api) ListStagedProductJobs(productGUID string) (map[string]string, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/jobs", productGUID), nil)
	if err != nil {
		return nil, fmt.Errorf("could not make api request to jobs endpoint: %s", err)
	}
	defer resp.Body.Close()

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

func (a Api) GetStagedProductJobResourceConfig(productGUID, jobGUID string) (JobProperties, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/jobs/%s/resource_config", productGUID, jobGUID), nil)
	if err != nil {
		return JobProperties{}, fmt.Errorf("could not make api request to resource_config endpoint: %s", err)
	}
	defer resp.Body.Close()

	var existingConfig JobProperties
	if err := json.NewDecoder(resp.Body).Decode(&existingConfig); err != nil {
		return JobProperties{}, err
	}

	return existingConfig, nil
}

func (a Api) UpdateStagedProductJobResourceConfig(productGUID, jobGUID string, jobProperties JobProperties) error {
	jsonPayload, err := json.Marshal(jobProperties)
	if err != nil {
		return err
	}

	_, err = a.sendAPIRequest("PUT", fmt.Sprintf("/api/v0/staged/products/%s/jobs/%s/resource_config", productGUID, jobGUID), jsonPayload)
	if err != nil {
		return fmt.Errorf("could not make api request to jobs resource_config endpoint: %s", err)
	}

	return nil
}
