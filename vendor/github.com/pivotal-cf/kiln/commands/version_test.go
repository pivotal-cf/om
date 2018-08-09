package commands_test

import (
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/kiln/commands"
	"github.com/pivotal-cf/kiln/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	var (
		logger  *fakes.Logger
		version commands.Version
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		version = commands.NewVersion(logger, "1.2.3-build.4")
	})

	Describe("Execute", func() {
		It("prints the version number", func() {
			err := version.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.PrintfCallCount()).To(Equal(1))
			format, args := logger.PrintfArgsForCall(0)
			Expect(format).To(Equal("kiln version %s\n"))
			Expect(args).To(Equal([]interface{}{"1.2.3-build.4"}))
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			version := commands.NewVersion(nil, "")
			Expect(version.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command prints the kiln release version number.",
				ShortDescription: "prints the kiln release version",
			}))
		})
	})
})
