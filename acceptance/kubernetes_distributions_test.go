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

var _ = Describe("kubernetes-distributions command", func() {
	var server *ghttp.Server

	BeforeEach(func() {
		server = createTLSServer()
		DeferCleanup(server.Close)
	})

	It("lists distributions from the library with product associations", func() {
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
						"guid": "rabbitmq-abc123",
						"identifier": "rabbitmq-on-k8s",
						"is_staged_for_deletion": false,
						"staged_kubernetes_distribution": {
							"identifier": "managed-k8s",
							"version": "0.2.0"
						},
						"deployed_kubernetes_distribution": null,
						"available_kubernetes_distributions": [
							{"identifier": "managed-k8s", "version": "0.1.0"},
							{"identifier": "managed-k8s", "version": "0.2.0"}
						]
					}],
					"kubernetes_distribution_library": [
						{"identifier": "unmanaged-k8s", "version": "0.1.0", "rank": 50, "label": "Unmanaged Kubernetes"},
						{"identifier": "managed-k8s", "version": "0.1.0", "rank": 1, "label": "Managed Kubernetes"},
						{"identifier": "managed-k8s", "version": "0.2.0", "rank": 1, "label": "Managed Kubernetes"}
					]
				}`),
			),
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "pass",
			"--skip-ssl-validation",
			"kubernetes-distributions",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(SatisfyAll(
			ContainSubstring("DISTRIBUTION"),
			ContainSubstring("VERSION"),
		))
		Expect(output).ToNot(ContainSubstring("STAGED"))
		Expect(output).ToNot(ContainSubstring("DEPLOYED"))
		Expect(output).To(ContainSubstring("unmanaged-k8s"))
		Expect(output).To(ContainSubstring("managed-k8s"))
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
			"kubernetes-distributions",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, 10*time.Second).Should(gexec.Exit(1))
		Eventually(session.Err).Should(gbytes.Say("kubernetes-distributions requires Ops Manager 3.3 or newer"))
	})
})
