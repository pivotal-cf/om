package network_test

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
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

			client, _ := network.NewUnauthenticatedClient(serverURL(server), true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)

			request, err := http.NewRequest("GET", "/path?query", strings.NewReader("request"))
			Expect(err).ToNot(HaveOccurred())

			response, err := client.Do(request)
			Expect(err).ToNot(HaveOccurred())

			Expect(response).ToNot(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusTeapot))

			body, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(Equal("response"))

			request, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(requestDump)))
			Expect(err).ToNot(HaveOccurred())

			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.String()).To(Equal("/path?query"))

			body, err = io.ReadAll(request.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(Equal("request"))
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

				client, err := network.NewUnauthenticatedClient(serverURL(server), false, pemCert, time.Duration(5)*time.Second, time.Duration(30)*time.Second)
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

				client, err := network.NewUnauthenticatedClient(serverURL(server), false, pemCert, time.Duration(5)*time.Second, time.Duration(30)*time.Second)
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

			client, _ := network.NewUnauthenticatedClient(serverURL(nonTLS12Server), true, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)

			req, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
			Expect(err).ToNot(HaveOccurred())

			_, err = client.Do(req)
			Expect(err).To(MatchError(ContainSubstring("protocol version not supported")))
		})

		Context("failure cases", func() {
			When("the target url is nil", func() {
				It("returns an error", func() {
					_, err := network.NewUnauthenticatedClient(nil, false, "", time.Duration(5)*time.Second, time.Duration(30)*time.Second)
					Expect(err).To(MatchError("expected a non-nil target"))
				})
			})
		})
	})
})

func serverURL(s *httptest.Server) *url.URL {
	u, _ := url.Parse(s.URL)
	return u
}
