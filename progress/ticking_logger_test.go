package progress_test

import (
	"fmt"
	"time"

	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/progress"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TickingLogger", func() {
	var (
		logger *fakes.Logger
		tl     *progress.TickingLogger
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		tl = progress.NewTickingLogger(logger, 10*time.Millisecond)
	})

	Describe("Start", func() {
		It("starts printing log lines", func() {
			tl.Start()

			Eventually(logger.PrintfCallCount).Should(BeNumerically(">", 0))
			Eventually(func() string {
				format, parts := logger.PrintfArgsForCall(logger.PrintfCallCount() - 1)
				return fmt.Sprintf(format, parts...)
			}, "5s").Should(ContainSubstring("2s elapsed"))
		})
	})

	Describe("Stop", func() {
		It("stops printing log lines", func() {
			tl.Start()

			Eventually(logger.PrintfCallCount).Should(BeNumerically(">", 0))

			tl.Stop()

			count := logger.PrintfCallCount()
			Consistently(logger.PrintfCallCount()).Should(Equal(count))
		})
	})
})
