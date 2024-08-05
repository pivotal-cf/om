package commands_test

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/extractor"
	"github.com/pivotal-cf/om/formcontent"

	. "github.com/onsi/ginkgo/v2"
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
		metadataExtractor.ExtractFromFileReturns(&extractor.Metadata{}, nil)
		logger = &fakes.Logger{}
	})

	It("uploads a product", func() {
		submission := formcontent.ContentSubmission{
			ContentLength: 10,
			Content:       io.NopCloser(strings.NewReader("")),
			ContentType:   "some content-type",
		}
		multipart.FinalizeReturns(submission)

		command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)

		err := executeCommand(command, []string{
			"--product", "/path/to/some-product.tgz",
		})
		Expect(err).ToNot(HaveOccurred())

		key, file := multipart.AddFileArgsForCall(0)
		Expect(key).To(Equal("product[file]"))
		Expect(file).To(Equal("/path/to/some-product.tgz"))
		Expect(fakeService.UploadAvailableProductArgsForCall(0)).To(Equal(api.UploadAvailableProductInput{
			ContentLength:   10,
			Product:         io.NopCloser(strings.NewReader("")),
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

	When("the polling interval is provided", func() {
		It("passes the value to the products service", func() {
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			err := executeCommand(command, []string{
				"--product", "/path/to/some-product.tgz",
				"--polling-interval", "48",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeService.UploadAvailableProductArgsForCall(0).PollingInterval).To(Equal(48))
		})
	})

	When("the same product is already present", func() {
		It("does nothing and exits gracefully", func() {
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			metadataExtractor.ExtractFromFileReturns(&extractor.Metadata{
				Name:    "cf",
				Version: "1.5.0",
			}, nil)
			fakeService.CheckProductAvailabilityStub = func(name, version string) (bool, error) {
				if name == "cf" && version == "1.5.0" {
					return true, nil
				}
				return false, errors.New("unknown")
			}

			err := executeCommand(command, []string{
				"--product", "/path/to/some-product.tgz",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(metadataExtractor.ExtractFromFileCallCount()).To(Equal(1))
			Expect(fakeService.UploadAvailableProductCallCount()).To(Equal(0))

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(Equal("product cf 1.5.0 is already uploaded, nothing to be done"))
		})
	})

	When("the --shasum flag is defined", func() {
		It("proceeds normally when the sha sums match", func() {
			file, err := os.CreateTemp("", "test-file.yaml")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(file.Name())

			_, err = file.WriteString("testing-shasum")
			Expect(err).ToNot(HaveOccurred())
			err = file.Close()
			Expect(err).ToNot(HaveOccurred())

			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			metadataExtractor.ExtractFromFileReturns(&extractor.Metadata{
				Name:    "cf",
				Version: "1.5.0",
			}, nil)
			fakeService.CheckProductAvailabilityStub = func(name, version string) (bool, error) {
				if name == "cf" && version == "1.5.0" {
					return true, nil
				}
				return false, errors.New("unknown")
			}

			err = executeCommand(command, []string{
				"--product", file.Name(),
				"--shasum", "2815ab9694a4a2cfd59424a734833010e143a0b2db20be3741507f177f289f44",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(metadataExtractor.ExtractFromFileCallCount()).To(Equal(1))
			Expect(fakeService.UploadAvailableProductCallCount()).To(Equal(0))

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(ContainSubstring("expected shasum matches product shasum."))
		})

		It("returns an error when the sha sums don't match", func() {
			file, err := os.CreateTemp("", "test-file.yaml")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(file.Name())

			_, err = file.WriteString("testing-shasum")
			Expect(err).ToNot(HaveOccurred())
			err = file.Close()
			Expect(err).ToNot(HaveOccurred())

			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			err = executeCommand(command, []string{
				"--product", file.Name(),
				"--shasum", "not-the-correct-shasum",
			})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("expected shasum not-the-correct-shasum does not match file shasum 2815ab9694a4a2cfd59424a734833010e143a0b2db20be3741507f177f289f44"))
		})

		It("fails when the file can not calculate a shasum", func() {
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			err := executeCommand(command, []string{
				"--product", "/path/to/testing.tgz",
				"--shasum", "not-the-correct-shasum",
			})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("open /path/to/testing.tgz: no such file or directory"))
		})
	})

	When("the --product-version flag is defined", func() {
		It("proceeds normally when the versions match", func() {
			file, err := os.CreateTemp("", "test-file.yaml")
			Expect(err).ToNot(HaveOccurred())
			err = file.Close()
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(file.Name())

			metadataExtractor.ExtractFromFileReturns(&extractor.Metadata{
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

			err = executeCommand(command, []string{
				"--product", file.Name(),
				"--product-version", "1.5.0",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(metadataExtractor.ExtractFromFileCallCount()).To(Equal(1))
			Expect(fakeService.UploadAvailableProductCallCount()).To(Equal(0))

			format, v := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, v...)).To(ContainSubstring("expected version matches product version."))
		})

		It("returns an error when the versions don't match", func() {
			file, err := os.CreateTemp("", "test-file.yaml")
			Expect(err).ToNot(HaveOccurred())
			err = file.Close()
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(file.Name())

			metadataExtractor.ExtractFromFileReturns(&extractor.Metadata{
				Name:    "cf",
				Version: "1.5.0",
			}, nil)
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			err = executeCommand(command, []string{
				"--product", file.Name(),
				"--product-version", "2.5.0",
			})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("expected version 2.5.0 does not match product version 1.5.0"))
		})
	})

	When("the product fails to upload the first time with a retryable error", func() {
		When("the product is now present", func() {
			It("succeeds", func() {
				stdout := gbytes.NewBuffer()
				logger := log.New(stdout, "", 0)

				command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)

				fakeService.UploadAvailableProductReturnsOnCall(0, api.UploadAvailableProductOutput{}, fmt.Errorf("some upload error: %w", io.EOF))
				fakeService.UploadAvailableProductReturnsOnCall(1, api.UploadAvailableProductOutput{}, nil)
				metadataExtractor.ExtractFromFileReturns(&extractor.Metadata{
					Name:    "cf",
					Version: "1.5.0",
				}, nil)
				fakeService.CheckProductAvailabilityReturnsOnCall(0, false, nil)
				fakeService.CheckProductAvailabilityReturnsOnCall(1, true, nil)

				err := executeCommand(command, []string{
					"--product", "/path/to/some-product.tgz",
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(metadataExtractor.ExtractFromFileCallCount()).To(Equal(1))
				Expect(fakeService.UploadAvailableProductCallCount()).To(Equal(1))

				Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("retrying product upload after error: some upload error: EOF")))
				Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("product cf 1.5.0 has been successfully uploaded")))
			})
		})

		It("tries again", func() {
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)

			fakeService.UploadAvailableProductReturnsOnCall(0, api.UploadAvailableProductOutput{}, fmt.Errorf("some upload error: %w", io.EOF))
			fakeService.UploadAvailableProductReturnsOnCall(1, api.UploadAvailableProductOutput{}, nil)

			err := executeCommand(command, []string{"--product", "/some/path"})
			Expect(err).ToNot(HaveOccurred())

			Expect(multipart.AddFileCallCount()).To(Equal(2))
			Expect(multipart.FinalizeCallCount()).To(Equal(2))
			Expect(multipart.ResetCallCount()).To(Equal(1))

			Expect(fakeService.UploadAvailableProductCallCount()).To(Equal(2))
		})
	})

	When("the product fails to upload three times", func() {
		It("returns an error", func() {
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)

			fakeService.CheckProductAvailabilityReturns(false, nil)
			fakeService.UploadAvailableProductReturns(api.UploadAvailableProductOutput{}, fmt.Errorf("some upload error: %w", io.EOF))

			err := executeCommand(command, []string{"--product", "/some/path"})

			Expect(multipart.AddFileCallCount()).To(Equal(3))
			Expect(multipart.FinalizeCallCount()).To(Equal(3))
			Expect(multipart.ResetCallCount()).To(Equal(2))

			Expect(err).To(MatchError(ContainSubstring("EOF")))

			Expect(fakeService.UploadAvailableProductCallCount()).To(Equal(3))
		})
	})

	When("extracting the product metadata returns an error", func() {
		It("returns an error", func() {
			metadataExtractor.ExtractFromFileReturns(&extractor.Metadata{}, errors.New("some error"))
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			err := executeCommand(command, []string{"--product", "/some/path"})
			Expect(err).To(MatchError("failed to extract product metadata: some error"))
		})
	})

	When("checking for product availability returns an error", func() {
		It("returns an error", func() {
			fakeService.CheckProductAvailabilityReturns(true, errors.New("some error"))
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			err := executeCommand(command, []string{"--product", "/some/path"})
			Expect(err).To(MatchError("failed to check product availability: some error"))
		})
	})

	When("adding the file fails", func() {
		It("returns an error", func() {
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			multipart.AddFileReturns(errors.New("bad file"))

			err := executeCommand(command, []string{"--product", "/some/path"})
			Expect(err).To(MatchError("failed to load product: bad file"))
		})
	})

	When("the product cannot be uploaded", func() {
		It("returns an error", func() {
			command := commands.NewUploadProduct(multipart, metadataExtractor, fakeService, logger)
			fakeService.UploadAvailableProductReturns(api.UploadAvailableProductOutput{}, errors.New("some product error"))

			err := executeCommand(command, []string{"--product", "/some/path"})
			Expect(err).To(MatchError("failed to upload product: some product error"))
		})
	})
})
