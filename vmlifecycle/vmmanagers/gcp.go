package vmmanagers

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type GCPCredential struct {
	ServiceAccount     string `yaml:"gcp_service_account,omitempty"`
	ServiceAccountName string `yaml:"gcp_service_account_name,omitempty"`
	Project            string `yaml:"project" validate:"required"`
	Region             string `yaml:"region" validate:"required"`
	Zone               string `yaml:"zone" validate:"required"`
}

type GCPConfig struct {
	GCPCredential `yaml:",inline"`
	VpcSubnet     string   `yaml:"vpc_subnet" validate:"required"`
	PublicIP      string   `yaml:"public_ip,omitempty"  validate:"omitempty,ip"`
	PrivateIP     string   `yaml:"private_ip" validate:"omitempty,ip"`
	VMName        string   `yaml:"vm_name"`
	Tags          string   `yaml:"tags,omitempty"`
	CPU           string   `yaml:"custom_cpu"`
	Memory        string   `yaml:"custom_memory"`
	BootDiskSize  string   `yaml:"boot_disk_size"`
	Scopes        []string `yaml:"scopes,omitempty"`
	SSHPublicKey  string   `yaml:"ssh_public_key"`
	Hostname      string   `yaml:"hostname"`
}

//go:generate counterfeiter -o ./fakes/gcloudRunner.go --fake-name GCloudRunner . gcloudRunner
type gcloudRunner interface {
	Execute(args []interface{}) (*bytes.Buffer, *bytes.Buffer, error)
}

type GCPVMManager struct {
	Config       *OpsmanConfigFilePayload
	ImageYaml    string
	State        StateInfo
	imageUriMap  imageGCPFormat
	gcloudRunner gcloudRunner
}

func NewGcpVMManager(config *OpsmanConfigFilePayload, imageYaml string, state StateInfo, gcloudRunner gcloudRunner) *GCPVMManager {
	return &GCPVMManager{
		ImageYaml:    imageYaml,
		State:        state,
		Config:       config,
		gcloudRunner: gcloudRunner,
	}
}

func (g *GCPVMManager) DeleteVM() error {
	err := validateIAASConfig(g.Config.OpsmanConfig.GCP.GCPCredential)
	if err != nil {
		return err
	}

	if g.State.IAAS != "gcp" {
		return fmt.Errorf("authentication file provided is for gcp, while the state file is for %s", g.State.IAAS)
	}

	err = g.authenticate()

	if err != nil {
		return err
	}

	err = g.setProject()
	if err != nil {
		return err
	}

	err = g.setComputeRegion()
	if err != nil {
		return err
	}

	_, err = g.vmExists()
	if err != nil {
		return err
	}

	err = g.deleteVM(g.State.ID)
	if err != nil {
		return err
	}

	err = g.deleteImage(fmt.Sprintf("%s-image", g.State.ID))
	if err != nil {
		return err
	}

	return nil
}

func (g *GCPVMManager) CreateVM() (Status, StateInfo, error) {
	if g.State.IAAS != "gcp" && g.State.IAAS != "" {
		return Unknown, StateInfo{}, fmt.Errorf("authentication file provided is for gcp, while the state file is for %s", g.State.IAAS)
	}

	err := validateIAASConfig(g.Config.OpsmanConfig.GCP)
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	if g.Config.OpsmanConfig.GCP.PublicIP == "" && g.Config.OpsmanConfig.GCP.PrivateIP == "" {
		return Unknown, StateInfo{}, errors.New("PublicIP and/or PrivateIP must be set")
	}

	if g.Config.OpsmanConfig.GCP.ServiceAccount == "" && g.Config.OpsmanConfig.GCP.ServiceAccountName == "" {
		return Unknown, StateInfo{}, errors.New("gcp_service_account or gcp_service_account_name must be set")
	}

	imageUriMap, err := loadImageYaml(g.ImageYaml)
	g.imageUriMap = imageUriMap
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	g.addDefaultConfigFields()

	err = g.authenticate()
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	err = g.setProject()
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	err = g.setComputeRegion()
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	exist, err := g.vmExists()
	if err != nil {
		return Unknown, StateInfo{}, err
	}
	if exist {
		return Exist, g.State, nil
	}

	err = g.forceRecreateImage()
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	return g.createVM(g.Config.OpsmanConfig.GCP.PublicIP)
}

func (g *GCPVMManager) authenticate() error {
	if len(g.Config.OpsmanConfig.GCP.ServiceAccount) == 0 {
		if len(g.Config.OpsmanConfig.GCP.ServiceAccountName) != 0 {
			_, _, err := g.gcloudRunner.Execute([]interface{}{
				"iam",
				"service-accounts",
				"describe",
				g.Config.OpsmanConfig.GCP.ServiceAccountName,
			})
			if err != nil {
				return fmt.Errorf("service_account_name error. The service account may not exist or this SDK dose not have permission to access: %s", err)
			}
		}

		return nil
	}

	serviceAccountFile, err := g.generateServiceAccountKeyFile()
	if err != nil {
		return err
	}

	_, _, err = g.gcloudRunner.Execute([]interface{}{
		"auth",
		"activate-service-account",
		"--key-file",
		serviceAccountFile,
	})

	if err != nil {
		return fmt.Errorf("gcloud error authenticating service account: %s", err)
	}

	return nil
}

func (g *GCPVMManager) setProject() error {
	_, _, err := g.gcloudRunner.Execute([]interface{}{
		"config",
		"set",
		"project",
		g.Config.OpsmanConfig.GCP.Project,
	})
	if err != nil {
		return fmt.Errorf("gcloud error setting project: %s", err)
	}
	return nil
}

func (g *GCPVMManager) setComputeRegion() error {
	_, _, err := g.gcloudRunner.Execute([]interface{}{
		"config",
		"set",
		"compute/region",
		g.Config.OpsmanConfig.GCP.Region,
	})
	if err != nil {
		return fmt.Errorf("gcloud error setting compute/region: %s", err)
	}
	return nil
}

func (g *GCPVMManager) forceRecreateImage() error {
	imageUri := "https://storage.googleapis.com/" + g.imageUriMap.get("us")

	// failing cases untested
	_, errBufWriter, err := g.gcloudRunner.Execute([]interface{}{
		"compute",
		"images",
		"delete",
		g.Config.OpsmanConfig.GCP.VMName + "-image",
		"--quiet",
	})

	if err != nil && !strings.Contains(errBufWriter.String(), "not found") {
		return fmt.Errorf("could not recreate the opsman image: %s", err)
	}

	_, _, err = g.gcloudRunner.Execute([]interface{}{
		"compute",
		"images",
		"create",
		g.Config.OpsmanConfig.GCP.VMName + "-image",
		"--source-uri=" + imageUri,
	})

	if err != nil {
		return fmt.Errorf("could not recreate the opsman image: %s", err)
	}

	return nil
}

func (g *GCPVMManager) createVM(publicIP string) (Status, StateInfo, error) {
	networkProperties := []string{
		fmt.Sprintf("subnet=%s", g.Config.OpsmanConfig.GCP.VpcSubnet),
	}
	if publicIP == "" {
		networkProperties = append(networkProperties, "no-address")
	} else {
		networkProperties = append(networkProperties, fmt.Sprintf("address=%s", publicIP))
	}
	if g.Config.OpsmanConfig.GCP.PrivateIP != "" {
		networkProperties = append(networkProperties, fmt.Sprintf("private-network-ip=%s", g.Config.OpsmanConfig.GCP.PrivateIP))
	}

	createVMArgs := []interface{}{
		"compute",
		"instances",
		"create",
		g.Config.OpsmanConfig.GCP.VMName,
		`--zone`, g.Config.OpsmanConfig.GCP.Zone,
		`--image`, g.Config.OpsmanConfig.GCP.VMName + "-image",
		`--custom-cpu`, g.Config.OpsmanConfig.GCP.CPU,
		`--custom-memory`, g.Config.OpsmanConfig.GCP.Memory,
		`--boot-disk-size`, g.Config.OpsmanConfig.GCP.BootDiskSize,
		`--network-interface`, strings.Join(networkProperties, ","),
	}

	if len(g.Config.OpsmanConfig.GCP.Tags) > 0 {
		createVMArgs = append(createVMArgs, "--tags")
		createVMArgs = append(createVMArgs, g.Config.OpsmanConfig.GCP.Tags)
	}

	if len(g.Config.OpsmanConfig.GCP.ServiceAccountName) > 0 {
		createVMArgs = append(createVMArgs, "--service-account")
		createVMArgs = append(createVMArgs, g.Config.OpsmanConfig.GCP.ServiceAccountName)
	}

	if len(g.Config.OpsmanConfig.GCP.Scopes) > 0 {
		createVMArgs = append(createVMArgs, "--scopes")
		createVMArgs = append(createVMArgs, strings.Join(g.Config.OpsmanConfig.GCP.Scopes, ","))
	}

	if len(g.Config.OpsmanConfig.GCP.SSHPublicKey) > 0 {
		createVMArgs = append(createVMArgs, "--metadata")
		createVMArgs = append(createVMArgs, fmt.Sprintf("ssh-keys=ubuntu:%s,block-project-ssh-keys=TRUE", g.Config.OpsmanConfig.GCP.SSHPublicKey))
	}

	if len(g.Config.OpsmanConfig.GCP.Hostname) > 0 {
		createVMArgs = append(
			createVMArgs,
			"--hostname",
			g.Config.OpsmanConfig.GCP.Hostname,
		)
	}

	_, errBufWriter, err := g.gcloudRunner.Execute(createVMArgs)

	errBufStr := errBufWriter.String()
	currentState := StateInfo{IAAS: "gcp", ID: g.Config.OpsmanConfig.GCP.VMName}
	if err == nil {
		return Success, currentState, nil
	}
	if strings.Contains(errBufStr, "already exists") {
		return Exist, currentState, nil
	}
	return Unknown, StateInfo{}, fmt.Errorf("gcloud error creating VM: %s", err)
}

func (g *GCPVMManager) deleteVM(id string) error {
	_, _, err := g.gcloudRunner.Execute([]interface{}{
		"compute", "instances", "delete", id,
		"--zone", g.Config.OpsmanConfig.GCP.Zone,
		"--quiet",
	})
	if err != nil {
		return fmt.Errorf("gcloud error deleting VM: %s", err)
	}
	return nil
}

func (g *GCPVMManager) generateServiceAccountKeyFile() (serviceAccountFileName string, err error) {
	serviceAccountFile, err := os.CreateTemp("", "key.yaml")
	if err != nil {
		return "", err
	}

	_, err = serviceAccountFile.WriteString(g.Config.OpsmanConfig.GCP.ServiceAccount)
	if err != nil {
		return "", err
	}
	err = serviceAccountFile.Close()
	if err != nil {
		return "", err
	}

	return serviceAccountFile.Name(), nil
}

func (g *GCPVMManager) addDefaultConfigFields() {
	if g.Config.OpsmanConfig.GCP.VMName == "" {
		g.Config.OpsmanConfig.GCP.VMName = "ops-manager-vm"
	}

	if g.Config.OpsmanConfig.GCP.CPU == "" {
		g.Config.OpsmanConfig.GCP.CPU = "2"
	}

	if g.Config.OpsmanConfig.GCP.Memory == "" {
		g.Config.OpsmanConfig.GCP.Memory = "8"
	}

	if g.Config.OpsmanConfig.GCP.BootDiskSize == "" {
		g.Config.OpsmanConfig.GCP.BootDiskSize = "100"
	}
}

func (g *GCPVMManager) vmExists() (vmExists bool, err error) {
	if g.State.ID == "" {
		return false, nil
	}

	_, _, err = g.gcloudRunner.Execute([]interface{}{
		"compute", "instances", "describe",
		g.State.ID,
		"--zone", g.Config.OpsmanConfig.GCP.Zone,
		"--format", "value(status)",
	})

	if err != nil {
		return false, fmt.Errorf("error: %s\n       Could not find VM with ID %q.\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile.", err, g.State.ID)
	}

	return true, nil
}

func (g *GCPVMManager) deleteImage(imageID string) error {
	_, stderr, err := g.gcloudRunner.Execute([]interface{}{
		"compute",
		"images",
		"delete",
		imageID,
		"--quiet",
	})

	if err != nil && !strings.Contains(stderr.String(), "not found") {
		return fmt.Errorf("could not delete image %s: %s", imageID, err)
	}
	return nil
}

type imageGCPFormat map[string]interface{}

func (f imageGCPFormat) get(s string) string {
	if v, ok := f[s]; ok {
		return v.(string)
	}
	return ""
}

func loadImageYaml(imageYaml string) (imageGCPFormat, error) {
	var imageUriMap imageGCPFormat
	imageUriMapByte, err := os.ReadFile(imageYaml)
	if err != nil {
		return imageUriMap, err
	}

	err = checkImageFileIsYaml(imageYaml)
	if err != nil {
		return imageUriMap, err
	}

	if err != nil {
		return imageUriMap, fmt.Errorf("could not read image file: %s", err)
	}

	err = yaml.Unmarshal(imageUriMapByte, &imageUriMap)
	if err != nil {
		return imageUriMap, fmt.Errorf("could not marshal image file: %s", err)
	}

	return imageUriMap, nil
}

func checkImageFileIsYaml(imageYaml string) (err error) {
	extension := filepath.Ext(imageYaml)

	if extension == ".yml" || extension == ".yaml" {
		return nil
	}

	return fmt.Errorf("ensure provided file %s is a .yml file", imageYaml)
}
