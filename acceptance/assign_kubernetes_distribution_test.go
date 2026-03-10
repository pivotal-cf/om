package acceptance

import (
	"net/http"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("assign-kubernetes-distribution command", func() {
	var server *ghttp.Server

	BeforeEach(func() {
		server = createTLSServer()
		DeferCleanup(server.Close)
	})

	It("successfully assigns a kubernetes distribution to a product", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/info"),
				ghttp.RespondWith(http.StatusOK, `{
					"info": {
						"version": "3.3.0"
					}
				}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/kubernetes_distribution_associations"),
				ghttp.RespondWith(http.StatusOK, `{
					"products": [{
						"guid": "kafka-abc123",
						"identifier": "kafka-on-k8s",
						"is_staged_for_deletion": false,
						"available_kubernetes_distributions": [
							{"identifier": "unmanaged-k8s", "version": "1.0.0"},
							{"identifier": "managed-k8s", "version": "1.23.0"}
						]
					}]
				}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("PATCH", "/api/v0/kubernetes_distribution_associations"),
				ghttp.VerifyJSON(`{
					"products": [{
						"guid": "kafka-abc123",
						"kubernetes_distribution": {
							"identifier": "unmanaged-k8s",
							"version": "1.0.0"
						}
					}]
				}`),
				ghttp.RespondWith(http.StatusOK, `{}`),
			),
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "pass",
			"--skip-ssl-validation",
			"assign-kubernetes-distribution",
			"--product", "kafka-on-k8s",
			"--distribution", "unmanaged-k8s:1.0.0",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		Eventually(session.Out).Should(gbytes.Say("assigned kubernetes distribution successfully"))
	})

	It("fails when Ops Manager version is too old", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/info"),
				ghttp.RespondWith(http.StatusOK, `{
					"info": {
						"version": "3.2.0"
					}
				}`),
			),
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "pass",
			"--skip-ssl-validation",
			"assign-kubernetes-distribution",
			"--product", "kafka-on-k8s",
			"--distribution", "unmanaged-k8s:1.0.0",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(1))
		Eventually(session.Err).Should(gbytes.Say("assign-kubernetes-distribution requires Ops Manager 3.3 or newer"))
	})
})
