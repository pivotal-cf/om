package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda/commands"
)

type DeployedProducts struct {
	presenter         Presenter
	diagnosticService diagnosticService
}

func NewDeployedProducts(presenter Presenter, diagnosticService diagnosticService) DeployedProducts {
	return DeployedProducts{
		presenter:         presenter,
		diagnosticService: diagnosticService,
	}
}

func (dp DeployedProducts) Execute(args []string) error {
	diagnosticReport, err := dp.diagnosticService.Report()
	if err != nil {
		return fmt.Errorf("failed to retrieve deployed products %s", err)
	}

	deployedProducts := diagnosticReport.DeployedProducts

	dp.presenter.PresentDeployedProducts(deployedProducts)

	return nil
}

func (dp DeployedProducts) Usage() commands.Usage {
	return commands.Usage{
		Description:      "This authenticated command lists all deployed products.",
		ShortDescription: "lists deployed products",
	}
}
