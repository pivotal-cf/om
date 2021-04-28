package vmmanagers

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pivotal-cf/om/vmlifecycle/runner"
)

type OpenstackCredential struct {
	Username           string `yaml:"username" validate:"required"`
	Password           string `yaml:"password" validate:"required"`
	AuthUrl            string `yaml:"auth_url" validate:"required"`
	Project            string `yaml:"project_name" validate:"required"`
	ProjectDomainName  string `yaml:"project_domain_name"`
	UserDomainName     string `yaml:"user_domain_name"`
	Insecure           bool   `yaml:"insecure"`
	IdentityAPIVersion int    `yaml:"identity_api_version"`
}

type OpenstackConfig struct {
	OpenstackCredential `yaml:",inline"`
	PublicIP            string `yaml:"public_ip" validate:"omitempty,ip"`
	PrivateIP           string `yaml:"private_ip" validate:"omitempty,ip"`
	VMName              string `yaml:"vm_name"`
	Flavor              string `yaml:"flavor"`
	NetID               string `yaml:"net_id" validate:"required"`
	SecurityGroup       string `yaml:"security_group_name" validate:"required"`
	KeyName             string `yaml:"key_pair_name" validate:"required"`
	AvailabilityZone    string `yaml:"availability_zone"`
}

//go:generate counterfeiter -o ./fakes/openstackRunner.go --fake-name OpenstackRunner . openstackRunner
type openstackRunner interface {
	Execute(args []interface{}) (*bytes.Buffer, *bytes.Buffer, error)
}

type OpenstackVMManager struct {
	Config *OpenstackConfig
	Image  string
	State  StateInfo
	runner openstackRunner
}

func NewOpenstackVMManager(config *OpsmanConfigFilePayload, image string, state StateInfo, openstackRunner openstackRunner) *OpenstackVMManager {
	return &OpenstackVMManager{
		Config: config.OpsmanConfig.Openstack,
		Image:  image,
		State:  state,
		runner: openstackRunner,
	}
}

func (o *OpenstackVMManager) DeleteVM() error {
	err := validateIAASConfig(o.Config.OpenstackCredential)
	if err != nil {
		return err
	}

	if o.State.IAAS != "openstack" {
		return fmt.Errorf("authentication file provided is for openstack, while the state file is for %s", o.State.IAAS)
	}

	imageID, err := o.getVMImage()
	if err != nil {
		return fmt.Errorf("openstack error deleting the vm: %s", err)
	}
	err = o.deleteVM()
	if err != nil {
		return fmt.Errorf("openstack error deleting the vm: %s", err)
	}

	err = o.deleteImage(imageID)
	if err != nil {
		return fmt.Errorf("openstack error deleting the vm: %s", err)
	}
	return nil
}

func (o *OpenstackVMManager) CreateVM() (Status, StateInfo, error) {
	if o.State.IAAS != "openstack" && o.State.IAAS != "" {
		return Unknown, StateInfo{}, fmt.Errorf("authentication file provided is for openstack, while the state file is for %s", o.State.IAAS)
	}

	err := validateIAASConfig(o.Config)
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	if o.Config.PublicIP == "" && o.Config.PrivateIP == "" {
		return Unknown, StateInfo{}, errors.New("PublicIP and/or PrivateIP must be set")
	}

	o.addDefaultConfigFields()

	if _, err := os.Stat(o.Image); os.IsNotExist(err) {
		return Unknown, StateInfo{}, fmt.Errorf("could not read image file: %s", err)
	}

	exist, err := o.vmExists()
	if err != nil {
		return Unknown, StateInfo{}, err
	}
	if exist {
		return Exist, o.State, nil
	}

	imageID, err := o.forceRecreateImage()
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	serverID, _, err := o.createVM(imageID)
	if err != nil {
		return Unknown, StateInfo{}, fmt.Errorf("openstack error creating the vm: %s", err)
	}

	fullState := StateInfo{IAAS: "openstack", ID: serverID}

	if o.Config.PublicIP != "" {
		err = o.attachIP(serverID)
		if err != nil {
			return Incomplete, fullState, err
		}
	}

	return Success, fullState, nil
}

func (o *OpenstackVMManager) forceRecreateImage() (string, error) {
	imageName := fmt.Sprintf("%s-image", o.Config.VMName)
	result, err := o.listExistingImages(imageName)
	if err != nil {
		return "", err
	}

	err = o.removeImageIfExists(result, imageName)
	if err != nil {
		return "", err
	}

	imageID, err := o.createImage(imageName)
	return imageID, checkFormatedError("openstack error could not create image: %s", err)
}

func (o *OpenstackVMManager) listExistingImages(imageName string) (string, error) {
	args := append(o.getAuthArguments(), `image`, `list`,
		`--name`, imageName,
		`--format`, `value`,
		`--column`, `ID`)

	stdout, _, err := o.runner.Execute(args)
	return cleanupString(stdout.String()), checkFormatedError("openstack error failed to list existing images. Could not create: %s", err)
}

func (o *OpenstackVMManager) removeImageIfExists(stdout string, imageName string) error {
	var err error
	if cleanupString(stdout) != "" {
		err = o.deleteImage(imageName)
	}
	return err
}

func (o *OpenstackVMManager) createImage(imageName string) (imageURI string, err error) {
	args := append(o.getAuthArguments(), `image`, `create`,
		`--format`, `value`, `--column`, `id`,
		`--file`, o.Image, imageName)

	stdout, _, err := o.runner.Execute(args)

	return cleanupString(stdout.String()), err
}

func (o *OpenstackVMManager) createVM(imageID string) (serverID string, state StateInfo, err error) {
	var nicStr string
	if o.Config.PrivateIP != "" {
		nicStr = fmt.Sprintf(`net-id=%s,v4-fixed-ip=%s`, o.Config.NetID, o.Config.PrivateIP)
	} else {
		nicStr = fmt.Sprintf(`net-id=%s`, o.Config.NetID)
	}

	args := append(o.getAuthArguments(), `server`, `create`,
		`--flavor`, o.Config.Flavor,
		`--image`, imageID,
		`--nic`, nicStr,
		`--security-group`, o.Config.SecurityGroup,
		`--key-name`, o.Config.KeyName,
		`--format`, `value`,
		`--column`, `id`,
		`--wait`,
	)

	if strings.TrimSpace(o.Config.AvailabilityZone) != "" {
		args = append(args, `--availability-zone`, o.Config.AvailabilityZone)
	}

	args = append(args, o.Config.VMName)

	stdout, _, err := o.runner.Execute(args)
	if err != nil {
		return "", StateInfo{}, err
	}
	return cleanupString(stdout.String()), StateInfo{}, nil
}

func (o *OpenstackVMManager) deleteVM() error {
	args := append(o.getAuthArguments(), `server`, `delete`, o.State.ID, `--wait`)
	_, _, err := o.runner.Execute(args)
	if err != nil {
		return fmt.Errorf("openstack error deleting VM: %s", err)
	}

	return nil
}

func (o *OpenstackVMManager) attachIP(serverID string) error {
	log.Println("Attaching Public IP to VM...")
	args := append(o.getAuthArguments(), `server`, `add`, `floating`, `ip`,
		serverID, o.Config.PublicIP)
	_, _, err := o.runner.Execute(args)
	if err != nil {
		return fmt.Errorf("openstack error attaching the IP address to VM: %s", err)
	}
	return nil
}

func (o *OpenstackVMManager) addDefaultConfigFields() {
	if o.Config.VMName == "" {
		o.Config.VMName = "ops-manager-vm"
	}
	if o.Config.Flavor == "" {
		o.Config.Flavor = "m1.xlarge"
	}
	if o.Config.IdentityAPIVersion == 0 {
		o.Config.IdentityAPIVersion = 3
	}
}

func (o *OpenstackVMManager) vmExists() (bool, error) {
	if o.State.ID == "" {
		return false, nil
	}

	args := append(o.getAuthArguments(), `server`, `show`, o.State.ID,
		`--column`, `status`,
		`--format`, `value`)

	stdout, _, err := o.runner.Execute(args)
	status := cleanupString(stdout.String())

	return status == "ACTIVE", checkFormatedError("VM ID in statefile does not exist. Please check your statefile and try again: %s", err)
}

func (o *OpenstackVMManager) getVMImage() (string, error) {
	args := append(o.getAuthArguments(), `server`, `show`, o.State.ID,
		`--column`, `image`,
		`--format`, `value`)
	stdout, _, err := o.runner.Execute(args)
	if err != nil {
		return "", fmt.Errorf(
			"%s\n       Could not find VM with ID %q.\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile.",
			err,
			o.State.ID,
		)
	}

	re := regexp.MustCompile("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")
	return re.FindString(stdout.String()), nil
}

func (o *OpenstackVMManager) deleteImage(imageID string) error {
	args := append(o.getAuthArguments(), `image`, `delete`, imageID)
	_, _, err := o.runner.Execute(args)

	return checkFormatedError("openstack error failed to remove existing image. Could not create: %s", err)
}

func (o *OpenstackVMManager) getAuthArguments() []interface{} {
	args := []interface{}{
		`--os-username`, runner.Redact(o.Config.Username),
		`--os-password`, runner.Redact(o.Config.Password),
		`--os-auth-url`, o.Config.AuthUrl,
		`--os-project-name`, o.Config.Project,
	}

	if o.Config.Insecure {
		args = append(args, `--insecure`)
	}

	if strings.TrimSpace(o.Config.ProjectDomainName) != "" {
		args = append(args, `--os-project-domain-name`, o.Config.ProjectDomainName)
	}

	if strings.TrimSpace(o.Config.UserDomainName) != "" {
		args = append(args, `--os-user-domain-name`, o.Config.UserDomainName)
	}

	if o.Config.IdentityAPIVersion >= 2 && o.Config.IdentityAPIVersion <= 3 {
		args = append(args, `--os-identity-api-version`, strconv.Itoa(o.Config.IdentityAPIVersion))
	}

	return args
}
