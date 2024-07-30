package acceptance

import (
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("activate certificate authority", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	It("activates a certificate authority", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities/some-id/activate"),
				ghttp.VerifyJSON(`{}`),
			),
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"activate-certificate-authority",
			"--id", "some-id",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal("Certificate authority 'some-id' activated\n"))
	})

	It("errors when the certificate authority does not exist", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities/missing-id/activate"),
				ghttp.VerifyJSON(`{}`),
				ghttp.RespondWith(http.StatusNotFound, `{"errors":["Certificate with specified guid not found"]}`),
			),
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"activate-certificate-authority",
			"--id", "missing-id",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(1))
		Expect(string(session.Out.Contents())).To(Not(ContainSubstring("Certificate authority 'missing-id' activated\n")))
	})
})
