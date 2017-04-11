package commands

import "fmt"

type DeployedProducts struct {
	tableWriter       tableWriter
	diagnosticService diagnosticService
}

func NewDeployedProducts(tableWriter tableWriter, diagnosticService diagnosticService) DeployedProducts {
	return DeployedProducts{
		tableWriter:       tableWriter,
		diagnosticService: diagnosticService,
	}
}

func (dp DeployedProducts) Execute(args []string) error {
	diagnosticReport, err := dp.diagnosticService.Report()
	if err != nil {
		return fmt.Errorf("failed to retrieve deployed products %s", err)
	}

	deployedProducts := diagnosticReport.DeployedProducts

	dp.tableWriter.SetHeader([]string{"Name", "Version"})

	for _, product := range deployedProducts {
		dp.tableWriter.Append([]string{product.Name, product.Version})
	}

	dp.tableWriter.Render()

	return nil
}

func (dp DeployedProducts) Usage() Usage {
	return Usage{
		Description:      "This authenticated command lists all deployed products.",
		ShortDescription: "lists deployed products",
	}
}
