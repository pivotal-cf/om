package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("installations command", func() {
	const tableOutput = `+----+-------------+-----------+--------------------------+--------------------------+
| ID |    USER     |  STATUS   |        STARTED AT        |       FINISHED AT        |
+----+-------------+-----------+--------------------------+--------------------------+
|  1 | some-user   | succeeded | 2017-05-24T23:38:37.316Z | 2017-05-24T23:55:56.106Z |
|  2 | some-user-2 | failed    | 2017-05-24T23:38:37.316Z | 2017-05-24T23:55:56.106Z |
|  3 | some-user-3 | running   | 2017-05-24T23:38:37.316Z |                          |
+----+-------------+-----------+--------------------------+--------------------------+
`

	const jsonOutput = `[
		{
			"user": "some-user",
			"finished_at": "2017-05-24T23:55:56.106Z",
			"started_at": "2017-05-24T23:38:37.316Z",
			"status": "succeeded",
			"id": 1
		},
		{
			"user": "some-user-2",
			"finished_at": "2017-05-24T23:55:56.106Z",
			"started_at": "2017-05-24T23:38:37.316Z",
			"status": "failed",
			"id": 2
		},
		{
			"user": "some-user-3",
			"started_at": "2017-05-24T23:38:37.316Z",
			"status": "running",
			"id": 3
		}
	]`

	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations"),
				ghttp.RespondWith(http.StatusOK, `{
					"installations": [{
						"user_name": "some-user",
						"finished_at": "2017-05-24T23:55:56.106Z",
						"started_at": "2017-05-24T23:38:37.316Z",
						"status": "succeeded",
						"id": 1
					}, {
						"user_name": "some-user-2",
						"finished_at": "2017-05-24T23:55:56.106Z",
						"started_at": "2017-05-24T23:38:37.316Z",
						"status": "failed",
						"id": 2
					}, {
						"user_name": "some-user-3",
						"finished_at": null,
						"started_at": "2017-05-24T23:38:37.316Z",
						"status": "running",
						"id": 3
					}]
				}`),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("displays a list of recent installation events", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"installations")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, "40s").Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	When("the --format flag is provided with json", func() {
		It("displays a list of recent installation events in json", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"installations",
				"--format", "json")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, "40s").Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})
})
