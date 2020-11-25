package network_test

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/pivotal-cf/om/network"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OAuthClient", func() {
	var (
		receivedRequest []byte
		authHeader      string
		callCount       int
		server          *httptest.Server
	)

	BeforeEach(func() {
		callCount = 0
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/uaa/oauth/token":
				callCount++
				var err error
				receivedRequest, err = httputil.DumpRequest(req, true)
				Expect(err).ToNot(HaveOccurred())

				w.Header().Set("Content-Type", "application/json")

				_, err = w.Write([]byte(`{
					"access_token": "some-opsman-token",
					"token_type": "bearer",
					"expires_in": 3600
					}`))
				Expect(err).ToNot(HaveOccurred())
			case "/some/path":
				authHeader = req.Header.Get("Authorization")

				http.SetCookie(w, &http.Cookie{
					Name:  "somecookie",
					Value: "somevalue",
				})

				w.WriteHeader(http.StatusNoContent)
			}
		}))
		server.Config.ErrorLog = log.New(GinkgoWriter, "", 0)
	})

	Describe("Do", func() {
		It("makes a request with authentication", func() {
			client, err := network.NewOAuthClient(server.URL, "opsman-username", "opsman-password", "", "", true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
			Expect(err).ToNot(HaveOccurred())

			Expect(callCount).To(Equal(0))

			req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
			Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			Expect(authHeader).To(Equal("Bearer some-opsman-token"))

			req, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(receivedRequest)))
			Expect(err).ToNot(HaveOccurred())
			Expect(req.Method).To(Equal("POST"))
			Expect(req.URL.Path).To(Equal("/uaa/oauth/token"))

			username, password, ok := req.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("opsman"))
			Expect(password).To(BeEmpty())

			err = req.ParseForm()
			Expect(err).ToNot(HaveOccurred())
			Expect(req.Form).To(Equal(url.Values{
				"grant_type":   []string{"password"},
				"username":     []string{"opsman-username"},
				"password":     []string{"opsman-password"},
				"token_format": []string{"opaque"},
				"client_id":    []string{"opsman"},
			}))
		})

		It("makes a request with client credentials", func() {
			client, err := network.NewOAuthClient(server.URL, "", "", "client_id", "client_secret", true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
			Expect(err).ToNot(HaveOccurred())

			Expect(callCount).To(Equal(0))

			req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
			Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			Expect(authHeader).To(Equal("Bearer some-opsman-token"))

			req, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(receivedRequest)))
			Expect(err).ToNot(HaveOccurred())
			Expect(req.Method).To(Equal("POST"))
			Expect(req.URL.Path).To(Equal("/uaa/oauth/token"))

			username, password, ok := req.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("client_id"))
			Expect(password).To(Equal("client_secret"))

			err = req.ParseForm()
			Expect(err).ToNot(HaveOccurred())
			Expect(req.Form).To(Equal(url.Values{
				"grant_type":   []string{"client_credentials"},
				"token_format": []string{"opaque"},
			}))
		})

		It("enforces minimum TLS version 1.2", func() {
			nonTLS12Server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
			nonTLS12Server.TLS.MaxVersion = tls.VersionTLS11
			nonTLS12Server.Config.ErrorLog = log.New(GinkgoWriter, "", 0)
			defer nonTLS12Server.Close()

			client, err := network.NewOAuthClient(nonTLS12Server.URL, "", "", "client_id", "client_secret", true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
			Expect(err).ToNot(HaveOccurred())

			req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
			Expect(err).ToNot(HaveOccurred())

			_, err = client.Do(req)
			Expect(err).To(MatchError(ContainSubstring("protocol version not supported")))
		})

		When("passing a url with no scheme", func() {
			It("defaults to HTTPS", func() {
				noScheme, err := url.Parse(server.URL)
				Expect(err).ToNot(HaveOccurred())

				noScheme.Scheme = ""
 				finalURL := noScheme.String()[2:] // removing leading "//"

				client, err := network.NewOAuthClient(finalURL, "opsman-username", "opsman-password", "", "", true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
				Expect(err).ToNot(HaveOccurred())

				req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
				Expect(err).ToNot(HaveOccurred())

				resp, err := client.Do(req)
				Expect(err).ToNot(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
			})
		})

		When("insecureSkipVerify is configured", func() {
			When("it is set to false", func() {
				It("throws an error for invalid certificates", func() {
					client, err := network.NewOAuthClient(server.URL, "opsman-username", "opsman-password", "", "", false, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
					Expect(err).ToNot(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).ToNot(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).To(MatchError(ContainSubstring("certificate signed by unknown authority")))
				})
			})

			When("it is set to true", func() {
				It("does not verify certificates", func() {
					client, err := network.NewOAuthClient(server.URL, "opsman-username", "opsman-password", "", "", true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
					Expect(err).ToNot(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).ToNot(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		When("supporting a ca cert", func() {
			It("loads from a string", func() {
				cert, err := x509.ParseCertificate(server.TLS.Certificates[0].Certificate[0])
				Expect(err).ToNot(HaveOccurred())
				pemCert := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}))

				client, err := network.NewOAuthClient(
					server.URL,
					"opsman-username", "opsman-password",
					"", "",
					false,
					pemCert,
					time.Duration(5)*time.Second, time.Duration(30)*time.Second,
				)

				Expect(err).ToNot(HaveOccurred())

				req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
				Expect(err).ToNot(HaveOccurred())

				_, err = client.Do(req)
				Expect(err).ToNot(HaveOccurred())
			})

			It("loads from a file", func() {
				cert, err := x509.ParseCertificate(server.TLS.Certificates[0].Certificate[0])
				Expect(err).ToNot(HaveOccurred())
				pemCert := writeFile(string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})))

				client, err := network.NewOAuthClient(
					server.URL,
					"opsman-username", "opsman-password",
					"", "",
					false,
					pemCert,
					time.Duration(5)*time.Second, time.Duration(30)*time.Second,
				)

				Expect(err).ToNot(HaveOccurred())

				req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
				Expect(err).ToNot(HaveOccurred())

				_, err = client.Do(req)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("an error occurs", func() {
			When("the initial token cannot be retrieved", func() {
				var badServer *httptest.Server

				BeforeEach(func() {
					badServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						w.WriteHeader(http.StatusInternalServerError)
					}))
					badServer.Config.ErrorLog = log.New(GinkgoWriter, "", 0)
				})

				It("returns an error", func() {
					client, err := network.NewOAuthClient(badServer.URL, "username", "password", "", "", true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
					Expect(err).ToNot(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).ToNot(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).To(MatchError(ContainSubstring("token could not be retrieved from target url")))
				})
			})

			When("the target url is empty", func() {
				It("returns an error", func() {
					client, err := network.NewOAuthClient("", "username", "password", "", "", false, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
					Expect(err).ToNot(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).ToNot(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).To(MatchError(ContainSubstring("")))
				})
			})
		})
	})
})
