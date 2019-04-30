package commands

import (
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"strings"
)

type AssignMultiStemcell struct {
	logger  logger
	service assignMultiStemcellService
	Options struct {
		ConfigFile  string   `long:"config"   short:"c"  description:"path to yml file for configuration (keys must match the following command line flags)"`
		ProductName string   `long:"product"  short:"p"  description:"name of Ops Manager tile to associate a stemcell to" required:"true"`
		Stemcells   []string `long:"stemcell" short:"s"  description:"associate a particular stemcell version to a tile (ie 'ubuntu-trusty:123.4')" required:"true"`
	}
}

//go:generate counterfeiter -o ./fakes/assign_multi_stemcell_service.go --fake-name AssignMultiStemcellService . assignMultiStemcellService
type assignMultiStemcellService interface {
	ListMultiStemcells() (api.ProductMultiStemcells, error)
	AssignMultiStemcell(input api.ProductMultiStemcells) error
	Info() (api.Info, error)
}

func NewAssignMultiStemcell(service assignMultiStemcellService, logger logger) AssignMultiStemcell {
	return AssignMultiStemcell{
		service: service,
		logger:  logger,
	}
}

func (as AssignMultiStemcell) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description: "This command will assign multiple already uploaded stemcells to a specific product in Ops Manager 2.6+.\n" +
			"It is recommended to use \"upload-stemcell --floating=false\" before using this command.",
		ShortDescription: "assigns multiple uploaded stemcells to a product in the targeted Ops Manager 2.6+",
		Flags:            as.Options,
	}
}

func (as AssignMultiStemcell) Execute(args []string) error {
	err := loadConfigFile(args, &as.Options, nil)
	if err != nil {
		return fmt.Errorf("could not parse assign-stemcell flags: %s", err)
	}

	err = as.validateOpsManVersion()
	if err != nil {
		return err
	}

	err = as.validateArgs()
	if err != nil {
		return fmt.Errorf("could not parse assign-stemcell arguments: %s", err)
	}

	as.logger.Printf("finding available stemcells for product: \"%s\"...", as.Options.ProductName)
	productStemcell, err := as.getProductStemcell()
	if err != nil {
		return err
	}

	if productStemcell.StagedForDeletion {
		return fmt.Errorf("could not assign stemcell: product \"%s\" is staged for deletion", as.Options.ProductName)
	}

	as.logger.Println("validating that stemcell exists in Ops Manager...")
	stemcells, err := as.validateStemcellVersion(productStemcell)
	if err != nil {
		return err
	}

	as.logger.Printf(
		"assigning stemcells: \"%s\" to product \"%s\"...\n",
		strings.Join(getAllStemcells(stemcells), ", "),
		as.Options.ProductName,
	)
	err = as.service.AssignMultiStemcell(api.ProductMultiStemcells{
		Products: []api.ProductMultiStemcell{
			{
				GUID:            productStemcell.GUID,
				StagedStemcells: stemcells,
			},
		},
	})
	if err != nil {
		return err
	}

	as.logger.Println("assigned stemcells successfully")
	return nil
}

func (as AssignMultiStemcell) validateArgs() error {
	for _, option := range as.Options.Stemcells {
		parts := strings.Split(option, ":")
		if len(parts) < 2 {
			return fmt.Errorf(`expected "--stemcell" format value as "operating-system=version"`)
		}
	}
	return nil
}

func (as AssignMultiStemcell) getProductStemcell() (api.ProductMultiStemcell, error) {
	var result api.ProductMultiStemcell

	productStemcells, err := as.service.ListMultiStemcells()
	if err != nil {
		return result, err
	}

	for _, productStemcell := range productStemcells.Products {
		if productStemcell.ProductName == as.Options.ProductName {
			return productStemcell, nil
		}
	}

	return result, fmt.Errorf("could not list product stemcell: product \"%s\" not found", as.Options.ProductName)
}

func (as *AssignMultiStemcell) validateStemcellVersion(productStemcell api.ProductMultiStemcell) ([]api.StemcellObject, error) {
	availableVersions := productStemcell.AvailableVersions

	if len(availableVersions) == 0 {
		return nil, fmt.Errorf("no stemcells are available for \"%s\". "+
			"minimum required stemcells are: %s. "+
			"upload-stemcell, and try again",
			as.Options.ProductName,
			strings.Join(getAllStemcells(productStemcell.RequiredStemcells), ", "),
		)
	}

	stemcellGroup := []api.StemcellObject{}
	for index, option := range as.Options.Stemcells {
		parts := strings.Split(option, ":")
		os, version := parts[0], parts[1]

		if version == "latest" {
			var runner int
			for i, available := range availableVersions {
				if os == available.OS {
					runner = i
				}
			}

			stemcellGroup = append(stemcellGroup, availableVersions[runner])

		} else {
			for _, available := range availableVersions {
				if os == available.OS && version == available.Version {
					stemcellGroup = append(stemcellGroup, available)
				}
			}
		}

		if len(stemcellGroup) < index+1 {
			listOfStemcells := strings.Join(getStemcellsForOS(availableVersions, os), ", ")
			if listOfStemcells == "" {
				return nil, fmt.Errorf(`stemcell version %s for %s not found in Ops Manager.
there are no available stemcells to for "%s"
upload-stemcell, and try again`, version, os, as.Options.ProductName)
			}
			return nil, fmt.Errorf(`stemcell version %s for %s not found in Ops Manager.
Available Stemcells for "%s": %s`, version, os, as.Options.ProductName, listOfStemcells)
		}
	}

	return stemcellGroup, nil

}

func (as AssignMultiStemcell) validateOpsManVersion() error {
	info, err := as.service.Info()
	if err != nil {
		return fmt.Errorf("cannot retrieve version of Ops Manager")
	}

	validVersion, err := info.VersionAtLeast(2, 6)
	if err != nil {
		return fmt.Errorf("could not determine version was 2.6+ compatible: %s", err)
	}

	if validVersion {
		return nil
	}

	return fmt.Errorf("this command can only be used with OpsManager 2.6+")
}

func getStemcellsForOS(availableStemcells []api.StemcellObject, os string) []string {
	var results []string

	for _, stemcell := range availableStemcells {
		if stemcell.OS == os {
			results = append(results, fmt.Sprintf("%s %s", stemcell.OS, stemcell.Version))
		}
	}

	return results
}

func getAllStemcells(stemcells []api.StemcellObject) []string {
	var results []string

	for _, stemcell := range stemcells {
		results = append(results, fmt.Sprintf("%s %s", stemcell.OS, stemcell.Version))
	}

	return results
}
