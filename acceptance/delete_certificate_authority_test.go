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

var _ = Describe("delete certificate authority", func() {
	var server *httptest.Server

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				_, err := w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/certificate_authorities/some-id":
				_, err := w.Write([]byte(`{}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/certificate_authorities/missing-id":
				w.WriteHeader(http.StatusNotFound)
				_, err := w.Write([]byte(`{"errors":["Certificate with specified guid not found"]}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/certificate_authorities/active-id":
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, err := w.Write([]byte(`{"errors":["Active certificates cannot be deleted"]}`))
				Expect(err).ToNot(HaveOccurred())
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

	It("deletes a certificate authority", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"delete-certificate-authority",
			"--id", "some-id",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal("Certificate authority 'some-id' deleted\n"))
	})

	When("the certificate authority does not exist", func() {
		It("errors", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"delete-certificate-authority",
				"--id", "missing-id",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(string(session.Err.Contents())).To(ContainSubstring("Certificate with specified guid not found"))
		})
	})

	When("the certificate authority is still active", func() {
		It("errors", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"delete-certificate-authority",
				"--id", "active-id",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(string(session.Err.Contents())).To(ContainSubstring("Active certificates cannot be deleted"))
		})
	})
})
