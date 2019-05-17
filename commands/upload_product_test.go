package commands_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/extractor"
	"github.com/pivotal-cf/om/formcontent"
	"github.com/pkg/errors"

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
			ContentLength: 10,
			Content:       ioutil.NopCloser(strings.NewReader("")),
			ContentType:   "some content-type",
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
			defer os.Remove(file.Name())

			_, err = file.WriteString("testing-shasum")
			Expect(err).ToNot(HaveOccurred())
			err = file.Close()
			Expect(err).ToNot(HaveOccurred())

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
				"--shasum", "2815ab9694a4a2cfd59424a734833010e143a0b2db20be3741507f177f289f44",
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
			defer os.Remove(file.Name())

			_, err = file.WriteString("testing-shasum")
			Expect(err).ToNot(HaveOccurred())
			err = file.Close()
			Expect(err).ToNot(HaveOccurred())

			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			err = command.Execute([]string{
				"--product", file.Name(),
				"--shasum", "not-the-correct-shasum",
			})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("expected shasum not-the-correct-shasum does not match file shasum 2815ab9694a4a2cfd59424a734833010e143a0b2db20be3741507f177f289f44"))
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

	Context("when the --product-version flag is defined", func() {
		It("proceeds normally when the versions match", func() {
			file, err := ioutil.TempFile("", "test-file.yaml")
			Expect(err).ToNot(HaveOccurred())
			err = file.Close()
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(file.Name())

			metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
				Name:    "cf",
				Version: "1.5.0",
			}, nil)
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			fakeService.CheckProductAvailabilityStub = func(name, version string) (bool, error) {
				if name == "cf" && version == "1.5.0" {
					return true, nil
				}
				return false, errors.New("unknown")
			}

			err = command.Execute([]string{
				"--product", file.Name(),
				"--product-version", "1.5.0",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(metadataExtractor.ExtractMetadataCallCount()).To(Equal(1))
			Expect(fakeService.UploadAvailableProductCallCount()).To(Equal(0))

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(ContainSubstring("expected version matches product version."))
		})

		It("returns an error when the versions don't match", func() {
			file, err := ioutil.TempFile("", "test-file.yaml")
			Expect(err).ToNot(HaveOccurred())
			err = file.Close()
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(file.Name())

			metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
				Name:    "cf",
				Version: "1.5.0",
			}, nil)
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			err = command.Execute([]string{
				"--product", file.Name(),
				"--product-version", "2.5.0",
			})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("expected version 2.5.0 does not match product version 1.5.0"))
		})
	})

	Context("when the product fails to upload the first time with a retryable error", func() {
		Context("when the product is now present", func() {
			It("does nothing and exits gracefully", func() {
				command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)

				fakeService.UploadAvailableProductReturnsOnCall(0, api.UploadAvailableProductOutput{}, errors.Wrap(io.EOF, "some upload error"))
				fakeService.UploadAvailableProductReturnsOnCall(1, api.UploadAvailableProductOutput{}, nil)
				metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
					Name:    "cf",
					Version: "1.5.0",
				}, nil)
				fakeService.CheckProductAvailabilityReturnsOnCall(0, false, nil)
				fakeService.CheckProductAvailabilityReturnsOnCall(1, true, nil)

				err := command.Execute([]string{
					"--product", "/path/to/some-product.tgz",
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(metadataExtractor.ExtractMetadataCallCount()).To(Equal(1))
				Expect(fakeService.UploadAvailableProductCallCount()).To(Equal(1))

				loggerPrintfCalls := logger.PrintfCallCount()

				format, v := logger.PrintfArgsForCall(loggerPrintfCalls - 2)
				Expect(fmt.Sprintf(format, v...)).To(Equal("retrying product upload after error: some upload error: EOF\n"))

				format, v = logger.PrintfArgsForCall(loggerPrintfCalls - 1)
				Expect(fmt.Sprintf(format, v...)).To(Equal("product cf 1.5.0 is already uploaded, nothing to be done."))
			})
		})

		It("tries again", func() {
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)

			fakeService.UploadAvailableProductReturnsOnCall(0, api.UploadAvailableProductOutput{}, errors.Wrap(io.EOF, "some upload error"))
			fakeService.UploadAvailableProductReturnsOnCall(1, api.UploadAvailableProductOutput{}, nil)

			err := command.Execute([]string{"--product", "/some/path"})
			Expect(err).NotTo(HaveOccurred())

			Expect(multipart.AddFileCallCount()).To(Equal(2))
			Expect(multipart.FinalizeCallCount()).To(Equal(2))
			Expect(multipart.ResetCallCount()).To(Equal(1))

			Expect(fakeService.UploadAvailableProductCallCount()).To(Equal(2))
		})
	})

	Context("when the product fails to upload three times", func() {
		It("returns an error", func() {
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)

			fakeService.CheckProductAvailabilityReturns(false, nil)
			fakeService.UploadAvailableProductReturns(api.UploadAvailableProductOutput{}, errors.Wrap(io.EOF, "some upload error"))

			err := command.Execute([]string{"--product", "/some/path"})

			Expect(multipart.AddFileCallCount()).To(Equal(3))
			Expect(multipart.FinalizeCallCount()).To(Equal(3))
			Expect(multipart.ResetCallCount()).To(Equal(2))

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("EOF"))

			Expect(fakeService.UploadAvailableProductCallCount()).To(Equal(3))
		})
	})

	Context("when config file is provided", func() {
		var configFile *os.File

		BeforeEach(func() {
			var err error
			configContent := `
product-version: 1.5.0
product: will-be-overridden-by-command-line
`
			configFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = configFile.WriteString(configContent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("reads configuration from config file", func() {
			file, err := ioutil.TempFile("", "test-file.yaml")
			Expect(err).ToNot(HaveOccurred())
			err = file.Close()
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(file.Name())

			metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
				Name:    "cf",
				Version: "1.5.0",
			}, nil)
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			fakeService.CheckProductAvailabilityStub = func(name, version string) (bool, error) {
				if name == "cf" && version == "1.5.0" {
					return true, nil
				}
				return false, errors.New("unknown")
			}

			err = command.Execute([]string{
				"--config", configFile.Name(),
				"--product", file.Name(),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(metadataExtractor.ExtractMetadataCallCount()).To(Equal(1))
			Expect(fakeService.UploadAvailableProductCallCount()).To(Equal(0))

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(ContainSubstring("expected version matches product version."))
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
