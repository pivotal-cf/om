package acceptance

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("disable_director_verifiers command", func() {
	var (
		server *httptest.Server
	)

	BeforeEach(func() {

		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()

			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				_, err := w.Write([]byte(`{
					"access_token": "some-opsman-token",
					"token_type": "bearer",
					"expires_in": 3600
				}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/director/verifiers/install_time":
				Expect(req.Method).To(Equal(http.MethodGet))

				_, err := w.Write([]byte(`{ "verifiers": [
					{ "type":"some-verifier-type", "enabled":true }
				]}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/director/verifiers/install_time/some-verifier-type":
				Expect(req.Method).To(Equal(http.MethodPut))
				body, err := ioutil.ReadAll(req.Body)
				Expect(err).ToNot(HaveOccurred())
				defer req.Body.Close()

				Expect(string(body)).To(Equal(`{ "enabled": false }`))

				_, err = w.Write([]byte(`{
					"type": "some-verifier-type",
					"enabled": false
				}`))
				Expect(err).ToNot(HaveOccurred())
			default:
				w.WriteHeader(http.StatusNotFound)
				_, err := w.Write([]byte(`{
					"errors": {
					  "base": [
						"No verifier on director with type '<missing-verifier>'"
					  ]
					}
				}`))
				Expect(err).ToNot(HaveOccurred())
			}
		}))
	})

	AfterEach(func() {
		server.Close()
	})

	It("disables any verifiers passed in if they exist", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"disable-director-verifiers",
			"--type", "some-verifier-type",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(Equal(`Disabling Director Verifiers...

The following verifiers were disabled:
- some-verifier-type
`))
	})

	It("errors if any verifiers passed in don't exist", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"disable-director-verifiers",
			"--type", "some-verifier-type",
			"-t", "another-verifier-type",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(1))

		Expect(string(session.Out.Contents())).To(Equal(`The following verifiers do not exist:
- another-verifier-type

No changes were made.

`))
	})
})
