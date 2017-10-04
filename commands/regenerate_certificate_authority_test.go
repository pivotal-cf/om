package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		command = commands.NewRegenerateCertificateAuthority(fakeCertificateAuthorityService, fakeLogger)
	})

	Describe("Execute", func() {
		It("makes a request to the Opsman to regenerate an inactive certificate authority", func() {
			err := command.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCertificateAuthorityService.GenerateCallCount()).To(Equal(1))
		})
	})
})
