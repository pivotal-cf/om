package baking_test

import (
	"errors"

	"github.com/pivotal-cf/kiln/builder"
	"github.com/pivotal-cf/kiln/internal/baking"
	"github.com/pivotal-cf/kiln/internal/baking/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PropertiesService", func() {
	Describe("FromDirectories", func() {
		var (
			logger  *fakes.Logger
			reader  *fakes.DirectoryReader
			service baking.PropertiesService
		)

		BeforeEach(func() {
			logger = &fakes.Logger{}
			reader = &fakes.DirectoryReader{}
			reader.ReadReturns([]builder.Part{
				{
					Name: "some-property",
					Metadata: builder.Metadata{
						"key": "value",
					},
				},
			}, nil)

			service = baking.NewPropertiesService(logger, reader)
		})

		It("parses the properties passed in a set of directories", func() {
			properties, err := service.FromDirectories([]string{"some-properties", "other-properties"})
			Expect(err).NotTo(HaveOccurred())
			Expect(properties).To(Equal(map[string]interface{}{
				"some-property": builder.Metadata{
					"key": "value",
				},
			}))

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			Expect(logger.PrintlnArgsForCall(0)).To(Equal([]interface{}{"Reading property blueprint files..."}))

			Expect(reader.ReadCallCount()).To(Equal(2))
			Expect(reader.ReadArgsForCall(0)).To(Equal("some-properties"))
			Expect(reader.ReadArgsForCall(1)).To(Equal("other-properties"))
		})

		Context("when the directories argument is empty", func() {
			It("returns nothing", func() {
				properties, err := service.FromDirectories(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(properties).To(BeNil())

				properties, err = service.FromDirectories([]string{})
				Expect(err).NotTo(HaveOccurred())
				Expect(properties).To(BeNil())
			})
		})

		Context("failure cases", func() {
			Context("when the reader fails", func() {
				It("returns an error", func() {
					reader.ReadReturns(nil, errors.New("failed to read"))

					_, err := service.FromDirectories([]string{"some-properties"})
					Expect(err).To(MatchError("failed to read"))
				})
			})
		})
	})
})
