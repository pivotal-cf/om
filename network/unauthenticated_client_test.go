package network_test

import (
	"bufio"
	"bytes"
	"io/ioutil"
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
				Expect(err).NotTo(HaveOccurred())

				w.WriteHeader(http.StatusTeapot)
				w.Write([]byte("response"))
			}))

			client := network.NewUnauthenticatedClient(server.URL, true, time.Duration(30)*time.Second)

			request, err := http.NewRequest("GET", "/path?query", strings.NewReader("request"))
			Expect(err).NotTo(HaveOccurred())

			response, err := client.Do(request)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).NotTo(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusTeapot))

			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(Equal("response"))

			request, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(requestDump)))
			Expect(err).NotTo(HaveOccurred())

			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.String()).To(Equal("/path?query"))

			body, err = ioutil.ReadAll(request.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(Equal("request"))
		})

		Context("when passing a url with no scheme", func() {
			It("defaults to HTTPS", func() {
				server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusTeapot)
					w.Write([]byte("response"))
				}))

				noScheme, err := url.Parse(server.URL)
				Expect(err).NotTo(HaveOccurred())

				noScheme.Scheme = ""
				finalURL := strings.Replace(noScheme.String(), "//", "", 1)

				client := network.NewUnauthenticatedClient(finalURL, true, time.Duration(30)*time.Second)
				Expect(err).NotTo(HaveOccurred())

				request, err := http.NewRequest("GET", "/some/path", strings.NewReader("request-body"))
				Expect(err).NotTo(HaveOccurred())

				response, err := client.Do(request)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).NotTo(BeNil())
				Expect(response.StatusCode).To(Equal(http.StatusTeapot))
			})
		})

		Context("failure cases", func() {
			Context("when the target url cannot be parsed", func() {
				It("returns an error", func() {
					client := network.NewUnauthenticatedClient("%%%", false, time.Duration(30)*time.Second)
					_, err := client.Do(&http.Request{})
					Expect(err).To(MatchError("could not parse target url: parse //%%%: invalid URL escape \"%%%\""))
				})
			})

			Context("when the target url is empty", func() {
				It("returns an error", func() {
					client := network.NewUnauthenticatedClient("", false, time.Duration(30)*time.Second)
					_, err := client.Do(&http.Request{})
					Expect(err).To(MatchError("target flag is required. Run `om help` for more info."))
				})
			})
		})
	})
})
