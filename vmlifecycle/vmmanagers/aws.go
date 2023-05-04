package vmmanagers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type AWSCredential struct {
	AWSInstanceProfile `yaml:",inline"`
	AccessKeyId        string `yaml:"access_key_id,omitempty"`
	SecretAccessKey    string `yaml:"secret_access_key,omitempty"`
	Region             string `yaml:"region" validate:"required"`
}

type AWSInstanceProfile struct {
	UseInstanceProfileDEPRECATED bool   `yaml:"use_instance_profile,omitempty"`
	AssumeRole                   string `yaml:"assume_role,omitempty"`
}

type AWSConfig struct {
	AWSCredential             `yaml:",inline"`
	VPCSubnetId               string            `yaml:"vpc_subnet_id" validate:"required"`
	SecurityGroupIdDEPRECATED string            `yaml:"security_group_id,omitempty"`
	SecurityGroupIds          []string          `yaml:"security_group_ids"`
	KeyPairName               string            `yaml:"key_pair_name" validate:"required"`
	IAMInstanceProfileName    string            `yaml:"iam_instance_profile_name" validate:"required"`
	PublicIP                  string            `yaml:"public_ip" validate:"omitempty,ip"`
	PrivateIP                 string            `yaml:"private_ip" validate:"omitempty,ip"`
	VMName                    string            `yaml:"vm_name"`
	BootDiskSize              string            `yaml:"boot_disk_size"`
	InstanceType              string            `yaml:"instance_type"`
	Tags                      map[string]string `yaml:"tags"`
}

//go:generate counterfeiter -o ./fakes/awsRunner.go --fake-name AwsRunner . awsRunner
type awsRunner interface {
	ExecuteWithEnvVars(env []string, args []interface{}) (*bytes.Buffer, *bytes.Buffer, error)
}

type AWSVMManager struct {
	stdout          io.Writer
	stderr          io.Writer
	Config          *OpsmanConfigFilePayload
	ImageYaml       string
	State           StateInfo
	runner          awsRunner
	pollingInterval time.Duration
}

func NewAWSVMManager(stdout, stderr io.Writer, config *OpsmanConfigFilePayload, imageYaml string, stateInfo StateInfo, awsRunner awsRunner, t time.Duration) *AWSVMManager {
	return &AWSVMManager{
		stdout:          stdout,
		stderr:          stderr,
		Config:          config,
		ImageYaml:       imageYaml,
		State:           stateInfo,
		runner:          awsRunner,
		pollingInterval: t,
	}
}

func (a *AWSVMManager) DeleteVM() error {
	iaasConfig := a.Config.OpsmanConfig.AWS
	err := validateIAASConfig(iaasConfig.AWSCredential)
	if err != nil {
		return err
	}

	err = iaasConfig.validateConfig()
	if err != nil {
		return err
	}

	if a.State.IAAS != "aws" {
		return fmt.Errorf("authentication file provided is for aws, while the state file is for %s", a.State.IAAS)
	}

	return a.deleteVM(a.State.ID)
}

func (a *AWSVMManager) CreateVM() (Status, StateInfo, error) {
	iaasConfig := a.Config.OpsmanConfig.AWS
	latestState := StateInfo{IAAS: "aws"}

	if a.State.IAAS != "aws" && a.State.IAAS != "" {
		return Unknown, latestState, fmt.Errorf("authentication file provided is for aws, while the state file is for %s", a.State.IAAS)
	}

	err := validateIAASConfig(iaasConfig)
	if err != nil {
		return Unknown, latestState, err
	}

	err = iaasConfig.validateConfig()
	if err != nil {
		return Unknown, latestState, err
	}

	if a.Config.OpsmanConfig.AWS.PublicIP == "" && a.Config.OpsmanConfig.AWS.PrivateIP == "" {
		return Unknown, latestState, errors.New("PublicIP and/or PrivateIP must be set")
	}

	ami, err := amiFromRegion(iaasConfig.Region, a.ImageYaml)
	if err != nil {
		return Unknown, latestState, err
	}

	a.addDefaultConfigFields()

	exist, err := a.vmExists()
	if err != nil {
		return Unknown, latestState, err
	}
	if exist {
		return Exist, a.State, nil
	}

	instanceID, err := a.createVM(ami)
	if err != nil {
		return Unknown, latestState, err
	}
	latestState.ID = instanceID

	volumeID, err := a.getVolumeID(instanceID)
	if err != nil {
		return Incomplete, latestState, err
	}

	err = a.modifyVolume(volumeID)
	if err != nil {
		return Incomplete, latestState, err
	}

	if iaasConfig.PublicIP != "" {
		addressID, err := a.getIPAddressID()
		if err != nil {
			return Incomplete, latestState, err
		}

		err = a.associateIP(addressID, instanceID)
		if err != nil {
			return Incomplete, latestState, err
		}
	}

	err = a.rebootVM(instanceID)
	if err != nil {
		return Incomplete, latestState, err
	}

	return Success, StateInfo{IAAS: "aws", ID: instanceID}, nil
}

func (a *AWSVMManager) AddEnvVars() []string {
	return a.addEnvVars()
}

func (a *AWSVMManager) addEnvVars() []string {
	aws := a.Config.OpsmanConfig.AWS

	if aws.AssumeRole != "" {
		return []string{
			fmt.Sprintf("AWS_DEFAULT_REGION=%s", aws.Region),
		}
	} else {
		return []string{
			fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", aws.AccessKeyId),
			fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", aws.SecretAccessKey),
			fmt.Sprintf("AWS_DEFAULT_REGION=%s", aws.Region),
		}
	}
}

func (a *AWSVMManager) addDefaultConfigFields() {
	if a.Config.OpsmanConfig.AWS.VMName == "" {
		a.Config.OpsmanConfig.AWS.VMName = "ops-manager-vm"
	}
	if a.Config.OpsmanConfig.AWS.BootDiskSize == "" {
		a.Config.OpsmanConfig.AWS.BootDiskSize = "200"
	}
	if a.Config.OpsmanConfig.AWS.InstanceType == "" {
		a.Config.OpsmanConfig.AWS.InstanceType = "m5.large"
	}
}

func (a *AWSVMManager) ExecuteWithInstanceProfile(env []string, args []interface{}) (*bytes.Buffer, *bytes.Buffer, error) {
	config := a.Config.OpsmanConfig.AWS

	if config.AssumeRole != "" {
		file, err := ioutil.TempFile("", "awsConfig")
		defer os.Remove(file.Name())
		if err != nil {
			return nil, nil, err
		}
		configStrTemplate := `
[profile p-automator-assume]
role_arn = %s
credential_source = Ec2InstanceMetadata
`

		_, err = fmt.Fprintf(file, configStrTemplate, config.AssumeRole)
		if err != nil {
			return nil, nil, err
		}
		env = append(env, fmt.Sprintf("AWS_CONFIG_FILE=%s", file.Name()))
		args = append(args, "--profile", "p-automator-assume")
	}
	return a.runner.ExecuteWithEnvVars(env, args)
}

func (a *AWSVMManager) createVM(ami string) (string, error) {
	config := a.Config.OpsmanConfig.AWS
	tagsObject := a.Config.OpsmanConfig.AWS.Tags
	tags := []string{fmt.Sprintf("{Key=Name,Value=%s}", config.VMName)}

	for key, value := range tagsObject {
		tags = append(tags, fmt.Sprintf("{Key=%s,Value=%s}", key, value))
	}

	sort.Strings(tags)

	args := []interface{}{
		"ec2", "run-instances",
		"--tag-specifications", fmt.Sprintf(
			"ResourceType=instance,Tags=[%s]",
			strings.Join(tags, ","),
		),
		"--image-id", ami,
		"--subnet-id", config.VPCSubnetId,
		"--security-group-ids",
	}
	for _, sgID := range config.SecurityGroupIds {
		args = append(args, sgID)
	}
	args = append(args,
		"--count", "1",
		"--instance-type", config.InstanceType,
		"--key-name", config.KeyPairName,
		"--no-associate-public-ip-address",
		"--iam-instance-profile", fmt.Sprintf("Name=%s", config.IAMInstanceProfileName),
		"--query", "Instances[0].InstanceId",
	)

	if config.PrivateIP != "" {
		args = append(args, "--private-ip-address", config.PrivateIP)
	}
	instanceID, _, err := a.ExecuteWithInstanceProfile(a.addEnvVars(), args)
	if err != nil {
		return "", fmt.Errorf("aws error creating the vm: %s", err)
	}

	return cleanupString(instanceID.String()), nil
}

func (a *AWSVMManager) getVolumeID(instanceID string) (volumeID string, err error) {
	// wait until available
	for {
		volumeID, _, err := a.ExecuteWithInstanceProfile(a.addEnvVars(),
			[]interface{}{
				"ec2", "describe-volumes",
				"--filters",
				fmt.Sprintf("Name=attachment.instance-id,Values=%s", instanceID),
				"Name=attachment.status,Values=attached",
				"Name=status,Values=in-use",
				"--query", "Volumes[0].VolumeId",
			})
		if err != nil {
			return "", fmt.Errorf("aws error finding volumeID of the root device: %s", err)
		}
		if !strings.Contains(volumeID.String(), "null") {
			return cleanupString(volumeID.String()), nil
		}
		_, _ = a.stderr.Write([]byte(fmt.Sprintf("volume not available yet, polling in %s\n", a.pollingInterval)))
		time.Sleep(a.pollingInterval)
	}
}

func (a *AWSVMManager) modifyVolume(volumeID string) error {
	// wait until available
	var currentRetryCount float64
	timeoutTime := time.Now().Add(time.Hour)
	for {
		_, _, err := a.ExecuteWithInstanceProfile(a.addEnvVars(),
			[]interface{}{
				"ec2", "modify-volume",
				"--volume-id", volumeID,
				"--size", a.Config.OpsmanConfig.AWS.BootDiskSize,
			})

		if err == nil {
			break
		}

		if !strings.Contains(err.Error(), "exit status 255") {
			return fmt.Errorf("aws error modifying size of root device: %s", err)
		}

		// Encrypted volume is not available
		waitTime := getExponentialWaitTime(currentRetryCount, a.pollingInterval)
		_, _ = a.stderr.Write([]byte(fmt.Sprintf("volume not available to configure yet, polling in %s\n", waitTime)))

		if time.Now().After(timeoutTime) {
			return errors.New("failed to modify VM disk volume within the hour allowed")
		}

		time.Sleep(waitTime)
		currentRetryCount += 1
	}

	return nil
}

func (a *AWSVMManager) getIPAddressID() (ipAddress string, err error) {
	allocationID, _, err := a.ExecuteWithInstanceProfile(a.addEnvVars(),
		[]interface{}{
			"ec2", "describe-addresses",
			`--filters`, fmt.Sprintf("Name=public-ip,Values=%s", a.Config.OpsmanConfig.AWS.PublicIP),
			`--query`, `Addresses[0].AllocationId`,
		})
	if err != nil {
		return "", fmt.Errorf("aws error finding public IP address: %s", err)
	}

	return cleanupString(allocationID.String()), nil
}

func (a *AWSVMManager) associateIP(addressID, instanceID string) error {
	_, _, err := a.ExecuteWithInstanceProfile(a.addEnvVars(),
		[]interface{}{
			"ec2", "associate-address",
			"--allocation-id", addressID,
			"--instance-id", instanceID,
		})
	if err != nil {
		return fmt.Errorf("aws error finding public IP address: %s", err)
	}

	return nil
}

func (a *AWSVMManager) rebootVM(instanceID string) error {
	_, _, err := a.ExecuteWithInstanceProfile(a.addEnvVars(),
		[]interface{}{
			"ec2", "stop-instances",
			"--instance-ids", instanceID,
		})
	if err != nil {
		return fmt.Errorf("aws error can not stop vm: %s", err)
	}

	// wait until vm stopped
	for {
		state, _, err := a.ExecuteWithInstanceProfile(a.addEnvVars(),
			[]interface{}{
				"ec2", "describe-instances",
				"--instance-ids", instanceID,
				"--query", "Reservations[*].Instances[*].State.Name",
			})
		if err != nil {
			return fmt.Errorf("aws error could not query the state of the rebooting vm: %s", err)
		}
		if strings.Contains(state.String(), "stopped") {
			break
		}
		time.Sleep(a.pollingInterval)
	}

	_, _, err = a.ExecuteWithInstanceProfile(a.addEnvVars(),
		[]interface{}{
			"ec2", "start-instances",
			"--instance-ids", instanceID,
		})

	if err != nil {
		return fmt.Errorf("aws error can not start vm: %s", err)
	}

	return nil
}

func (a *AWSVMManager) deleteVM(instanceID string) error {
	_, err := a.vmExists()
	if err != nil {
		return fmt.Errorf("aws error %s", err)
	}

	_, errBuffWriter, err := a.ExecuteWithInstanceProfile(a.addEnvVars(),
		[]interface{}{
			"ec2", "terminate-instances",
			"--instance-ids", instanceID,
		})
	if err != nil {
		errStr := errBuffWriter.String()
		if strings.Contains(errStr, "InvalidInstanceID.NotFound") {
			return fmt.Errorf("%s\n       Could not find VM with ID %q.\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile.", err, a.State.ID)
		}

		return fmt.Errorf("aws error could not terminate vm: %s", err)
	}

	// retry 200 times. It does not guarantee the vm is fully terminated if 200 trials exceeds, do as best effort.
	for i := 0; i < 200; i++ {
		time.Sleep(a.pollingInterval)
		exist, err := a.vmExists()
		if err != nil {
			return fmt.Errorf("aws error could query vm status: %s", err)
		}
		if !exist {
			break
		}
	}

	return nil
}

func (a *AWSVMManager) vmExists() (vmExists bool, err error) {
	if a.State.ID == "" {
		return false, nil
	}

	var vmStatus, errBufWriter *bytes.Buffer
	vmStatus, errBufWriter, err = a.ExecuteWithInstanceProfile(a.addEnvVars(),
		[]interface{}{
			"ec2", "describe-instances",
			"--instance-ids", a.State.ID,
			"--query", "Reservations[*].Instances[*].State.Name",
		})

	if err == nil {
		if strings.Contains(vmStatus.String(), "terminated") || strings.Contains(vmStatus.String(), "[]") {
			return false, nil
		}
		return true, nil
	}

	errStr := errBufWriter.String()
	if strings.Contains(errStr, "InvalidInstanceID.Malformed") {
		return false, fmt.Errorf("instance ID %s is malformed. Please check your statefile and try again", a.State.ID)
	}
	if strings.Contains(errStr, "InvalidInstanceID.NotFound") {
		return false, fmt.Errorf("%s\n       Could not find VM with ID %q.\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile.", err, a.State.ID)
	}

	return false, fmt.Errorf("an unexpected error occurred: %s", err)
}

func (a *AWSConfig) ValidateConfig() error {
	return a.validateConfig()
}

func (a *AWSConfig) validateConfig() error {
	err := a.validateInstanceProfileConfig()
	if err != nil {
		return err
	}

	err = a.validateDeprecations()
	if err != nil {
		return err
	}

	return nil
}

func (a *AWSConfig) validateInstanceProfileConfig() error {
	if a.AssumeRole != "" && (a.AccessKeyId != "" || a.SecretAccessKey != "") {
		return errors.New("Assume Role only works when using an instance profile for AWS authentication")
	}

	return nil
}

func (a *AWSConfig) validateDeprecations() error {
	if a.SecurityGroupIdDEPRECATED == "" && len(a.SecurityGroupIds) == 0 {
		return errors.New("security_groups_ids is required")
	}
	if a.SecurityGroupIdDEPRECATED != "" && len(a.SecurityGroupIds) > 0 {
		return errors.New(`security_groups_id is DEPRECATED. Cannot use "security_group_id" and "security_group_ids" together. Use "security_groups_ids" instead.`)
	}
	if a.SecurityGroupIdDEPRECATED != "" && len(a.SecurityGroupIds) == 0 {
		a.SecurityGroupIds = []string{a.SecurityGroupIdDEPRECATED}
	}

	return nil
}

func amiFromRegion(region, imageFile string) (imageURI string, err error) {
	var images map[string]string

	contents, err := ioutil.ReadFile(imageFile)
	if err != nil {
		return "", err
	}

	err = checkImageFileIsYaml(imageFile)
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(contents, &images)
	if err != nil {
		return "", err
	}

	image, ok := images[region]
	if !ok {
		return "", errors.New("could not find a image uri for region")
	}

	return image, nil
}

func cleanupString(s string) string {
	return strings.Trim(strings.TrimSpace(s), "\"")
}

func getExponentialWaitTime(currentRetryCount float64, pollingMultiplier time.Duration) time.Duration {
	maxTime := 2 * time.Minute
	waitTime := time.Duration(math.Pow(2, currentRetryCount)) * pollingMultiplier

	if waitTime > maxTime {
		waitTime = maxTime
	}

	return waitTime
}
