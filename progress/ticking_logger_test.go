package progress_test

import (
	"time"

	"github.com/pivotal-cf/om/progress"
	"github.com/pivotal-cf/om/progress/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TickingLogger", func() {
	var (
		liveWriter *fakes.LiveWriter
		tl         *progress.TickingLogger
	)

	BeforeEach(func() {
		liveWriter = &fakes.LiveWriter{}
		tl = progress.NewTickingLogger(liveWriter, 10*time.Millisecond)
	})

	Describe("Start", func() {
		It("starts printing log lines", func() {
			tl.Start()

			Eventually(liveWriter.WriteCallCount).Should(BeNumerically(">", 0))
			Eventually(func() string {
				buffer := liveWriter.WriteArgsForCall(liveWriter.WriteCallCount() - 1)
				return string(buffer)
			}, "5s").Should(ContainSubstring("2s elapsed"))
		})
	})

	Describe("Stop", func() {
		It("stops printing log lines", func() {
			tl.Start()

			Eventually(liveWriter.WriteCallCount).Should(BeNumerically(">", 0))

			tl.Stop()

			count := liveWriter.WriteCallCount()
			Consistently(liveWriter.WriteCallCount()).Should(Equal(count))
		})
	})
})
