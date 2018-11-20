package commands

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type AssignStemcell struct {
	logger  logger
	service assignStemcellService
	Options struct {
		ProductName     string `long:"product" short:"p"           description:"name of Ops Manager tile to associate a stemcell to" required:"true"`
		StemcellVersion string `long:"stemcell-version" short:"s"  description:"associate a particular stemcell version to a tile." default:"latest"`
	}
}

//go:generate counterfeiter -o ./fakes/assign_stemcell_service.go --fake-name AssignStemcellService . assignStemcellService
type assignStemcellService interface {
	ListStemcells() (api.ProductStemcells, error)
	AssignStemcell(input api.ProductStemcells) error
}

func NewAssignStemcell(service assignStemcellService, logger logger) AssignStemcell {
	return AssignStemcell{
		service: service,
		logger:  logger,
	}
}

func (as AssignStemcell) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description: "This command will assign an already uploaded stemcell to a specific product in Ops Manager.\n" +
		"It is recommended to use \"upload-stemcell --floating=false\" before using this command.",
		ShortDescription: "assigns an uploaded stemcell to a product in the targeted Ops Manager",
		Flags:            as.Options,
	}
}

func (as AssignStemcell) Execute(args []string) error {
	if _, err := jhanda.Parse(&as.Options, args); err != nil {
		return fmt.Errorf("could not parse assign-stemcell flags: %s", err)
	}

	productStemcell, err := as.getProductStemcell()
	if err != nil {
		return err
	}

	if productStemcell.StagedForDeletion {
		return fmt.Errorf("could not assign stemcell: product \"%s\" is staged for deletion", as.Options.ProductName)
	}

	stemcellVersion, err := as.validateStemcellVersion(productStemcell)
	if err != nil {
		return err
	}

	return as.service.AssignStemcell(api.ProductStemcells{
		Products: []api.ProductStemcell{
			{
				GUID:                  productStemcell.GUID,
				StagedStemcellVersion: stemcellVersion,
			},
		},
	})
}

func (as AssignStemcell) getProductStemcell() (api.ProductStemcell, error) {
	var result api.ProductStemcell

	productStemcells, err := as.service.ListStemcells()
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

func (as *AssignStemcell) validateStemcellVersion(productStemcell api.ProductStemcell) (string, error) {
	availableVersions := productStemcell.AvailableVersions

	if len(availableVersions) == 0 {
		return "", fmt.Errorf("no stemcells are available for \"%s\". "+
			"minimum required stemcell version is: %s. "+
			"upload-stemcell, and try again",
			as.Options.ProductName,
			productStemcell.RequiredStemcellVersion)
	}

	if as.Options.StemcellVersion == "latest" {
		return availableVersions[len(availableVersions)-1], nil
	}

	for _, version := range availableVersions {
		if as.Options.StemcellVersion == version {
			return as.Options.StemcellVersion, nil
		}
	}

	return "", fmt.Errorf(`stemcell version %s not found in Ops Manager. 
	Available Stemcells for "%s": %s`, as.Options.StemcellVersion, as.Options.ProductName, strings.Join(availableVersions, ", "))
}
