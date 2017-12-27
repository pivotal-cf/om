package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/presenters"
)

type StagedProducts struct {
	presenter         presenters.Presenter
	diagnosticService diagnosticService
}

func NewStagedProducts(presenter presenters.Presenter, diagnosticService diagnosticService) StagedProducts {
	return StagedProducts{
		presenter:         presenter,
		diagnosticService: diagnosticService,
	}
}

func (sp StagedProducts) Execute(args []string) error {
	diagnosticReport, err := sp.diagnosticService.Report()
	if err != nil {
		return fmt.Errorf("failed to retrieve staged products %s", err)
	}

	stagedProducts := diagnosticReport.StagedProducts

	sp.presenter.PresentStagedProducts(stagedProducts)
	return nil
}

func (sp StagedProducts) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists all staged products.",
		ShortDescription: "lists staged products",
	}
}
