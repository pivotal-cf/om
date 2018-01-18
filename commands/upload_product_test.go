package commands_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/formcontent"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UploadProduct", func() {
	var (
		productsService *fakes.ProductUploader
		extractor       *fakes.Extractor
		multipart       *fakes.Multipart
		logger          *fakes.Logger
	)

	BeforeEach(func() {
		multipart = &fakes.Multipart{}
		productsService = &fakes.ProductUploader{}
		extractor = &fakes.Extractor{}
		logger = &fakes.Logger{}
	})

	It("uploads a product", func() {
		submission := formcontent.ContentSubmission{
			Length:      10,
			Content:     ioutil.NopCloser(strings.NewReader("")),
			ContentType: "some content-type",
		}
		multipart.FinalizeReturns(submission, nil)

		command := commands.NewUploadProduct(multipart, extractor, productsService, logger)

		err := command.Execute([]string{
			"--product", "/path/to/some-product.tgz",
		})
		Expect(err).NotTo(HaveOccurred())

		key, file := multipart.AddFileArgsForCall(0)
		Expect(key).To(Equal("product[file]"))
		Expect(file).To(Equal("/path/to/some-product.tgz"))
		Expect(productsService.UploadArgsForCall(0)).To(Equal(api.UploadProductInput{
			ContentLength:   10,
			Product:         ioutil.NopCloser(strings.NewReader("")),
			ContentType:     "some content-type",
			PollingInterval: 1,
		}))

		Expect(multipart.FinalizeCallCount()).To(Equal(1))

		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("processing product"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("beginning product upload to Ops Manager"))

		format, v = logger.PrintfArgsForCall(2)
		Expect(fmt.Sprintf(format, v...)).To(Equal("finished upload"))
	})

	Context("when the polling interval is provided", func() {
		It("passes the value to the products service", func() {
			command := commands.NewUploadProduct(multipart, extractor, productsService, logger)
			err := command.Execute([]string{
				"--product", "/path/to/some-product.tgz",
				"--polling-interval", "48",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(productsService.UploadArgsForCall(0).PollingInterval).To(Equal(48))
		})
	})

	Context("when the same product is already present", func() {
		It("does nothing and exits gracefully", func() {
			command := commands.NewUploadProduct(multipart, extractor, productsService, logger)
			extractor.ExtractMetadataReturns("cf", "1.5.0", nil)
			productsService.CheckProductAvailabilityStub = func(name, version string) (bool, error) {
				if name == "cf" && version == "1.5.0" {
					return true, nil
				}
				return false, errors.New("unknown")
			}

			err := command.Execute([]string{
				"--product", "/path/to/some-product.tgz",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(extractor.ExtractMetadataCallCount()).To(Equal(1))
			Expect(productsService.UploadCallCount()).To(Equal(0))

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(Equal("product cf 1.5.0 is already uploaded, nothing to be done."))
		})
	})

	Context("failure cases", func() {
		Context("when an unkwown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewUploadProduct(multipart, extractor, productsService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse upload-product flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when the product flag is not provided", func() {
			It("returns an error", func() {
				command := commands.NewUploadProduct(multipart, extractor, productsService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not parse upload-product flags: missing required flag \"--product\""))
			})
		})

		Context("when extracting the product metadata returns an error", func() {
			It("returns an error", func() {
				extractor.ExtractMetadataReturns("", "", errors.New("some error"))
				command := commands.NewUploadProduct(multipart, extractor, productsService, logger)
				err := command.Execute([]string{"--product", "/some/path"})
				Expect(err).To(MatchError("failed to extract product metadata: some error"))
			})
		})

		Context("when checking for product availability returns an error", func() {
			It("returns an error", func() {
				productsService.CheckProductAvailabilityReturns(true, errors.New("some error"))
				command := commands.NewUploadProduct(multipart, extractor, productsService, logger)
				err := command.Execute([]string{"--product", "/some/path"})
				Expect(err).To(MatchError("failed to check product availability: some error"))
			})
		})

		Context("when adding the file fails", func() {
			It("returns an error", func() {
				command := commands.NewUploadProduct(multipart, extractor, productsService, logger)
				multipart.AddFileReturns(errors.New("bad file"))

				err := command.Execute([]string{"--product", "/some/path"})
				Expect(err).To(MatchError("failed to load product: bad file"))
			})
		})

		Context("when the product cannot be uploaded", func() {
			It("returns and error", func() {
				command := commands.NewUploadProduct(multipart, extractor, productsService, logger)
				productsService.UploadReturns(api.UploadProductOutput{}, errors.New("some product error"))

				err := command.Execute([]string{"--product", "/some/path"})
				Expect(err).To(MatchError("failed to upload product: some product error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewUploadProduct(nil, nil, nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command attempts to upload a product to the Ops Manager",
				ShortDescription: "uploads a given product to the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
