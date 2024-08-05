package acceptance

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/onsi/gomega/ghttp"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("upload-stemcell command", func() {
	var (
		stemcellName string
		server       *ghttp.Server
	)

	createStemcell := func(filename string) (string, string) {
		dir, err := os.MkdirTemp("", "")
		Expect(err).ToNot(HaveOccurred())

		err = os.MkdirAll(filepath.Join(dir, "stemcells"), 0777)
		Expect(err).ToNot(HaveOccurred())

		path := filepath.Join(dir, "stemcells", filename)
		err = os.WriteFile(path, []byte("content so validation does not fail"), 0777)
		Expect(err).ToNot(HaveOccurred())
		return path, dir
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
			filename, _ := createStemcell("stemcell.tgz")
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "pass",
				"--skip-ssl-validation",
				"upload-stemcell",
				"--stemcell", filename,
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, 10*time.Second).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("processing stemcell"))
			Eventually(session.Out).Should(gbytes.Say("beginning stemcell upload to Ops Manager"))
			Eventually(session.Out).Should(gbytes.Say("finished upload"))

			Expect(stemcellName).To(Equal("stemcell.tgz"))
		})

		When("a config file is provided", func() {
			It("successfully sends the stemcell to the Ops Manager", func() {
				filename, _ := createStemcell("stemcell.tgz")
				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--username", "some-username",
					"--password", "pass",
					"--skip-ssl-validation",
					"upload-stemcell",
					"--config", writeFile(fmt.Sprintf("{stemcell: %s, shasum: 33d5f6335e83364e11760878afad059fffd6a2729ae53691c87cc349a784de92}", filename)),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session, 10*time.Second).Should(gexec.Exit(0))
				Eventually(session.Out).Should(gbytes.Say("expected shasum matches stemcell shasum."))

				Expect(stemcellName).To(Equal("stemcell.tgz"))
			})

		})

		When("the stemcell name has the `download-product` prefix", func() {
			When("a relative path", func() {
				It("successfully sends the stemcell to the Ops Manager", func() {
					filename, dir := createStemcell("[ubuntu-xenial,97.88]stemcell.tgz")
					command := exec.Command(pathToMain,
						"--target", server.URL(),
						"--username", "some-username",
						"--password", "pass",
						"--skip-ssl-validation",
						"upload-stemcell",
						"--stemcell", "stemcells/[ubuntu-xenial,97.88]stemcell.tgz",
					)
					command.Dir = dir

					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())

					Eventually(session, 10*time.Second).Should(gexec.Exit(0))
					Eventually(session.Out).Should(gbytes.Say("processing stemcell"))
					Eventually(session.Out).Should(gbytes.Say("beginning stemcell upload to Ops Manager"))
					Eventually(session.Out).Should(gbytes.Say("finished upload"))

					Expect(stemcellName).To(Equal("stemcell.tgz"))
					Expect(filename).To(BeAnExistingFile())
				})
			})

			When("a absolute path", func() {
				It("successfully sends the stemcell to the Ops Manager", func() {
					filename, _ := createStemcell("[ubuntu-xenial,97.88]stemcell.tgz")
					command := exec.Command(pathToMain,
						"--target", server.URL(),
						"--username", "some-username",
						"--password", "pass",
						"--skip-ssl-validation",
						"upload-stemcell",
						"--stemcell", filename,
					)

					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())

					Eventually(session, 10*time.Second).Should(gexec.Exit(0))
					Eventually(session.Out).Should(gbytes.Say("processing stemcell"))
					Eventually(session.Out).Should(gbytes.Say("beginning stemcell upload to Ops Manager"))
					Eventually(session.Out).Should(gbytes.Say("finished upload"))

					Expect(stemcellName).To(Equal("stemcell.tgz"))
					Expect(filename).To(BeAnExistingFile())
				})
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
			filename, _ := createStemcell("stemcell.tgz")
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"upload-stemcell",
				"--stemcell", filename,
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

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
				emptyContent, err := os.CreateTemp("", "")
				Expect(err).ToNot(HaveOccurred())

				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"upload-stemcell",
					"--stemcell", emptyContent.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

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
				Expect(err).ToNot(HaveOccurred())

				Eventually(session, 10*time.Second).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say(`no such file or directory`))
			})
		})
	})
})
