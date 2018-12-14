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

var _ = Describe("stage-product command", func() {
	var (
		stageRequest       string
		stageRequestMethod string
		server             *httptest.Server
	)

	Context("when the same type of product is not already deployed", func() {
		BeforeEach(func() {
			server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				var responseString string
				w.Header().Set("Content-Type", "application/json")

				switch req.URL.Path {
				case "/api/v0/installations":
					w.Write([]byte(`{"installations": []}`))
				case "/uaa/oauth/token":
					responseString = `{
						"access_token": "some-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`
				case "/api/v0/available_products":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					responseString = `[{
						"name": "cf",
						"product_version": "1.8.7-build.3"
					}]`
				case "/api/v0/deployed/products":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					responseString = `[{
						"type": "bosh",
						"installation_name": "bosh-some-other-guid",
						"guid": "bosh-some-other-guid"
					}]`
				case "/api/v0/staged/products":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					if req.Method == "GET" {
						responseString = `[]`
					} else {
						responseString = `{}`
						stageRequestMethod = req.Method
					}
					reqBody, err := ioutil.ReadAll(req.Body)
					Expect(err).NotTo(HaveOccurred())
					stageRequest = string(reqBody)
				case "/api/v0/diagnostic_report":
					responseString = `{}`
				default:
					out, err := httputil.DumpRequest(req, true)
					Expect(err).NotTo(HaveOccurred())
					Fail(fmt.Sprintf("unexpected request: %s", out))
				}

				w.Write([]byte(responseString))
			}))
		})

		AfterEach(func() {
			server.Close()
		})

		It("successfully stages a product to the Ops Manager", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"stage-product",
				"--product-name", "cf",
				"--product-version", "1.8.7-build.3",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("staging cf"))
			Eventually(session.Out).Should(gbytes.Say("finished staging"))

			Expect(stageRequestMethod).To(Equal("POST"))
			Expect(stageRequest).To(MatchJSON(`{
					"name": "cf",
					"product_version": "1.8.7-build.3"
			}`))
		})
	})

	Context("when the same type of product is already deployed", func() {
		BeforeEach(func() {
			server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				var responseString string
				w.Header().Set("Content-Type", "application/json")

				switch req.URL.Path {
				case "/api/v0/installations":
					w.Write([]byte(`{"installations": []}`))
				case "/uaa/oauth/token":
					responseString = `{
						"access_token": "some-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`
				case "/api/v0/available_products":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					responseString = `[{
						"name": "cf",
						"product_version": "1.8.7-build.3"
					},
					{
						"name": "cf",
						"product_version": "1.8.5-build.1"
					}]`
				case "/api/v0/deployed/products":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					responseString = `[{
						"type": "cf",
						"installation_name": "cf-some-guid",
						"guid": "cf-some-guid"
					},
					{
						"type": "bosh",
						"installation_name": "bosh-some-other-guid",
						"guid": "bosh-some-other-guid"
					}]`
				case "/api/v0/staged/products":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					if req.Method == "GET" {
						responseString = `[]`
					}
					reqBody, err := ioutil.ReadAll(req.Body)
					Expect(err).NotTo(HaveOccurred())
					stageRequest = string(reqBody)
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
				case "/api/v0/diagnostic_report":
					responseString = `{}`
				default:
					out, err := httputil.DumpRequest(req, true)
					Expect(err).NotTo(HaveOccurred())
					Fail(fmt.Sprintf("unexpected request: %s", out))
				}

				w.Write([]byte(responseString))
			}))
		})

		AfterEach(func() {
			server.Close()
		})

		It("successfully stages a product to the Ops Manager", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"stage-product",
				"--product-name", "cf",
				"--product-version", "1.8.7-build.3",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("staging cf"))
			Eventually(session.Out).Should(gbytes.Say("finished staging"))

			Expect(stageRequestMethod).To(Equal("PUT"))
			Expect(stageRequest).To(MatchJSON(`{
					"to_version": "1.8.7-build.3"
			}`))
		})
	})

	Context("when the same type of product is already staged", func() {
		BeforeEach(func() {
			server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				var responseString string
				w.Header().Set("Content-Type", "application/json")

				switch req.URL.Path {
				case "/api/v0/installations":
					w.Write([]byte(`{"installations": []}`))
				case "/uaa/oauth/token":
					responseString = `{
						"access_token": "some-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`
				case "/api/v0/available_products":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					responseString = `[{
						"name": "cf",
						"product_version": "1.8.7-build.3"
					},
					{
						"name": "cf",
						"product_version": "1.8.5-build.1"
					}]`
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
				case "/api/v0/deployed/products":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					responseString = `[]`
					reqBody, err := ioutil.ReadAll(req.Body)
					Expect(err).NotTo(HaveOccurred())
					stageRequest = string(reqBody)
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
				case "/api/v0/diagnostic_report":
					responseString = `{}`
				default:
					out, err := httputil.DumpRequest(req, true)
					Expect(err).NotTo(HaveOccurred())
					Fail(fmt.Sprintf("unexpected request: %s", out))
				}

				w.Write([]byte(responseString))
			}))
		})

		AfterEach(func() {
			server.Close()
		})

		It("successfully stages a product to the Ops Manager", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"stage-product",
				"--product-name", "cf",
				"--product-version", "1.8.7-build.3",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("staging cf"))
			Eventually(session.Out).Should(gbytes.Say("finished staging"))

			Expect(stageRequestMethod).To(Equal("PUT"))
			Expect(stageRequest).To(MatchJSON(`{
					"to_version": "1.8.7-build.3"
			}`))
		})
	})

	Context("when an error occurs", func() {
		BeforeEach(func() {
			server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				var responseString string
				w.Header().Set("Content-Type", "application/json")

				switch req.URL.Path {
				case "/api/v0/installations":
					w.Write([]byte(`{"installations": []}`))
				case "/uaa/oauth/token":
					responseString = `{
						"access_token": "some-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`
				case "/api/v0/available_products":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					responseString = `[{
						"name": "cf",
						"product_version": "1.8.7-build.3"
					}]`
				case "/api/v0/deployed/products":
					auth := req.Header.Get("Authorization")
					if auth != "Bearer some-opsman-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					responseString = `[{
						"type": "bosh",
						"installation_name": "bosh-some-other-guid",
						"guid": "bosh-some-other-guid"
					}]`
				case "/api/v0/staged/products":
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
				case "/api/v0/diagnostic_report":
					responseString = `{}`
				default:
					out, err := httputil.DumpRequest(req, true)
					Expect(err).NotTo(HaveOccurred())
					Fail(fmt.Sprintf("unexpected request: %s", out))
				}

				w.Write([]byte(responseString))
			}))
		})

		AfterEach(func() {
			server.Close()
		})

		Context("when the product is not available", func() {
			It("returns an error", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL,
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"stage-product",
					"--product-name", "bosh",
					"--product-version", "2.0",
				)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say("cannot find product bosh 2.0"))
			})
		})
	})
})
