package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("create VM extension", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	It("creates a VM extension in OpsMan", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v0/staged/vm_extensions/some-vm-extension"),
				ghttp.VerifyJSON(`{
					"name": "some-vm-extension",
					"cloud_properties": {
						"iam_instance_profile": "some-iam-profile",
						"elbs": [
							"some-elb"
						]
					}
				}`),
			),
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"create-vm-extension",
			"--name", "some-vm-extension",
			"--cloud-properties", "{ \"iam_instance_profile\": \"some-iam-profile\", \"elbs\": [\"some-elb\"] }",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal("VM Extension 'some-vm-extension' created/updated\n"))
	})
})
