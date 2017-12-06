package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pending_changes command", func() {
	var (
		server *httptest.Server
	)

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
			"action": "update"
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

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
			case "/api/v0/staged/pending_changes":
				w.Write([]byte(`{
					"product_changes": [{
						"guid": "some-product-1",
						"errands": [
							{"post_deploy": "true", "pre_delete": true, "name": "smoke-tests"},
							{"post_deploy": "false", "pre_delete": false, "name": "deploy-autoscaling"}
						],
						"action": "update"
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
				}`))
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	It("lists the pending changes belonging to the product", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"pending-changes")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	Context("when JSON format is requrested", func() {
		It("lists the pending changes in JSON format", func() {
			command := exec.Command(pathToMain,
				"--format", "json",
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"pending-changes")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})

})
