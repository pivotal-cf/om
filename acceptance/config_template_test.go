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
	. "github.com/onsi/ginkgo/extensions/table"
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
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v2/products/example-product/releases/24/pivnet_resource_eula_acceptance"),
					ghttp.RespondWith(http.StatusOK, ""),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v2/products/product-24/releases/32/product_files/21/download"),
					ghttp.RespondWith(http.StatusFound, "", map[string][]string{
						"Location": {fmt.Sprintf("%s/download_file/product.pivotal", server.URL())},
					}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("HEAD", "/download_file/product.pivotal"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/download_file/product.pivotal"),
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

		When("the metadata contains a required collection that contains a cert", func() {
			It("renders the cert fields appropriately in the product.yml", func() {
				outputDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())

				productSlug, metadataName, productVersion := "example-product", "example-product", "1.0-build.0"

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

				productYMLFile := filepath.Join(outputDir, metadataName, productVersion, "product.yml")
				Expect(productYMLFile).To(BeAnExistingFile())

				productYMLBytes, err := ioutil.ReadFile(productYMLFile)
				Expect(err).ToNot(HaveOccurred())

				expectedYAML := `.properties.example_required_cert_collection:
    value:
    - required_collection_cert:
        cert_pem: ((example_required_cert_collection_0_certificate))
        private_key_pem: ((example_required_cert_collection_0_privatekey))
`
				Expect(string(productYMLBytes)).To(ContainSubstring(expectedYAML))
			})
		})

		When("the metadata contains a required collection that contains a cert", func() {
			It("renders the cert fields appropriately in the product.yml", func() {
				outputDir, err := ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())

				productSlug, metadataName, productVersion := "example-product", "example-product", "1.0-build.0"

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

				productYMLFile := filepath.Join(outputDir, metadataName, productVersion, "product.yml")
				Expect(productYMLFile).To(BeAnExistingFile())

				productYMLBytes, err := ioutil.ReadFile(productYMLFile)
				Expect(err).ToNot(HaveOccurred())

				expectedYAML := `.properties.required_secret_collection:
    value:
    - password:
        secret: ((required_secret_collection_0_password))
`
				Expect(string(productYMLBytes)).To(ContainSubstring(expectedYAML))
			})
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
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v2/products/another-example-product/releases/14/pivnet_resource_eula_acceptance"),
					ghttp.RespondWith(http.StatusOK, ""),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v2/products/product-14/releases/14/product_files/1/download"),
					ghttp.RespondWith(http.StatusFound, "", map[string][]string{
						"Location": {fmt.Sprintf("%s/download_file/product.pivotal", server.URL())},
					}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("HEAD", "/download_file/product.pivotal"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/download_file/product.pivotal"),
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
					"--file-glob", "product.pivotal",
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

var _ = Describe("config-template output", func() {
	DescribeTable("has the same output as historically cached", func(pivnetSlug, version, glob, metadataName string) {
		pivnetToken := os.Getenv("OM_PIVNET_TOKEN")
		if pivnetToken == "" {
			Skip("OM_PIVNET_TOKEN not specified")
		}

		outputDir, err := ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		command := exec.Command("go", "run", "../main.go",
			"config-template",
			"--output-directory", outputDir,
			"--pivnet-product-slug", pivnetSlug,
			"--product-version", version,
			"--pivnet-api-token", pivnetToken,
			"--file-glob", fmt.Sprintf("%s", glob),
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, "20s", "2s").Should(gexec.Exit(0))

		command = exec.Command("git", "diff",
			filepath.Join(outputDir, metadataName),
			filepath.Join("../configtemplate/generator/fixtures", metadataName),
		)

		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, "10s", "2s").Should(gexec.Exit(0))
	},
		Entry("SRT - for broad coverage", "elastic-runtime", "2.8.6", "*srt*", "cf"),
		Entry("Spring data - for required secret collections", "p-dataflow", "1.6.6", "*.pivotal", "p-dataflow"),
	)
})
