package acceptance

import (
	"archive/zip"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/onsi/gomega/ghttp"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("import-installation command", func() {
	var (
		installation string
		passphrase   string
		content      *os.File
		server       *ghttp.Server
	)

	createZipFile := func(files []struct{ Name, Body string }) *os.File {
		tmpFile, err := os.CreateTemp("", "")
		w := zip.NewWriter(tmpFile)

		Expect(err).ToNot(HaveOccurred())
		for _, file := range files {
			f, err := w.Create(file.Name)
			if err != nil {
				Expect(err).ToNot(HaveOccurred())
			}
			_, err = f.Write([]byte(file.Body))
			if err != nil {
				Expect(err).ToNot(HaveOccurred())
			}
		}
		err = w.Close()
		Expect(err).ToNot(HaveOccurred())

		return tmpFile
	}

	BeforeEach(func() {
		content = createZipFile([]struct{ Name, Body string }{
			{"installation.yml", ""},
		})

		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	It("successfully uploads an installation to the Ops Manager", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/login/ensure_availability"),
				ghttp.RespondWith(http.StatusFound, "", map[string][]string{"Location": {"/setup"}}),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v0/installation_asset_collection"),
				http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					defer GinkgoRecover()

					err := req.ParseMultipartForm(100)
					Expect(err).ToNot(HaveOccurred())

					installation = req.MultipartForm.File["installation[file]"][0].Filename
					passphrase = req.MultipartForm.Value["passphrase"][0]
				}),
				ghttp.RespondWith(http.StatusOK, `{}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/login/ensure_availability"),
				ghttp.RespondWith(http.StatusFound, "", map[string][]string{"Location": {"/auth/cloudfoundry"}}),
			),
		)

		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--decryption-passphrase", "fake-passphrase",
			"--skip-ssl-validation",
			"import-installation",
			"--polling-interval", "0",
			"--installation", content.Name(),
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session, 5).Should(gexec.Exit(0))
		Eventually(session.Out, 5).Should(gbytes.Say("processing installation"))
		Eventually(session.Out, 5).Should(gbytes.Say("beginning installation import to Ops Manager"))
		Eventually(session.Out, 5).Should(gbytes.Say("finished import"))

		Expect(installation).To(Equal(filepath.Base(content.Name())))
		Expect(passphrase).To(Equal("fake-passphrase"))
	})

	When("the ops manager is already configured", func() {
		It("returns an error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/login/ensure_availability"),
					ghttp.RespondWith(http.StatusFound, "", map[string][]string{"Location": {"/auth/cloudfoundry"}}),
				),
			)

			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--decryption-passphrase", "fake-passphrase",
				"--skip-ssl-validation",
				"import-installation",
				"--installation", content.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, 5).Should(gexec.Exit(0))
			Eventually(session.Out, 5).Should(gbytes.Say("Ops Manager is already configured"))
		})
	})

	When("the content cannot be read", func() {
		BeforeEach(func() {
			err := os.Remove(content.Name())
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--decryption-passphrase", "fake-passphrase",
				"--skip-ssl-validation",
				"import-installation",
				"--installation", content.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, 5).Should(gexec.Exit(1))
			Eventually(session.Err, 5).Should(gbytes.Say(`does not exist. Please check the name and try again.`))
		})
	})
})
