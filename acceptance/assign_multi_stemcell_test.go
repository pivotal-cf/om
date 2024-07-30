package acceptance

import (
	"net/http"
	"os/exec"
	"time"

	"github.com/onsi/gomega/ghttp"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("assign-multi-stemcell command", func() {
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
				ghttp.VerifyRequest("GET", "/api/v0/info"),
				ghttp.RespondWith(http.StatusOK, `{
					"info": {
						"version": "2.6.0"
					}
				}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/stemcell_associations"),
				ghttp.RespondWith(http.StatusOK, `{
					"products": [{
						"guid": "cf-guid",
						"identifier": "cf",
						"available_stemcells": [{
							"os": "ubuntu-trusty",
							"version": "1234.5"
						}, {
							"os": "ubuntu-trusty",
							"version": "1234.57"
						}]
					}]
				}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("PATCH", "/api/v0/stemcell_associations"),
				ghttp.VerifyJSON(`{
					"products": [{
						"guid": "cf-guid",
						"staged_stemcells": [{
							"os": "ubuntu-trusty",
							"version": "1234.57"
						}]
					}]
				}`),
			),
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "pass",
			"--skip-ssl-validation",
			"assign-multi-stemcell",
			"--product", "cf",
			"--stemcell", "ubuntu-trusty:latest",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say(`assigning stemcells: "ubuntu-trusty 1234.57" to product "cf"`))
		Eventually(session.Out).Should(gbytes.Say("assigned stemcells successfully"))
	})
})
