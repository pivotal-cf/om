package commands_test

import (
	"bytes"
	"errors"

	"github.com/pivotal-cf/om/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LogWriter", func() {
	Describe("Flush", func() {
		var (
			buffer *bytes.Buffer
			writer *commands.LogWriter
		)

		BeforeEach(func() {
			buffer = bytes.NewBuffer([]byte{})
			writer = commands.NewLogWriter(buffer)
		})

		It("writes the given log lines", func() {
			err := writer.Flush("logs-1\nlogs-2\nlogs-3\n")
			Expect(err).NotTo(HaveOccurred())

			Expect(buffer.String()).To(Equal("logs-1\nlogs-2\nlogs-3\n"))

			err = writer.Flush("logs-1\nlogs-2\nlogs-3\nlogs-4\nlogs-5\n")
			Expect(err).NotTo(HaveOccurred())

			Expect(buffer.String()).To(Equal("logs-1\nlogs-2\nlogs-3\nlogs-4\nlogs-5\n"))
		})

		Context("when an error occurs", func() {
			Context("when the writer fails to copy", func() {
				It("returns an error", func() {
					writer = commands.NewLogWriter(errorWriter{})
					err := writer.Flush("logs-1\nlogs-2\nlogs-3\n")
					Expect(err).To(MatchError("failed to write"))
				})
			})
		})
	})
})

type errorWriter struct{}

func (e errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("failed to write")
}
