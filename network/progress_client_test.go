package network_test

import (
	"errors"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/network"
	"github.com/pivotal-cf/om/network/fakes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProgressClient", func() {
	var (
		client         *fakes.HttpClient
		progressClient network.ProgressClient
		buffer         *gbytes.Buffer
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		buffer = gbytes.NewBuffer()

		progressClient = network.NewProgressClient(client, buffer)
	})

	Describe("Do", func() {
		It("makes a request to upload the product to the Ops Manager", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				io.Copy(ioutil.Discard, req.Body)
				req.Body.Close()
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
				}, nil
			}

			req, err := http.NewRequest("POST", "/some/endpoint", strings.NewReader("some content"))
			Expect(err).ToNot(HaveOccurred())

			resp, err := progressClient.Do(req)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			rawRespBody, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(rawRespBody)).To(Equal(`{}`))

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("POST"))
			Expect(request.URL.Path).To(Equal("/some/endpoint"))

			Eventually(buffer).Should(gbytes.Say("===] 100.00%"))
		})

		It("makes a request to download the product to the Ops Manager", func() {
			client.DoReturns(&http.Response{
				StatusCode:    http.StatusOK,
				ContentLength: int64(len([]byte("fake-server-response"))),
				Body:          ioutil.NopCloser(strings.NewReader("fake-server-response")),
			}, nil)

			req, err := http.NewRequest("GET", "/some/endpoint", nil)
			Expect(err).ToNot(HaveOccurred())

			resp, err := progressClient.Do(req)
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			rawRespBody, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(rawRespBody)).To(Equal("fake-server-response"))
			Eventually(buffer).Should(gbytes.Say("===]"))

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/some/endpoint"))
		})

		When("an error occurs", func() {
			When("the client errors performing the request", func() {
				It("returns an error", func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						req.Body.Close()
						return &http.Response{}, errors.New("some client error")
					}

					req, err := http.NewRequest("POST", "/some/endpoint", strings.NewReader("some content"))
					Expect(err).ToNot(HaveOccurred())

					_, err = progressClient.Do(req)
					Expect(err).To(MatchError("some client error"))
				})
			})

			When("server responds with timeout error before upload has finished", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusRequestTimeout,
						Body:       ioutil.NopCloser(strings.NewReader(`something from nginx probably xml`)),
					}, nil)

					var req *http.Request
					req, err := http.NewRequest("POST", "/some/endpoint", strings.NewReader("some content"))
					Expect(err).ToNot(HaveOccurred())

					var (
						resp *http.Response
						done = make(chan bool)
					)
					go func() {
						resp, err = progressClient.Do(req)
						close(done)
					}()

					Eventually(done, 3).Should(BeClosed())
					Expect(resp.StatusCode).To(Equal(http.StatusRequestTimeout))
				})
			})
		})
	})
})
