package vmlifecyclecommands

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/pivotal-cf/om/interpolate"
	"github.com/pivotal-cf/om/vmlifecycle/extractopsmansemver"
	"github.com/pivotal-cf/om/vmlifecycle/runner"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type opsmanVersion struct {
	Info struct {
		Version string `yaml:"version"`
	} `yaml:"info"`
}

type processingError int

const (
	ERROR     processingError = 2
	DOWNGRADE processingError = 1
	EQUAL     processingError = 0
	UPGRADE   processingError = -1
)

type target struct {
	Target               string `yaml:"target"`
	DecryptionPassphrase string `yaml:"decryption-passphrase"`
}

type UpgradeOpsman struct {
	stdout             io.Writer
	stderr             io.Writer
	pollingInterval    time.Duration
	timeout            time.Duration
	Recreate           bool `long:"recreate" description:"Force recreate the Ops Manager VM"`
	ImportInstallation struct {
		Installation string `long:"installation" description:"Path to installation" required:"true"`
		EnvFile      string `long:"env-file" description:"Ops Manager Environment File" required:"true"`
	} `group:"om import-installation flags"`
	CreateVM *CreateVM `group:"vm management flags"`
	DeleteVM *DeleteVM `no-flag:"true"`
	omRunner *runner.Runner
}

func NewUpgradeOpsman(stdout, stderr io.Writer, createVM *CreateVM, deleteVM *DeleteVM, omRunner *runner.Runner, interval time.Duration, timeout time.Duration) *UpgradeOpsman {
	return &UpgradeOpsman{
		stdout:          stdout,
		stderr:          stderr,
		CreateVM:        createVM,
		DeleteVM:        deleteVM,
		omRunner:        omRunner,
		pollingInterval: interval,
		timeout:         timeout,
	}
}

func (n *UpgradeOpsman) Execute(args []string) error {
	if err := n.validate(); err != nil {
		return err
	}

	n.setup()

	outBuf, errBuf, err := n.omRunner.Execute([]interface{}{"--env", n.ImportInstallation.EnvFile, "--skip-ssl-validation", "curl", "--path", "/api/v0/info"})
	if err != nil {
		if e, ok := err.(*exec.Error); ok && e.Err == exec.ErrNotFound {
			return err
		}
		err := n.checkCredentials(errBuf.String())
		if err != nil {
			return err
		}

		return n.createAndImport()
	}

	compare, err := n.compareOpsmanVersions(outBuf)
	if compare == ERROR {
		return err
	}
	if compare == UPGRADE {
		return n.upgradeOpsman()
	}
	if compare == DOWNGRADE {
		return fmt.Errorf("downgrading is not supported by Ops Manager")
	}
	if compare == EQUAL {
		if !n.Recreate {
			_, _ = n.stdout.Write([]byte("the same version of opsman installed, nothing to change\n"))

			return nil
		}

		_, _ = n.stdout.Write([]byte("recreating the opsman VM\n"))

		return n.upgradeOpsman()
	}

	return fmt.Errorf("unexpected error in upgrading opsman")
}

func (n *UpgradeOpsman) setup() {
	n.DeleteVM.StateFile = n.CreateVM.StateFile
	n.DeleteVM.Config = n.CreateVM.Config
	n.DeleteVM.VarsFile = n.CreateVM.VarsFile
	n.DeleteVM.VarsEnv = n.CreateVM.VarsEnv
}

func (n *UpgradeOpsman) checkCredentials(errInfo string) error {
	if strings.Contains(strings.ToLower(errInfo), "bad credential") {
		return fmt.Errorf("could not authenticate with Ops Manager %s", errInfo)
	}
	return nil
}

func (n *UpgradeOpsman) createAndImport() error {
	_, _ = n.stdout.Write([]byte("creating the new opsman vm\n"))
	err := n.CreateVM.Execute([]string{})
	if err != nil {
		return fmt.Errorf("could not create the vm: %s", err)
	}

	_, _ = n.stdout.Write([]byte("importing the old installation\n"))
	return n.pollImportInstallation()
}

func (n *UpgradeOpsman) compareOpsmanVersions(outBuf *bytes.Buffer) (processingError, error) {
	versionJson := outBuf.Bytes()

	version := opsmanVersion{}

	err := yaml.Unmarshal(versionJson, &version)
	if err != nil {
		return ERROR, fmt.Errorf("fatal error %s", err)
	}
	if version.Info.Version == "" {
		return ERROR, fmt.Errorf("info struct could not be parsed")
	}

	versionToInstall, err := extractopsmansemver.Do(n.CreateVM.ImageFile) //[ops-manager,2.2.0]OpsManager2.2-build.296onGCP.yml
	if err != nil {
		return ERROR, fmt.Errorf("the file name '%s' for the Ops Manager image needs to contain the original version number as downloaded from Pivnet (ie 'OpsManager2.2-build.296onGCP.yml')", n.CreateVM.ImageFile)
	}
	versionInstalled, err := extractopsmansemver.Do(version.Info.Version)
	if err != nil {
		log.Fatal(err)
	}

	return processingError(versionInstalled.Compare(versionToInstall)), nil
}

func (n *UpgradeOpsman) upgradeOpsman() error {
	_, _ = n.stdout.Write([]byte("deleting the old opsman vm\n"))
	err := n.DeleteVM.Execute([]string{})
	if err != nil {
		return err
	}

	_, _ = n.stdout.Write([]byte("creating the new opsman vm\n"))
	err = n.CreateVM.Execute([]string{})
	if err != nil {
		return err
	}

	time.Sleep(n.pollingInterval)

	_, _ = n.stdout.Write([]byte("importing the old installation\n"))
	return n.pollImportInstallation()
}

func (n *UpgradeOpsman) pollImportInstallation() error {
	var err error

	start := time.Now()
	for {
		_, _, err = n.omRunner.Execute([]interface{}{
			"--env", n.ImportInstallation.EnvFile,
			"--skip-ssl-validation",
			"import-installation",
			"--installation", n.ImportInstallation.Installation,
		})
		if err == nil {
			break
		}

		current := time.Now()
		if current.Sub(start) > n.timeout {
			return fmt.Errorf("exceeded %s waiting for opsman to respond", n.timeout)
		}

		_, _ = n.stdout.Write([]byte(fmt.Sprintf("could not reach opsman, polling again in %s, the cause: %s\n", n.pollingInterval, err)))
		time.Sleep(n.pollingInterval)
	}

	return nil
}

func (n *UpgradeOpsman) validate() error {
	if _, err := os.Stat(n.CreateVM.Config); err != nil {
		return fmt.Errorf("could not open config file (%s): %s", n.CreateVM.Config, err)
	}

	if _, err := os.Stat(n.ImportInstallation.EnvFile); err != nil {
		return fmt.Errorf("could not open env file (%s): %s", n.ImportInstallation.EnvFile, err)
	}

	if _, err := os.Stat(n.CreateVM.ImageFile); err != nil {
		return fmt.Errorf("could not open image file (%s): %s", n.CreateVM.ImageFile, err)
	}

	if _, err := os.Stat(n.ImportInstallation.Installation); err != nil {
		return fmt.Errorf("could not open installation file (%s): %s", n.ImportInstallation.Installation, err)
	}

	if _, err := os.Stat(n.CreateVM.StateFile); err != nil {
		return fmt.Errorf("could not open state file (%s): %s", n.CreateVM.StateFile, err)
	}

	for _, varFile := range n.CreateVM.VarsFile {
		if _, err := os.Stat(varFile); err != nil {
			return fmt.Errorf("could not open vars file (%s): %s", varFile, err)
		}
	}

	if zipper, err := zip.OpenReader(n.ImportInstallation.Installation); err != nil {
		return fmt.Errorf("file: \"%s\" is not a valid zip file", n.ImportInstallation.Installation)
	} else {
		defer zipper.Close()
		found := false
		for _, f := range zipper.File {
			if f.Name == "installation.yml" {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("file: \"%s\" is not a valid installation file", n.ImportInstallation.Installation)
		}
	}

	// validate opsman config & vars
	opsmanConfig := vmmanagers.OpsmanConfigFilePayload{}
	opsmanContent, err := interpolate.Execute(interpolate.Options{
		TemplateFile:  n.CreateVM.Config,
		VarsFiles:     n.CreateVM.VarsFile,
		EnvironFunc:   os.Environ,
		VarsEnvs:      n.CreateVM.VarsEnv,
		ExpectAllKeys: true,
	})
	if err != nil {
		return fmt.Errorf("could not load config file (%s): %s", n.CreateVM.Config, err)
	}

	err = yaml.UnmarshalStrict(opsmanContent, &opsmanConfig)
	if err != nil {
		return fmt.Errorf("could not unmarshal config file (%s): %s", n.CreateVM.Config, err)
	}

	err = GuardAgainstMissingOpsmanConfiguration(opsmanContent, n.CreateVM.Config)
	if err != nil {
		return err
	}

	err = vmmanagers.ValidateOpsManConfig(&opsmanConfig)
	if err != nil {
		return fmt.Errorf("could not validate config file (%s): %s", n.CreateVM.Config, err)
	}

	_, err = extractopsmansemver.Do(n.CreateVM.ImageFile) //[ops-manager,2.2.0]OpsManager2.2-build.296onGCP.yml
	if err != nil {
		return fmt.Errorf("the file name '%s' for the Ops Manager image needs to contain the original version number as downloaded from Pivnet (ie 'OpsManager2.2-build.296onGCP.yml')", n.CreateVM.ImageFile)
	}

	return n.validateImportInstallationConfig()
}

func (n *UpgradeOpsman) validateImportInstallationConfig() error {
	target := target{}
	content, err := ioutil.ReadFile(n.ImportInstallation.EnvFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(content, &target)
	if err != nil {
		return err
	}

	if target.Target == "" {
		target.Target = os.Getenv("OM_TARGET")
	}

	if target.Target == "" {
		return fmt.Errorf("target is a required field in the env configuration")
	}

	if target.DecryptionPassphrase == "" {
		target.DecryptionPassphrase = os.Getenv("OM_DECRYPTION_PASSPHRASE")
	}

	if target.DecryptionPassphrase == "" {
		return fmt.Errorf("decryption-passphrase is a required field in the env configuration")
	}

	return nil
}
