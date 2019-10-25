package acceptance

import (
	"bytes"
	"fmt"
	"github.com/onsi/gomega/ghttp"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("config-template command", func() {

	var server *ghttp.Server

	AfterEach(func() {
		server.Close()
	})

	When("there is only one .pivotal file for the product version", func() {
		BeforeEach(func() {
			pivotalFile := createPivotalFile("[example-product,1.10.1]example*pivotal", "./fixtures/example-product.yml")
			contents, err := ioutil.ReadFile(pivotalFile)
			Expect(err).ToNot(HaveOccurred())
			modTime := time.Now()

			var fakePivnetMetadataResponse []byte

			fixtureMetadata, err := os.Open("fixtures/example-product.yml")
			defer fixtureMetadata.Close()

			Expect(err).ToNot(HaveOccurred())

			_, err = fixtureMetadata.Read(fakePivnetMetadataResponse)
			Expect(err).ToNot(HaveOccurred())

			server = ghttp.NewTLSServer()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases"),
					ghttp.RespondWith(http.StatusOK, `{
  "releases": [
    {
      "id": 24,
      "version": "1.0-build.0"
    }
  ]
}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases/24/product_files"),
					ghttp.RespondWith(http.StatusOK, `{
  "product_files": [
  {
    "id": 1,
    "aws_object_key": "product.pivotal",
    "_links": {
      "download": {
        "href": "http://example.com/api/v2/products/product-24/releases/32/product_files/21/download"
      }
    }
  }
]
}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v2/products/example-product/releases/24/pivnet_resource_eula_acceptance"),
					ghttp.RespondWith(http.StatusOK, ""),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("HEAD", "/api/v2/products/product-24/releases/32/product_files/21/download"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/product-24/releases/32/product_files/21/download"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
			)
		})

		It("writes a config template subdir for the product in the output directory", func() {
			outputDir, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())

			productSlug, productVersion := "example-product", "1.0-build.0"

			command := exec.Command(pathToMain,
				"config-template",
				"--output-directory", outputDir,
				"--pivnet-product-slug", productSlug,
				"--product-version", productVersion,
				"--pivnet-api-token", "token",
				"--pivnet-disable-ssl",
			)
			command.Env = []string{fmt.Sprintf("HTTP_PROXY=%s", server.URL())}

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, "10s").Should(gexec.Exit(0))

			productDir := filepath.Join(outputDir, "example-product", "1.0-build.0")
			Expect(productDir).To(BeADirectory())
		})
	})

	When("there is more than one .pivotal file for a product version", func() {
		BeforeEach(func() {
			pivotalFile := createPivotalFile("[example-product,1.10.1]example*pivotal", "./fixtures/example-product.yml")
			contents, err := ioutil.ReadFile(pivotalFile)
			Expect(err).ToNot(HaveOccurred())
			modTime := time.Now()

			var fakePivnetMetadataResponse []byte

			fixtureMetadata, err := os.Open("fixtures/example-product.yml")
			defer fixtureMetadata.Close()

			Expect(err).ToNot(HaveOccurred())

			_, err = fixtureMetadata.Read(fakePivnetMetadataResponse)
			Expect(err).ToNot(HaveOccurred())

			server = ghttp.NewTLSServer()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/another-example-product/releases"),
					ghttp.RespondWith(http.StatusOK, `{
  "releases": [
    {
      "id": 14,
      "version": "1.0-build.0"
    }
  ]
}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/another-example-product/releases/14/product_files"),
					ghttp.RespondWith(http.StatusOK, `{
  "product_files": [
  {
    "id": 1,
    "aws_object_key": "product.pivotal",
    "_links": {
      "download": {
        "href": "http://example.com/api/v2/products/product-14/releases/14/product_files/1/download"
      }
    }
  },
  {
	"id": 2,
    "aws_object_key": "product-lite.pivotal",
    "_links": {
      "download": {
        "href": "http://example.com/api/v2/products/product-14/releases/14/product_files/2/download"
      }
    }
  }
]
}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v2/products/another-example-product/releases/14/pivnet_resource_eula_acceptance"),
					ghttp.RespondWith(http.StatusOK, ""),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("HEAD", "/api/v2/products/product-14/releases/14/product_files/1/download"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/product-14/releases/14/product_files/1/download"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("HEAD", "/api/v2/products/product-14/releases/14/product_files/2/download"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/product-14/releases/14/product_files/2/download"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
			)
		})
		Context("and the user has not provided a product file glob", func() {
			It("errors because the default glob did not match", func() {
				outputDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())

				productSlug, productVersion := "another-example-product", "1.0-build.0"

				command := exec.Command(pathToMain,
					"config-template",
					"--output-directory", outputDir,
					"--pivnet-product-slug", productSlug,
					"--product-version", productVersion,
					"--pivnet-api-token", "token",
					"--pivnet-disable-ssl",
				)
				command.Env = []string{fmt.Sprintf("HTTP_PROXY=%s", server.URL())}

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session, "10s").Should(gexec.Exit(1))
			})
		})
		Context("and the user has provided a glob with a unique match", func() {
			It("writes a config template subdir for the product in the output directory", func() {
				outputDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())

				productSlug, productVersion := "another-example-product", "1.0-build.0"

				command := exec.Command(pathToMain,
					"config-template",
					"--output-directory", outputDir,
					"--pivnet-product-slug", productSlug,
					"--product-version", productVersion,
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "product.pivotal",
					"--pivnet-disable-ssl",
				)
				command.Env = []string{fmt.Sprintf("HTTP_PROXY=%s", server.URL())}

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session, "10s").Should(gexec.Exit(0))
				productDir := filepath.Join(outputDir, "example-product", "1.0-build.0")
				Expect(productDir).To(BeADirectory())
			})
		})
	})
})
