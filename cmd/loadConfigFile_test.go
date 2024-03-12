package cmd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var fakeOptions map[string]interface{}

var _ = Describe("parseOptions", func() {
	BeforeEach(func() {
		fakeOptions = map[string]interface{}{"test-name": `"pas`}
	})

	It("Should parse values with double quotes not escaping it", func() {
		expected := []string{`--test-name="pas`}
		val, err := parseOptions(fakeOptions)
		Expect(err).To(BeNil())
		Expect(val).To(Equal(expected))

	})
})
