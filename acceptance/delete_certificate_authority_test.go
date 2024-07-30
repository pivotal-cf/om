package acceptance

import (
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("delete certificate authority", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	It("deletes a certificate authority", func() {
		server.AppendHandlers(
			ghttp.VerifyRequest("DELETE", "/api/v0/certificate_authorities/some-id"),
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"delete-certificate-authority",
			"--id", "some-id",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal("Certificate authority 'some-id' deleted\n"))
	})

	When("the certificate authority does not exist", func() {
		It("errors", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/certificate_authorities/missing-id"),
					ghttp.RespondWith(http.StatusNotFound, `{
						"errors": [
							"Certificate with specified guid not found"
						]
					}`),
				),
			)

			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"delete-certificate-authority",
				"--id", "missing-id",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(string(session.Err.Contents())).To(ContainSubstring("Certificate with specified guid not found"))
		})
	})

	When("the certificate authority is still active", func() {
		It("errors", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/certificate_authorities/active-id"),
					ghttp.RespondWith(http.StatusUnprocessableEntity, `{
						"errors": [
							"Active certificates cannot be deleted"
						]
					}`),
				),
			)

			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"delete-certificate-authority",
				"--id", "active-id",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(string(session.Err.Contents())).To(ContainSubstring("Active certificates cannot be deleted"))
		})
	})
})
