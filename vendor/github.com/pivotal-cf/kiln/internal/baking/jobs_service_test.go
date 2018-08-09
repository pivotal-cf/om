package baking_test

import (
	"errors"

	"github.com/pivotal-cf/kiln/builder"
	"github.com/pivotal-cf/kiln/internal/baking"
	"github.com/pivotal-cf/kiln/internal/baking/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JobsService", func() {
	Describe("FromDirectories", func() {
		var (
			logger  *fakes.Logger
			reader  *fakes.DirectoryReader
			service baking.JobsService
		)

		BeforeEach(func() {
			logger = &fakes.Logger{}
			reader = &fakes.DirectoryReader{}
			reader.ReadReturns([]builder.Part{
				{
					Name: "some-job",
					Metadata: builder.Metadata{
						"key": "value",
					},
				},
			}, nil)

			service = baking.NewJobsService(logger, reader)
		})

		It("parses the jobs passed in a set of directories", func() {
			jobs, err := service.FromDirectories([]string{"some-jobs", "other-jobs"})
			Expect(err).NotTo(HaveOccurred())
			Expect(jobs).To(Equal(map[string]interface{}{
				"some-job": builder.Metadata{
					"key": "value",
				},
			}))

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			Expect(logger.PrintlnArgsForCall(0)).To(Equal([]interface{}{"Reading jobs files..."}))

			Expect(reader.ReadCallCount()).To(Equal(2))
			Expect(reader.ReadArgsForCall(0)).To(Equal("some-jobs"))
			Expect(reader.ReadArgsForCall(1)).To(Equal("other-jobs"))
		})

		Context("when the directories argument is empty", func() {
			It("returns nothing", func() {
				jobs, err := service.FromDirectories(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(jobs).To(BeNil())

				jobs, err = service.FromDirectories([]string{})
				Expect(err).NotTo(HaveOccurred())
				Expect(jobs).To(BeNil())
			})
		})

		Context("failure cases", func() {
			Context("when the reader fails", func() {
				It("returns an error", func() {
					reader.ReadReturns(nil, errors.New("failed to read"))

					_, err := service.FromDirectories([]string{"some-jobs"})
					Expect(err).To(MatchError("failed to read"))
				})
			})
		})
	})
})
