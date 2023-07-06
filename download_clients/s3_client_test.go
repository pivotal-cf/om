package download_clients_test

import (
	"archive/zip"
	"io/ioutil"
	"log"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
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
				ProductPath:     "some/directory",
			}

			client, err := download_clients.NewS3Client(config, stderr)
			Expect(err).ToNot(HaveOccurred())

			fakeClient := fakes.FakeAWSS3Client{}
			out := s3.ListObjectsV2Output{
				Name: aws.String("bucket"),
				Contents: []types.Object{
					{
						Key: aws.String("some/directory/product-slug-1.2.3.tgz"),
					},
				},
			}
			fakeClient.ListObjectsV2Returns(&out, nil)
			client.Client = &fakeClient

			It("finds the stupid thing", func() {
				products, err := client.GetAllProductVersions("product-slug")
				Expect(err).ToNot(HaveOccurred())

				Expect(products).To(ContainElement("1.2.3"))
			})
		})
	})

	Describe("GetLatestProductFile", func() {})
	Describe("DownloadProductToFile", func() {})
	Describe("GetLatestStemcellForProduct", func() {})
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
