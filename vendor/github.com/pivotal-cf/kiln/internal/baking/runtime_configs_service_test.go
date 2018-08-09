package baking_test

import (
	"errors"

	"github.com/pivotal-cf/kiln/builder"
	"github.com/pivotal-cf/kiln/internal/baking"
	"github.com/pivotal-cf/kiln/internal/baking/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuntimeConfigsService", func() {
	Describe("FromDirectories", func() {
		var (
			logger  *fakes.Logger
			reader  *fakes.DirectoryReader
			service baking.RuntimeConfigsService
		)

		BeforeEach(func() {
			logger = &fakes.Logger{}
			reader = &fakes.DirectoryReader{}
			reader.ReadReturns([]builder.Part{
				{
					Name: "some-runtime-config",
					Metadata: builder.Metadata{
						"key": "value",
					},
				},
			}, nil)

			service = baking.NewRuntimeConfigsService(logger, reader)
		})

		It("parses the runtime configs passed in a set of directories", func() {
			runtimeConfigs, err := service.FromDirectories([]string{"some-runtime-configs", "other-runtime-configs"})
			Expect(err).NotTo(HaveOccurred())
			Expect(runtimeConfigs).To(Equal(map[string]interface{}{
				"some-runtime-config": builder.Metadata{
					"key": "value",
				},
			}))

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			Expect(logger.PrintlnArgsForCall(0)).To(Equal([]interface{}{"Reading runtime config files..."}))

			Expect(reader.ReadCallCount()).To(Equal(2))
			Expect(reader.ReadArgsForCall(0)).To(Equal("some-runtime-configs"))
			Expect(reader.ReadArgsForCall(1)).To(Equal("other-runtime-configs"))
		})

		Context("when the directories argument is empty", func() {
			It("returns nothing", func() {
				runtimeConfigs, err := service.FromDirectories(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(runtimeConfigs).To(BeNil())

				runtimeConfigs, err = service.FromDirectories([]string{})
				Expect(err).NotTo(HaveOccurred())
				Expect(runtimeConfigs).To(BeNil())

				Expect(logger.PrintlnCallCount()).To(Equal(0))
				Expect(reader.ReadCallCount()).To(Equal(0))
			})
		})

		Context("failure cases", func() {
			Context("when the reader fails", func() {
				It("returns an error", func() {
					reader.ReadReturns(nil, errors.New("failed to read"))

					_, err := service.FromDirectories([]string{"some-runtime-configs"})
					Expect(err).To(MatchError("failed to read"))
				})
			})
		})
	})
})
