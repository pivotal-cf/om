package network_test

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/pivotal-cf/om/network"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnauthenticatedClient", func() {
	Describe("Do", func() {
		It("makes requests without any authentication", func() {
			var requestDump []byte
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				var err error
				requestDump, err = httputil.DumpRequest(req, true)
				Expect(err).ToNot(HaveOccurred())

				w.WriteHeader(http.StatusTeapot)
				_, err = w.Write([]byte("response"))
				Expect(err).ToNot(HaveOccurred())
			}))
			server.Config.ErrorLog = log.New(GinkgoWriter, "", 0)

			client, _ := network.NewUnauthenticatedClient(server.URL, true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)

			request, err := http.NewRequest("GET", "/path?query", strings.NewReader("request"))
			Expect(err).ToNot(HaveOccurred())

			response, err := client.Do(request)
			Expect(err).ToNot(HaveOccurred())

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusTeapot))

			body, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(Equal("response"))

			request, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(requestDump)))
			Expect(err).ToNot(HaveOccurred())

			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.String()).To(Equal("/path?query"))

			body, err = ioutil.ReadAll(request.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(Equal("request"))
		})

		When("passing a url with no scheme", func() {
			It("defaults to HTTPS", func() {
				server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusTeapot)
					_, err := w.Write([]byte("response"))
					Expect(err).ToNot(HaveOccurred())
				}))
				server.Config.ErrorLog = log.New(GinkgoWriter, "", 0)

				noScheme, err := url.Parse(server.URL)
				Expect(err).ToNot(HaveOccurred())

				noScheme.Scheme = ""
				finalURL := strings.Replace(noScheme.String(), "//", "", 1)

				client, _ := network.NewUnauthenticatedClient(finalURL, true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
				Expect(err).ToNot(HaveOccurred())

				request, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
				Expect(err).ToNot(HaveOccurred())

				response, err := client.Do(request)
				Expect(err).ToNot(HaveOccurred())

				Expect(response).ToNot(BeNil())
				Expect(response.StatusCode).To(Equal(http.StatusTeapot))
			})
		})

		When("supporting a ca cert", func() {
			var server *httptest.Server
			BeforeEach(func() {
				server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				server.Config.ErrorLog = log.New(GinkgoWriter, "", 0)
			})

			It("loads from a string", func() {
				cert, err := x509.ParseCertificate(server.TLS.Certificates[0].Certificate[0])
				Expect(err).ToNot(HaveOccurred())
				pemCert := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}))

				client, err := network.NewUnauthenticatedClient(server.URL, false, pemCert, time.Duration(5)*time.Second, time.Duration(30)*time.Second)
				Expect(err).ToNot(HaveOccurred())

				request, err := http.NewRequest("GET", "/path?query", strings.NewReader("request"))
				Expect(err).ToNot(HaveOccurred())

				_, err = client.Do(request)
				Expect(err).ToNot(HaveOccurred())
			})

			It("loads from a file", func() {
				cert, err := x509.ParseCertificate(server.TLS.Certificates[0].Certificate[0])
				Expect(err).ToNot(HaveOccurred())
				pemCert := writeFile(string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})))

				client, err := network.NewUnauthenticatedClient(server.URL, false, pemCert, time.Duration(5)*time.Second, time.Duration(30)*time.Second)
				Expect(err).ToNot(HaveOccurred())

				request, err := http.NewRequest("GET", "/path?query", strings.NewReader("request"))
				Expect(err).ToNot(HaveOccurred())

				_, err = client.Do(request)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		It("enforces minimum TLS version 1.2", func() {
			nonTLS12Server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
			nonTLS12Server.TLS.MaxVersion = tls.VersionTLS11
			nonTLS12Server.Config.ErrorLog = log.New(GinkgoWriter, "", 0)
			defer nonTLS12Server.Close()

			client, _ := network.NewUnauthenticatedClient(nonTLS12Server.URL, true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)

			req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
			Expect(err).ToNot(HaveOccurred())

			_, err = client.Do(req)
			Expect(err).To(MatchError(ContainSubstring("protocol version not supported")))
		})

		Context("failure cases", func() {
			When("the target url is empty", func() {
				It("returns an error", func() {
					client, _ := network.NewUnauthenticatedClient("", false, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
					_, err := client.Do(&http.Request{})
					Expect(err).To(MatchError("target flag is required, run `om help` for more info"))
				})
			})
		})
	})
})
