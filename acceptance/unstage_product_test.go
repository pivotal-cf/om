package acceptance

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("unstage-product command", func() {
	var (
		stageRequest       string
		stageRequestMethod string
		server             *httptest.Server
	)

	Context("when the product is staged", func() {
		BeforeEach(func() {
			server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				var responseString string
				w.Header().Set("Content-Type", "application/json")

				switch req.URL.Path {
				case "/uaa/oauth/token":
					responseString = `{
						"access_token": "some-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`
				case "/api/v0/staged/products":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					if req.Method == "GET" {
						responseString = `[]`
						responseString = `[{
							"type": "cf",
							"guid": "cf-some-guid"
						},
						{
							"type": "bosh",
							"guid": "bosh-some-other-guid"
						}]`
					}
				case "/api/v0/staged/products/cf-some-guid":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					responseString = `{}`
					stageRequestMethod = req.Method
					reqBody, err := ioutil.ReadAll(req.Body)
					Expect(err).NotTo(HaveOccurred())
					stageRequest = string(reqBody)
				default:
					out, err := httputil.DumpRequest(req, true)
					Expect(err).NotTo(HaveOccurred())
					Fail(fmt.Sprintf("unexpected request: %s", out))
				}

				w.Write([]byte(responseString))
			}))
		})

		It("successfully unstages a product from the Ops Manager", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"unstage-product",
				"--product-name", "cf",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("unstaging cf"))
			Eventually(session.Out).Should(gbytes.Say("finished unstaging"))

			Expect(stageRequestMethod).To(Equal("DELETE"))
		})
	})

	Context("when the product is not staged", func() {
		BeforeEach(func() {
			server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				var responseString string
				w.Header().Set("Content-Type", "application/json")

				switch req.URL.Path {
				case "/uaa/oauth/token":
					responseString = `{
						"access_token": "some-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`
				case "/api/v0/staged/products":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					responseString = `[]`
					stageRequestMethod = req.Method
					reqBody, err := ioutil.ReadAll(req.Body)
					Expect(err).NotTo(HaveOccurred())
					stageRequest = string(reqBody)
				default:
					out, err := httputil.DumpRequest(req, true)
					Expect(err).NotTo(HaveOccurred())
					Fail(fmt.Sprintf("unexpected request: %s", out))
				}

				w.Write([]byte(responseString))
			}))
		})

		It("returns an error", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"unstage-product",
				"--product-name", "cf",
			)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("product is not staged: cf"))
		})
	})
})
