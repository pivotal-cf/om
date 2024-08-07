package commands_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("RegenerateCertificates", func() {
	var (
		fakeService *fakes.RegenerateCertificatesService
		fakeLogger  *fakes.Logger
		command     *commands.RegenerateCertificates
	)

	BeforeEach(func() {
		fakeService = &fakes.RegenerateCertificatesService{}
		fakeLogger = &fakes.Logger{}
		command = commands.NewRegenerateCertificates(fakeService, fakeLogger)
	})

	Describe("Execute", func() {
		It("makes a request to the Opsman to regenerate the certificates generated by an inactive certificate authority", func() {
			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeService.RegenerateCertificatesCallCount()).To(Equal(1))

			Expect(fakeLogger.PrintfCallCount()).To(Equal(1))
			format, content := fakeLogger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("Certificates regenerated.\n"))
		})
	})
})
