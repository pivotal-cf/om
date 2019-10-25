package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("stage-product command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations"),
				ghttp.RespondWith(http.StatusOK, `{"installations": []}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
				ghttp.RespondWith(http.StatusOK, `{}`),
			),
		)
	})

	When("the same type of product is not already deployed", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
					ghttp.RespondWith(http.StatusOK, `[{
						"type": "bosh",
						"installation_name": "bosh-some-other-guid",
						"guid": "bosh-some-other-guid"
					}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/available_products"),
					ghttp.RespondWith(http.StatusOK, `[{
						"name": "cf",
						"product_version": "1.8.7-build.3"
					}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/staged/products"),
					ghttp.VerifyJSON(`{"name":"cf","product_version":"1.8.7-build.3"}`),
				),
			)
		})

		It("successfully stages a product to the Ops Manager", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"stage-product",
				"--product-name", "cf",
				"--product-version", "1.8.7-build.3",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("staging cf"))
			Eventually(session.Out).Should(gbytes.Say("finished staging"))
		})
	})

	When("the same type of product is already deployed", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
					ghttp.RespondWith(http.StatusOK, `[{
						"type": "cf",
						"installation_name": "cf-some-guid",
						"guid": "cf-some-guid"
					}, {
						"type": "bosh",
						"installation_name": "bosh-some-other-guid",
						"guid": "bosh-some-other-guid"
					}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/available_products"),
					ghttp.RespondWith(http.StatusOK, `[{
						"name": "cf",
						"product_version": "1.8.7-build.3"
					}, {
						"name": "cf",
						"product_version": "1.8.5-build.1"
					}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/cf-some-guid"),
					ghttp.VerifyJSON(` {"to_version": "1.8.7-build.3"}`),
				),
			)
		})

		It("successfully stages a product to the Ops Manager", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"stage-product",
				"--product-name", "cf",
				"--product-version", "1.8.7-build.3",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("staging cf"))
			Eventually(session.Out).Should(gbytes.Say("finished staging"))
		})
	})

	When("the same type of product is already staged", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
					ghttp.RespondWith(http.StatusOK, `[]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/available_products"),
					ghttp.RespondWith(http.StatusOK, `[{
						"name": "cf",
						"product_version": "1.8.6-build.3"
					}, {
						"name": "cf",
						"product_version": "1.8.5-build.1"
					}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[{
						"type": "cf",
						"guid": "cf-some-guid"
					}, {
						"type": "bosh",
						"guid": "bosh-some-other-guid"
					}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/cf-some-guid"),
					ghttp.VerifyJSON(` {"to_version": "1.8.6-build.3"}`),
				),
			)
		})

		It("successfully stages a product to the Ops Manager", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"stage-product",
				"--product-name", "cf",
				"--product-version", "1.8.6-build.3",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("staging cf"))
			Eventually(session.Out).Should(gbytes.Say("finished staging"))
		})
	})

	When("the product is not available", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
					ghttp.RespondWith(http.StatusOK, `[{
							"type": "bosh",
							"installation_name": "bosh-some-other-guid",
							"guid": "bosh-some-other-guid"
						}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/available_products"),
					ghttp.RespondWith(http.StatusOK, `[{
							"name": "cf",
							"product_version": "1.8.7-build.3"
						}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/cf-some-guid"),
					ghttp.VerifyJSON(` {"to_version": "1.8.6-build.3"}`),
				),
			)
		})

		It("returns an error", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"stage-product",
				"--product-name", "bosh",
				"--product-version", "2.0",
			)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("cannot find product bosh 2.0"))
		})
	})
})
