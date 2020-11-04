package vmmanagers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/blang/semver"
	"github.com/pivotal-cf/om/vmlifecycle/extractopsmansemver"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"fmt"

	"archive/tar"
)

type VcenterCredential struct {
	URL      string `yaml:"url" validate:"required"`
	Username string `yaml:"username" validate:"required"`
	Password string `yaml:"password" validate:"required"`
}

type Vcenter struct {
	VcenterCredential `yaml:",inline"`
	Datacenter        string `yaml:"datacenter" validate:"required"`
	Datastore         string `yaml:"datastore" validate:"required"`
	Insecure          string `yaml:"insecure"`
	CACert            string `yaml:"ca_cert,omitempty"`
	ResourcePool      string `yaml:"resource_pool" validate:"required"`
	HostDEPRECATED    string `yaml:"host,omitempty"`
	Folder            string `yaml:"folder"`
}

type VsphereConfig struct {
	Vcenter      `yaml:"vcenter"`
	DiskType     string `yaml:"disk_type" validate:"required"`
	PrivateIP    string `yaml:"private_ip" validate:"required,ip"`
	DNS          string `yaml:"dns" validate:"required"`
	NTP          string `yaml:"ntp" validate:"required"`
	SSHPassword  string `yaml:"ssh_password,omitempty"`
	SSHPublicKey string `yaml:"ssh_public_key"`
	Hostname     string `yaml:"hostname" validate:"required"`
	Network      string `yaml:"network"  validate:"required"`
	Netmask      string `yaml:"netmask" validate:"required"`
	Gateway      string `yaml:"gateway" validate:"required"`
	VMName       string `yaml:"vm_name"`
	Memory       string `yaml:"memory"`
	CPU          string `yaml:"cpu"`
}

type ovaJSONConfig struct {
	DiskProvisioning   string
	IPAllocationPolicy string
	PropertyMapping    []propertyMapping
	NetworkMapping     []networkMapping
	Annotation         string
	PowerOn            bool
	InjectOvfEnv       bool
	WaitForIP          bool
	Name               string
}

type propertyMapping struct {
	Key, Value string
}

type networkMapping struct {
	Name, Network string
}

//go:generate counterfeiter -o ./fakes/govcRunner.go --fake-name GovcRunner . govcRunner
type govcRunner interface {
	ExecuteWithEnvVars(env []string, args []interface{}) (*bytes.Buffer, *bytes.Buffer, error)
}

type VsphereVMManager struct {
	ImageOVA string
	State    StateInfo
	Config   *OpsmanConfigFilePayload
	runner   govcRunner
}

const DefaultMemory = "8"
const DefaultCPU = "1"

func NewVsphereVMManager(config *OpsmanConfigFilePayload, imageFileName string, state StateInfo, govcRunner govcRunner) *VsphereVMManager {
	return &VsphereVMManager{
		ImageOVA: imageFileName,
		State:    state,
		Config:   config,
		runner:   govcRunner,
	}
}

func (v *VsphereVMManager) DeleteVM() error {
	err := validateIAASConfig(v.Config.OpsmanConfig.Vsphere.Vcenter.VcenterCredential)
	if err != nil {
		return err
	}

	env, err := v.addEnvVars()
	if err != nil {
		return err
	}

	if v.State.IAAS != "vsphere" {
		return fmt.Errorf("authentication file provided is for vsphere, while the state file is for %s", v.State.IAAS)
	}

	_, err = v.vmExists(env)
	if err != nil {
		return err
	}

	err = v.powerOffVM(env, v.State.ID)
	if err != nil {
		return err
	}

	return v.deleteVM(env, v.State.ID)
}

func (v *VsphereVMManager) CreateVM() (Status, StateInfo, error) {
	if v.State.IAAS != "vsphere" && v.State.IAAS != "" {
		return Unknown, StateInfo{}, fmt.Errorf("authentication file provided is for vsphere, while the state file is for %s", v.State.IAAS)
	}

	err := v.validateVsphereConfig()
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	err = validateIAASConfig(v.Config.OpsmanConfig.Vsphere)
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	v.addDefaultConfigFields()

	env, err := v.addEnvVars()
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	exist, err := v.vmExists(env)
	if err != nil {
		return Unknown, StateInfo{}, err
	}
	if exist {
		return Exist, v.State, nil
	}

	err = v.validateImage()
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	optionFilename, err := v.createOptionsFile()
	if err != nil {
		return Unknown, StateInfo{}, err
	}

	ipath := v.createIpath()

	errBufWriter, err := v.createVM(env, optionFilename)
	fullState := StateInfo{IAAS: "vsphere", ID: ipath}

	if err != nil {
		errBufStr := errBufWriter.String()
		if strings.Contains(errBufStr, "already exists") {
			return Exist, fullState, nil
		}

		return Unknown, StateInfo{}, fmt.Errorf("govc error: %s", err)
	}

	if v.Config.OpsmanConfig.Vsphere.Memory != DefaultMemory || v.Config.OpsmanConfig.Vsphere.CPU != DefaultCPU {
		err := v.updateVM(env, ipath)
		if err != nil {
			return Incomplete, fullState, err
		}
	}

	return Success, fullState, nil
}

func (v *VsphereVMManager) createOptionsFile() (optionsFileName string, err error) {
	options := ovaJSONConfig{
		DiskProvisioning:   v.Config.OpsmanConfig.Vsphere.DiskType,
		IPAllocationPolicy: "",
		PropertyMapping: []propertyMapping{
			{
				Key:   "ip0",
				Value: v.Config.OpsmanConfig.Vsphere.PrivateIP,
			},
			{
				Key:   "netmask0",
				Value: v.Config.OpsmanConfig.Vsphere.Netmask,
			},
			{
				Key:   "gateway",
				Value: v.Config.OpsmanConfig.Vsphere.Gateway,
			},
			{
				Key:   "DNS",
				Value: v.Config.OpsmanConfig.Vsphere.DNS,
			},
			{
				Key:   "ntp_servers",
				Value: v.Config.OpsmanConfig.Vsphere.NTP,
			},
			{
				Key:   "admin_password",
				Value: v.Config.OpsmanConfig.Vsphere.SSHPassword,
			},
			{
				Key:   "public_ssh_key",
				Value: v.Config.OpsmanConfig.Vsphere.SSHPublicKey,
			},
			{
				Key:   "custom_hostname",
				Value: v.Config.OpsmanConfig.Vsphere.Hostname,
			},
		},
		NetworkMapping: []networkMapping{{
			Name:    "Network 1",
			Network: v.Config.OpsmanConfig.Vsphere.Network,
		}},
		Annotation:   "Ops Manager for Pivotal Cloud Foundry\ninstalls and manages PCF products and services.",
		PowerOn:      true,
		InjectOvfEnv: false,
		WaitForIP:    false,
		Name:         v.Config.OpsmanConfig.Vsphere.VMName,
	}

	optionsBytes, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("options failed to marshal: %s", err)
	}

	optionsFile, err := ioutil.TempFile("", "options.json")
	if err != nil {
		return "", fmt.Errorf("could not create temp option file: %s", err)
	}

	err = ioutil.WriteFile(optionsFile.Name(), optionsBytes, 0644)
	if err != nil {
		return "", fmt.Errorf("could not write options to file: %s", err)
	}

	err = optionsFile.Close()
	if err != nil {
		return "", fmt.Errorf("could not write options to file: %s", err)
	}

	return optionsFile.Name(), nil
}

func (v *VsphereVMManager) addEnvVars() (envVarsList []string, err error) {
	if v.Config.OpsmanConfig.Vsphere.Vcenter.HostDEPRECATED != "" {
		log.Println("vcenter \"host\" is DEPRECATED. Platform Automation cannot guarantee the location of the VM, given the nature of vSphere")
	}

	env := append(os.Environ(),
		"GOVC_URL="+v.Config.OpsmanConfig.Vsphere.Vcenter.URL,
		"GOVC_USERNAME="+v.Config.OpsmanConfig.Vsphere.Vcenter.Username,
		"GOVC_PASSWORD="+v.Config.OpsmanConfig.Vsphere.Vcenter.Password,
		"GOVC_DATASTORE="+v.Config.OpsmanConfig.Vsphere.Vcenter.Datastore,
		"GOVC_DATACENTER="+v.Config.OpsmanConfig.Vsphere.Vcenter.Datacenter,
		"GOVC_INSECURE="+v.Config.OpsmanConfig.Vsphere.Vcenter.Insecure,
		"GOVC_NETWORK="+v.Config.OpsmanConfig.Vsphere.Network,
		"GOVC_RESOURCE_POOL="+v.Config.OpsmanConfig.Vsphere.Vcenter.ResourcePool,
		"GOVC_HOST="+v.Config.OpsmanConfig.Vsphere.Vcenter.HostDEPRECATED,
		"GOVC_FOLDER="+v.Config.OpsmanConfig.Vsphere.Vcenter.Folder,
		"GOMAXPROCS=1",
	)

	if v.Config.OpsmanConfig.Vsphere.Vcenter.CACert != "" {
		caCertFile, err := ioutil.TempFile("", "ca.crt")
		if err != nil {
			return []string{}, fmt.Errorf("could not create temp file for ca cert: %s", err)
		}
		_, err = caCertFile.WriteString(v.Config.OpsmanConfig.Vsphere.Vcenter.CACert)
		if err != nil {
			return []string{}, fmt.Errorf("could not write cert to the cert file: %s", err)
		}
		env = append(env, "GOVC_TLS_CA_CERTS="+caCertFile.Name())
	}

	return env, nil
}

func (v *VsphereVMManager) validateImage() error {
	imageFile, err := os.Open(v.ImageOVA)
	if err != nil {
		return fmt.Errorf("could not read image file: %s", err)
	}
	imageTar := tar.NewReader(imageFile)

	for {
		fileHeader, err := imageTar.Next()
		if err != nil {
			return fmt.Errorf("could not validate image-file format of %s. Is your image an OVA file? %s", v.ImageOVA, err.Error())

		}
		if strings.Contains(fileHeader.Name, ".ovf") {
			return nil
		}
	}
}

func (v *VsphereVMManager) createVM(env []string, optionFilename string) (errorBuffer *bytes.Buffer, err error) {
	_, errBufWriter, err := v.runner.ExecuteWithEnvVars(env, []interface{}{"import.ova",
		"-options=" + optionFilename,
		v.ImageOVA})

	return errBufWriter, checkFormatedError("govc error: %s", err)
}

func (v *VsphereVMManager) addDefaultConfigFields() {
	if v.Config.OpsmanConfig.Vsphere.Vcenter.Insecure == "" {
		v.Config.OpsmanConfig.Vsphere.Vcenter.Insecure = "0"
	}
	if v.Config.OpsmanConfig.Vsphere.CPU == "" {
		v.Config.OpsmanConfig.Vsphere.CPU = "1"
	}
	if v.Config.OpsmanConfig.Vsphere.Memory == "" {
		v.Config.OpsmanConfig.Vsphere.Memory = "8"
	}
	if v.Config.OpsmanConfig.Vsphere.VMName == "" {
		v.Config.OpsmanConfig.Vsphere.VMName = "ops-manager-vm"
	}
}

func (v *VsphereVMManager) validateVsphereConfig() error {
	var errs []string

	opsmanVersion, _ := extractopsmansemver.Do(v.ImageOVA)
	if opsmanVersion.GTE(semver.MustParse("2.6.0")) {
		if v.Config.OpsmanConfig.Vsphere.SSHPublicKey == "" {
			errs = append(errs, "'ssh_public_key' is required for OpsManager 2.6+")
		}

		if v.Config.OpsmanConfig.Vsphere.SSHPassword != "" {
			errs = append(errs, "'ssh_password' cannot be used with OpsManager 2.6+")
		}
	} else {
		if v.Config.OpsmanConfig.Vsphere.SSHPassword == "" && v.Config.OpsmanConfig.Vsphere.SSHPublicKey == "" {
			errs = append(errs, "'ssh_password' or 'ssh_public_key' must be set")
		}
	}

	if v.Config.OpsmanConfig.Vsphere.Insecure == "0" {
		if v.Config.OpsmanConfig.Vsphere.CACert == "" {
			errs = append(errs, "'ca_cert' is required if 'insecure' is set to 0 (secure)")
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

func (v *VsphereVMManager) updateVM(env []string, ipath string) error {
	log.Println("Setting Memory and CPU for the VM...")
	err := v.powerOffVM(env, ipath)
	if err != nil {
		return fmt.Errorf("govc error: could not turn off VM")
	}

	err = v.setVMProperties(env, ipath)
	if err != nil {
		return fmt.Errorf("govc error: could not reassign memory and CPU")
	}

	err = v.powerOnVM(env, ipath)
	if err != nil {
		return fmt.Errorf("govc error: could not turn on VM")
	}

	log.Println("Memory and CPU set")
	return nil
}

func (v *VsphereVMManager) powerOffVM(env []string, vmID string) error {
	_, stderr, err := v.runner.ExecuteWithEnvVars(env, []interface{}{"vm.power",
		"-off=true",
		fmt.Sprintf("-vm.ipath=%s", vmID)})

	if err != nil && strings.Contains(stderr.String(), "attempted operation cannot be performed in the current state (Powered off)") {

		return nil
	}

	return checkFormatedError("govc error: %s", err)
}

func (v *VsphereVMManager) setVMProperties(env []string, ipath string) error {
	memory, err := v.translateGBToMB()
	if err != nil {
		return fmt.Errorf("could not parse memory as an integer: %v", err)
	}
	_, _, err = v.runner.ExecuteWithEnvVars(env, []interface{}{"vm.change",
		"-vm.ipath=" + ipath,
		"-m=" + memory,
		"-c=" + v.Config.OpsmanConfig.Vsphere.CPU,
	})
	return err
}

func (v *VsphereVMManager) powerOnVM(env []string, vmID string) error {
	_, _, err := v.runner.ExecuteWithEnvVars(env, []interface{}{"vm.power",
		"-on=true",
		fmt.Sprintf(`-vm.ipath=%s`, vmID),
	})
	return err
}

func (v *VsphereVMManager) deleteVM(env []string, vmID string) error {
	_, _, err := v.runner.ExecuteWithEnvVars(env, []interface{}{`vm.destroy`,
		fmt.Sprintf(`-vm.ipath=%s`, vmID),
	})

	return checkFormatedError("govc error: %s", err)
}

func (v *VsphereVMManager) createIpath() string {
	datacenterPrefix := fmt.Sprintf("/%s/vm/", v.Config.OpsmanConfig.Vsphere.Datacenter)
	folder := strings.Replace(v.Config.OpsmanConfig.Vsphere.Folder, datacenterPrefix, "", 1)

	return strings.Join([]string{
		"",
		v.Config.OpsmanConfig.Vsphere.Datacenter,
		"vm",
		folder,
		v.Config.OpsmanConfig.Vsphere.VMName,
	}, "/")
}

func (v *VsphereVMManager) vmExists(env []string) (vmExists bool, err error) {
	if v.State.ID == "" {
		return false, nil
	}

	_, _, err = v.runner.ExecuteWithEnvVars(env, []interface{}{"vm.info", "-vm.ipath=" + v.State.ID})
	if err != nil {
		return false, fmt.Errorf("error: %s\n       Could not find VM with ID %q.\n       To fix, ensure the VM ID in the statefile matches a VM that exists.\n       If the VM has already been deleted, delete the contents of the statefile.", err, v.State.ID)
	}

	return true, nil
}

func (v *VsphereVMManager) translateGBToMB() (string, error) {
	temp, err := strconv.Atoi(v.Config.OpsmanConfig.Vsphere.Memory)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(temp * 1024), nil
}
