package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("StagedProducts", func() {
	var (
		tableWriter       *fakes.TableWriter
		diagnosticService *fakes.DiagnosticService
		command           commands.StagedProducts
	)

	BeforeEach(func() {
		tableWriter = &fakes.TableWriter{}
		diagnosticService = &fakes.DiagnosticService{}
		command = commands.NewStagedProducts(tableWriter, diagnosticService)
	})

	It("lists the staged products", func() {
		diagnosticService.ReportReturns(api.DiagnosticReport{
			StagedProducts: []api.DiagnosticProduct{
				{
					Name:    "some-product",
					Version: "some-version",
				},
				{
					Name:    "acme-product",
					Version: "version-infinity",
				},
			},
		}, nil)

		err := command.Execute([]string{})
		Expect(err).NotTo(HaveOccurred())

		Expect(diagnosticService.ReportCallCount()).To(Equal(1))

		Expect(tableWriter.SetHeaderCallCount()).To(Equal(1))
		Expect(tableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"Name", "Version"}))

		Expect(tableWriter.AppendCallCount()).To(Equal(2))
		Expect(tableWriter.AppendArgsForCall(0)).To(Equal([]string{"some-product", "some-version"}))
		Expect(tableWriter.AppendArgsForCall(1)).To(Equal([]string{"acme-product", "version-infinity"}))
	})

	Context("failure cases", func() {
		Context("when fetching the diagnostic report fails", func() {
			It("returns an error", func() {
				diagnosticService.ReportReturns(api.DiagnosticReport{}, errors.New("beep boop"))

				err := command.Execute([]string{})
				Expect(err).To(MatchError("failed to retrieve staged products beep boop"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewStagedProducts(nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This authenticated command lists all staged products.",
				ShortDescription: "lists staged products",
			}))
		})
	})
})
