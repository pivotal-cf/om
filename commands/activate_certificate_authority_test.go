package commands_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("ActivateCertificateAuthority", func() {
	var (
		fakeService *fakes.ActivateCertificateAuthorityService
		fakeLogger  *fakes.Logger
		command     *commands.ActivateCertificateAuthority
	)

	BeforeEach(func() {
		fakeService = &fakes.ActivateCertificateAuthorityService{}
		fakeLogger = &fakes.Logger{}
		command = commands.NewActivateCertificateAuthority(fakeService, fakeLogger)
	})

	Describe("Execute", func() {
		It("activates the specified certificate authority", func() {
			err := command.Execute([]string{
				"--id", "some-certificate-authority-id",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.ActivateCertificateAuthorityCallCount()).To(Equal(1))
			Expect(fakeService.ActivateCertificateAuthorityArgsForCall(0)).To(Equal(api.ActivateCertificateAuthorityInput{
				GUID: "some-certificate-authority-id",
			}))

			Expect(fakeLogger.PrintfCallCount()).To(Equal(1))
			format, content := fakeLogger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("Certificate authority 'some-certificate-authority-id' activated\n"))
		})

		Context("failure cases", func() {
			When("the service fails to activate a certificate", func() {
				It("returns an error", func() {
					fakeService.ActivateCertificateAuthorityReturns(errors.New("failed to activate certificate"))

					err := command.Execute([]string{
						"--id", "some-certificate-authority-id",
					})
					Expect(err).To(MatchError("failed to activate certificate"))
				})
			})

			When("an unknown flag is provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--badflag"})
					Expect(err).To(MatchError("could not parse activate-certificate-authority flags: flag provided but not defined: -badflag"))
				})
			})

			When("the id flag is not provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{})
					Expect(err).To(MatchError("could not parse activate-certificate-authority flags: missing required flag \"--id\""))
				})
			})
		})
	})
})
