package network_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pivotal-cf/om/network"
	"github.com/pivotal-cf/om/network/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProgressClient", func() {
	var (
		client         *fakes.HttpClient
		progressBar    *fakes.ProgressBar
		liveWriter     *fakes.LiveWriter
		progressClient network.ProgressClient
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		liveWriter = &fakes.LiveWriter{}
		progressBar = &fakes.ProgressBar{}

		progressClient = network.NewProgressClient(client, progressBar, liveWriter)
	})

	Describe("Do", func() {
		It("makes a request to upload the product to the Ops Manager", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("{}")),
			}, nil)

			progressBar.NewProxyReaderReturns(ioutil.NopCloser(strings.NewReader("some content")))

			req, err := http.NewRequest("POST", "/some/endpoint", strings.NewReader("some content"))
			Expect(err).NotTo(HaveOccurred())

			req = req.WithContext(context.WithValue(req.Context(), "polling-interval", time.Second))

			resp, err := progressClient.Do(req)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			rawRespBody, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(rawRespBody)).To(Equal("{}"))

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("POST"))
			Expect(request.URL.Path).To(Equal("/some/endpoint"))

			rawReqBody, err := ioutil.ReadAll(req.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(rawReqBody)).To(Equal("some content"))

			Expect(progressBar.SetTotal64CallCount()).To(Equal(1))
			Expect(progressBar.SetTotal64ArgsForCall(0)).To(Equal(int64(12)))

			Expect(progressBar.StartCallCount()).To(Equal(1))
			Expect(progressBar.FinishCallCount()).To(Equal(1))
		})

		It("logs while waiting for a response from the Ops Manager", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				_, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				time.Sleep(20 * time.Millisecond)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("some-installation")),
				}, nil
			}

			progressBar.NewProxyReaderReturns(ioutil.NopCloser(strings.NewReader("some-fake-installation")))

			req, err := http.NewRequest("POST", "/some/endpoint", strings.NewReader("some content"))
			Expect(err).NotTo(HaveOccurred())

			req = req.WithContext(context.WithValue(req.Context(), "polling-interval", 10*time.Millisecond))

			_, err = progressClient.Do(req)
			Expect(err).NotTo(HaveOccurred())

			By("starting the live log writer", func() {
				Expect(liveWriter.StartCallCount()).To(Equal(1))
			})

			By("writing to the live log writer", func() {
				Expect(liveWriter.WriteCallCount()).To(BeNumerically("~", 3, 1))
				Expect(string(liveWriter.WriteArgsForCall(0))).To(ContainSubstring("10ms elapsed"))
				Expect(string(liveWriter.WriteArgsForCall(1))).To(ContainSubstring("20ms elapsed"))
			})

			By("flushing the live log writer", func() {
				Expect(liveWriter.StopCallCount()).To(Equal(1))
			})
		})

		It("makes a request to download the product to the Ops Manager", func() {
			client.DoReturns(&http.Response{
				StatusCode:    http.StatusOK,
				ContentLength: int64(len([]byte("fake-server-response"))),
				Body:          ioutil.NopCloser(strings.NewReader("fake-server-response")),
			}, nil)

			progressBar.NewProxyReaderReturns(ioutil.NopCloser(strings.NewReader("fake-wrapper-response")))

			req, err := http.NewRequest("GET", "/some/endpoint", nil)
			Expect(err).NotTo(HaveOccurred())

			req = req.WithContext(context.WithValue(req.Context(), "polling-interval", time.Second))

			resp, err := progressClient.Do(req)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			rawRespBody, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(rawRespBody)).To(Equal("fake-wrapper-response"))

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/some/endpoint"))

			Expect(progressBar.SetTotal64CallCount()).To(Equal(1))
			Expect(progressBar.SetTotal64ArgsForCall(0)).To(Equal(int64(len([]byte("fake-server-response")))))

			Expect(progressBar.StartCallCount()).To(Equal(1))
			Expect(progressBar.FinishCallCount()).To(Equal(1))
		})

		Context("when the polling interval is greater than 1", func() {
			It("logs at the correct interval", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					ioutil.ReadAll(req.Body)
					defer req.Body.Close()

					time.Sleep(50 * time.Millisecond)

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("some-installation")),
					}, nil
				}

				progressBar.NewProxyReaderReturns(ioutil.NopCloser(strings.NewReader("some-fake-installation")))

				req, err := http.NewRequest("POST", "/some/endpoint", strings.NewReader("some content"))
				Expect(err).NotTo(HaveOccurred())

				req = req.WithContext(context.WithValue(req.Context(), "polling-interval", 20*time.Millisecond))

				_, err = progressClient.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(liveWriter.StartCallCount()).To(Equal(1))
				Expect(liveWriter.WriteCallCount()).To(BeNumerically("~", 2, 1))
				Expect(string(liveWriter.WriteArgsForCall(0))).To(ContainSubstring("20ms elapsed"))
				Expect(string(liveWriter.WriteArgsForCall(1))).To(ContainSubstring("40ms elapsed"))
				Expect(liveWriter.StopCallCount()).To(Equal(1))
			})
		})

		Context("when the polling interval is greater than the time it takes upload the product", func() {
			It("logs at the correct interval", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					ioutil.ReadAll(req.Body)
					defer req.Body.Close()

					time.Sleep(100 * time.Millisecond)

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("some-installation")),
					}, nil
				}

				progressBar.NewProxyReaderReturns(ioutil.NopCloser(strings.NewReader("some-fake-installation")))

				req, err := http.NewRequest("POST", "/some/endpoint", strings.NewReader("some content"))
				Expect(err).NotTo(HaveOccurred())

				_, err = progressClient.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(liveWriter.StartCallCount()).To(Equal(1))
				Expect(liveWriter.WriteCallCount()).To(Equal(1))
				Expect(liveWriter.StopCallCount()).To(Equal(1))
			})
		})

		Context("when an error occurs", func() {
			Context("when the client errors performing the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some client error"))

					req, err := http.NewRequest("POST", "/some/endpoint", strings.NewReader("some content"))
					Expect(err).NotTo(HaveOccurred())

					_, err = progressClient.Do(req)
					Expect(err).To(MatchError("some client error"))
				})
			})

			Context("when server responds with timeout error before upload has finished", func() {
				It("returns an error", func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusRequestTimeout,
							Body:       ioutil.NopCloser(strings.NewReader(`something from nginx probably xml`)),
						}, nil
					}

					var req *http.Request
					req, err := http.NewRequest("POST", "/some/endpoint", strings.NewReader("some content"))
					Expect(err).NotTo(HaveOccurred())

					var (
						resp *http.Response
						done = make(chan bool)
					)
					go func() {
						resp, err = progressClient.Do(req)
						close(done)
					}()

					Eventually(done).Should(BeClosed())
					Expect(resp.StatusCode).To(Equal(http.StatusRequestTimeout))
				})
			})
		})
	})
})
