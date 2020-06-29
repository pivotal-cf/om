package acceptance

import (
	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/cmd"
	"io/ioutil"
	"net/http"
	"time"
)

var _ = XDescribe("CheckProductUploaded", func() {
	var (
		pivnetServer              *ghttp.Server
		opsmanServer              *ghttp.Server
		downloadProductConfigFile string
	)

	BeforeEach(func() {
		pivnetServer = createTLSServer()
		downloadProductConfigFile = writeFile(fmt.Sprintf(`---
pivnet-api-token: "token"
file-glob: "*.pivotal"
pivnet-product-slug: pivnet-slug
pivnet-disable-ssl: true
pivnet-host: %s
product-version: 1.10.1`, pivnetServer.URL()))

		opsmanServer = createTLSServer()
	})

	AfterEach(func() {
		pivnetServer.Close()
		opsmanServer.Close()
	})

	When("Pivnet borks", func() {
		BeforeEach(func() {
			pivnetServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/pivnet-slug/releases"),
					ghttp.RespondWith(http.StatusNotFound, ``),
				),
			)
		})

		It("returns an error", func() {
			err := cmd.Main(GinkgoWriter, GinkgoWriter, "", "10ms",
				[]string{
					"om", "--target", opsmanServer.URL(),
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"check-product-uploaded",
					"-c", downloadProductConfigFile,
				},
			)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not parse response from Pivnet"))
		})
	})

	When("product metadata cannot be parsed", func() {
		BeforeEach(func() {
			pivotalFile := createPivotalFile("[pivnet-slug,1.10.1]example*pivotal", writeFile(`%*&`))
			contents, err := ioutil.ReadFile(pivotalFile)
			Expect(err).ToNot(HaveOccurred())
			modTime := time.Now()

			pivnetServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/pivnet-slug/releases"),
					ghttp.RespondWith(http.StatusOK, `{
  "releases": [
    {
      "id": 24,
      "version": "1.10.1"
    }
  ]
}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/pivnet-slug/releases/24/product_files"),
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
}`, pivnetServer.URL())),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v2/products/pivnet-slug/releases/24/pivnet_resource_eula_acceptance"),
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

		It("returns an error", func() {
			err := cmd.Main(GinkgoWriter, GinkgoWriter, "", "10ms",
				[]string{
					"om", "--target", opsmanServer.URL(),
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"check-product-uploaded",
					"-c", downloadProductConfigFile,
				},
			)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not parse metadata from Pivnet product"))
		})
	})

	When("the products exists on Pivnet", func() {
		BeforeEach(func() {
			pivotalFile := createPivotalFile("[pivnet-slug,1.10.1]example*pivotal", "./fixtures/example-product.yml")
			contents, err := ioutil.ReadFile(pivotalFile)
			Expect(err).ToNot(HaveOccurred())
			modTime := time.Now()

			pivnetServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/pivnet-slug/releases"),
					ghttp.RespondWith(http.StatusOK, `{
  "releases": [
    {
      "id": 24,
      "version": "1.10.1"
    }
  ]
}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/pivnet-slug/releases/24/product_files"),
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
}`, pivnetServer.URL())),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v2/products/pivnet-slug/releases/24/pivnet_resource_eula_acceptance"),
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
		When("product is already on OpsManager", func() {
			BeforeEach(func() {
				opsmanServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/available_products"),
						ghttp.RespondWith(http.StatusOK, `[{
						"name": "example-product",
						"product_version": "1.2.3"
					}, {
						"name":"example-product",
						"product_version":"1.0-build.0"
					}]`),
					),
				)
			})

			It("returns exit code success", func() {
				err := cmd.Main(GinkgoWriter, GinkgoWriter, "", "10ms",
					[]string{
						"om", "--target", opsmanServer.URL(),
						"--username", "some-username",
						"--password", "some-password",
						"--skip-ssl-validation",
						"check-product-uploaded",
						"-c", downloadProductConfigFile,
					},
				)

				Expect(pivnetServer.ReceivedRequests()).To(HaveLen(5))
				Expect(opsmanServer.ReceivedRequests()).To(HaveLen(2))

				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("products cannot be found on OpsManager", func() {
			BeforeEach(func() {
				opsmanServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/available_products"),
						ghttp.RespondWith(http.StatusNotFound, `[]`),
					),
				)
			})

			It("returns exit code failure", func() {
				err := cmd.Main(GinkgoWriter, GinkgoWriter, "", "10ms",
					[]string{
						"om", "--target", opsmanServer.URL(),
						"--username", "some-username",
						"--password", "some-password",
						"--skip-ssl-validation",
						"check-product-uploaded",
						"-c", downloadProductConfigFile,
					},
				)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find products on OpsManager"))
			})
		})

		When("product is not OpsManager", func() {
			BeforeEach(func() {
				opsmanServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/available_products"),
						ghttp.RespondWith(http.StatusOK, `[]`),
					),
				)
			})

			It("returns exit code failure", func() {
				err := cmd.Main(GinkgoWriter, GinkgoWriter, "", "10ms",
					[]string{
						"om", "--target", opsmanServer.URL(),
						"--username", "some-username",
						"--password", "some-password",
						"--skip-ssl-validation",
						"check-product-uploaded",
						"-c", downloadProductConfigFile,
					},
				)

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
