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
			First  bool              `short:"1" long:"first"                 description:"the first flag"`
			Second flags.StringSlice `short:"2" long:"second" default:"true" description:"the second flag"`
			Third  string            `          long:"third"                 description:"the third flag"`
		}{})
		Expect(err).NotTo(HaveOccurred())
		Expect(usage).To(Equal(strings.TrimSpace(`
--first, -1   bool               the first flag
--second, -2  string (variadic)  the second flag (default: true)
--third       string             the third flag
`)))
	})

	Context("when the receiver passed is not a struct", func() {
		It("returns an error", func() {
			_, err := flags.Usage(123)
			Expect(err).To(MatchError("unexpected pointer to non-struct type int"))
		})
	})
})
