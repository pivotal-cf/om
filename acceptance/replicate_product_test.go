package acceptance

import (
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("replicate-product command", func() {
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
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/available_products"),
				ghttp.RespondWith(http.StatusOK, `[{
					"name": "p-isolation-segment",
					"product_version": "10.4.0-build.7"
				}]`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v0/staged/products"),
				ghttp.VerifyJSON(`{"name":"p-isolation-segment","product_version":"10.4.0-build.7","replicate":true,"replica_suffix":"fun-suffix-2"}`),
				ghttp.RespondWith(http.StatusOK, ``),
			),
		)
	})

	It("successfully replicates a product in Ops Manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"replicate-product",
			"--product-name", "p-isolation-segment",
			"--product-version", "10.4.0-build.7",
			"--replica-suffix", "fun-suffix-2",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say("replicating p-isolation-segment"))
		Eventually(session.Out).Should(gbytes.Say("finished replicating"))
	})

	When("a --config option is passed", func() {
		It("loads product name, version, and replica-suffix from the config file", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"replicate-product",
				"--config", writeFile(`
product-name: p-isolation-segment
product-version: 10.4.0-build.7
replica-suffix: fun-suffix-2
`),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("replicating p-isolation-segment"))
			Eventually(session.Out).Should(gbytes.Say("finished replicating"))
		})
	})
})
