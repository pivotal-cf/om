package flags_test

import (
	"github.com/pivotal-cf/om/flags"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Global", func() {
	Describe("Parse", func() {
		It("parses the global flags", func() {
			global := flags.NewGlobal()
			args, err := global.Parse([]string{"-h", "-v", "command", "--command-flag"})
			Expect(err).NotTo(HaveOccurred())

			Expect(global.Help.Value).To(BeTrue())
			Expect(global.Version.Value).To(BeTrue())

			Expect(args).To(Equal([]string{"command", "--command-flag"}))
		})

		Context("when an unknown flag is given", func() {
			It("returns an error", func() {
				global := flags.NewGlobal()
				_, err := global.Parse([]string{"-?"})
				Expect(err).To(MatchError(ContainSubstring("flag provided but not defined: -?")))
			})
		})
	})
})
