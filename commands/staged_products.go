package commands

import "fmt"

type StagedProducts struct {
	tableWriter       tableWriter
	diagnosticService diagnosticService
}

func NewStagedProducts(tableWriter tableWriter, diagnosticService diagnosticService) StagedProducts {
	return StagedProducts{
		tableWriter:       tableWriter,
		diagnosticService: diagnosticService,
	}
}

func (sp StagedProducts) Execute(args []string) error {
	diagnosticReport, err := sp.diagnosticService.Report()
	if err != nil {
		return fmt.Errorf("failed to retrieve staged products %s", err)
	}

	stagedProducts := diagnosticReport.StagedProducts

	sp.tableWriter.SetHeader([]string{"Name", "Version"})

	for _, product := range stagedProducts {
		sp.tableWriter.Append([]string{product.Name, product.Version})
	}

	sp.tableWriter.Render()
	return nil
}

func (sp StagedProducts) Usage() Usage {
	return Usage{
		Description:      "This authenticated command lists all staged products.",
		ShortDescription: "lists staged products",
	}
}
