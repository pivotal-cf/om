package network_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/onsi/gomega/ghttp"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/pivotal-cf/om/network"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OAuthClient", func() {
	var (
		server     *ghttp.Server
	)

	BeforeEach(func() {
		server = ghttp.NewTLSServer()
	})

	Describe("Do", func() {
		It("makes a request with authentication", func() {
			server.RouteToHandler("POST", "/uaa/oauth/token", ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth("opsman", ""),
				ghttp.VerifyForm(url.Values{
					"grant_type":   []string{"password"},
					"username":     []string{"opsman-username"},
					"password":     []string{"opsman-password"},
					"token_format": []string{"opaque"},
					"client_id":    []string{"opsman"},
				}),
				ghttp.RespondWith(http.StatusOK, `{
					"access_token": "some-opsman-token",
					"token_type": "bearer",
					"expires_in": 3600
					}`, http.Header{
					"Content-Type": []string{"application/json"},
				}),
			))
			server.AppendHandlers(ghttp.RespondWith(http.StatusOK, nil))

			client, err := network.NewOAuthClient(server.URL(), "opsman-username", "opsman-password", "", "", true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
			Expect(err).ToNot(HaveOccurred())

			req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
			Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("makes a request with client credentials", func() {
			server.RouteToHandler("POST", "/uaa/oauth/token", ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth("client_id", "client_secret"),
				ghttp.VerifyForm(url.Values{
					"grant_type":   []string{"client_credentials"},
					"token_format": []string{"opaque"},
				}),
				ghttp.RespondWith(http.StatusOK, `{
					"access_token": "some-opsman-token",
					"token_type": "bearer",
					"expires_in": 3600
					}`, http.Header{
					"Content-Type": []string{"application/json"},
				}),
			))
			server.AppendHandlers(ghttp.RespondWith(http.StatusOK, nil))

			client, err := network.NewOAuthClient(server.URL(), "", "", "client_id", "client_secret", true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
			Expect(err).ToNot(HaveOccurred())

			req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
			Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(req)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
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
				setupBasicOauth(server)

				noScheme, err := url.Parse(server.URL())
				Expect(err).ToNot(HaveOccurred())

				noScheme.Scheme = ""
				finalURL := noScheme.String()[2:] // removing leading "//"

				client, err := network.NewOAuthClient(finalURL, "opsman-username", "opsman-password", "", "", true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
				Expect(err).ToNot(HaveOccurred())

				req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
				Expect(err).ToNot(HaveOccurred())

				resp, err := client.Do(req)
				Expect(err).ToNot(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		When("insecureSkipVerify is configured", func() {
			When("it is set to false", func() {
				It("throws an error for invalid certificates", func() {
					client, err := network.NewOAuthClient(server.URL(), "opsman-username", "opsman-password", "", "", false, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
					Expect(err).ToNot(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).ToNot(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).To(MatchError(ContainSubstring("certificate signed by unknown authority")))
				})
			})

			When("it is set to true", func() {
				It("does not verify certificates", func() {
					setupBasicOauth(server)

					client, err := network.NewOAuthClient(server.URL(), "opsman-username", "opsman-password", "", "", true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
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
				setupBasicOauth(server)

				cert, err := x509.ParseCertificate(server.HTTPTestServer.TLS.Certificates[0].Certificate[0])
				Expect(err).ToNot(HaveOccurred())
				pemCert := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}))

				client, err := network.NewOAuthClient(
					server.URL(),
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
				setupBasicOauth(server)

				cert, err := x509.ParseCertificate(server.HTTPTestServer.TLS.Certificates[0].Certificate[0])
				Expect(err).ToNot(HaveOccurred())
				pemCert := writeFile(string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})))

				client, err := network.NewOAuthClient(
					server.URL(),
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

func setupBasicOauth(server *ghttp.Server) {
	server.RouteToHandler("POST", "/uaa/oauth/token", ghttp.CombineHandlers(
		ghttp.VerifyBasicAuth("opsman", ""),
		ghttp.VerifyForm(url.Values{
			"grant_type":   []string{"password"},
			"username":     []string{"opsman-username"},
			"password":     []string{"opsman-password"},
			"token_format": []string{"opaque"},
			"client_id":    []string{"opsman"},
		}),
		ghttp.RespondWith(http.StatusOK, `{
					"access_token": "some-opsman-token",
					"token_type": "bearer",
					"expires_in": 3600
					}`, http.Header{
			"Content-Type": []string{"application/json"},
		}),
	))
	server.AppendHandlers(ghttp.RespondWith(http.StatusOK, nil))
}
