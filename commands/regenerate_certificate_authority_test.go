package commands_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("RegenerateCertificateAuthority", func() {
	var (
		fakeCertificateAuthorityService *fakes.CertificateAuthorityRegenerator
		fakeLogger                      *fakes.Logger
		command                         commands.RegenerateCertificateAuthority
	)

	BeforeEach(func() {
		fakeCertificateAuthorityService = &fakes.CertificateAuthorityRegenerator{}
		fakeLogger = &fakes.Logger{}
		command = commands.NewRegenerateCertificateAuthority(fakeCertificateAuthorityService, fakeLogger)
	})

	Describe("Execute", func() {
		It("makes a request to the Opsman to regenerate an inactive certificate authority", func() {
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCertificateAuthorityService.RegenerateCallCount()).To(Equal(1))

			Expect(fakeLogger.PrintfCallCount()).To(Equal(1))
			format, content := fakeLogger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("Certificate authority regenerated.\n"))
		})
	})

	Describe("Usage", func() {
		It("returns usage info", func() {
			usage := command.Usage()
			Expect(usage).To(Equal(jhanda.Usage{
				Description:      "This authenticated command regenerates certificate authority on the Ops Manager",
				ShortDescription: "regenerates a certificate authority on the Opsman",
			}))
		})
	})
})
