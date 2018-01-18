package jhanda_test

import (
	"strings"

	"github.com/pivotal-cf/jhanda"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Usage", func() {
	It("returns a formatted version of the flag set usage", func() {
		usage, err := jhanda.PrintUsage(struct {
			Second []string `short:"2" long:"second" required:"true" default:"true"  description:"the second flag"`
			Third  string   `          long:"third"                                  description:"the third flag"`
			First  bool     `short:"1" long:"first"  required:"true"                 description:"the first flag"`
		}{})
		Expect(err).NotTo(HaveOccurred())
		Expect(usage).To(Equal(strings.TrimSpace(`
--first, -1   bool (required)              the first flag
--second, -2  string (required, variadic)  the second flag (default: true)
--third       string                       the third flag
`)))
	})

	Context("when the receiver passed is not a struct", func() {
		It("returns an error", func() {
			_, err := jhanda.PrintUsage(123)
			Expect(err).To(MatchError("unexpected pointer to non-struct type int"))
		})
	})
})
