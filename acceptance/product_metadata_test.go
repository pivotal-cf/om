package acceptance

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("product-metadata command", func() {
	Context("local file", func() {
		var productFile *os.File

		BeforeEach(func() {
			var err error
			productFile, err = ioutil.TempFile("", "fake-tile")
			Expect(err).ToNot(HaveOccurred())
			z := zip.NewWriter(productFile)

			f, err := z.Create("metadata/fake-tile.yml")
			Expect(err).ToNot(HaveOccurred())

			_, err = f.Write([]byte(`
name: fake-tile
product_version: 1.2.3
`))
			Expect(err).ToNot(HaveOccurred())

			Expect(z.Close()).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(productFile.Name())).To(Succeed())
		})

		It("shows product name from tile metadata file", func() {
			command := exec.Command(pathToMain,
				"product-metadata",
				"-p", productFile.Name(),
				"--product-name",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("fake-tile"))
		})

		It("shows product version from tile metadata file", func() {
			command := exec.Command(pathToMain,
				"product-metadata",
				"-p", productFile.Name(),
				"--product-version",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(ContainSubstring("1.2.3"))
		})
	})

	Context("file on pivnet", func() {
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

				server = ghttp.NewTLSServer()
				server.RouteToHandler("GET", "/api/v2/products/example-product/releases",
					ghttp.RespondWith(http.StatusOK, `{
  "releases": [
    {
      "id": 24,
      "version": "1.0-build.0"
    }
  ]
}`),
				)
				server.RouteToHandler("GET", "/api/v2/products/example-product/releases/24",
					ghttp.RespondWith(http.StatusOK, `{"id":24}`),
				)
				server.RouteToHandler("GET", "/api/v2/products/example-product/releases/24/product_files",
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{
  "product_files": [
  {
    "id": 1,
    "aws_object_key": "product.pivotal",
    "_links": {
      "download": {
        "href": "%s/api/v2/products/product-24/releases/32/product_files/21/download"
      }
    }
  }
]
}`, server.URL())),
				)
				server.RouteToHandler("GET", "/api/v2/products/example-product/releases/24/file_groups",
					ghttp.RespondWith(http.StatusOK, `{}`),
				)
				server.RouteToHandler("POST", "/api/v2/products/example-product/releases/24/pivnet_resource_eula_acceptance",
					ghttp.RespondWith(http.StatusOK, ""),
				)
				server.RouteToHandler("POST", "/api/v2/products/product-24/releases/32/product_files/21/download",
					ghttp.RespondWith(http.StatusFound, "", map[string][]string{
						"Location": {fmt.Sprintf("%s/download_file/product.pivotal", server.URL())},
					}),
				)
				server.RouteToHandler("HEAD", "/download_file/product.pivotal",
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				)
				server.RouteToHandler("GET", "/download_file/product.pivotal",
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				)
			})

			It("shows product name and version from the pivnet metadata", func() {
				productSlug, productVersion := "example-product", "1.0-build.0"

				command := exec.Command(pathToMain,
					"product-metadata",
					"--pivnet-product-slug", productSlug,
					"--pivnet-product-version", productVersion,
					"--pivnet-api-token", "token",
					"--pivnet-disable-ssl",
					"--pivnet-host", server.URL(),
					"--product-name",
					"--product-version",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session, "10s").Should(gexec.Exit(0))

				Expect(session.Out).To(gbytes.Say("example-product\n1.0-build.0"))
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
				defer func() { _ = fixtureMetadata.Close() }()

				Expect(err).ToNot(HaveOccurred())

				_, err = fixtureMetadata.Read(fakePivnetMetadataResponse)
				Expect(err).ToNot(HaveOccurred())

				server = ghttp.NewTLSServer()
				server.RouteToHandler("GET", "/api/v2/products/another-example-product/releases",
					ghttp.RespondWith(http.StatusOK, `{
  "releases": [
    {
      "id": 14,
      "version": "1.0-build.0"
    }
  ]
}`))
				server.RouteToHandler("GET", "/api/v2/products/another-example-product/releases/14",
					ghttp.RespondWith(http.StatusOK, `{"id":14}`),
				)
				server.RouteToHandler("GET", "/api/v2/products/another-example-product/releases/14/product_files",
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{
  "product_files": [
  {
    "id": 1,
    "aws_object_key": "product.pivotal",
    "_links": {
      "download": {
        "href": "%s/api/v2/products/product-14/releases/14/product_files/1/download"
      }
    }
  },
  {
	"id": 2,
    "aws_object_key": "product-lite.pivotal",
    "_links": {
      "download": {
        "href": "%s/api/v2/products/product-14/releases/14/product_files/2/download"
      }
    }
  }
]
}`, server.URL(), server.URL())),
				)
				server.RouteToHandler("GET", "/api/v2/products/another-example-product/releases/14/file_groups",
					ghttp.RespondWith(http.StatusOK, `{}`),
				)
				server.RouteToHandler("POST", "/api/v2/products/another-example-product/releases/14/pivnet_resource_eula_acceptance",
					ghttp.RespondWith(http.StatusOK, ""),
				)
				server.RouteToHandler("POST", "/api/v2/products/product-14/releases/14/product_files/1/download",
					ghttp.RespondWith(http.StatusFound, "", map[string][]string{
						"Location": {fmt.Sprintf("%s/download_file/product.pivotal", server.URL())},
					}),
				)
				server.RouteToHandler("HEAD", "/download_file/product.pivotal",
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				)
				server.RouteToHandler("GET", "/download_file/product.pivotal",
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				)
			})
			Context("and the user has not provided a product file glob", func() {
				It("errors because the default glob did not match", func() {
					productSlug, productVersion := "another-example-product", "1.0-build.0"

					command := exec.Command(pathToMain,
						"product-metadata",
						"--pivnet-product-slug", productSlug,
						"--pivnet-product-version", productVersion,
						"--pivnet-api-token", "token",
						"--pivnet-disable-ssl",
						"--pivnet-host", server.URL(),
						"--product-name",
						"--product-version",
					)

					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())

					Eventually(session, "10s").Should(gexec.Exit(1))
				})
			})
			Context("and the user has provided a glob with a unique match", func() {
				It("shows product name and version from the pivnet metadata", func() {
					productSlug, productVersion := "another-example-product", "1.0-build.0"

					command := exec.Command(pathToMain,
						"product-metadata",
						"--pivnet-product-slug", productSlug,
						"--pivnet-product-version", productVersion,
						"--pivnet-api-token", "token",
						"--file-glob", "product.pivotal",
						"--pivnet-disable-ssl",
						"--pivnet-host", server.URL(),
						"--product-name",
						"--product-version",
					)

					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())

					Eventually(session, "10s").Should(gexec.Exit(0))

					Expect(session.Out).To(gbytes.Say("example-product\n1.0-build.0"))
				})
			})
		})
	})
})
