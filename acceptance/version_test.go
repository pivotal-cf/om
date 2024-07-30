package acceptance

import (
	"fmt"
	"time"

	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-cf/om/cmd"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("version command", func() {
	var version string
	var buffer *gbytes.Buffer

	BeforeEach(func() {
		version = fmt.Sprintf("v0.0.0-dev.%d", time.Now().Unix())
		buffer = gbytes.NewBuffer()
	})

	When("given the -v short flag", func() {
		It("returns the binary version", func() {
			err := cmd.Main(buffer, buffer, version, "1ms", []string{"om", "-v"})
			Expect(err).ToNot(HaveOccurred())

			Expect(buffer).To(gbytes.Say(fmt.Sprintf("%s\n", version)))
		})
	})

	When("given the --version long flag", func() {
		It("returns the binary version", func() {
			err := cmd.Main(buffer, buffer, version, "1ms", []string{"om", "--version"})
			Expect(err).ToNot(HaveOccurred())

			Expect(buffer).To(gbytes.Say(fmt.Sprintf("%s\n", version)))
		})
	})

	When("given the version command", func() {
		It("returns the binary version", func() {
			err := cmd.Main(buffer, buffer, version, "1ms", []string{"om", "version"})
			Expect(err).ToNot(HaveOccurred())

			Expect(buffer).To(gbytes.Say(fmt.Sprintf("%s\n", version)))
		})
	})
})
