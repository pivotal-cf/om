package baking_test

import (
	"errors"

	"github.com/pivotal-cf/kiln/builder"
	"github.com/pivotal-cf/kiln/internal/baking"
	"github.com/pivotal-cf/kiln/internal/baking/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BOSHVariablesService", func() {
	Describe("FromDirectories", func() {
		var (
			service baking.BOSHVariablesService
			logger  *fakes.Logger
			reader  *fakes.DirectoryReader
		)

		BeforeEach(func() {
			logger = &fakes.Logger{}
			reader = &fakes.DirectoryReader{}

			service = baking.NewBOSHVariablesService(logger, reader)

			reader.ReadReturns([]builder.Part{
				{
					Name: "some-key",
					Metadata: builder.Metadata{
						"type": "user",
						"options": map[string]interface{}{
							"username": "some-username",
						},
					},
				},
			}, nil)

			service = baking.NewBOSHVariablesService(logger, reader)
		})

		It("parses template variables from a collection of files", func() {
			variables, err := service.FromDirectories([]string{"some-bosh-variables"})
			Expect(err).NotTo(HaveOccurred())
			Expect(variables).To(Equal(map[string]interface{}{
				"some-key": builder.Metadata{
					"type": "user",
					"options": map[string]interface{}{
						"username": "some-username",
					},
				},
			}))
		})

		Context("failure cases", func() {
			Context("when the directories argument is empty", func() {
				It("returns nothing", func() {
					boshVariables, err := service.FromDirectories([]string{})
					Expect(err).NotTo(HaveOccurred())
					Expect(boshVariables).To(BeNil())

					boshVariables, err = service.FromDirectories(nil)
					Expect(err).NotTo(HaveOccurred())
					Expect(boshVariables).To(BeNil())

					Expect(logger.PrintlnCallCount()).To(Equal(0))
					Expect(reader.ReadCallCount()).To(Equal(0))
				})
			})

			Context("failure cases", func() {
				Context("when the reader fails", func() {
					It("returns an error", func() {
						reader.ReadReturns(nil, errors.New("failed to read"))

						_, err := service.FromDirectories([]string{"some-bosh-variables"})
						Expect(err).To(MatchError("failed to read"))
					})
				})
			})
		})
	})
})
