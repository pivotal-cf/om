package network_test

import (
	"bufio"
	"bytes"
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
		receivedCookies []*http.Cookie
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
				Expect(err).NotTo(HaveOccurred())

				w.Header().Set("Content-Type", "application/json")

				w.Write([]byte(`{
					"access_token": "some-opsman-token",
					"token_type": "bearer",
					"expires_in": 3600
					}`))
			case "/some/path":
				authHeader = req.Header.Get("Authorization")

				http.SetCookie(w, &http.Cookie{
					Name:  "somecookie",
					Value: "somevalue",
				})

				w.WriteHeader(http.StatusNoContent)
				w.Write([]byte("response"))
			default:
				receivedCookies = req.Cookies()
			}
		}))
	})

	Describe("Do", func() {
		It("makes a request with authentication", func() {
			client, err := network.NewOAuthClient(server.URL, "opsman-username", "opsman-password", "", "", true, false, time.Duration(30)*time.Second)
			Expect(err).NotTo(HaveOccurred())

			Expect(callCount).To(Equal(0))

			req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
			Expect(err).NotTo(HaveOccurred())

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			Expect(authHeader).To(Equal("Bearer some-opsman-token"))

			req, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(receivedRequest)))
			Expect(req.Method).To(Equal("POST"))
			Expect(req.URL.Path).To(Equal("/uaa/oauth/token"))

			username, password, ok := req.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("opsman"))
			Expect(password).To(BeEmpty())

			err = req.ParseForm()
			Expect(err).NotTo(HaveOccurred())
			Expect(req.Form).To(Equal(url.Values{
				"grant_type": []string{"password"},
				"username":   []string{"opsman-username"},
				"password":   []string{"opsman-password"},
			}))
		})

		It("makes a request with client credentials", func() {
			client, err := network.NewOAuthClient(server.URL, "", "", "client_id", "client_secret", true, false, time.Duration(30)*time.Second)
			Expect(err).NotTo(HaveOccurred())

			Expect(callCount).To(Equal(0))

			req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
			Expect(err).NotTo(HaveOccurred())

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			Expect(authHeader).To(Equal("Bearer some-opsman-token"))

			req, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(receivedRequest)))
			Expect(req.Method).To(Equal("POST"))
			Expect(req.URL.Path).To(Equal("/uaa/oauth/token"))

			username, password, ok := req.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("client_id"))
			Expect(password).To(Equal("client_secret"))

			err = req.ParseForm()
			Expect(err).NotTo(HaveOccurred())
			Expect(req.Form).To(Equal(url.Values{
				"grant_type": []string{"client_credentials"},
			}))
		})

		Context("when passing a url with no scheme", func() {
			It("defaults to HTTPS", func() {
				noScheme, err := url.Parse(server.URL)
				Expect(err).NotTo(HaveOccurred())

				noScheme.Scheme = ""
				finalURL := noScheme.String()

				client, err := network.NewOAuthClient(finalURL, "opsman-username", "opsman-password", "", "", true, false, time.Duration(30)*time.Second)
				Expect(err).NotTo(HaveOccurred())

				req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
				Expect(err).NotTo(HaveOccurred())

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
			})
		})

		Context("when insecureSkipVerify is configured", func() {
			Context("when it is set to false", func() {
				It("throws an error for invalid certificates", func() {
					client, err := network.NewOAuthClient(server.URL, "opsman-username", "opsman-password", "", "", false, false, time.Duration(30)*time.Second)
					Expect(err).NotTo(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).NotTo(HaveOccurred())

					_, err = client.Do(req)
					Expect(err.Error()).To(HaveSuffix("certificate signed by unknown authority"))
				})
			})

			Context("when it is set to true", func() {
				It("does not verify certificates", func() {
					client, err := network.NewOAuthClient(server.URL, "opsman-username", "opsman-password", "", "", true, false, time.Duration(30)*time.Second)
					Expect(err).NotTo(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).NotTo(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when includeCookies is configured", func() {
			Context("when it is set to true", func() {
				It("has a cookie jar", func() {
					client, err := network.NewOAuthClient(server.URL, "opsman-username", "opsman-password", "", "", true, true, time.Duration(30)*time.Second)
					Expect(err).NotTo(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).NotTo(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).NotTo(HaveOccurred())

					req, err = http.NewRequest("GET", "/some/different/path", strings.NewReader("request-body"))
					Expect(err).NotTo(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).NotTo(HaveOccurred())

					Expect(receivedCookies).To(HaveLen(1))
					Expect(receivedCookies[0].Name).To(Equal("somecookie"))
				})
			})

			Context("when it is false", func() {
				It("does not collect any of the cookies", func() {
					client, err := network.NewOAuthClient(server.URL, "opsman-username", "opsman-password", "", "", true, false, time.Duration(30)*time.Second)
					Expect(err).NotTo(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).NotTo(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).NotTo(HaveOccurred())

					req, err = http.NewRequest("GET", "/some/different/path", strings.NewReader("request-body"))
					Expect(err).NotTo(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).NotTo(HaveOccurred())

					Expect(receivedCookies).To(HaveLen(0))
				})
			})
		})

		Context("when an error occurs", func() {
			Context("when the initial token cannot be retrieved", func() {
				var badServer *httptest.Server

				BeforeEach(func() {
					badServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						w.WriteHeader(http.StatusInternalServerError)
					}))
				})

				It("returns an error", func() {
					client, err := network.NewOAuthClient(badServer.URL, "username", "password", "", "", true, false, time.Duration(30)*time.Second)
					Expect(err).NotTo(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).NotTo(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).To(MatchError(ContainSubstring("token could not be retrieved from target url: oauth2: cannot fetch token: 500")))
				})
			})

			Context("when the target url is empty", func() {
				It("returns an error", func() {
					client, err := network.NewOAuthClient("", "username", "password", "", "", false, false, time.Duration(30)*time.Second)
					Expect(err).NotTo(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).NotTo(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).To(MatchError(ContainSubstring("")))
				})
			})
		})
	})
})
