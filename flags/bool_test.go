package flags_test

import (
	"github.com/pivotal-cf/om/flags"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BoolVar", func() {
	Describe("Help", func() {
		It("returns a string representing the description of the flag", func() {
			b := flags.NewBool("?", "query", false, " poses a question")
			Expect(b.Help()).To(Equal("-?, --query  poses a question"))
		})
	})
})
