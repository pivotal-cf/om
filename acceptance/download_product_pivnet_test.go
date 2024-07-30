package acceptance

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

var _ = Describe("download-product command", func() {
	When("downloading from Pivnet", func() {
		var server *ghttp.Server

		BeforeEach(func() {
			pivotalFile := createPivotalFile("[pivnet-product,1.10.1]example*pivotal", "./fixtures/example-product.yml")
			pivotalContents, err := ioutil.ReadFile(pivotalFile)
			Expect(err).ToNot(HaveOccurred())
			modTime := time.Now()

			var fakePivnetMetadataResponse []byte

			fixtureMetadata, err := os.Open("fixtures/example-product.yml")
			defer func() { _ = fixtureMetadata.Close() }()

			Expect(err).ToNot(HaveOccurred())

			_, err = fixtureMetadata.Read(fakePivnetMetadataResponse)
			Expect(err).ToNot(HaveOccurred())

			server = ghttp.NewTLSServer()

			server.RouteToHandler("GET", "/api/v2/products/pivnet-product/releases",
				ghttp.RespondWith(http.StatusOK, `{
  "releases": [
    {
      "id": 24,
      "version": "1.10.1"
    }
  ]
}`))
			server.RouteToHandler("GET", "/api/v2/products/pivnet-product/releases/24",
				ghttp.RespondWith(http.StatusOK, `{"id":24}`),
			)
			server.RouteToHandler("POST", "/api/v2/products/pivnet-product/releases/24/pivnet_resource_eula_acceptance",
				ghttp.RespondWith(http.StatusOK, nil),
			)
			server.RouteToHandler("GET", "/api/v2/products/pivnet-product/releases/24/product_files",
				ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{
					 "product_files": [
					 {
					   "id": 1,

					   "aws_object_key": "pivnet-product.pivotal",
					   "_links": {
					     "download": {
					       "href": "%s/api/v2/products/pivnet-product/releases/24/product_files/1/download"
					     }
					   }
					 }
					]
					}`, server.URL())),
			)
			server.RouteToHandler("GET", "/api/v2/products/pivnet-product/releases/24/file_groups",
				ghttp.RespondWith(http.StatusOK, `{}`),
			)
			server.RouteToHandler("GET", "/api/v2/products/pivnet-product/releases/24/product_files/1",
				ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{
					"product_file": {
					   "id": 1,
						"_links": {
							"download": {
								"href":"%s/api/v2/products/pivnet-product/releases/24/product_files/1/download"
							}
						}
					}
					}`, server.URL())),
			)
			server.RouteToHandler("POST", "/api/v2/products/pivnet-product/releases/24/product_files/1/download",
				ghttp.RespondWith(http.StatusFound, `{}`, http.Header{"Location": {fmt.Sprintf("%s/api/v2/products/pivnet-product/releases/24/product_files/1/download", server.URL())}}),
			)
			server.RouteToHandler("HEAD", "/api/v2/products/pivnet-product/releases/24/product_files/1/download",
				func(w http.ResponseWriter, r *http.Request) {
					http.ServeContent(w, r, "download", modTime, bytes.NewReader(pivotalContents))
				},
			)
			server.RouteToHandler("GET", "/api/v2/products/pivnet-product/releases/24/product_files/1/download",
				func(w http.ResponseWriter, r *http.Request) {
					http.ServeContent(w, r, "download", modTime, bytes.NewReader(pivotalContents))
				},
			)
			server.RouteToHandler("GET", "/api/v2/products/pivnet-product/releases/24/dependencies",
				ghttp.RespondWith(http.StatusOK, `{
					"dependencies": [{
						"release": {
							"id": 24,
							"version": "100.00",
							"product": {"slug": "xenial-stemcells"}
						}
					}]
				}`),
			)

			stemcellContents := `stemcell contents that should do nothing`
			server.RouteToHandler("GET", "/api/v2/products/xenial-stemcells/releases",
				ghttp.RespondWith(http.StatusOK, `{
					  "releases": [
						{
						  "id": 24,
						  "version": "100.00"
						}
					  ]
					}`),
			)
			server.RouteToHandler("GET", "/api/v2/products/xenial-stemcells/releases/24",
				ghttp.RespondWith(http.StatusOK, `{"id":24}`),
			)
			server.RouteToHandler("POST", "/api/v2/products/xenial-stemcells/releases/24/pivnet_resource_eula_acceptance",
				ghttp.RespondWith(http.StatusOK, nil),
			)
			server.RouteToHandler("GET", "/api/v2/products/xenial-stemcells/releases/24/product_files",
				ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{
					 "product_files": [
						 {
						   "id": 1,
	
						   "aws_object_key": "light-bosh-stemcell-621.77-google-kvm-ubuntu-xenial-go_agent.tgz",
						   "_links": {
							 "download": {
							   "href": "%s/api/v2/products/xenial-stemcells/releases/24/product_files/1/download"
							 }
						   }
						 },
						 {
						   "id": 1,
	
						   "aws_object_key": "bosh-stemcell-621.77-google-kvm-ubuntu-xenial-go_agent.tgz",
						   "_links": {
							 "download": {
							   "href": "%s/api/v2/products/xenial-stemcells/releases/24/product_files/1/download"
							 }
						   }
						 }
					]
				}`, server.URL(), server.URL())),
			)
			server.RouteToHandler("GET", "/api/v2/products/xenial-stemcells/releases/24/file_groups",
				ghttp.RespondWith(http.StatusOK, `{}`),
			)
			server.RouteToHandler("GET", "/api/v2/products/xenial-stemcells/releases/24/product_files/1",
				ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{
					"product_file": {
					   "id": 1,
						"_links": {
							"download": {
								"href":"%s/api/v2/products/xenial-stemcells/releases/24/product_files/1/download"
							}
						}
					}
					}`, server.URL())),
			)
			server.RouteToHandler("POST", "/api/v2/products/xenial-stemcells/releases/24/product_files/1/download",
				ghttp.RespondWith(http.StatusFound, `{}`, http.Header{"Location": {fmt.Sprintf("%s/api/v2/products/xenial-stemcells/releases/24/product_files/1/download", server.URL())}}),
			)
			server.RouteToHandler("HEAD", "/api/v2/products/xenial-stemcells/releases/24/product_files/1/download",
				func(w http.ResponseWriter, r *http.Request) {
					http.ServeContent(w, r, "download", modTime, strings.NewReader(stemcellContents))
				},
			)
			server.RouteToHandler("GET", "/api/v2/products/xenial-stemcells/releases/24/product_files/1/download",
				func(w http.ResponseWriter, r *http.Request) {
					http.ServeContent(w, r, "download", modTime, strings.NewReader(stemcellContents))
				},
			)
		})

		AfterEach(func() {
			server.Close()
		})

		It("downloads the product", func() {
			tmpDir, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			command := exec.Command(pathToMain, "download-product",
				"--pivnet-api-token", "token",
				"--file-glob", "pivnet-product.pivotal",
				"--pivnet-product-slug", "pivnet-product",
				"--pivnet-disable-ssl",
				"--pivnet-host", server.URL(),
				"--product-version", "1.10.1",
				"--stemcell-iaas", "google",
				"--output-directory", tmpDir,
			)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(session.Err).To(gbytes.Say(`attempting to download the file.*pivnet-product.pivotal.*from source pivnet`))
			Expect(session.Err).To(gbytes.Say(`Writing a list of downloaded artifact to download-file.json`))

			Expect(filepath.Join(tmpDir, "pivnet-product.pivotal")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "light-bosh-stemcell-621.77-google-kvm-ubuntu-xenial-go_agent.tgz")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "pivnet-product.pivotal.partial")).ToNot(BeAnExistingFile())
		})

		When("the product and stemcell are already on the OpsManager", func() {
			It("does nothing", func() {
				opsmanServer := createTLSServer()
				opsmanServer.RouteToHandler("GET", "/api/v0/available_products",
					ghttp.RespondWith(http.StatusOK, `[{
						"name": "example-product",
						"product_version": "1.0-build.0"
					}]`),
				)
				opsmanServer.RouteToHandler("GET", "/api/v0/diagnostic_report",
					ghttp.RespondWith(http.StatusOK, `{
						"stemcells": ["light-bosh-stemcell-621.77-google-kvm-ubuntu-xenial-go_agent.tgz"]
					}`),
				)
				opsmanServer.RouteToHandler("GET", "/api/v0/info",
					ghttp.RespondWith(http.StatusOK, `{"info":{"version":"2.4.0"}}`),
				)

				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())
				command := exec.Command(pathToMain,
					"-k",
					"--target", opsmanServer.URL(),
					"--username", "some-username",
					"--password", "some-password",
					"download-product",
					"--pivnet-api-token", "token",
					"--file-glob", "pivnet-product.pivotal",
					"--pivnet-product-slug", "pivnet-product",
					"--pivnet-disable-ssl",
					"--pivnet-host", server.URL(),
					"--product-version", "1.10.1",
					"--stemcell-iaas", "google",
					"--output-directory", tmpDir,
					"--check-already-uploaded",
				)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, "10s").Should(gexec.Exit(0))

				Expect(filepath.Join(tmpDir, "pivnet-product.pivotal")).ToNot(BeAnExistingFile())
				Expect(filepath.Join(tmpDir, "light-bosh-stemcell-621.77-google-kvm-ubuntu-xenial-go_agent.tgz")).ToNot(BeAnExistingFile())
				Expect(filepath.Join(tmpDir, "pivnet-product.pivotal.partial")).ToNot(BeAnExistingFile())
			})
		})
	})
})

func fileContents(paths ...string) []byte {
	path := filepath.Join(paths...)
	contents, err := ioutil.ReadFile(filepath.Join(path))
	Expect(err).ToNot(HaveOccurred())
	return contents
}

func createPivotalFile(productFileName, metadataFilename string) string {
	tempfile, err := ioutil.TempFile("", productFileName)
	Expect(err).ToNot(HaveOccurred())

	zipper := zip.NewWriter(tempfile)
	file, err := zipper.Create("metadata/props.yml")
	Expect(err).ToNot(HaveOccurred())

	contents, err := ioutil.ReadFile(metadataFilename)
	Expect(err).ToNot(HaveOccurred())

	_, err = file.Write(contents)
	Expect(err).ToNot(HaveOccurred())

	Expect(zipper.Close()).ToNot(HaveOccurred())
	return tempfile.Name()
}

func uploadGCSFile(localFile, serviceAccountKey, bucketName, objectName string) {
	f, err := os.Open(localFile)
	Expect(err).ToNot(HaveOccurred())

	defer f.Close()

	cxt := context.Background()
	creds, err := google.CredentialsFromJSON(cxt, []byte(serviceAccountKey), storage.ScopeReadWrite)
	Expect(err).ToNot(HaveOccurred())

	client, err := storage.NewClient(cxt, option.WithCredentials(creds))
	Expect(err).ToNot(HaveOccurred())

	wc := client.Bucket(bucketName).Object(objectName).NewWriter(cxt)
	_, err = io.Copy(wc, f)
	Expect(err).ToNot(HaveOccurred())

	err = wc.Close()
	Expect(err).ToNot(HaveOccurred())
}
