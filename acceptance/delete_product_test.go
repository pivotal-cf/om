package acceptance

import (
	"os/exec"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-product command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.VerifyRequest(
				"DELETE",
				"/api/v0/available_products",
				"product_name=some-product&version=1.2.3-build.4",
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("deletes the speecified product", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"delete-product",
			"--product-name", "some-product",
			"--product-version", "1.2.3-build.4",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
	})
})
