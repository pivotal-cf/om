package acceptance

import (
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/ghttp"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("staged-products command", func() {
	const tableOutput = `+----------------+------------------+
|      NAME      |     VERSION      |
+----------------+------------------+
| acme-product-1 | 1.13.0-build.100 |
| acme-product-2 | 1.8.9-build.1    |
+----------------+------------------+
`

	const jsonOutput = `[
		{"name":"acme-product-1","version":"1.13.0-build.100"},
		{"name":"acme-product-2","version":"1.8.9-build.1"}
	]`

	const diagnosticReport = `{
		"added_products": {
			"staged": [
				{"name":"acme-product-1","version":"1.13.0-build.100"},
				{"name":"acme-product-2","version":"1.8.9-build.1"}
			]
		}
	}`

	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
				ghttp.RespondWith(http.StatusOK, diagnosticReport),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("lists the staged products on Ops Manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"staged-products",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	When("json format is requested", func() {
		It("lists the staged products on Ops Manager", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"staged-products",
				"--format", "json",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})
})
