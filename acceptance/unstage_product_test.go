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

var _ = Describe("unstage-product command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	When("the product is staged", func() {
		BeforeEach(func() {
			server.AppendHandlers(
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
					ghttp.VerifyRequest("DELETE", "/api/v0/staged/products/cf-some-guid"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)
		})

		It("successfully unstages a product from the Ops Manager", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"unstage-product",
				"--product-name", "cf",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("unstaging cf"))
			Eventually(session.Out).Should(gbytes.Say("finished unstaging"))
		})
	})

	When("the product is not staged", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/staged/products/cf-some-guid"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)
		})

		AfterEach(func() {
			server.Close()
		})

		It("returns an error", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"unstage-product",
				"--product-name", "cf",
			)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("product is not staged: cf"))
		})
	})
})
