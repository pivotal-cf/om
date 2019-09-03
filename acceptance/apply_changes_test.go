package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("apply-changes command", func() {
	var (
		server                       *httptest.Server
		installationsStatusCallCount int
		installationsLogsCallCount   int
		logLines                     string
	)

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				username := req.FormValue("username")

				if username == "some-username" {
					_, err := w.Write([]byte(`{
						"access_token": "some-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`))
					Expect(err).ToNot(HaveOccurred())
				} else {
					_, err := w.Write([]byte(`{
						"access_token": "some-running-install-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`))
					Expect(err).ToNot(HaveOccurred())
				}
			case "/api/v0/installations":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" && auth != "Bearer some-running-install-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				if auth == "Bearer some-running-install-opsman-token" {
					_, err := w.Write([]byte(`{ "installations": [ { "id": 42, "status": "running", "started_at": "2017-03-02T06:50:32.370Z" } ] }`))
					Expect(err).ToNot(HaveOccurred())
				} else {
					_, err := w.Write([]byte(`{ "install": { "id": 42 } }`))
					Expect(err).ToNot(HaveOccurred())
				}
			case "/api/v0/installations/42":
				if installationsStatusCallCount == 3 {
					_, err := w.Write([]byte(`{ "status": "succeeded" }`))
					Expect(err).ToNot(HaveOccurred())
					return
				}

				installationsStatusCallCount++
				_, err := w.Write([]byte(`{ "status": "running" }`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/installations/42/logs":
				if installationsLogsCallCount != 3 {
					logLines += fmt.Sprintf("something logged for call #%d\n", installationsLogsCallCount)
				}

				_, err := w.Write([]byte(fmt.Sprintf(`{ "logs": %q }`, logLines)))
				Expect(err).ToNot(HaveOccurred())
				installationsLogsCallCount++
			case "/api/v0/staged/products":
				_, err := w.Write([]byte(`[{"guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"}]`))
				Expect(err).ToNot(HaveOccurred())
				return
			case "/api/v0/deployed/products":
				_, err := w.Write([]byte(`[]`))
				Expect(err).ToNot(HaveOccurred())
				return
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	AfterEach(func() {
		server.Close()
	})

	It("successfully applies the changes to the Ops Manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"apply-changes")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "5s").Should(gexec.Exit(0))

		Expect(installationsStatusCallCount).To(Equal(3))
		Expect(installationsStatusCallCount).To(Equal(3))

		Expect(session.Out).To(gbytes.Say("attempting to apply changes to the targeted Ops Manager"))
		Expect(session.Out).To(gbytes.Say("something logged for call #0"))
		Expect(session.Out).To(gbytes.Say("something logged for call #1"))
		Expect(session.Out).To(gbytes.Say("something logged for call #2"))
	})

	It("successfully re-attaches to an existing deployment", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--username", "some-running-install-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"apply-changes",
			"--reattach")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "5s").Should(gexec.Exit(0))

		Expect(installationsStatusCallCount).To(Equal(3))

		Expect(session.Out).To(gbytes.Say(`found already running installation... re-attaching \(Installation ID: 42, Started: Thu Mar  2 06:50:32 UTC 2017\)`))
		Expect(session.Out).To(gbytes.Say("something logged for call #0"))
		Expect(session.Out).To(gbytes.Say("something logged for call #1"))
		Expect(session.Out).To(gbytes.Say("something logged for call #2"))
	})

})
