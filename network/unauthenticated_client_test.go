package network_test

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"

	"github.com/pivotal-cf/om/network"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
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

		Context("failure cases", func() {
			Context("when the target url cannot be parsed", func() {
				It("returns an error", func() {
					client := network.NewUnauthenticatedClient("%%%", false, time.Duration(30)*time.Second)
					_, err := client.Do(&http.Request{})
					Expect(err).To(MatchError("could not parse target url: parse %%%: invalid URL escape \"%%%\""))
				})
			})
		})
	})

	Describe("RoundTrip", func() {
		It("makes a single http transaction without any authentication", func() {
			var requestDump []byte
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				var err error
				requestDump, err = httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())

				w.Header().Set("Location", "/redirect")
				w.WriteHeader(http.StatusFound)
				w.Write([]byte("response"))
			}))

			client := network.NewUnauthenticatedClient(server.URL, true, time.Duration(30)*time.Second)

			request, err := http.NewRequest("GET", "/path?query", strings.NewReader("request"))
			Expect(err).NotTo(HaveOccurred())

			response, err := client.RoundTrip(request)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).NotTo(BeNil())
			Expect(response.StatusCode).To(Equal(http.StatusFound))
			Expect(response.Header.Get("Location")).To(Equal("/redirect"))

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

		Context("failure cases", func() {
			Context("when the target url cannot be parsed", func() {
				It("returns an error", func() {
					client := network.NewUnauthenticatedClient("%%%", false, time.Duration(30)*time.Second)
					_, err := client.RoundTrip(&http.Request{})
					Expect(err).To(MatchError("could not parse target url: parse %%%: invalid URL escape \"%%%\""))
				})
			})
		})
	})
})
