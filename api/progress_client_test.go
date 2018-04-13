package api_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("ProgressClient", func() {
	var (
		client         *fakes.HttpClient
		bar            *fakes.Progress
		liveWriter     *fakes.LiveWriter
		progressClient api.ProgressClient
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		liveWriter = &fakes.LiveWriter{}
		bar = &fakes.Progress{}

		progressClient = api.NewProgressClient(client, bar, liveWriter, 1)
	})

	Describe("Do", func() {
		It("makes a request to upload the product to the Ops Manager", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("{}")),
			}, nil)

			bar.NewBarReaderReturns(ioutil.NopCloser(strings.NewReader("some content")))

			req, err := http.NewRequest("POST", "/some/endpoint", strings.NewReader("some content"))
			Expect(err).NotTo(HaveOccurred())

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

			Expect(bar.SetTotalCallCount()).To(Equal(1))
			Expect(bar.SetTotalArgsForCall(0)).To(Equal(int64(12)))

			Expect(bar.KickoffCallCount()).To(Equal(1))
			Expect(bar.EndCallCount()).To(Equal(1))
		})

		It("logs while waiting for a response from the Ops Manager", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				_, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				time.Sleep(5 * time.Second)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("some-installation")),
				}, nil
			}

			bar.NewBarReaderReturns(ioutil.NopCloser(strings.NewReader("some-fake-installation")))

			req, err := http.NewRequest("POST", "/some/endpoint", strings.NewReader("some content"))
			Expect(err).NotTo(HaveOccurred())

			_, err = progressClient.Do(req)
			Expect(err).NotTo(HaveOccurred())

			By("starting the live log writer", func() {
				Expect(liveWriter.StartCallCount()).To(Equal(1))
			})

			By("writing to the live log writer", func() {
				Expect(liveWriter.WriteCallCount()).To(BeNumerically("~", 5, 1))
				for i := 0; i < liveWriter.WriteCallCount(); i++ {
					Expect(string(liveWriter.WriteArgsForCall(i))).To(ContainSubstring(fmt.Sprintf("%ds elapsed", i+1)))
				}
			})

			By("flushing the live log writer", func() {
				Expect(liveWriter.StopCallCount()).To(Equal(1))
			})
		})

		Context("when the polling interval is greater than 1", func() {
			BeforeEach(func() {
				progressClient = api.NewProgressClient(client, bar, liveWriter, 2)
			})

			It("logs at the correct interval", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					time.Sleep(5 * time.Second)

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("some-installation")),
					}, nil
				}

				bar.NewBarReaderReturns(ioutil.NopCloser(strings.NewReader("some-fake-installation")))

				req, err := http.NewRequest("POST", "/some/endpoint", strings.NewReader("some content"))
				Expect(err).NotTo(HaveOccurred())

				_, err = progressClient.Do(req)
				Expect(err).NotTo(HaveOccurred())

				By("starting the live log writer", func() {
					Expect(liveWriter.StartCallCount()).To(Equal(1))
				})

				By("writing to the live log writer", func() {
					Expect(liveWriter.WriteCallCount()).To(BeNumerically("~", 2, 1))

					for i := 0; i < liveWriter.WriteCallCount(); i++ {
						Expect(string(liveWriter.WriteArgsForCall(i))).To(ContainSubstring(fmt.Sprintf("%ds elapsed", 2*(i+1))))
					}
				})

				By("flushing the live log writer", func() {
					Expect(liveWriter.StopCallCount()).To(Equal(1))
				})
			})
		})

		Context("when the polling interval is greater than the time it takes upload the product", func() {
			It("logs at the correct interval", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					time.Sleep(100 * time.Millisecond)

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("some-installation")),
					}, nil
				}

				bar.NewBarReaderReturns(ioutil.NopCloser(strings.NewReader("some-fake-installation")))

				req, err := http.NewRequest("POST", "/some/endpoint", strings.NewReader("some content"))
				Expect(err).NotTo(HaveOccurred())

				_, err = progressClient.Do(req)
				Expect(err).NotTo(HaveOccurred())

				By("starting the live log writer", func() {
					Expect(liveWriter.StartCallCount()).To(Equal(1))
				})

				By("writing to the live log writer", func() {
					Expect(liveWriter.WriteCallCount()).To(Equal(0))
				})

				By("flushing the live log writer", func() {
					Expect(liveWriter.StopCallCount()).To(Equal(1))
				})
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

					bar.GetCurrentReturns(0)
					bar.GetTotalReturns(100)

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
