package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("installations command", func() {
	var server *httptest.Server

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

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/api/v0/installations":
				w.Write([]byte(`{
				  "installations": [
				    {
				      "user_name": "some-user",
				      "finished_at": "2017-05-24T23:55:56.106Z",
				      "started_at": "2017-05-24T23:38:37.316Z",
				      "status": "succeeded",
				      "id": 1
				    },
				    {
				      "user_name": "some-user-2",
				      "finished_at": "2017-05-24T23:55:56.106Z",
				      "started_at": "2017-05-24T23:38:37.316Z",
				      "status": "failed",
				      "id": 2
				    },
				    {
				      "user_name": "some-user-3",
				      "finished_at": null,
				      "started_at": "2017-05-24T23:38:37.316Z",
				      "status": "running",
				      "id": 3
				    }
				  ]
				}`))
			case "/uaa/oauth/token":
				w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
				}`))
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	It("displays a list of recent installation events", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"installations")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "40s").Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(Equal(tableOutput))
	})

	Context("when the --format flag is provided with json", func() {
		It("displays a list of recent installation events in json", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"--format", "json",
				"installations")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "40s").Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(MatchJSON(jsonOutput))
		})
	})
})
