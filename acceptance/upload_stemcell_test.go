package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("upload-stemcell command", func() {
	var (
		stemcellName string
		server       *ghttp.Server
	)

	createStemcell := func(filename string) string {
		dir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		path := filepath.Join(dir, filename)
		err = ioutil.WriteFile(path, []byte("content so validation does not fail"), 0777)
		Expect(err).NotTo(HaveOccurred())
		return path
	}

	BeforeEach(func() {
		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	When("the stemcell upload is successful", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/info"),
					ghttp.RespondWith(http.StatusOK, `{
					"info": {
						"version": "2.1-build.79"
					}
				}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/stemcells"),
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						err := req.ParseMultipartForm(100)
						Expect(err).ToNot(HaveOccurred())

						stemcellName = req.MultipartForm.File["stemcell[file]"][0].Filename

						_, err = w.Write([]byte(`{}`))
						Expect(err).ToNot(HaveOccurred())
					}),
				),
			)
		})

		It("successfully sends the stemcell to the Ops Manager", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "pass",
				"--skip-ssl-validation",
				"upload-stemcell",
				"--stemcell", createStemcell("stemcell.tgz"),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 10*time.Second).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("processing stemcell"))
			Eventually(session.Out).Should(gbytes.Say("beginning stemcell upload to Ops Manager"))
			Eventually(session.Out).Should(gbytes.Say("finished upload"))

			Expect(stemcellName).To(Equal("stemcell.tgz"))
		})

		When("the stemcell name has the `download-product` prefix", func() {
			It("successfully sends the stemcell to the Ops Manager", func() {
				filename := createStemcell("[ubuntu-xenial,97.88]stemcell.tgz")
				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--username", "some-username",
					"--password", "pass",
					"--skip-ssl-validation",
					"upload-stemcell",
					"--stemcell", filename,
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 10*time.Second).Should(gexec.Exit(0))
				Eventually(session.Out).Should(gbytes.Say("processing stemcell"))
				Eventually(session.Out).Should(gbytes.Say("beginning stemcell upload to Ops Manager"))
				Eventually(session.Out).Should(gbytes.Say("finished upload"))

				Expect(stemcellName).To(Equal("stemcell.tgz"))
				Expect(filename).To(BeAnExistingFile())
			})
		})
	})

	When("the stemcell already exists", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
					ghttp.RespondWith(http.StatusOK, `{
						"stemcells": [
							"bosh-stemcell-3215-vsphere-esxi-ubuntu-trusty-go_agent.tgz",
							"stemcell.tgz"
						]
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/info"),
					ghttp.RespondWith(http.StatusOK, `{
					"info": {
						"version": "2.1-build.79"
					}
				}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/stemcells"),
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						err := req.ParseMultipartForm(100)
						Expect(err).ToNot(HaveOccurred())

						stemcellName = req.MultipartForm.File["stemcell[file]"][0].Filename

						_, err = w.Write([]byte(`{}`))
						Expect(err).ToNot(HaveOccurred())
					}),
				),
			)
		})

		It("exits early with no error", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-stemcell",
				"--stemcell", createStemcell("stemcell.tgz"),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 10*time.Second).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("stemcell has already been uploaded"))
		})
	})

	When("an error occurs", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/info"),
					ghttp.RespondWith(http.StatusOK, `{
						"info": {
							"version": "2.1-build.79"
						}
					}`),
				),
			)
		})
		When("the content to upload is empty", func() {
			It("returns an error", func() {
				emptyContent, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"upload-stemcell",
					"--stemcell", emptyContent.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 10*time.Second).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say("failed to upload stemcell: file provided has no content"))
			})
		})

		When("the content cannot be read", func() {
			It("returns an error", func() {

				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"upload-stemcell",
					"--stemcell", "/unknown/path/whatever",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 10*time.Second).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say(`no such file or directory`))
			})
		})

		When("the server returns EOF during upload", func() {
			var uploadCallCount int

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/stemcells"),
						http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
							uploadCallCount++

							server.CloseClientConnections()
							time.Sleep(1 * time.Second)
							return
						}),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/stemcells"),
						http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
							uploadCallCount++

							err := req.ParseMultipartForm(100)
							Expect(err).ToNot(HaveOccurred())

							stemcellName = req.MultipartForm.File["stemcell[file]"][0].Filename

							_, err = w.Write([]byte(`{}`))
							Expect(err).ToNot(HaveOccurred())
						}),
					),
				)
			})

			It("retries the upload", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"upload-stemcell",
					"--stemcell", createStemcell("stemcell.tgz"),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(0))

				Expect(uploadCallCount).To(Equal(2))
			})
		})
	})
})
