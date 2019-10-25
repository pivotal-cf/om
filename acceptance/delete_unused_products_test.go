package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-unused-products command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.VerifyRequest("DELETE", "/api/v0/available_products"),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("successfully deletes unused products in Ops Manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"delete-unused-products",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, 5).Should(gexec.Exit(0))
		Eventually(session.Out, 5).Should(gbytes.Say("trashing unused products"))
		Eventually(session.Out, 5).Should(gbytes.Say("done"))
	})
})
