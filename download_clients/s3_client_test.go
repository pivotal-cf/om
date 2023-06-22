package download_clients_test

import (
	"archive/zip"
	"io/ioutil"
	"log"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/download_clients"
	"github.com/pivotal-cf/om/download_clients/fakes"
)

var _ = Describe("S3Client", func() {
	var stderr *log.Logger

	BeforeEach(func() {
		stderr = log.New(GinkgoWriter, "", 0)
	})

	Describe("NewS3Client", func() {
		When("auth type is not set and key/secret are missing", func() {
			It("defaults to accesskey and fails", func() {
				_, err := download_clients.NewS3Client(download_clients.S3Configuration{}, stderr)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Name", func() {
		It("returns the name of the client", func() {
			config := download_clients.S3Configuration{
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				Bucket:          "bucket",
			}
			client, err := download_clients.NewS3Client(config, stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(client.Name()).To(ContainSubstring("s3://bucket"))
		})
	})

	Describe("GetAllProductVersions", func() {
		When("the bucket contains a product", func() {
			config := download_clients.S3Configuration{
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				Bucket:          "bucket",
			}

			client, err := download_clients.NewS3Client(config, stderr)
			Expect(err).ToNot(HaveOccurred())

			fakeClient := fakes.FakeAWSS3Client{}
			bucket := "bucket"
			fileName := "some-thing-v1.2.3"
			out := s3.ListObjectsV2Output{
				Name: &bucket,
				Contents: []types.Object{
					{
						Key: &fileName,
					},
				},
			}
			fakeClient.ListObjectsV2Returns(&out, nil)
			client.Client = &fakeClient

			It("finds the stupid thing", func() {
				products, err := client.GetAllProductVersions("product-slug")
				Expect(err).ToNot(HaveOccurred())

				Expect(products).To(ContainElement("some-thing-v1.2.3"))
			})
		})
	})

	Describe("GetLatestProductFile", func() {})
	Describe("DownloadProductToFile", func() {})
	Describe("GetLatestStemcellForProduct", func() {})

	// Describe("property validation and defaults", func() {
	// 	DescribeTable("required property validation", func(param string) {
	// 		stower := &mockStower{}
	// 		config := download_clients.S3Configuration{}
	// 		_, err := download_clients.NewS3Client(stower, config, stderr)
	// 		Expect(err).To(MatchError(ContainSubstring("Field validation for '%s' failed on the 'required' tag", param)))
	// 	},
	// 		Entry("requires Bucket", "Bucket"),
	// 		Entry("requires RegionName", "RegionName"),
	// 	)

	// 	It("defaults optional properties", func() {
	// 		config := download_clients.S3Configuration{
	// 			Bucket:          "bucket",
	// 			AccessKeyID:     "access-key-id",
	// 			SecretAccessKey: "secret-access-key",
	// 			RegionName:      "region",
	// 			Endpoint:        "endpoint",
	// 		}
	// 		stower := &mockStower{itemsList: []mockItem{}}
	// 		client, err := download_clients.NewS3Client(stower, config, stderr)
	// 		Expect(err).ToNot(HaveOccurred())

	// 		retrievedDisableSSLValue, retrievedValuePresence := client.Config.Config("disable_ssl")
	// 		Expect(retrievedValuePresence).To(Equal(true))
	// 		Expect(retrievedDisableSSLValue).To(Equal("false"))

	// 		retrievedAuthTypeValue, retrievedValuePresence := client.Config.Config("auth_type")
	// 		Expect(retrievedValuePresence).To(Equal(true))
	// 		Expect(retrievedAuthTypeValue).To(Equal("accesskey"))
	// 	})

	// 	When("both region and endpoint are given", func() {
	// 		It("returns an error if they do not match", func() {
	// 			config := download_clients.S3Configuration{
	// 				Bucket:          "bucket",
	// 				AccessKeyID:     "access-key-id",
	// 				SecretAccessKey: "secret-access-key",
	// 				RegionName:      "wrongRegion",
	// 				Endpoint:        "endpoint",
	// 			}
	// 			stower := &mockStower{itemsList: []mockItem{}}
	// 			_, err := download_clients.NewS3Client(stower, config, stderr)
	// 			Expect(err).ToNot(HaveOccurred())
	// 		})
	// 	})
	// 	When("AuthType is set", func() {
	// 		var config download_clients.S3Configuration
	// 		BeforeEach(func() {
	// 			config = download_clients.S3Configuration{
	// 				Bucket:          "bucket",
	// 				AccessKeyID:     "access-key-id",
	// 				SecretAccessKey: "secret-access-key",
	// 				RegionName:      "wrongRegion",
	// 				Endpoint:        "endpoint",
	// 				AuthType:        "fakeAuthType",
	// 			}
	// 		})

	// 		It("passes the auth_type down to stow", func() {
	// 			stower := &mockStower{itemsList: []mockItem{}}
	// 			client, err := download_clients.NewS3Client(stower, config, stderr)
	// 			Expect(err).ToNot(HaveOccurred())

	// 			retrievedAuthTypeValue, retrievedValuePresence := client.Config.Config("auth_type")
	// 			Expect(retrievedValuePresence).To(Equal(true))
	// 			Expect(retrievedAuthTypeValue).To(Equal("fakeAuthType"))

	// 		})

	// 		When("AuthType is 'iam' and the id/secret are not provided", func() {
	// 			BeforeEach(func() {
	// 				config.AuthType = "iam"
	// 				config.AccessKeyID = ""
	// 				config.SecretAccessKey = ""
	// 			})

	// 			It("does not raise a validation error", func() {
	// 				stower := &mockStower{itemsList: []mockItem{}}
	// 				_, err := download_clients.NewS3Client(stower, config, stderr)
	// 				Expect(err).ToNot(HaveOccurred())
	// 			})
	// 		})
	// 		When("AuthType is accesskey/default and the id/secret are not provided", func() {
	// 			BeforeEach(func() {
	// 				config.AuthType = "accesskey"
	// 				config.AccessKeyID = ""
	// 				config.SecretAccessKey = ""

	// 			})

	// 			It("raises a validation error", func() {
	// 				stower := &mockStower{itemsList: []mockItem{}}
	// 				_, err := download_clients.NewS3Client(stower, config, stderr)
	// 				Expect(err).To(HaveOccurred())
	// 			})
	// 		})
	// 	})
	// })

	// It("returns an error on stower failure", func() {
	// 	dialError := errors.New("dial error")
	// 	itemsList := []mockItem{{}}
	// 	stower := newMockStower(itemsList)
	// 	stower.dialError = dialError

	// 	config := download_clients.S3Configuration{
	// 		Bucket:          "bucket",
	// 		AccessKeyID:     "access-key-id",
	// 		SecretAccessKey: "secret-access-key",
	// 		RegionName:      "region",
	// 		Endpoint:        "endpoint",
	// 	}

	// 	client, err := download_clients.NewS3Client(stower, config, stderr)
	// 	Expect(err).ToNot(HaveOccurred())

	// 	_, err = client.GetAllProductVersions("product-slug")
	// 	Expect(err).To(HaveOccurred())
	// 	Expect(err).To(Equal(dialError))
	// })
})

func createPivotalFile(productFileName, stemcellName, stemcellVersion string) string {
	tempfile, err := ioutil.TempFile("", productFileName)
	Expect(err).ToNot(HaveOccurred())

	zipper := zip.NewWriter(tempfile)
	file, err := zipper.Create("metadata/props.yml")
	Expect(err).ToNot(HaveOccurred())

	contents, err := ioutil.ReadFile("./fixtures/example-product-metadata.yml")
	Expect(err).ToNot(HaveOccurred())

	context := struct {
		StemcellName    string
		StemcellVersion string
	}{
		StemcellName:    stemcellName,
		StemcellVersion: stemcellVersion,
	}

	tmpl, err := template.New("example-product").Parse(string(contents))
	Expect(err).ToNot(HaveOccurred())

	err = tmpl.Execute(file, context)
	Expect(err).ToNot(HaveOccurred())

	Expect(zipper.Close()).ToNot(HaveOccurred())
	return tempfile.Name()
}
