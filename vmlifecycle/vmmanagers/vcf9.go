package vmmanagers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// VCF9Credential represents VCF 9.0 authentication credentials
type VCF9Credential struct {
	Endpoint      string `yaml:"endpoint" validate:"required"`
	APIToken      string `yaml:"api_token" validate:"required"`
	TenantName    string `yaml:"tenant_name" validate:"required"`
	CACertificate string `yaml:"ca_certificate"` // Path to CA certificate file
}

// VCF9Config represents the configuration for VCF 9.0 VM creation
type VCF9Config struct {
	VCF9Credential `yaml:",inline"`
	ContextName    string `yaml:"context_name" validate:"required"`
	Namespace      string `yaml:"namespace" validate:"required"`
	Project        string `yaml:"project" validate:"required"`
	VMName         string `yaml:"vm_name"`
	VMClass        string `yaml:"vm_class"`
	StorageClass   string `yaml:"storage_class"`
	CPU            int    `yaml:"cpu"`
	Memory         string `yaml:"memory"`    // e.g., "4Gi"
	DiskSize       string `yaml:"disk_size"` // e.g., "40Gi"
	NetworkName    string `yaml:"network_name"`

	// SSH and Cloud-Init Configuration
	SSHPublicKey string `yaml:"ssh_public_key"` // SSH public key for VM access

	// Option 1: Use pre-uploaded image (recommended when you don't have vCenter admin access)
	ImageName string `yaml:"image_name"` // VMImage name (e.g., "vmi-5822d624965b8c561")

	// Option 2: Upload OVA to content library (requires vCenter admin access)
	ContentLibrary  string `yaml:"content_library"`  // Content library name for OVA upload
	VCenterEndpoint string `yaml:"vcenter_endpoint"` // vCenter endpoint for content library operations
	VCenterUsername string `yaml:"vcenter_username"` // vCenter username
	VCenterPassword string `yaml:"vcenter_password"` // vCenter password
}

// VCF9VMManager manages VMs using VCF 9.0 CLI and kubectl
type VCF9VMManager struct {
	Config        *OpsmanConfigFilePayload
	ImageOVA      string // OVA file to upload to content library
	State         StateInfo
	uploadedImage string // VMImage name after OVA upload
	vcfRunner     vcfRunner
	kubectlRunner kubectlRunner
	govcRunner    vcf9GovcRunner // For content library operations
	stdout        io.Writer
	stderr        io.Writer
}

//go:generate counterfeiter -o ./fakes/vcfRunner.go --fake-name VcfRunner . vcfRunner
type vcfRunner interface {
	ExecuteWithEnvVars(env []string, args []interface{}) (*bytes.Buffer, *bytes.Buffer, error)
}

//go:generate counterfeiter -o ./fakes/kubectlRunner.go --fake-name KubectlRunner . kubectlRunner
type kubectlRunner interface {
	ExecuteWithEnvVars(env []string, args []interface{}) (*bytes.Buffer, *bytes.Buffer, error)
}

//go:generate counterfeiter -o ./fakes/vcf9GovcRunner.go --fake-name VCF9GovcRunner . vcf9GovcRunner
type vcf9GovcRunner interface {
	ExecuteWithEnvVars(env []string, args []interface{}) (*bytes.Buffer, *bytes.Buffer, error)
}

// NewVCF9VMManager creates a new VCF 9.0 VM manager
func NewVCF9VMManager(config *OpsmanConfigFilePayload, imageOVA string, state StateInfo, stdout, stderr io.Writer, vcfRunner vcfRunner, kubectlRunner kubectlRunner, govcRunner vcf9GovcRunner) *VCF9VMManager {
	return &VCF9VMManager{
		Config:        config,
		ImageOVA:      imageOVA,
		State:         state,
		stdout:        stdout,
		stderr:        stderr,
		vcfRunner:     vcfRunner,
		kubectlRunner: kubectlRunner,
		govcRunner:    govcRunner,
	}
}

// CreateVM creates a VM using VCF 9.0 CLI and kubectl
func (v *VCF9VMManager) CreateVM() (Status, StateInfo, error) {
	if v.State.IAAS != "VCF9" && v.State.IAAS != "" {
		return Unknown, StateInfo{}, fmt.Errorf("authentication file provided is for VCF9, while the state file is for %s", v.State.IAAS)
	}

	err := validateIAASConfig(v.Config.OpsmanConfig.VCF9)
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	v.addDefaultConfigFields()

	// Step 1: Create VCF context
	_, _ = v.stdout.Write([]byte("Creating VCF context...\n"))
	err = v.createVCFContext()
	if err != nil {
		return Unknown, StateInfo{}, fmt.Errorf("failed to create VCF context: %w", err)
	}

	// Step 2: Switch to the context
	_, _ = v.stdout.Write([]byte("Switching to VCF context...\n"))
	err = v.useVCFContext()
	if err != nil {
		return Unknown, StateInfo{}, fmt.Errorf("failed to switch VCF context: %w", err)
	}

	// Step 3: Get or upload image
	var imageName string
	if v.Config.OpsmanConfig.VCF9.ImageName != "" {
		// Option 1: Use pre-uploaded image by name
		_, _ = v.stdout.Write([]byte(fmt.Sprintf("Using pre-uploaded VMImage: %s\n", v.Config.OpsmanConfig.VCF9.ImageName)))
		imageName = v.Config.OpsmanConfig.VCF9.ImageName
	} else {
		// Option 2: Upload OVA to content library
		// Validate OVA file
		err = v.validateImage()
		if err != nil {
			return Unknown, StateInfo{}, err
		}

		_, _ = v.stdout.Write([]byte("Checking content library and uploading OVA if necessary...\n"))
		imageName, err = v.uploadOVAToContentLibrary()
		if err != nil {
			return Unknown, StateInfo{}, fmt.Errorf("failed to upload OVA: %w", err)
		}
		_, _ = v.stdout.Write([]byte(fmt.Sprintf("OVA uploaded successfully as VMImage: %s\n", imageName)))
	}
	v.uploadedImage = imageName

	// Step 4: Check if VM already exists
	exist, err := v.vmExists()
	if err != nil {
		return Unknown, StateInfo{}, err
	}
	if exist {
		return Exist, v.State, nil
	}

	// Step 5: Create VM manifest file
	manifestFile, err := v.createVMManifest()
	if err != nil {
		return Unknown, StateInfo{}, fmt.Errorf("failed to create VM manifest: %w", err)
	}
	defer os.Remove(manifestFile)

	// Step 6: Create VM using kubectl
	_, _ = v.stdout.Write([]byte("Creating VM using kubectl...\n"))
	vmName, err := v.createVMWithKubectl(manifestFile)
	if err != nil {
		return Unknown, StateInfo{}, fmt.Errorf("failed to create VM: %w", err)
	}

	fullState := StateInfo{IAAS: "VCF9", ID: vmName}

	// Step 7: Wait for VM to be ready
	_, _ = v.stdout.Write([]byte("Waiting for VM to be ready...\n"))
	err = v.waitForVMReady(vmName)
	if err != nil {
		return Incomplete, fullState, fmt.Errorf("VM created but not ready: %w", err)
	}

	return Success, fullState, nil
}

// DeleteVM deletes a VM using VCF 9.0 CLI and kubectl
func (v *VCF9VMManager) DeleteVM() error {
	err := validateIAASConfig(v.Config.OpsmanConfig.VCF9.VCF9Credential)
	if err != nil {
		return err
	}

	if v.State.IAAS != "VCF9" {
		return fmt.Errorf("authentication file provided is for VCF9, while the state file is for %s", v.State.IAAS)
	}

	// Switch to the context
	err = v.useVCFContext()
	if err != nil {
		return fmt.Errorf("failed to switch VCF context: %w", err)
	}

	// Check if VM exists
	_, err = v.vmExists()
	if err != nil {
		return err
	}

	// Delete VM
	_, _ = v.stdout.Write([]byte(fmt.Sprintf("Deleting VM %s...\n", v.State.ID)))
	err = v.deleteVMWithKubectl(v.State.ID)
	if err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	return nil
}

// createVCFContext creates a VCF context using vcf CLI
func (v *VCF9VMManager) createVCFContext() error {
	config := v.Config.OpsmanConfig.VCF9

	// First, try to delete the context if it exists (to handle expired tokens)
	deleteArgs := []interface{}{
		"context", "delete", config.ContextName,
		"-y", // Skip confirmation
	}
	_, _, _ = v.vcfRunner.ExecuteWithEnvVars([]string{}, deleteArgs)
	// Ignore errors - context might not exist, which is fine

	// Now create the context with fresh credentials
	args := []interface{}{
		"context", "create", config.ContextName,
		"-e", config.Endpoint,
		"--api-token", config.APIToken,
		"--tenant-name", config.TenantName,
	}

	// Add CA certificate if provided
	if config.CACertificate != "" {
		args = append(args, "--ca-certificate", config.CACertificate)
	}

	_, stderr, err := v.vcfRunner.ExecuteWithEnvVars([]string{}, args)
	if err != nil {
		return fmt.Errorf("vcf context create failed: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// useVCFContext switches to the specified VCF context
func (v *VCF9VMManager) useVCFContext() error {
	config := v.Config.OpsmanConfig.VCF9

	// Build context string: contextName:namespace:project
	contextString := fmt.Sprintf("%s:%s:%s",
		config.ContextName,
		config.Namespace,
		config.Project,
	)

	args := []interface{}{
		"context", "use", contextString,
	}

	_, stderr, err := v.vcfRunner.ExecuteWithEnvVars([]string{}, args)
	if err != nil {
		return fmt.Errorf("vcf context use failed: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// vmExists checks if a VM with the given name exists
func (v *VCF9VMManager) vmExists() (bool, error) {
	vmName := v.State.ID
	if vmName == "" {
		vmName = v.Config.OpsmanConfig.VCF9.VMName
	}

	args := []interface{}{
		"get", "vm", vmName,
		"-o", "name",
	}

	stdout, stderr, err := v.kubectlRunner.ExecuteWithEnvVars([]string{}, args)
	if err != nil {
		if strings.Contains(stderr.String(), "not found") || strings.Contains(stderr.String(), "NotFound") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check VM existence: %w, stderr: %s", err, stderr.String())
	}

	// If we got output, VM exists
	if strings.TrimSpace(stdout.String()) != "" {
		v.State.ID = vmName
		v.State.IAAS = "VCF9"
		return true, nil
	}

	return false, nil
}

// getImageName returns the VMImage name from the uploaded OVA
func (v *VCF9VMManager) getImageName() (string, error) {
	if v.uploadedImage == "" {
		return "", errors.New("OVA has not been uploaded yet")
	}
	return v.uploadedImage, nil
}

// validateImage validates that the OVA file exists and is valid
func (v *VCF9VMManager) validateImage() error {
	_, err := os.Stat(v.ImageOVA)
	if err != nil {
		return fmt.Errorf("could not read OVA file: %w", err)
	}
	return nil
}

// findExistingLibraryItem checks if an item with the given name already exists in the content library
func (v *VCF9VMManager) findExistingLibraryItem(itemName string, env []string) (bool, error) {
	// If govc is not available, we can't check the content library
	if v.govcRunner == nil {
		return false, nil
	}

	config := v.Config.OpsmanConfig.VCF9

	// List all items in the content library using govc
	args := []interface{}{
		"library.ls",
		fmt.Sprintf("/%s/", config.ContentLibrary),
	}

	stdout, _, err := v.govcRunner.ExecuteWithEnvVars(env, args)
	if err != nil {
		// If the library doesn't exist or is empty, that's okay
		return false, nil
	}

	// Parse output to find matching item
	lines := strings.Split(stdout.String(), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Check if the line contains our item name
		if strings.Contains(line, itemName) {
			return true, nil
		}
	}

	return false, nil
}

// uploadOVAToContentLibrary uploads an OVA file to the vCenter content library
// and returns the VMImage name that can be used with kubectl
func (v *VCF9VMManager) uploadOVAToContentLibrary() (string, error) {
	config := v.Config.OpsmanConfig.VCF9

	// Check if govc is available
	if v.govcRunner == nil {
		return "", errors.New("govc CLI is required for OVA upload mode but was not found in PATH.\n" +
			"Please install govc:\n" +
			"  macOS:   brew install govmomi/tap/govc\n" +
			"  Linux:   curl -L -o - \"https://github.com/vmware/govmomi/releases/latest/download/govc_$(uname -s)_$(uname -m).tar.gz\" | tar -C /usr/local/bin -xvzf - govc\n" +
			"  Windows: choco install govc\n" +
			"Or use YAML mode with pre-uploaded images instead.")
	}

	// Validate required fields for OVA upload
	if config.ContentLibrary == "" {
		return "", errors.New("content_library must be specified in config for OVA upload")
	}
	if config.VCenterEndpoint == "" {
		return "", errors.New("vcenter_endpoint must be specified in config for OVA upload")
	}
	if config.VCenterUsername == "" {
		return "", errors.New("vcenter_username must be specified in config for OVA upload")
	}
	if config.VCenterPassword == "" {
		return "", errors.New("vcenter_password must be specified in config for OVA upload")
	}

	// Generate a deterministic name for the library item based on OVA filename
	// This allows us to detect if the same OVA has already been uploaded
	baseName := strings.TrimSuffix(filepath.Base(v.ImageOVA), filepath.Ext(v.ImageOVA))
	itemName := baseName

	// Set up govc environment
	env := []string{
		fmt.Sprintf("GOVC_URL=%s", config.VCenterEndpoint),
		fmt.Sprintf("GOVC_USERNAME=%s", config.VCenterUsername),
		fmt.Sprintf("GOVC_PASSWORD=%s", config.VCenterPassword),
		"GOVC_INSECURE=true", // TODO: Make this configurable based on ca_certificate
	}

	// Check if item already exists in content library
	exists, err := v.findExistingLibraryItem(itemName, env)
	if err != nil {
		return "", fmt.Errorf("failed to check for existing library item: %w", err)
	}

	if exists {
		_, _ = v.stdout.Write([]byte(fmt.Sprintf("Library item '%s' already exists in content library '%s'. Skipping upload.\n", itemName, config.ContentLibrary)))
	} else {
		// Upload OVA to content library
		_, _ = v.stdout.Write([]byte(fmt.Sprintf("Uploading OVA as library item: %s\n", itemName)))

		// Import OVA to content library using govc
		args := []interface{}{
			"library.import",
			"-n", itemName,
			config.ContentLibrary,
			v.ImageOVA,
		}

		stdout, stderr, err := v.govcRunner.ExecuteWithEnvVars(env, args)
		if err != nil {
			return "", fmt.Errorf("failed to upload OVA to content library: %w, stderr: %s", err, stderr.String())
		}

		_, _ = v.stdout.Write([]byte(fmt.Sprintf("OVA uploaded successfully: %s\n", stdout.String())))
	}

	// Wait for VMImage CRD to be created by VM Operator
	// The VM Operator watches the content library and creates VMImage resources
	_, _ = v.stdout.Write([]byte("Waiting for VMImage to be created by VM Operator...\n"))

	vmiName, err := v.waitForVMImage(itemName)
	if err != nil {
		return "", fmt.Errorf("OVA uploaded but VMImage not created: %w", err)
	}

	return vmiName, nil
}

// waitForVMImage waits for a VMImage resource to be created for the uploaded content library item
func (v *VCF9VMManager) waitForVMImage(itemName string) (string, error) {
	maxAttempts := 30
	sleepDuration := 10 * time.Second

	for i := 0; i < maxAttempts; i++ {
		// List all VMImages and look for one matching our item name
		args := []interface{}{
			"get", "vmimage",
			"-o", "jsonpath={range .items[*]}{.metadata.name}{\"\\t\"}{.status.name}{\"\\n\"}{end}",
		}

		stdout, _, err := v.kubectlRunner.ExecuteWithEnvVars([]string{}, args)
		if err == nil {
			// Parse output to find matching VMImage
			lines := strings.Split(stdout.String(), "\n")
			for _, line := range lines {
				parts := strings.Split(line, "\t")
				if len(parts) >= 2 {
					vmiName := strings.TrimSpace(parts[0])
					displayName := strings.TrimSpace(parts[1])

					// Check if display name matches our item name
					if strings.Contains(displayName, itemName) || strings.Contains(vmiName, itemName) {
						_, _ = v.stdout.Write([]byte(fmt.Sprintf("Found VMImage: %s\n", vmiName)))
						return vmiName, nil
					}
				}
			}
		}

		_, _ = v.stdout.Write([]byte(fmt.Sprintf("VMImage not ready yet, waiting... (attempt %d/%d)\n", i+1, maxAttempts)))
		time.Sleep(sleepDuration)
	}

	return "", fmt.Errorf("timeout waiting for VMImage to be created for item '%s'", itemName)
}

// createVMManifest creates a VM manifest YAML file
func (v *VCF9VMManager) createVMManifest() (string, error) {
	config := v.Config.OpsmanConfig.VCF9

	// Get image name from YAML file
	imageName, err := v.getImageName()
	if err != nil {
		return "", err
	}

	manifest := map[string]interface{}{
		"apiVersion": "vmoperator.vmware.com/v1alpha3",
		"kind":       "VirtualMachine",
		"metadata": map[string]interface{}{
			"name":      config.VMName,
			"namespace": config.Namespace,
		},
		"spec": map[string]interface{}{
			"className":    config.VMClass,
			"imageName":    imageName,
			"storageClass": config.StorageClass,
			"powerState":   "PoweredOn",
		},
	}

	// Add bootstrap configuration with cloud-init and SSH key
	if config.SSHPublicKey != "" {
		manifest["spec"].(map[string]interface{})["bootstrap"] = map[string]interface{}{
			"cloudInit": map[string]interface{}{
				"cloudConfig": map[string]interface{}{
					"defaultUserEnabled": false,
				},
				"sshAuthorizedKeys": []string{
					config.SSHPublicKey,
				},
			},
		}
	}

	// Add network if specified
	if config.NetworkName != "" {
		manifest["spec"].(map[string]interface{})["networkInterfaces"] = []map[string]interface{}{
			{
				"networkName": config.NetworkName,
			},
		}
	}

	// Note: VCF9 VMs don't use PVCs for disk management
	// The disk size is managed by the VM class and storage class
	// Unlike vSphere, we don't need to create separate PVCs

	// Marshal to YAML
	yamlData, err := yaml.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal VM manifest: %w", err)
	}

	// Write to temporary file
	tmpFile, err := os.CreateTemp("", "vm-manifest-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = tmpFile.Write(yamlData)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write manifest: %w", err)
	}

	tmpFile.Close()
	return tmpFile.Name(), nil
}

// createVMWithKubectl creates a VM using kubectl
func (v *VCF9VMManager) createVMWithKubectl(manifestFile string) (string, error) {
	args := []interface{}{
		"create", "-f", manifestFile,
	}

	stdout, stderr, err := v.kubectlRunner.ExecuteWithEnvVars([]string{}, args)
	if err != nil {
		return "", fmt.Errorf("kubectl create failed: %w, stderr: %s", err, stderr.String())
	}

	// Extract VM name from output (e.g., "virtualmachine.vmoperator.vmware.com/ops-manager-vm created")
	output := stdout.String()
	if strings.Contains(output, "created") {
		return v.Config.OpsmanConfig.VCF9.VMName, nil
	}

	return "", fmt.Errorf("unexpected kubectl output: %s", output)
}

// waitForVMReady waits for the VM to be in a ready state
func (v *VCF9VMManager) waitForVMReady(vmName string) error {
	maxAttempts := 60
	sleepDuration := 10 * time.Second

	for i := 0; i < maxAttempts; i++ {
		// Check if VM is powered on
		powerStateArgs := []interface{}{
			"get", "vm", vmName,
			"-o", "jsonpath={.status.powerState}",
		}

		stdout, _, err := v.kubectlRunner.ExecuteWithEnvVars([]string{}, powerStateArgs)
		if err == nil {
			powerState := strings.TrimSpace(stdout.String())
			if powerState == "PoweredOn" {
				// Also check if VirtualMachineCreated condition is true
				conditionArgs := []interface{}{
					"get", "vm", vmName,
					"-o", "jsonpath={.status.conditions[?(@.type=='VirtualMachineCreated')].status}",
				}

				condStdout, _, condErr := v.kubectlRunner.ExecuteWithEnvVars([]string{}, conditionArgs)
				if condErr == nil {
					condStatus := strings.TrimSpace(condStdout.String())
					if condStatus == "True" {
						_, _ = v.stdout.Write([]byte("VM is powered on and ready!\n"))
						return nil
					}
				}
			}
			_, _ = v.stdout.Write([]byte(fmt.Sprintf("VM power state: %s, waiting for VM to be created...\n", powerState)))
		} else {
			_, _ = v.stdout.Write([]byte(fmt.Sprintf("Waiting for VM status... (attempt %d/%d)\n", i+1, maxAttempts)))
		}

		time.Sleep(sleepDuration)
	}

	return errors.New("timeout waiting for VM to be ready")
}

// deleteVMWithKubectl deletes a VM using kubectl
func (v *VCF9VMManager) deleteVMWithKubectl(vmName string) error {
	args := []interface{}{
		"delete", "vm", vmName,
		"--wait=true",
	}

	_, stderr, err := v.kubectlRunner.ExecuteWithEnvVars([]string{}, args)
	if err != nil {
		if strings.Contains(stderr.String(), "not found") || strings.Contains(stderr.String(), "NotFound") {
			return nil // VM already deleted
		}
		return fmt.Errorf("kubectl delete failed: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// addDefaultConfigFields sets default values for optional configuration fields
func (v *VCF9VMManager) addDefaultConfigFields() {
	if v.Config.OpsmanConfig.VCF9.VMName == "" {
		v.Config.OpsmanConfig.VCF9.VMName = "ops-manager-vm"
	}
	if v.Config.OpsmanConfig.VCF9.VMClass == "" {
		v.Config.OpsmanConfig.VCF9.VMClass = "best-effort-medium"
	}
	if v.Config.OpsmanConfig.VCF9.StorageClass == "" {
		v.Config.OpsmanConfig.VCF9.StorageClass = "vsan-default-storage-policy"
	}
	if v.Config.OpsmanConfig.VCF9.Memory == "" {
		v.Config.OpsmanConfig.VCF9.Memory = "8Gi"
	}
	if v.Config.OpsmanConfig.VCF9.DiskSize == "" {
		v.Config.OpsmanConfig.VCF9.DiskSize = "160Gi"
	}
	if v.Config.OpsmanConfig.VCF9.CPU == 0 {
		v.Config.OpsmanConfig.VCF9.CPU = 2
	}
}
