package baking_test

import (
	"errors"

	"github.com/pivotal-cf/kiln/builder"
	"github.com/pivotal-cf/kiln/internal/baking"
	"github.com/pivotal-cf/kiln/internal/baking/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StemcellService", func() {
	Describe("FromTarball", func() {
		var (
			logger  *fakes.Logger
			reader  *fakes.PartReader
			service baking.StemcellService
		)

		BeforeEach(func() {
			logger = &fakes.Logger{}
			reader = &fakes.PartReader{}
			reader.ReadReturns(builder.Part{
				Metadata: builder.StemcellManifest{
					Version:         "some-version",
					OperatingSystem: "some-os",
				},
			}, nil)

			service = baking.NewStemcellService(logger, reader)
		})

		It("parses the stemcell passed as a tarball", func() {
			stemcell, err := service.FromTarball("some-stemcell-tarball")
			Expect(err).NotTo(HaveOccurred())
			Expect(stemcell).To(Equal(builder.StemcellManifest{
				Version:         "some-version",
				OperatingSystem: "some-os",
			}))

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			Expect(logger.PrintlnArgsForCall(0)).To(Equal([]interface{}{"Reading stemcell manifest..."}))

			Expect(reader.ReadCallCount()).To(Equal(1))
			Expect(reader.ReadArgsForCall(0)).To(Equal("some-stemcell-tarball"))
		})

		Context("when there is no tarball to parse", func() {
			It("returns nothing", func() {
				stemcell, err := service.FromTarball("")
				Expect(err).NotTo(HaveOccurred())
				Expect(stemcell).To(BeNil())

				Expect(logger.PrintlnCallCount()).To(Equal(0))
				Expect(reader.ReadCallCount()).To(Equal(0))
			})
		})

		Context("failure cases", func() {
			Context("when the reader fails", func() {
				It("returns an error", func() {
					reader.ReadReturns(builder.Part{}, errors.New("failed to read"))

					_, err := service.FromTarball("some-stemcell-tarball")
					Expect(err).To(MatchError("failed to read"))
				})
			})
		})
	})
})
