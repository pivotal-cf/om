package flags_test

import (
	"strings"

	"github.com/pivotal-cf/jhanda/flags"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Usage", func() {
	It("returns a formatted version of the flag set usage", func() {
		usage, err := flags.Usage(struct {
			First  bool   `short:"1" long:"first"                 description:"the first flag"`
			Second bool   `short:"2" long:"second" default:"true" description:"the second flag"`
			Third  string `          long:"third"                 description:"the third flag"`
		}{})
		Expect(err).NotTo(HaveOccurred())
		Expect(usage).To(Equal(strings.TrimSpace(`
-1, --first   bool    the first flag
-2, --second  bool    the second flag (default: true)
--third       string  the third flag
`)))
	})

	Context("when the receiver passed is not a struct", func() {
		It("returns an error", func() {
			_, err := flags.Usage(123)
			Expect(err).To(MatchError("unexpected pointer to non-struct type int"))
		})
	})
})
