package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("assign-stemcell command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	It("successfully sends the stemcell to the Ops Manager", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/stemcell_assignments"),
				ghttp.RespondWith(http.StatusOK, `{
					"products": [{
						"guid": "cf-guid",
						"identifier": "cf",
						"available_stemcell_versions": [
							"1234.5, 1234.9"
						]
					}]
				}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("PATCH", "/api/v0/stemcell_assignments"),
				ghttp.VerifyJSON(` {
					"products": [{
						"guid": "cf-guid",
						"staged_stemcell_version": "1234.5, 1234.9"
					}]
				}`),
			),
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "pass",
			"--skip-ssl-validation",
			"assign-stemcell",
			"--product", "cf",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say("assigned stemcell successfully"))
	})
})
