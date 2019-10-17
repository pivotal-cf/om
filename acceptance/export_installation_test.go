package acceptance

import (
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("export-installation command", func() {
	var (
		server         *ghttp.Server
		outputFileName string
	)

	BeforeEach(func() {
		tempFile, err := ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())
		outputFileName = tempFile.Name()

		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installation_asset_collection"),
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					defer GinkgoRecover()

					time.Sleep(1010 * time.Millisecond)
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
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, 5).Should(gexec.Exit(0))
		Expect(session.Err).To(gbytes.Say("exporting installation"))
		Expect(session.Err).To(gbytes.Say("waiting for response"))
		Expect(session.Err).To(gbytes.Say(`100(\.\d+)?`))
		Expect(session.Err).To(gbytes.Say("finished exporting installation"))

		content, err := ioutil.ReadFile(outputFileName)
		Expect(err).NotTo(HaveOccurred())
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
			Expect(err).NotTo(HaveOccurred())

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
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 3).Should(gexec.Exit(1))
			Eventually(session.Err, 3).Should(gbytes.Say(`.*request canceled \(Client\.Timeout exceeded while awaiting headers\)`))
		})
	})
})
