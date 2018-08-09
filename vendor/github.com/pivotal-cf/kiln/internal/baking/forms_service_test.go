package baking_test

import (
	"errors"

	"github.com/pivotal-cf/kiln/builder"
	"github.com/pivotal-cf/kiln/internal/baking"
	"github.com/pivotal-cf/kiln/internal/baking/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FormsService", func() {
	Describe("FromDirectories", func() {
		var (
			logger  *fakes.Logger
			reader  *fakes.DirectoryReader
			service baking.FormsService
		)

		BeforeEach(func() {
			logger = &fakes.Logger{}
			reader = &fakes.DirectoryReader{}
			reader.ReadReturnsOnCall(1, []builder.Part{
				{
					File: "some-form-file",
					Name: "some-form-name",
					Metadata: map[string]interface{}{
						"some-key": "some-value",
					},
				},
			}, nil)

			service = baking.NewFormsService(logger, reader)
		})

		It("parses the forms passed in a set of directories", func() {
			forms, err := service.FromDirectories([]string{"some-forms", "other-forms"})
			Expect(err).NotTo(HaveOccurred())

			Expect(forms).To(Equal(map[string]interface{}{
				"some-form-name": map[string]interface{}{
					"some-key": "some-value",
				},
			}))

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			Expect(logger.PrintlnArgsForCall(0)).To(Equal([]interface{}{"Reading form files..."}))

			Expect(reader.ReadCallCount()).To(Equal(2))
			Expect(reader.ReadArgsForCall(0)).To(Equal("some-forms"))
			Expect(reader.ReadArgsForCall(1)).To(Equal("other-forms"))
		})

		Context("when there are no directories to parse", func() {
			It("returns nothing", func() {
				forms, err := service.FromDirectories([]string{})
				Expect(err).NotTo(HaveOccurred())
				Expect(forms).To(BeNil())

				forms, err = service.FromDirectories(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(forms).To(BeNil())

				Expect(logger.PrintlnCallCount()).To(Equal(0))
				Expect(reader.ReadCallCount()).To(Equal(0))
			})
		})

		Context("failure cases", func() {
			Context("when the reader fails", func() {
				It("returns an error", func() {
					reader.ReadReturns(nil, errors.New("failed to read"))

					_, err := service.FromDirectories([]string{"some-forms"})
					Expect(err).To(MatchError("failed to read"))
				})
			})
		})
	})
})
