package commands_test

import (
	"bytes"
	"errors"

	"github.com/pivotal-cf/om/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type badWriter struct{}

func (bw badWriter) Write([]byte) (int, error) {
	return 0, errors.New("failed to write")
}

var _ = Describe("Version", func() {
	Describe("Execute", func() {
		It("prints the version to the output", func() {
			output := bytes.NewBuffer([]byte{})
			version := commands.NewVersion("v1.2.3", output)

			err := version.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(output).To(ContainSubstring("v1.2.3\n"))
		})

		When("the output cannot be written to", func() {
			It("returns an error", func() {

				version := commands.NewVersion("v1.2.3", badWriter{})

				err := version.Execute([]string{})
				Expect(err).To(MatchError("could not print version: failed to write"))
			})
		})
	})
})
