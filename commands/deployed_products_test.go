package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("DeployedProducts", func() {
	var (
		tableWriter       *fakes.TableWriter
		diagnosticService *fakes.DiagnosticService
		command           commands.DeployedProducts
	)

	BeforeEach(func() {
		tableWriter = &fakes.TableWriter{}
		diagnosticService = &fakes.DiagnosticService{}
		command = commands.NewDeployedProducts(tableWriter, diagnosticService)
	})

	It("lists the deployed products", func() {
		diagnosticService.ReportReturns(api.DiagnosticReport{
			DeployedProducts: []api.DiagnosticProduct{
				{
					Name:    "nonsense-product",
					Version: "nonsense-number",
				},
				{
					Name:    "acme-product",
					Version: "googleplex",
				},
			},
		}, nil)

		err := command.Execute([]string{})
		Expect(err).NotTo(HaveOccurred())

		Expect(diagnosticService.ReportCallCount()).To(Equal(1))

		Expect(tableWriter.SetHeaderCallCount()).To(Equal(1))
		Expect(tableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"Name", "Version"}))

		Expect(tableWriter.AppendCallCount()).To(Equal(2))
		Expect(tableWriter.AppendArgsForCall(0)).To(Equal([]string{"nonsense-product", "nonsense-number"}))
		Expect(tableWriter.AppendArgsForCall(1)).To(Equal([]string{"acme-product", "googleplex"}))
	})

	Context("failure cases", func() {
		Context("when fetching the diagnostic report fails", func() {
			It("returns an error", func() {
				diagnosticService.ReportReturns(api.DiagnosticReport{}, errors.New("beep boop"))

				err := command.Execute([]string{})
				Expect(err).To(MatchError("failed to retrieve deployed products beep boop"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewDeployedProducts(nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This authenticated command lists all deployed products.",
				ShortDescription: "lists deployed products",
			}))
		})
	})
})
