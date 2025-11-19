package download_test

import (
	"io/ioutil"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pivotal-cf/go-pivnet/v7/download"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Progress", func() {
	var (
		b download.Bar
	)

	BeforeEach(func() {
		b = download.NewBar()
		b.SetOutput(GinkgoWriter)
	})

	It("handles concurrent writes without racing", func() {
		total := 10

		b.SetTotal(int64(total))
		b.Output = ioutil.Discard
		b.Kickoff()

		var g errgroup.Group
		for i := 0; i < total; i++ {
			g.Go(func() error {
				time.Sleep(10 * time.Millisecond)

				_, err := b.Write([]byte("a"))
				return err
			})
		}

		err := g.Wait()
		Expect(err).NotTo(HaveOccurred())
	})
})
