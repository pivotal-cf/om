package acceptance

import (
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/ghttp"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("pending_changes command", func() {
	const tableOutput = `+----------------+---------+--------------------+
|    PRODUCT     | ACTION  |      ERRANDS       |
+----------------+---------+--------------------+
| some-product-1 | update  | smoke-tests        |
|                |         | deploy-autoscaling |
| some-product-2 | install | deploy-broker      |
| some-product-3 | delete  | delete-broker      |
+----------------+---------+--------------------+
`

	const jsonOutput = `[
		{
			"guid": "some-product-1",
			"errands": [
				{"post_deploy": "true", "pre_delete": true, "name": "smoke-tests"},
				{"post_deploy": "false", "pre_delete": false, "name": "deploy-autoscaling"}
			],
			"action": "update",
			"an-extra-filed": "is preserved"
		},
		{
			"guid": "some-product-2",
			"errands": [
				{"post_deploy": "when-changed", "name": "deploy-broker"}
			],
			"action": "install"
		},
		{
			"guid": "some-product-3",
			"errands": [
				{"post_deploy": "when-changed", "name": "delete-broker"}
			],
			"action": "delete"
		}
	]`

	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/staged/pending_changes"),
				ghttp.RespondWith(http.StatusOK, `{
					"product_changes": [{
						"guid": "some-product-1",
						"errands": [
							{"post_deploy": "true", "pre_delete": true, "name": "smoke-tests"},
							{"post_deploy": "false", "pre_delete": false, "name": "deploy-autoscaling"}
						],
						"action": "update",
						"an-extra-filed": "is preserved"
					},
					{
						"guid": "some-product-2",
						"errands": [
							{"post_deploy": "when-changed", "name": "deploy-broker"}
						],
						"action": "install"
					},
					{
						"guid": "some-product-3",
						"errands": [
							{"post_deploy": "when-changed", "name": "delete-broker"}
						],
						"action": "delete"
					}]
				}`),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("lists the pending changes belonging to the product", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"pending-changes")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	When("JSON format is requested", func() {
		It("lists the pending changes in JSON format", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"pending-changes",
				"--format", "json")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})

})
