package commands_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/extractor"
	"github.com/fredwangwang/formcontent"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UploadProduct", func() {
	var (
		fakeService       *fakes.UploadProductService
		metadataExtractor *fakes.MetadataExtractor
		multipart         *fakes.Multipart
		logger            *fakes.Logger
	)

	BeforeEach(func() {
		multipart = &fakes.Multipart{}
		fakeService = &fakes.UploadProductService{}
		metadataExtractor = &fakes.MetadataExtractor{}
		logger = &fakes.Logger{}
	})

	It("uploads a product", func() {
		submission := formcontent.ContentSubmission{
			ContentLength:      10,
			Content:     ioutil.NopCloser(strings.NewReader("")),
			ContentType: "some content-type",
		}
		multipart.FinalizeReturns(submission)

		command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)

		err := command.Execute([]string{
			"--product", "/path/to/some-product.tgz",
		})
		Expect(err).NotTo(HaveOccurred())

		key, file := multipart.AddFileArgsForCall(0)
		Expect(key).To(Equal("product[file]"))
		Expect(file).To(Equal("/path/to/some-product.tgz"))
		Expect(fakeService.UploadAvailableProductArgsForCall(0)).To(Equal(api.UploadAvailableProductInput{
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
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			err := command.Execute([]string{
				"--product", "/path/to/some-product.tgz",
				"--polling-interval", "48",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeService.UploadAvailableProductArgsForCall(0).PollingInterval).To(Equal(48))
		})
	})

	Context("when the same product is already present", func() {
		It("does nothing and exits gracefully", func() {
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
				Name:    "cf",
				Version: "1.5.0",
			}, nil)
			fakeService.CheckProductAvailabilityStub = func(name, version string) (bool, error) {
				if name == "cf" && version == "1.5.0" {
					return true, nil
				}
				return false, errors.New("unknown")
			}

			err := command.Execute([]string{
				"--product", "/path/to/some-product.tgz",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(metadataExtractor.ExtractMetadataCallCount()).To(Equal(1))
			Expect(fakeService.UploadAvailableProductCallCount()).To(Equal(0))

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(Equal("product cf 1.5.0 is already uploaded, nothing to be done."))
		})
	})

	Context("when the --shasum flag is defined", func() {
		It("proceeds normally when the sha sums match", func() {
			file, err := ioutil.TempFile("", "test-file.yaml")
			Expect(err).ToNot(HaveOccurred())

			file.Close()
			defer os.Remove(file.Name())

			file.WriteString("testing-shasum")

			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
				Name:    "cf",
				Version: "1.5.0",
			}, nil)
			fakeService.CheckProductAvailabilityStub = func(name, version string) (bool, error) {
				if name == "cf" && version == "1.5.0" {
					return true, nil
				}
				return false, errors.New("unknown")
			}

			err = command.Execute([]string{
				"--product", file.Name(),
				"--shasum", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(metadataExtractor.ExtractMetadataCallCount()).To(Equal(1))
			Expect(fakeService.UploadAvailableProductCallCount()).To(Equal(0))

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(ContainSubstring("expected shasum matches product shasum."))
		})

		It("returns an error when the sha sums don't match", func() {
			file, err := ioutil.TempFile("", "test-file.yaml")
			Expect(err).ToNot(HaveOccurred())

			file.Close()
			defer os.Remove(file.Name())

			file.WriteString("testing-shasum")

			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			err = command.Execute([]string{
				"--product", file.Name(),
				"--shasum", "not-the-correct-shasum",
			})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("expected shasum not-the-correct-shasum does not match file shasum e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"))
		})

		It("fails when the file can not calculate a shasum", func() {
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			err := command.Execute([]string{
				"--product", "/path/to/testing.tgz",
				"--shasum", "not-the-correct-shasum",
			})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("open /path/to/testing.tgz: no such file or directory"))
		})
	})

	Context("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse upload-product flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when the product flag is not provided", func() {
			It("returns an error", func() {
				command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not parse upload-product flags: missing required flag \"--product\""))
			})
		})

		Context("when extracting the product metadata returns an error", func() {
			It("returns an error", func() {
				metadataExtractor.ExtractMetadataReturns(extractor.Metadata{}, errors.New("some error"))
				command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
				err := command.Execute([]string{"--product", "/some/path"})
				Expect(err).To(MatchError("failed to extract product metadata: some error"))
			})
		})

		Context("when checking for product availability returns an error", func() {
			It("returns an error", func() {
				fakeService.CheckProductAvailabilityReturns(true, errors.New("some error"))
				command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
				err := command.Execute([]string{"--product", "/some/path"})
				Expect(err).To(MatchError("failed to check product availability: some error"))
			})
		})

		Context("when adding the file fails", func() {
			It("returns an error", func() {
				command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
				multipart.AddFileReturns(errors.New("bad file"))

				err := command.Execute([]string{"--product", "/some/path"})
				Expect(err).To(MatchError("failed to load product: bad file"))
			})
		})

		Context("when the product cannot be uploaded", func() {
			It("returns an error", func() {
				command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
				fakeService.UploadAvailableProductReturns(api.UploadAvailableProductOutput{}, errors.New("some product error"))

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
