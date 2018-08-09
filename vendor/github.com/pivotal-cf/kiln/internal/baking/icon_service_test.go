package baking_test

import (
	"io/ioutil"
	"os"

	"github.com/pivotal-cf/kiln/internal/baking"
	"github.com/pivotal-cf/kiln/internal/baking/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IconService", func() {
	Describe("Encode", func() {
		var (
			path    string
			logger  *fakes.Logger
			service baking.IconService
		)

		BeforeEach(func() {
			file, err := ioutil.TempFile("", "icon")
			Expect(err).NotTo(HaveOccurred())

			path = file.Name()

			_, err = file.WriteString("this is some data")
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Close()).To(Succeed())

			logger = &fakes.Logger{}
			service = baking.NewIconService(logger)
		})

		AfterEach(func() {
			Expect(os.Remove(path)).To(Succeed())
		})

		It("encodes a icon to base64 given a path", func() {
			encoding, err := service.Encode(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(encoding).To(Equal("dGhpcyBpcyBzb21lIGRhdGE="))

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			Expect(logger.PrintlnArgsForCall(0)).To(Equal([]interface{}{"Encoding icon..."}))
		})

		Context("when the icon path is empty", func() {
			It("returns nothing", func() {
				encoding, err := service.Encode("")
				Expect(err).NotTo(HaveOccurred())
				Expect(encoding).To(Equal(""))
			})
		})

		Context("failure cases", func() {
			Context("when the icon does not exist", func() {
				It("returns an error", func() {
					_, err := service.Encode("missing-icon.png")
					Expect(err).To(MatchError(ContainSubstring("open missing-icon.png: no such file or directory")))
				})
			})
		})
	})
})
