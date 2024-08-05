package acceptance

import (
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("export-installation command", func() {
	var (
		server         *ghttp.Server
		outputFileName string
	)

	BeforeEach(func() {
		tempFile, err := os.CreateTemp("", "")
		Expect(err).ToNot(HaveOccurred())
		outputFileName = tempFile.Name()

		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installation_asset_collection"),
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					defer GinkgoRecover()

					time.Sleep(2 * time.Second)
					_, err := w.Write([]byte("some-installation"))
					Expect(err).ToNot(HaveOccurred())
				}),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("successfully exports the installation of the ops-manager", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"export-installation",
			"--output-file", outputFileName,
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, 5).Should(gexec.Exit(0))
		Expect(session.Err).To(gbytes.Say("exporting installation"))
		Expect(session.Err).To(gbytes.Say("finished exporting installation"))

		content, err := os.ReadFile(outputFileName)
		Expect(err).ToNot(HaveOccurred())
		Expect(content).To(Equal([]byte("some-installation")))
	})

	When("the output file cannot be written to", func() {
		It("returns an error", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"export-installation",
				"--output-file", "fake-dir/fake-file",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, 5).Should(gexec.Exit(1))
			Eventually(session.Err, 5).Should(gbytes.Say("cannot create output file:"))
		})
	})

	When("the request takes longer than specified timeout", func() {
		It("returns an error", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"--request-timeout", "1",
				"export-installation",
				"--output-file", outputFileName,
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, 3).Should(gexec.Exit(1))
			Eventually(session.Err, 3).Should(gbytes.Say(`Client.Timeout exceeded while awaiting headers`))
		})
	})
})
