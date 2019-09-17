package download_clients_test

import (
	"errors"
	"fmt"
	"github.com/graymeta/stow"
	"github.com/pivotal-cf/om/download_clients"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

// should delete most of s3_client_test (validation stays)
// this file should test the interface methods.

var _ = Describe("stowClient", func() {
	Describe("GetAllProductVersions", func() {
		When("there are multiple files of the same 'version', differing by beta version", func() {
			var (
				stower *mockStower
			)

			BeforeEach(func() {
				itemsList := []mockItem{
					newMockItem("[product-slug,1.0.0-beta.1]someproductfile.zip"),
					newMockItem("[product-slug,1.0.0-beta.2]someproductfile.zip"),
					newMockItem("[product-slug,1.1.1]somefile-0.0.2.zip"),
					newMockItem("[product-slug,1.1.1]someotherfile-0.0.2.zip"),
				}

				stower = newMockStower(itemsList)
			})

			It("reports all versions, including the beta versions", func() {
				client := download_clients.NewStowClient(stower, "", nil, nil, "", "", "")

				versions, err := client.GetAllProductVersions("product-slug")
				Expect(err).ToNot(HaveOccurred())

				Expect(versions).To(Equal([]string{
					"1.0.0-beta.1",
					"1.0.0-beta.2",
					"1.1.1",
				}))
			})
		})

		DescribeTable("the path variable", func(productPath string) {
			var (
				stower *mockStower
			)

			itemsList := []mockItem{
				newMockItem("/some-path/nested-path/[product-slug,8.8.8]someproductfile.zip"),
				newMockItem("/some-path/[product-slug,1.0.0-beta.1]someproductfile.zip"),
				newMockItem("some-path/[product-slug,1.2.3]someproductfile.zip"),
				newMockItem("[product-slug,7.7.7]someotherfile-0.0.2.zip"),
				newMockItem("/some-path/[product-slug,1.1.1]someotherfile-0.0.2.zip"),
				newMockItem("/some-path/[product-slug,1.1.1]with-another-right-bracket-]0.0.2.zip"),
			}

			stower = newMockStower(itemsList)
			client := download_clients.NewStowClient(stower, "", nil, nil, productPath, "", "")

			versions, err := client.GetAllProductVersions("product-slug")
			Expect(err).ToNot(HaveOccurred())

			Expect(versions).To(Equal([]string{
				"1.0.0-beta.1",
				"1.2.3",
				"1.1.1",
			}))
		},
			Entry("with a leading and trailing slash", "/some-path/"),
			Entry("with a leading and without a trailing slash", "/some-path"),
			Entry("without a leading slash", "some-path/"),
			Entry("without a leading or trailing slash", "some-path"),
		)

		When("the container returns 'expected element type <Error>", func() {
			var (
				stower *mockStower
			)

			BeforeEach(func() {
				location := mockLocation{
					containerError: errors.New("expected element type <Error> but have StowErrorType"),
				}
				stower = &mockStower{
					location: location,
				}
			})
			It("returns an error, containing endpoint information, saying S3 could not be reached", func() {
				client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, nil, "", "", "")

				_, err := client.GetAllProductVersions("someslug")
				Expect(err.Error()).To(ContainSubstring("could not reach provided endpoint and bucket 'endpoint/bucket': expected element type <Error> but have StowErrorType"))
			})
		})

		When("zero files match the slug", func() {
			var stower *mockStower

			BeforeEach(func() {
				itemsList := []mockItem{
					newMockItem("product-slug-1.0.0-pcf-vsphere-2.1-build.341.ova"),
					newMockItem("product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova"),
				}

				stower = newMockStower(itemsList)
			})

			It("gives an error message indicating the key and value that were not matched", func() {
				client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, nil, "", "", "")

				_, err := client.GetAllProductVersions("someslug")
				Expect(err.Error()).To(ContainSubstring("no files matching pivnet-product-slug someslug found"))
			})
		})
	})

	Describe("GetLatestProductFile", func() {
		It("returns a file artifact", func() {
			itemsList := []mockItem{
				newMockItem("[product-slug,1.0.0]pcf-vsphere-2.1-build.341.ova"),
				newMockItem("[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList)
			client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, nil, "", "", "")

			fileArtifact, err := client.GetLatestProductFile("product-slug", "1.1.1", "*vsphere*ova")
			Expect(err).ToNot(HaveOccurred())
			Expect(fileArtifact.Name()).To(Equal("[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"))
		})

		It("removes the prefix when globbing", func() {
			itemsList := []mockItem{
				newMockItem("[product-slug,1.0.0]pcf-vsphere-2.1-build.341.ova"),
				newMockItem("[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList)
			client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, nil, "", "", "")

			fileArtifact, err := client.GetLatestProductFile("product-slug", "1.1.1", "pcf-vsphere*ova")
			Expect(err).ToNot(HaveOccurred())
			Expect(fileArtifact.Name()).To(Equal("[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"))
		})

		It("errors when two files match the same glob", func() {
			itemsList := []mockItem{
				newMockItem("[product-slug,1.0.0]pcf-vsphere-2.1-build.341.ova"),
				newMockItem("[product-slug,1.1.1]pcf-vsphere-2.1-build.345.ova"),
				newMockItem("[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList)
			client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, nil, "", "", "")

			_, err := client.GetLatestProductFile("product-slug", "1.1.1", "*vsphere*ova")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the glob '*vsphere*ova' matches multiple files. Write your glob to match exactly one of the following"))
		})

		It("errors when zero prefixed files match the glob", func() {
			itemsList := []mockItem{
				newMockItem("[product-slug,1.0.0]pcf-vsphere-2.1-build.341.ova"),
				newMockItem("[product-slug,1.1.1]pcf-vsphere-2.1-build.345.ova"),
				newMockItem("[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList)
			client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, nil, "", "", "")

			_, err := client.GetLatestProductFile("product-slug", "1.1.1", "*.zip")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the glob '*.zip' matches no file"))
		})

		DescribeTable("the item exists in the path in the bucket", func(productPath string) {
			itemsList := []mockItem{
				newMockItem("/some-path/nested/[product-slug,1.0.0]pcf-vsphere-2.1-build.341.ova"),
				newMockItem("/some-path/[product-slug,1.0.0]pcf-vsphere-2.1-build.341.ova"),
				newMockItem("some-path/[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"),
				newMockItem("[product-slug,7.7.7]pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList)
			client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, nil, productPath, "", "")

			fileArtifact, err := client.GetLatestProductFile("product-slug", "1.1.1", "*vsphere*ova")
			Expect(err).ToNot(HaveOccurred())
			Expect(fileArtifact.Name()).To(Equal("some-path/[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"))
		},
			Entry("with a leading and trailing slash", "/some-path/"),
			Entry("with a leading and without a trailing slash", "/some-path"),
			Entry("without a leading slash", "some-path/"),
			Entry("without a leading or trailing slash", "some-path"),
		)
	})

	Describe("DownloadProductToFile", func() {
		var file *os.File
		var fileContents = "hello world"

		BeforeEach(func() {
			var err error
			file, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = file.WriteString(fileContents)
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Close()).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := os.Remove(file.Name())
			Expect(err).ToNot(HaveOccurred())
		})

		It("writes to a file when the file exists", func() {
			item := newMockItem(file.Name())
			container := mockContainer{item: item}
			location := mockLocation{container: &container}
			stower := &mockStower{
				location:  location,
				itemsList: []mockItem{item},
			}

			client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, GinkgoWriter, "", "", "")

			file, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			err = client.DownloadProductToFile(createPivnetFileArtifact(), file)
			Expect(err).ToNot(HaveOccurred())

			contents, err := ioutil.ReadFile(file.Name())
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(Equal([]byte(fileContents)))
		})

		It("returns a helpful error if the InvalidSignature is returned by container", func() {
			location := mockLocation{
				containerError: errors.New("expected element type <Error> but have StowErrorType"),
			}
			stower := &mockStower{
				location: location,
			}

			file, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, GinkgoWriter, "", "", "")

			err = client.DownloadProductToFile(createPivnetFileArtifact(), file)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not reach provided endpoint and bucket 'endpoint/bucket': expected element type <Error> but have StowErrorType"))
		})

		It("errors when cannot open file", func() {
			item := newMockItem(file.Name())
			item.fileError = errors.New("could not open file")
			container := mockContainer{item: item}
			location := mockLocation{container: &container}
			stower := &mockStower{
				location:  location,
				itemsList: []mockItem{item},
			}

			client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, GinkgoWriter, "", "", "")

			err := client.DownloadProductToFile(createPivnetFileArtifact(), file)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetLatestStemcellForProduct", func() {
		When("the bucket has stemcells that product can used", func() {
			DescribeTable("returns the latest stemcell", func(stemcellName, stemcellProductName, stemcellPath string) {
				exampleTileFileName := createPivotalFile(
					"[example-product,1.0-build.0]example*pivotal",
					stemcellName,
					"97.28",
				)

				stower := &mockStower{
					itemsList: []mockItem{
						newMockItem(fmt.Sprintf("%s[%s,97.28]stemcell.tgz", stemcellPath, stemcellProductName)),
						newMockItem(fmt.Sprintf("%s[%s,97.10]stemcell.tgz", stemcellPath, stemcellProductName)),
						newMockItem(fmt.Sprintf("%s[%s,97.101]stemcell.tgz", stemcellPath, stemcellProductName)),
						newMockItem(fmt.Sprintf("%s[%s,97.asdf]stemcell.tgz", stemcellPath, stemcellProductName)),
					},
				}

				client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, GinkgoWriter, "", stemcellPath, "")

				stemcell, err := client.GetLatestStemcellForProduct(nil, exampleTileFileName)
				Expect(err).ToNot(HaveOccurred())

				Expect(stemcell.Version()).To(Equal("97.101"))
				Expect(stemcell.Slug()).To(Equal(stemcellProductName))
			},
				Entry("supporting xenial", "ubuntu-xenial", "stemcells-ubuntu-xenial", ""),
				Entry("supporting trusty", "ubuntu-trusty", "stemcells", ""),
				Entry("supporting windows2016", "windows2016", "stemcells-windows-server", ""),
				Entry("supporting windows1803", "windows1803", "stemcells-windows-server", ""),
				Entry("supporting windows2019", "windows2019", "stemcells-windows-server", ""),
				Entry("supporting stemcellpath", "ubuntu-xenial", "stemcells-ubuntu-xenial", "/some-path/"),
			)
		})

		Context("failure cases", func() {
			It("errors with malformed stemcell version in the product", func() {
				exampleTileFileName := createPivotalFile(
					"[example-product,1.0-build.0]example*pivotal",
					"ubuntu-xenial",
					"bad-version",
				)

				stower := &mockStower{
					itemsList: []mockItem{},
				}

				client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, GinkgoWriter, "", "", "")

				_, err := client.GetLatestStemcellForProduct(nil, exampleTileFileName)
				Expect(err).To(MatchError("versioning of stemcell dependency in unexpected format: \"major.minor\" or \"major\". the following version could not be parsed: bad-version"))
			})

			It("errors when the product file does not have stemcell information", func() {
				client := download_clients.NewStowClient(nil, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, GinkgoWriter, "", "", "")

				_, err := client.GetLatestStemcellForProduct(nil, "./fixtures/example-product.yml")
				Expect(err).To(HaveOccurred())
			})

			It("errors when there are no available stemcell versions on s3", func() {
				exampleTileFileName := createPivotalFile(
					"[example-product,1.0-build.0]example*pivotal",
					"ubuntu-xenial",
					"97.28",
				)

				stower := &mockStower{
					itemsList: []mockItem{},
				}

				client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, GinkgoWriter, "", "", "blobstore")

				_, err := client.GetLatestStemcellForProduct(nil, exampleTileFileName)
				Expect(err).To(MatchError("could not find stemcells on blobstore: bucket contains no files"))
			})

			It("errors when cannot get latest stemcell version", func() {
				exampleTileFileName := createPivotalFile(
					"[example-product,1.0-build.0]example*pivotal",
					"ubuntu-xenial",
					"97.28",
				)

				stower := &mockStower{
					itemsList: []mockItem{
						newMockItem("[stemcells-ubuntu-xenial,96.28]stemcell.tgz"),
						newMockItem("[stemcells-ubuntu-xenial,96.54]stemcell.tgz"),
						newMockItem("[stemcells-ubuntu-xenial,96.10]stemcell.tgz"),
					},
				}

				client := download_clients.NewStowClient(stower, "bucket", stow.ConfigMap{"endpoint": "endpoint"}, GinkgoWriter, "", "", "blobstore")

				_, err := client.GetLatestStemcellForProduct(nil, exampleTileFileName)
				Expect(err).To(MatchError("no versions could be found equal to or greater than 97.28"))
			})
		})
	})
})
