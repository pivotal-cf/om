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
			default:
				authHeader = req.Header.Get("Authorization")

				w.WriteHeader(http.StatusNoContent)
				w.Write([]byte("response"))
			}
		}))
	})

	Describe("Do", func() {
		It("makes a request with authentication", func() {
			client, err := network.NewOAuthClient(server.URL, "opsman-username", "opsman-password", true, time.Duration(30)*time.Second)
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
				"client_id":  []string{"opsman"},
				"grant_type": []string{"password"},
				"username":   []string{"opsman-username"},
				"password":   []string{"opsman-password"},
			}))
		})

		Context("when insecureSkipVerify is configured", func() {
			Context("when it is set to false", func() {
				It("throws an error for invalid certificates", func() {
					client, err := network.NewOAuthClient(server.URL, "opsman-username", "opsman-password", false, time.Duration(30)*time.Second)
					Expect(err).NotTo(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).NotTo(HaveOccurred())

					_, err = client.Do(req)
					Expect(err.Error()).To(HaveSuffix("certificate signed by unknown authority"))
				})
			})

			Context("when it is set to true", func() {
				It("does not verify certificates", func() {
					client, err := network.NewOAuthClient(server.URL, "opsman-username", "opsman-password", true, time.Duration(30)*time.Second)
					Expect(err).NotTo(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).NotTo(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when an error occurs", func() {
			Context("when the initial token cannot be retrieved", func() {
				It("returns an error", func() {
					client, err := network.NewOAuthClient("%%%", "username", "password", false, time.Duration(30)*time.Second)
					Expect(err).NotTo(HaveOccurred())

					req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
					Expect(err).NotTo(HaveOccurred())

					_, err = client.Do(req)
					Expect(err).To(MatchError("token could not be retrieved from target url: parse %%%/uaa/oauth/token: invalid URL escape \"%%%\""))
				})
			})
		})
	})

	Describe("RoundTrip", func() {
		It("makes a request with authentication", func() {
			client, err := network.NewOAuthClient(server.URL, "opsman-username", "opsman-password", true, time.Duration(30)*time.Second)
			Expect(err).NotTo(HaveOccurred())

			req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
			Expect(err).NotTo(HaveOccurred())

			resp, err := client.RoundTrip(req)
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
				"client_id":  []string{"opsman"},
				"grant_type": []string{"password"},
				"username":   []string{"opsman-username"},
				"password":   []string{"opsman-password"},
			}))
		})
	})

	Context("when an error occurs", func() {
		Context("when the initial token cannot be retrieved", func() {
			It("returns an error", func() {
				client, err := network.NewOAuthClient("%%%", "username", "password", false, time.Duration(30)*time.Second)
				Expect(err).NotTo(HaveOccurred())

				req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
				Expect(err).NotTo(HaveOccurred())

				_, err = client.RoundTrip(req)
				Expect(err).To(MatchError("token could not be retrieved from target url: parse %%%/uaa/oauth/token: invalid URL escape \"%%%\""))
			})
			It("retries", func() {
				client, err := network.NewOAuthClient("http://240.255.255.255:12345", "opsman-username", "opsman-password", true, time.Duration(2)*time.Second)
				Expect(err).NotTo(HaveOccurred())

				var req http.Request
				start := time.Now()

				_, err = client.Do(&req)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("token could not be retrieved from target url: Post http://240.255.255.255:12345/uaa/oauth/token: dial tcp 240.255.255.255:12345: i/o timeout"))

				duration := time.Now().Sub(start)
				Expect(duration).To(BeNumerically(">", 90 * time.Second))
			})
		})
	})
})
