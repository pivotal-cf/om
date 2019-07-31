package api

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// TODO: add omitempty everywhere
type JobProperties struct {
	Instances              interface{}  `json:"instances,omitempty" yaml:"instances,omitempty"`
	PersistentDisk         *Disk        `json:"persistent_disk,omitempty" yaml:"persistent_disk,omitempty"`
	InstanceType           InstanceType `json:"instance_type,omitempty" yaml:"instance_type,omitempty"`
	InternetConnected      *bool        `json:"internet_connected,omitempty" yaml:"internet_connected,omitempty"`
	LBNames                []string     `json:"elb_names,omitempty" yaml:"elb_names,omitempty"`
	NSX                    *NSX         `json:"nsx,omitempty" yaml:"nsx,omitempty"`
	NSXT                   *NSXT        `json:"nsxt,omitempty" yaml:"nsxt,omitempty"`
	Pre27NSXSecurityGroups []string     `json:"nsx_security_groups,omitempty" yaml:"nsx_security_groups,omitempty"`
	Pre27NSXLBS            []Pre27NSXLB `json:"nsx_lbs,omitempty" yaml:"nsx_lbs,omitempty"`
	FloatingIPs            string       `json:"floating_ips,omitempty" yaml:"floating_ips,omitempty"`
	AdditionalVMExtensions []string     `json:"additional_vm_extensions,omitempty" yaml:"additional_vm_extensions,omitempty"`
}

type NSX struct {
	SecurityGroups []string     `json:"security_groups" yaml:"security_groups"`
	LBS            []Pre27NSXLB `json:"lbs" yaml:"lbs"`
}

type NSXT struct {
	NSGroups []string `json:"ns_groups" yaml:"ns_groups"`
	VIFType  *string  `json:"vif_type" yaml:"vif_type"`
	LB       NSXTLB   `json:"lb" yaml:"lb"`
}

type NSXTLB struct {
	ServerPools []ServerPool `json:"server_pools" yaml:"server_pools"`
}

type Pre27NSXLB struct {
	EdgeName      string `json:"edge_name" yaml:"edge_name"`
	PoolName      string `json:"pool_name" yaml:"pool_name"`
	SecurityGroup string `json:"security_group" yaml:"security_group"`
	Port          int    `json:"port" yaml:"port"`
}

type ServerPool struct {
	Name string `json:"name" yaml:"name"`
	Port int    `json:"port" yaml:"port"`
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

func (a Api) UpdateStagedProductJobResourceConfig(productGUID, jobGUID string, jobProperties JobProperties) error {
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
