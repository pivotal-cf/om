package commands

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"time"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type ApplyChanges struct {
	service        applyChangesService
	pendingService pendingChangesService
	logger         logger
	logWriter      logWriter
	waitDuration   time.Duration
	Options        struct {
		Config                string   `short:"c"   long:"config"               description:"path to yml file containing errand configuration (see docs/apply-changes/README.md for format)"`
		IgnoreWarnings        bool     `short:"i"   long:"ignore-warnings"      description:"ignore issues reported by Ops Manager when applying changes"`
		SkipDeployProducts    bool     `short:"sdp" long:"skip-deploy-products" description:"skip deploying products when applying changes - just update the director"`
		SkipUnchangedProducts bool     `short:"sup" long:"skip-unchanged-products"         description:"skip deploying unchanged products - just run changed or new products --skip-unchanged-products (OM 2.2+)"`
		ProductNames          []string `short:"n"   long:"product-name"         description:"name of the product(s) to deploy, cannot be used in conjunction with --skip-deploy-products (OM 2.2+)"`
	}
}

//go:generate counterfeiter -o ./fakes/apply_changes_service.go --fake-name ApplyChangesService . applyChangesService
type applyChangesService interface {
	CreateInstallation(bool, bool, []string, api.ApplyErrandChanges) (api.InstallationsServiceOutput, error)
	GetInstallation(id int) (api.InstallationsServiceOutput, error)
	GetInstallationLogs(id int) (api.InstallationsServiceOutput, error)
	Info() (api.Info, error)
	RunningInstallation() (api.InstallationsServiceOutput, error)
	ListInstallations() ([]api.InstallationsServiceOutput, error)
}

//go:generate counterfeiter -o ./fakes/log_writer.go --fake-name LogWriter . logWriter
type logWriter interface {
	Flush(logs string) error
}

func NewApplyChanges(service applyChangesService, pendingService pendingChangesService, logWriter logWriter, logger logger, waitDuration time.Duration) ApplyChanges {
	return ApplyChanges{
		service:        service,
		pendingService: pendingService,
		logger:         logger,
		logWriter:      logWriter,
		waitDuration:   waitDuration,
	}
}

func (ac ApplyChanges) Execute(args []string) error {
	if _, err := jhanda.Parse(&ac.Options, args); err != nil {
		return fmt.Errorf("could not parse apply-changes flags: %s", err)
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

	changedProducts := []string{}
	deployProducts := !ac.Options.SkipDeployProducts

	if len(ac.Options.ProductNames) > 0 {
		if ac.Options.SkipDeployProducts {
			return fmt.Errorf("product-name flag can not be passed with the skip-deploy-products flag")
		}
		if ac.Options.SkipUnchangedProducts {
			return fmt.Errorf("product-name flag can not be passed with the skip-unchanged-products flag")
		}
		info, err := ac.service.Info()
		if err != nil {
			return fmt.Errorf("could not retrieve info from targetted ops manager: %v", err)
		}
		if !info.VersionAtLeast(2, 2) {
			return fmt.Errorf("--product-name is only available with Ops Manager 2.2 or later: you are running %s", info.Version)
		}
		changedProducts = append(changedProducts, ac.Options.ProductNames...)
	}

	if ac.Options.SkipUnchangedProducts {
		s, err := ac.pendingService.ListStagedPendingChanges()
		if err != nil {
			return fmt.Errorf("could not check for any pending changes installation: %s", err)
		}
		info, err := ac.service.Info()
		if err != nil {
			return fmt.Errorf("could not retrieve info from targetted ops manager: %v", err)
		}
		if !info.VersionAtLeast(2, 2) {
			return fmt.Errorf("skip-unchanged-products is only available with Ops Manager 2.2 or later: you are running %s", info.Version)
		}
		for _, p := range s.ChangeList {
			ac.logger.Printf("Found product: %s with action of: %s", p.GUID, p.Action)
			if p.Action != "unchanged" {
				changedProducts = append(changedProducts, p.GUID)
				ac.logger.Printf("Adding %s to ProductNames", p.GUID)
			}
		}
		if len(changedProducts) <= 0 {
			deployProducts = false
		}
	}

	installation, err := ac.service.RunningInstallation()
	if err != nil {
		return fmt.Errorf("could not check for any already running installation: %s", err)
	}

	if installation == (api.InstallationsServiceOutput{}) {
		ac.logger.Printf("attempting to apply changes to the targeted Ops Manager")
		installation, err = ac.service.CreateInstallation(ac.Options.IgnoreWarnings, deployProducts, changedProducts, errands)
		if err != nil {
			return fmt.Errorf("installation failed to trigger: %s", err)
		}
	} else {
		startedAtFormatted := installation.StartedAt.Format(time.UnixDate)
		ac.logger.Printf("found already running installation...re-attaching (Installation ID: %d, Started: %s)", installation.ID, startedAtFormatted)
	}

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

func (ac ApplyChanges) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command kicks off an install of any staged changes on the Ops Manager.",
		ShortDescription: "triggers an install on the Ops Manager targeted",
		Flags:            ac.Options,
	}
}
