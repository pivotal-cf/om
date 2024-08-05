package vmmanagers

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/pivotal-cf/om/vmlifecycle/runner"

	"gopkg.in/yaml.v2"
)

type AzureCredential struct {
	TenantID       string `yaml:"tenant_id"       validate:"required"`
	SubscriptionID string `yaml:"subscription_id" validate:"required"`
	ClientID       string `yaml:"client_id"       validate:"required"`
	ClientSecret   string `yaml:"client_secret"   validate:"required"`
	Location       string `yaml:"location"        validate:"required"`
	ResourceGroup  string `yaml:"resource_group"  validate:"required"`
	StorageAccount string `yaml:"storage_account" validate:"required"`
	StorageKey     string `yaml:"storage_key"`
}

type AzureConfig struct {
	AzureCredential            `yaml:",inline"`
	CloudName                  string `yaml:"cloud_name"`
	Container                  string `yaml:"container"`
	VPCSubnetDEPRECATED        string `yaml:"vpc_subnet,omitempty"`
	SubnetID                   string `yaml:"subnet_id" validate:"required"`
	NSG                        string `yaml:"network_security_group"`
	VMName                     string `yaml:"vm_name"`
	SSHPublicKey               string `yaml:"ssh_public_key" validate:"required"`
	BootDiskSize               string `yaml:"boot_disk_size"`
	PrivateIP                  string `yaml:"private_ip" validate:"omitempty,ip"`
	PublicIP                   string `yaml:"public_ip" validate:"omitempty,ip" `
	UseUnmanagedDiskDEPRECATED string `yaml:"use_unmanaged_disk,omitempty"`
	UseManagedDisk             string `yaml:"use_managed_disk"`
	StorageSKU                 string `yaml:"storage_sku,omitempty"`
	VMSize                     string `yaml:"vm_size"`
	Tags                       string `yaml:"tags"`
}

type AzureVMInfo struct {
	NetworkProfile AzureNetworkProfile `json:"networkProfile"`
	StorageProfile AzureStorageProfile `json:"storageProfile"`
}

type AzureNetworkProfile struct {
	NetworkInterfaces []ID `json:"networkInterfaces"`
}

type ID struct {
	ID string `json:"id"`
}

type AzureStorageProfile struct {
	ImageReference ID          `json:"imageReference"`
	OSDisk         AzureOSDisk `json:"osDisk"`
}

type AzureOSDisk struct {
	ManagedDisk ID  `json:"managedDisk"`
	VHD         VHD `json:"vhd"`
}

type VHD struct {
	URI string `json:"uri"`
}

//go:generate counterfeiter -o ./fakes/azureRunner.go --fake-name AzureRunner . azureRunner
type azureRunner interface {
	ExecuteWithEnvVars(env []string, args []interface{}) (*bytes.Buffer, *bytes.Buffer, error)
}

type AzureVMManager struct {
	stdout          io.Writer
	stderr          io.Writer
	Config          *OpsmanConfigFilePayload
	ImageYaml       string
	State           StateInfo
	runner          azureRunner
	pollingInterval time.Duration
}

func NewAzureVMManager(stdout, stderr io.Writer, config *OpsmanConfigFilePayload, imageYaml string, state StateInfo, azureRunner azureRunner, t time.Duration) *AzureVMManager {
	return &AzureVMManager{
		stdout:          stdout,
		stderr:          stderr,
		Config:          config,
		ImageYaml:       imageYaml,
		State:           state,
		runner:          azureRunner,
		pollingInterval: t,
	}
}

func (a *AzureVMManager) DeleteVM() error {
	err := validateIAASConfig(a.Config.OpsmanConfig.Azure.AzureCredential)
	if err != nil {
		return err
	}

	a.addDefaultConfigFields()

	err = a.authenticate()
	if err != nil {
		return err
	}

	if a.State.IAAS != "azure" {
		return fmt.Errorf("authentication file provided is for azure, while the state file is for %s", a.State.IAAS)
	}

	_, err = a.vmExists()
	if err != nil {
		return err
	}

	azVMInfo, err := a.getVMInfo()
	if err != nil {
		return err
	}

	var errs []string
	err = a.deleteVM()
	if err != nil {
		errs = append(errs, err.Error())
	}

	err = a.deleteDisk(azVMInfo)
	if err != nil {
		errs = append(errs, err.Error())
	}

	err = a.deleteNics(azVMInfo.NetworkProfile.NetworkInterfaces)
	if err != nil {
		errs = append(errs, err.Error())
	}

	if azVMInfo.StorageProfile.ImageReference.ID != "" {
		err = a.deleteImage(azVMInfo.StorageProfile.ImageReference.ID)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

func (a *AzureVMManager) CreateVM() (Status, StateInfo, error) {
	if a.State.IAAS != "azure" && a.State.IAAS != "" {
		return Unknown, StateInfo{}, fmt.Errorf("authentication file provided is for azure, while the state file is for %s", a.State.IAAS)
	}

	err := a.validateDeprecatedVars()
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	if a.Config.OpsmanConfig.Azure.UseManagedDisk != "" {
		_, err := strconv.ParseBool(a.Config.OpsmanConfig.Azure.UseManagedDisk)
		if err != nil {
			return Unknown, StateInfo{}, fmt.Errorf("expected \"use_managed_disk\" to be a boolean. Got: %s. %s", a.Config.OpsmanConfig.Azure.UseManagedDisk, err)
		}
	}

	err = validateIAASConfig(a.Config.OpsmanConfig.Azure)
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	if a.Config.OpsmanConfig.Azure.PublicIP == "" && a.Config.OpsmanConfig.Azure.PrivateIP == "" {
		return Unknown, StateInfo{}, errors.New("PublicIP and/or PrivateIP must be set")
	}

	a.addDefaultConfigFields()

	imageSourceURL, err := a.getImage()
	if err != nil {
		return Unknown, StateInfo{}, fmt.Errorf("azure error: %s", err)
	}

	err = a.authenticate()
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	exist, err := a.vmExists()
	if err != nil {
		return Unknown, StateInfo{}, err
	}
	if exist {
		return Exist, a.State, nil
	}

	imageName := a.generateImageName(imageSourceURL)

	err = a.copyImageBlob(imageSourceURL, imageName)
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	defer func() {
		_ = a.deleteBlobs()
	}()

	imageUri, err := a.getImageUri(imageName)
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	publicIP, err := a.getPublicIP()
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	imageId, err := a.createImage(imageUri, imageName)
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	err = a.createVM(imageUri, imageId, publicIP)

	if err != nil {
		return Unknown, StateInfo{}, err
	}

	return Success, StateInfo{IAAS: "azure", ID: a.Config.OpsmanConfig.Azure.VMName}, nil
}

func (a *AzureVMManager) addEnvVars() []string {
	azure := a.Config.OpsmanConfig.Azure

	if azure.StorageKey != "" {
		return []string{
			fmt.Sprintf("AZURE_STORAGE_KEY=%s", azure.StorageKey),
			fmt.Sprintf("AZURE_STORAGE_ACCOUNT=%s", azure.StorageAccount),
		}
	}
	return []string{
		fmt.Sprintf("AZURE_STORAGE_ACCOUNT=%s", azure.StorageAccount),
	}
}

func (a *AzureVMManager) getImage() (imageURL string, err error) {
	var images map[string]string

	contents, err := ioutil.ReadFile(a.ImageYaml)
	if err != nil {
		return "", err
	}

	err = checkImageFileIsYaml(a.ImageYaml)
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(contents, &images)
	if err != nil {
		return "", err
	}

	for k, v := range images {
		images[strings.Replace(k, "_", "", -1)] = v
	}

	if url, ok := images[a.Config.OpsmanConfig.Azure.Location]; ok {
		return url, nil
	}

	for _, url := range images {
		return url, nil
	}
	return "", nil
}

func (a *AzureVMManager) generateImageName(image string) string {
	digest := md5.New()
	_, _ = digest.Write([]byte(image))
	return fmt.Sprintf("opsman-image-%s", hex.EncodeToString(digest.Sum(nil)))
}

func (a *AzureVMManager) generateDiskName(disk string) string {
	digest := md5.New()
	_, _ = digest.Write([]byte(disk))
	return fmt.Sprintf("opsman-disk-%s", hex.EncodeToString(digest.Sum(nil)))
}

func (a *AzureVMManager) deleteVM() error {
	azure := a.Config.OpsmanConfig.Azure

	_, _, err := a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{
			"vm", "delete",
			"--yes",
			"--name", a.State.ID,
			"--resource-group", azure.ResourceGroup,
		})
	if err != nil {
		return fmt.Errorf("azure error deleting vm: %s", err)
	}

	return nil
}

func (a *AzureVMManager) authenticate() error {
	azure := a.Config.OpsmanConfig.Azure

	_, _, err := a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{
			"cloud", "set",
			"--name", azure.CloudName,
		})
	if err != nil {
		return fmt.Errorf("azure error: %s", err)
	}

	_, _, err = a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{
			"login", "--service-principal",
			"-u", azure.ClientID,
			"-p", runner.Redact(azure.ClientSecret),
			"--tenant", azure.TenantID,
		})
	if err != nil {
		return fmt.Errorf("azure error: %s", err)
	}

	_, _, err = a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{"account", "set", "--subscription", azure.SubscriptionID},
	)

	return checkFormatedError("azure error: %s", err)
}

func (a *AzureVMManager) copyImageBlob(imageSourceURL string, imageName string) error {
	azure := a.Config.OpsmanConfig.Azure

	_, _, err := a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{
			"storage", "blob", "copy", "start",
			"--source-uri", imageSourceURL,
			"--destination-container", azure.Container,
			"--destination-blob", imageName + ".vhd",
		})
	if err != nil {
		return fmt.Errorf("azure error: %s", err)
	}

	for {
		stdout, stderr, err := a.runner.ExecuteWithEnvVars(
			a.addEnvVars(),
			[]any{
				"storage", "blob", "show",
				"--name", imageName + ".vhd",
				"--container-name", azure.Container,
				"--query", "properties.copy.status",
			},
		)

		switch {
		case err != nil && strings.Contains(stderr.String(), "ErrorCode:PendingCopyOperation"):
			// Continue polling
		case err != nil:
			return fmt.Errorf("azure error: %s", err)
		case strings.TrimSpace(stdout.String()) == `"success"`:
			return nil // Operation is finished
		}

		_, _ = a.stderr.Write([]byte(fmt.Sprintf("blob not ready yet, polling in %s\n", a.pollingInterval)))
		time.Sleep(a.pollingInterval)
	}
}

func (a *AzureVMManager) addDefaultConfigFields() {
	if a.Config.OpsmanConfig.Azure.VMName == "" {
		a.Config.OpsmanConfig.Azure.VMName = "ops-manager-vm"
	}
	if a.Config.OpsmanConfig.Azure.BootDiskSize == "" {
		a.Config.OpsmanConfig.Azure.BootDiskSize = "200"
	}
	if a.Config.OpsmanConfig.Azure.CloudName == "" {
		a.Config.OpsmanConfig.Azure.CloudName = "AzureCloud"
	}
	if a.Config.OpsmanConfig.Azure.UseManagedDisk == "" {
		a.Config.OpsmanConfig.Azure.UseManagedDisk = "true"
	}
	if a.Config.OpsmanConfig.Azure.Container == "" {
		a.Config.OpsmanConfig.Azure.Container = "opsmanagerimage"
	}
	if a.Config.OpsmanConfig.Azure.VMSize == "" {
		a.Config.OpsmanConfig.Azure.VMSize = "Standard_DS2_v2"
	}
	if a.Config.OpsmanConfig.Azure.StorageSKU == "" {
		a.Config.OpsmanConfig.Azure.StorageSKU = "Standard_LRS"
	}
}

func (a *AzureVMManager) getImageUri(imageName string) (imageURI string, err error) {
	out, _, err := a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{
			"storage", "blob", "url",
			"--name", imageName + ".vhd",
			"--container-name", a.Config.OpsmanConfig.Azure.Container,
		})

	return cleanupString(out.String()), checkFormatedError("azure error: %s", err)
}

func (a *AzureVMManager) getPublicIP() (publicIP string, err error) {
	if a.Config.OpsmanConfig.Azure.PublicIP == "" {
		return "", nil
	}

	out, _, err := a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{
			"network", "public-ip", "list",
			"--query", fmt.Sprintf("[?ipAddress == '%s'].name | [0]", a.Config.OpsmanConfig.Azure.PublicIP),
		})

	if out.String() == "" {
		return "", fmt.Errorf("could not find resource assignment for public ip address %s", a.Config.OpsmanConfig.Azure.PublicIP)
	}

	return cleanupString(out.String()), checkFormatedError("azure error: %s", err)
}

func (a *AzureVMManager) createImage(imageUri string, imageName string) (string, error) {
	azure := a.Config.OpsmanConfig.Azure

	if azure.UseUnmanagedDiskDEPRECATED == "true" || azure.UseManagedDisk == "false" {
		return "", nil
	}
	out, _, err := a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{
			"image", "create",
			"--resource-group", azure.ResourceGroup,
			"--name", imageName,
			"--source", imageUri,
			"--location", azure.Location,
			"--os-type", "Linux",
			"--query", "id",
		})
	if err != nil {
		return "", fmt.Errorf("azure error: %s", err)
	}

	return cleanupString(out.String()), nil
}

func (a *AzureVMManager) createVM(imageUrl string, imageID string, publicIP string) error {
	azure := a.Config.OpsmanConfig.Azure

	file, err := ioutil.TempFile("", "ssh-key")
	if err != nil {
		return fmt.Errorf("azure error: %s", err)
	}
	_, _ = file.WriteString(azure.SSHPublicKey)
	file.Close()

	args := []interface{}{
		"vm", "create",
		"--name", azure.VMName,
		"--resource-group", azure.ResourceGroup,
		"--location", azure.Location,
		"--os-disk-size-gb", azure.BootDiskSize,
		"--admin-username", "ubuntu",
		"--size", azure.VMSize,
		"--ssh-key-value", file.Name(),
		"--subnet", azure.SubnetID,
		"--nsg", azure.NSG,
		"--public-ip-address", publicIP,
	}

	if azure.PrivateIP != "" {
		args = append(args, "--private-ip-address", azure.PrivateIP)
	}

	if azure.UseUnmanagedDiskDEPRECATED == "true" || azure.UseManagedDisk == "false" {
		args = append(args,
			"--use-unmanaged-disk",
			"--os-disk-name", a.generateDiskName(imageUrl),
			"--os-type", "Linux",
			"--storage-container-name", azure.Container,
			"--storage-account", azure.StorageAccount,
			"--image", imageUrl,
		)
	} else {
		args = append(args,
			"--storage-sku", azure.StorageSKU,
			"--image", imageID,
		)
	}

	if azure.Tags != "" {
		args = append(args, "--tags", azure.Tags)
	}

	_, _, err = a.runner.ExecuteWithEnvVars(a.addEnvVars(), args)

	return checkFormatedError("azure error: %s", err)
}

func (a *AzureVMManager) getVMInfo() (AzureVMInfo, error) {
	azure := a.Config.OpsmanConfig.Azure

	out, _, err := a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{
			"vm", "show",
			"--name", a.State.ID,
			"--resource-group", azure.ResourceGroup,
		})
	if err != nil {
		return AzureVMInfo{}, fmt.Errorf("azure error getting vm info: %s", err)
	}

	azVMInfo := AzureVMInfo{}
	err = json.Unmarshal(out.Bytes(), &azVMInfo)
	if err != nil {
		return AzureVMInfo{}, fmt.Errorf("azure error unmarshalling vm info: %s", err)
	}

	return azVMInfo, nil
}

func (a *AzureVMManager) deleteNics(ids []ID) error {
	args := []interface{}{
		"network", "nic", "delete",
		"--ids",
	}

	for _, id := range ids {
		args = append(args, id.ID)
	}

	_, _, err := a.runner.ExecuteWithEnvVars(a.addEnvVars(), args)
	if err != nil {
		return fmt.Errorf("azure error deleting nic: %s", err)
	}

	return nil
}

func (a *AzureVMManager) deleteImage(id string) error {
	_, _, err := a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{
			"image", "delete",
			"--ids", id,
		})
	if err != nil {
		return fmt.Errorf("azure error deleting image: %s", err)
	}

	return nil
}

func (a *AzureVMManager) vmExists() (vmExists bool, err error) {
	if a.State.ID == "" {
		return false, nil
	}

	vmID, errBuffWriter, err := a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{
			"vm", "show",
			"--name", a.State.ID,
			"--resource-group", a.Config.OpsmanConfig.Azure.ResourceGroup,
			"--query", "id",
		})

	if err == nil {
		if cleanupString(vmID.String()) == "" { // state file has id, but vm not exist case
			return false, fmt.Errorf("Could not find VM with ID %q.\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile.", a.State.ID)
		}
		return true, nil
	}

	if errBuffWriter != nil {
		errStr := errBuffWriter.String()
		if strings.Contains(errStr, fmt.Sprintf("The Resource 'Microsoft.Compute/virtualMachines/%s' under resource group '%s' was not found.", a.State.ID, a.Config.OpsmanConfig.Azure.ResourceGroup)) {
			return false, fmt.Errorf("Could not find VM with ID %q.\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile.", a.State.ID)
		}
	}

	return false, err
}

func (a *AzureVMManager) deleteBlobs() error {
	_, _, err := a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{
			"storage", "blob", "delete-batch",
			"--source", a.Config.OpsmanConfig.Azure.Container,
			"--pattern", "opsman-image-*.vhd",
		})

	return err
}

func (a *AzureVMManager) deleteDisk(vmInfo AzureVMInfo) error {
	if vmInfo.StorageProfile.OSDisk.ManagedDisk.ID != "" {
		return a.deleteManagedDisk(vmInfo.StorageProfile.OSDisk.ManagedDisk.ID)
	}
	if vmInfo.StorageProfile.OSDisk.VHD.URI != "" {
		return a.deleteUnmanagedDisk(vmInfo.StorageProfile.OSDisk.VHD.URI)
	}

	_, err := fmt.Fprintln(a.stderr, "could not find disk to cleanup. Doing Nothing.")
	return err
}

func (a *AzureVMManager) deleteUnmanagedDisk(diskURI string) error {
	splitURI := strings.Split(diskURI, "/")
	container := splitURI[len(splitURI)-2]
	diskName := splitURI[len(splitURI)-1]
	_, _, err := a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{
			"storage", "blob", "delete",
			"--container-name", container,
			"--name", diskName,
		})
	return err
}

func (a *AzureVMManager) deleteManagedDisk(id string) error {
	_, _, err := a.runner.ExecuteWithEnvVars(a.addEnvVars(),
		[]interface{}{
			"disk", "delete",
			"--yes",
			"--ids", id,
		})
	if err != nil {
		return fmt.Errorf("azure error deleting disk: %s", err)
	}

	return nil
}

func (a *AzureVMManager) validateDeprecatedVars() error {
	if a.Config.OpsmanConfig.Azure.VPCSubnetDEPRECATED != "" && a.Config.OpsmanConfig.Azure.SubnetID != "" {
		return errors.New("\"vpc_subnet\" is DEPRECATED. Cannot use \"vpc_subnet\" and \"subnet_id\" together. Use \"subnet_id\" instead")
	}

	if a.Config.OpsmanConfig.Azure.VPCSubnetDEPRECATED != "" {
		log.Println("\"vpc_subnet\" is DEPRECATED. Please use \"subnet_id\"")
		a.Config.OpsmanConfig.Azure.SubnetID = a.Config.OpsmanConfig.Azure.VPCSubnetDEPRECATED
	}

	if a.Config.OpsmanConfig.Azure.UseManagedDisk != "" && a.Config.OpsmanConfig.Azure.UseUnmanagedDiskDEPRECATED != "" {
		return errors.New("\"use_unmanaged_disk\" is DEPRECATED. Cannot use \"use_unmanaged_disk\" and \"use_managed_disk\" together. Use \"use_managed_disk\" instead")
	}

	if a.Config.OpsmanConfig.Azure.UseUnmanagedDiskDEPRECATED != "" {
		unmanagedDiskBool, err := strconv.ParseBool(a.Config.OpsmanConfig.Azure.UseUnmanagedDiskDEPRECATED)
		if err != nil {
			return fmt.Errorf("expected use_unmanaged_disk to be a boolean. Got: %s. %s", a.Config.OpsmanConfig.Azure.UseUnmanagedDiskDEPRECATED, err)
		}
		log.Printf("\"use_unmanaged_disk\" is DEPRECATED. Please use \"use_managed_disk:%v\"", !unmanagedDiskBool)
	}

	return nil
}
