package baking_test

import (
	"errors"

	"github.com/pivotal-cf/kiln/builder"
	"github.com/pivotal-cf/kiln/internal/baking"
	"github.com/pivotal-cf/kiln/internal/baking/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstanceGroupsService", func() {
	Describe("FromDirectories", func() {
		var (
			logger  *fakes.Logger
			reader  *fakes.DirectoryReader
			service baking.InstanceGroupsService
		)

		BeforeEach(func() {
			logger = &fakes.Logger{}
			reader = &fakes.DirectoryReader{}
			reader.ReadReturns([]builder.Part{
				{
					Name: "some-instance-group",
					Metadata: builder.Metadata{
						"key": "value",
					},
				},
			}, nil)

			service = baking.NewInstanceGroupsService(logger, reader)
		})

		It("parses the instance groups passed in a set of directories", func() {
			instanceGroups, err := service.FromDirectories([]string{"some-instance-groups", "other-instance-groups"})
			Expect(err).NotTo(HaveOccurred())
			Expect(instanceGroups).To(Equal(map[string]interface{}{
				"some-instance-group": builder.Metadata{
					"key": "value",
				},
			}))

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			Expect(logger.PrintlnArgsForCall(0)).To(Equal([]interface{}{"Reading instance group files..."}))

			Expect(reader.ReadCallCount()).To(Equal(2))
			Expect(reader.ReadArgsForCall(0)).To(Equal("some-instance-groups"))
			Expect(reader.ReadArgsForCall(1)).To(Equal("other-instance-groups"))
		})

		Context("when the directories argument is empty", func() {
			It("returns nothing", func() {
				instanceGroups, err := service.FromDirectories([]string{})
				Expect(err).NotTo(HaveOccurred())
				Expect(instanceGroups).To(BeNil())

				instanceGroups, err = service.FromDirectories(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(instanceGroups).To(BeNil())

				Expect(logger.PrintlnCallCount()).To(Equal(0))
				Expect(reader.ReadCallCount()).To(Equal(0))
			})
		})

		Context("failure cases", func() {
			Context("when the reader fails", func() {
				It("returns an error", func() {
					reader.ReadReturns(nil, errors.New("failed to read"))

					_, err := service.FromDirectories([]string{"some-instance-groups"})
					Expect(err).To(MatchError("failed to read"))
				})
			})
		})
	})
})
