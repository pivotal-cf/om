package commands_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/formcontent"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UploadProduct", func() {
	var (
		productService *fakes.ProductService
		multipart      *fakes.Multipart
		logger         *fakes.OtherLogger
	)

	BeforeEach(func() {
		multipart = &fakes.Multipart{}
		productService = &fakes.ProductService{}
		logger = &fakes.OtherLogger{}
	})

	It("uploads a product", func() {
		submission := formcontent.ContentSubmission{
			Length:      10,
			Content:     ioutil.NopCloser(strings.NewReader("")),
			ContentType: "some content-type",
		}
		multipart.CreateReturns(submission, nil)

		command := commands.NewUploadProduct(multipart, productService, logger)

		err := command.Execute([]string{
			"--product", "/path/to/some-product.tgz",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(multipart.CreateArgsForCall(0)).To(Equal("/path/to/some-product.tgz"))
		Expect(productService.UploadArgsForCall(0)).To(Equal(api.UploadProductInput{
			ContentLength: 10,
			Product:       ioutil.NopCloser(strings.NewReader("")),
			ContentType:   "some content-type",
		}))

		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("processing product"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("beginning product upload to Ops Manager"))

		format, v = logger.PrintfArgsForCall(2)
		Expect(fmt.Sprintf(format, v...)).To(Equal("finished upload"))
	})

	Context("failure cases", func() {
		Context("when an unkwown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewUploadProduct(multipart, productService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse upload-product flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when the file cannot be opened", func() {
			It("returns an error", func() {
				command := commands.NewUploadProduct(multipart, productService, logger)
				multipart.CreateReturns(formcontent.ContentSubmission{}, errors.New("bad file"))

				err := command.Execute([]string{"--product", "/some/path"})
				Expect(err).To(MatchError("failed to load product: bad file"))
			})
		})

		Context("when the product cannot be uploaded", func() {
			It("returns and error", func() {
				command := commands.NewUploadProduct(multipart, productService, logger)
				productService.UploadReturns(api.UploadProductOutput{}, errors.New("some product error"))

				err := command.Execute([]string{"--product", "/some/path"})
				Expect(err).To(MatchError("failed to upload product: some product error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewUploadProduct(nil, nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This command attempts to upload a product to the Ops Manager",
				ShortDescription: "uploads a given product to the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
