package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/om/api"
)

type ApplyChanges struct {
	service        applyChangesService
	pendingService pendingChangesService
	logger         logger
	logWriter      logWriter
	waitDuration   time.Duration
	Options        struct {
		Config             string   `short:"c"   long:"config"               description:"path to yml file containing errand configuration (see docs/apply-changes/README.md for format)"`
		IgnoreWarnings     bool     `short:"i"   long:"ignore-warnings"      description:"For convenience. Use other commands to disable particular verifiers if they are inappropriate."`
		Reattach           bool     `long:"reattach" description:"reattach to an already running apply changes (if available)"`
		RecreateVMs        bool     `long:"recreate-vms" description:"recreate all vms"`
		SkipDeployProducts bool     `short:"s" long:"skip-deploy-products" description:"skip deploying products when applying changes - just update the director"`
		ProductNames       []string `short:"n"   long:"product-name"         description:"name of the product(s) to deploy, cannot be used in conjunction with --skip-deploy-products (OM 2.2+)"`
	}
}

//counterfeiter:generate -o ./fakes/apply_changes_service.go --fake-name ApplyChangesService . applyChangesService
type applyChangesService interface {
	CreateInstallation(bool, bool, []string, api.ApplyErrandChanges) (api.InstallationsServiceOutput, error)
	GetInstallation(id int) (api.InstallationsServiceOutput, error)
	GetInstallationLogs(id int) (api.InstallationsServiceOutput, error)
	Info() (api.Info, error)
	RunningInstallation() (api.InstallationsServiceOutput, error)
	ListInstallations() ([]api.InstallationsServiceOutput, error)
	UpdateStagedDirectorProperties(api.DirectorProperties) error
}

//counterfeiter:generate -o ./fakes/log_writer.go --fake-name LogWriter . logWriter
type logWriter interface {
	Flush(logs string) error
}

func NewApplyChanges(service applyChangesService, pendingService pendingChangesService, logWriter logWriter, logger logger, waitDuration time.Duration) *ApplyChanges {
	return &ApplyChanges{
		service:        service,
		pendingService: pendingService,
		logger:         logger,
		logWriter:      logWriter,
		waitDuration:   waitDuration,
	}
}

func (ac ApplyChanges) Execute(args []string) error {
	if ac.Options.RecreateVMs && ac.Options.Reattach {
		return errors.New("--recreate-vms cannot be used with --reattach because it requires the ability to update a director property")
	}

	errands := api.ApplyErrandChanges{}

	if ac.Options.Config != "" {
		fh, err := os.Open(ac.Options.Config)
		if err != nil {
			return fmt.Errorf("could not load config: %s", err)
		}
		defer fh.Close()
		err = yaml.NewDecoder(fh).Decode(&errands)
		if err != nil {
			return fmt.Errorf("could not parse %s: %s", ac.Options.Config, err)
		}
	}

	var changedProducts []string
	if len(ac.Options.ProductNames) > 0 {
		if ac.Options.SkipDeployProducts {
			return errors.New("product-name flag can not be passed with the skip-deploy-products flag")
		}
		info, err := ac.service.Info()
		if err != nil {
			return fmt.Errorf("could not retrieve info from targetted ops manager: %v", err)
		}
		if ok, err := info.VersionAtLeast(2, 2); !ok {
			return fmt.Errorf("--product-name is only available with Ops Manager 2.2 or later: you are running %s. Error: %w", info.Version, err)
		}
		changedProducts = ac.Options.ProductNames
	}

	installation, err := ac.service.RunningInstallation()
	if err != nil {
		return fmt.Errorf("could not check for any already running installation: %s", err)
	}

	if installation != (api.InstallationsServiceOutput{}) {
		startedAtFormatted := installation.StartedAt.Format(time.UnixDate)

		if ac.Options.Reattach {
			ac.logger.Printf("found already running installation... re-attaching (Installation ID: %d, Started: %s)", installation.ID, startedAtFormatted)
			err = ac.waitForApplyChangesCompletion(installation)
			ac.logger.Printf("found already running installation... re-attaching (Installation ID: %d, Started: %s)", installation.ID, startedAtFormatted)

			return err
		} else {
			ac.logger.Printf("found already running installation... not re-attaching (Installation ID: %d, Started: %s)", installation.ID, startedAtFormatted)
			return errors.New("apply changes is already running, use \"--reattach\" to enable reattaching")
		}
	}

	if ac.Options.RecreateVMs {
		var config struct {
			DirectorConfiguration struct {
				DirectorRecreate bool `json:"bosh_director_recreate_on_next_deploy,omitempty"`
				ProductRecreate  bool `json:"bosh_recreate_on_next_deploy,omitempty"`
			} `json:"director_configuration"`
		}

		if len(ac.Options.ProductNames) > 0 {
			ac.logger.Println("setting director to recreate all VMs for the following products:")
			sort.Strings(ac.Options.ProductNames)

			for _, product := range ac.Options.ProductNames {
				ac.logger.Printf("- %s", product)
			}
			ac.logger.Println("this will also recreate the director vm if there are changes")
			config.DirectorConfiguration.ProductRecreate = true
		} else if ac.Options.SkipDeployProducts {
			ac.logger.Println("setting director to recreate director vm (available in Ops Manager 2.9+)")
			config.DirectorConfiguration.DirectorRecreate = true
		} else {
			ac.logger.Println("setting director to recreate all vms (available in Ops Manager 2.9+)")
			config.DirectorConfiguration.ProductRecreate = true
			config.DirectorConfiguration.DirectorRecreate = true
		}

		info, err := ac.service.Info()
		if err != nil {
			return err
		}

		versionAtLeast29, err := info.VersionAtLeast(2, 9)
		if err != nil {
			return err
		}

		if !versionAtLeast29 {
			config.DirectorConfiguration.ProductRecreate = true
			config.DirectorConfiguration.DirectorRecreate = false
		}

		payload, _ := json.Marshal(config)
		err = ac.service.UpdateStagedDirectorProperties(api.DirectorProperties(string(payload)))
		if err != nil {
			return fmt.Errorf("could not set director to recreate VMS: %s", err)
		}
	}

	ac.logger.Printf("attempting to apply changes to the targeted Ops Manager")
	installation, err = ac.service.CreateInstallation(ac.Options.IgnoreWarnings, !ac.Options.SkipDeployProducts, changedProducts, errands)
	if err != nil {
		return fmt.Errorf("installation failed to trigger: %s", err)
	}

	return ac.waitForApplyChangesCompletion(installation)
}

func (ac ApplyChanges) waitForApplyChangesCompletion(installation api.InstallationsServiceOutput) error {
	for {
		current, err := ac.service.GetInstallation(installation.ID)
		if err != nil {
			return fmt.Errorf("installation failed to get status: %s", err)
		}

		install, err := ac.service.GetInstallationLogs(installation.ID)
		if err != nil {
			return fmt.Errorf("installation failed to get logs: %s", err)
		}

		err = ac.logWriter.Flush(install.Logs)
		if err != nil {
			return fmt.Errorf("installation failed to flush logs: %s", err)
		}

		if current.Status == api.StatusSucceeded {
			return nil
		} else if current.Status == api.StatusFailed {
			return errors.New("installation was unsuccessful")
		}

		time.Sleep(ac.waitDuration)
	}
}
