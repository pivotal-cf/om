package download_test

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/pivotal-cf/go-pivnet/v7/download"
	"github.com/pivotal-cf/go-pivnet/v7/download/fakes"

	"fmt"
	"math"
	"net"
	"os"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet/v7/logger/loggerfakes"
)

type EOFReader struct{}

func (e EOFReader) Read(p []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

type ConnectionResetReader struct{}

func (e ConnectionResetReader) Read(p []byte) (int, error) {
	return 0, &net.OpError{Err: fmt.Errorf(syscall.ECONNRESET.Error())}
}

type NetError struct {
	error
}

func (ne NetError) Temporary() bool {
	return true
}

func (ne NetError) Timeout() bool {
	return true
}

type ReaderThatDoesntRead struct{}

func (r ReaderThatDoesntRead) Read(p []byte) (int, error) {
	for {
		time.Sleep(time.Second)
	}
}

var _ = Describe("Downloader", func() {
	var (
		httpClient          *fakes.HTTPClient
		ranger              *fakes.Ranger
		bar                 *fakes.Bar
		downloadLinkFetcher *fakes.DownloadLinkFetcher
	)

	BeforeEach(func() {
		httpClient = &fakes.HTTPClient{}
		ranger = &fakes.Ranger{}
		bar = &fakes.Bar{}

		bar.NewProxyReaderStub = func(reader io.Reader) io.Reader { return reader }

		downloadLinkFetcher = &fakes.DownloadLinkFetcher{}
		downloadLinkFetcher.NewDownloadLinkStub = func() (string, error) {
			return "https://example.com/some-file", nil
		}
	})

	Describe("Get", func() {
		It("writes the product to the given location", func() {
			ranger.BuildRangeReturns([]download.Range{
				download.NewRange(
					0,
					9,
					http.Header{"Range": []string{"bytes=0-9"}},
				),
				download.NewRange(
					10,
					19,
					http.Header{"Range": []string{"bytes=10-19"}},
				),
			}, nil)

			var receivedRequests []*http.Request
			var m = &sync.Mutex{}
			httpClient.DoStub = func(req *http.Request) (*http.Response, error) {
				m.Lock()
				receivedRequests = append(receivedRequests, req)
				m.Unlock()

				switch req.Header.Get("Range") {
				case "bytes=0-9":
					return &http.Response{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(strings.NewReader("fake produ")),
					}, nil
				case "bytes=10-19":
					return &http.Response{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(strings.NewReader("ct content")),
					}, nil
				default:
					return &http.Response{
						StatusCode:    http.StatusOK,
						ContentLength: 10,
						Request: &http.Request{
							URL: &url.URL{
								Scheme: "https",
								Host:   "example.com",
								Path:   "some-file",
							},
						},
					}, nil
				}
			}

			downloader := download.Client{
				Logger:     &loggerfakes.FakeLogger{},
				HTTPClient: httpClient,
				Ranger:     ranger,
				Bar:        bar,
				Timeout:    5 * time.Millisecond,
			}

			tmpFile, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			tmpLocation, err := download.NewFileInfo(tmpFile)
			Expect(err).NotTo(HaveOccurred())

			err = downloader.Get(tmpLocation, downloadLinkFetcher, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			content, err := ioutil.ReadAll(tmpFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(Equal("fake product content"))

			Expect(ranger.BuildRangeCallCount()).To(Equal(1))
			Expect(ranger.BuildRangeArgsForCall(0)).To(Equal(int64(10)))

			Expect(bar.SetTotalArgsForCall(0)).To(Equal(int64(10)))
			Expect(bar.KickoffCallCount()).To(Equal(1))

			Expect(httpClient.DoCallCount()).To(Equal(3))

			methods := []string{
				receivedRequests[0].Method,
				receivedRequests[1].Method,
				receivedRequests[2].Method,
			}
			urls := []string{
				receivedRequests[0].URL.String(),
				receivedRequests[1].URL.String(),
				receivedRequests[2].URL.String(),
			}
			headers := []string{
				receivedRequests[1].Header.Get("Range"),
				receivedRequests[2].Header.Get("Range"),
			}
			refererHeaders := []string{
				receivedRequests[0].Header.Get("Referer"),
				receivedRequests[1].Header.Get("Referer"),
				receivedRequests[2].Header.Get("Referer"),
			}

			Expect(methods).To(ConsistOf([]string{"HEAD", "GET", "GET"}))
			Expect(urls).To(ConsistOf([]string{"https://example.com/some-file", "https://example.com/some-file", "https://example.com/some-file"}))
			Expect(headers).To(ConsistOf([]string{"bytes=0-9", "bytes=10-19"}))
			Expect(refererHeaders).To(ConsistOf([]string{
				"https://go-pivnet.network.tanzu.vmware.com",
				"https://go-pivnet.network.tanzu.vmware.com",
				"https://go-pivnet.network.tanzu.vmware.com",
			}))

			Expect(bar.FinishCallCount()).To(Equal(1))
		})
	})

	Context("when a retryable error occurs", func() {
		var (
			responses      []*http.Response
			responseErrors []error
			tmpFile        *os.File
		)

		JustBeforeEach(func() {
			defaultErrors := []error{nil}
			defaultResponses := []*http.Response{
				{
					Request: &http.Request{
						URL: &url.URL{
							Scheme: "https",
							Host:   "example.com",
							Path:   "some-file",
						},
					},
				},
			}

			responseErrors = append(defaultErrors, responseErrors...)
			responses = append(defaultResponses, responses...)

			httpClient.DoStub = func(req *http.Request) (*http.Response, error) {
				count := httpClient.DoCallCount() - 1
				return responses[count], responseErrors[count]
			}

			ranger.BuildRangeReturns([]download.Range{download.NewRange(0, 15, http.Header{})}, nil)

			downloader := download.Client{
				Logger:     &loggerfakes.FakeLogger{},
				HTTPClient: httpClient,
				Ranger:     ranger,
				Bar:        bar,
				Timeout:    5 * time.Millisecond,
			}

			var err error
			tmpFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			tmpLocation, err := download.NewFileInfo(tmpFile)
			Expect(err).NotTo(HaveOccurred())

			err = downloader.Get(tmpLocation, downloadLinkFetcher, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an unexpected EOF", func() {
			BeforeEach(func() {
				responseErrors = []error{nil, nil}
				responses = []*http.Response{
					{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(io.MultiReader(strings.NewReader("some"), EOFReader{})),
					},
					{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(strings.NewReader("something")),
					},
				}
			})

			It("successfully retries the download", func() {
				stats, err := tmpFile.Stat()
				Expect(err).NotTo(HaveOccurred())

				Expect(stats.Size()).To(BeNumerically(">", 0))

				Expect(bar.AddArgsForCall(0)).To(Equal(-4))

				content, err := ioutil.ReadAll(tmpFile)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(content)).To(Equal("something"))
			})
		})

		Context("when there is a temporary network error", func() {
			BeforeEach(func() {
				responses = []*http.Response{
					{
						StatusCode: http.StatusPartialContent,
					},
					{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(strings.NewReader("something")),
					},
				}
				responseErrors = []error{NetError{errors.New("whoops")}, nil}
			})

			It("successfully retries the download", func() {
				stats, err := tmpFile.Stat()
				Expect(err).NotTo(HaveOccurred())

				Expect(stats.Size()).To(BeNumerically(">", 0))
			})
		})

		Context("when the connection is reset", func() {
			BeforeEach(func() {
				responses = []*http.Response{
					{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(io.MultiReader(strings.NewReader("some"), ConnectionResetReader{})),
					},
					{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(strings.NewReader("something")),
					},
				}

				responseErrors = []error{nil, nil}

			})

			It("successfully retries the download", func() {
				stats, err := tmpFile.Stat()
				Expect(err).NotTo(HaveOccurred())

				Expect(stats.Size()).To(BeNumerically(">", 0))
				Expect(bar.AddArgsForCall(0)).To(Equal(-4))

				content, err := ioutil.ReadAll(tmpFile)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(content)).To(Equal("something"))
			})
		})

		Context("when there is a timeout", func() {
			BeforeEach(func() {
				responses = []*http.Response{
					{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(io.MultiReader(strings.NewReader("some"), ReaderThatDoesntRead{})),
					},
					{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(strings.NewReader("something")),
					},
				}

				responseErrors = []error{nil, nil}

			})

			It("retries", func() {
				stats, err := tmpFile.Stat()
				Expect(err).NotTo(HaveOccurred())

				Expect(stats.Size()).To(BeNumerically(">", 0))
				Expect(bar.AddArgsForCall(0)).To(Equal(-4))

				content, err := ioutil.ReadAll(tmpFile)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(content)).To(Equal("something"))
			})
		})
	})

	Context("when an error occurs", func() {
		Context("when the disk is out of memory", func() {
			It("returns an error", func() {
				tooBig := int64(math.MaxInt64)
				responses := []*http.Response{
					{
						Request: &http.Request{
							URL: &url.URL{
								Scheme: "https",
								Host:   "example.com",
								Path:   "some-file",
							},
						},
						ContentLength: tooBig,
					},
				}
				errors := []error{nil, nil}

				httpClient.DoStub = func(req *http.Request) (*http.Response, error) {
					count := httpClient.DoCallCount() - 1
					return responses[count], errors[count]
				}

				downloader := download.Client{
					Logger:     &loggerfakes.FakeLogger{},
					HTTPClient: httpClient,
					Ranger:     ranger,
					Bar:        bar,
					Timeout:    5 * time.Millisecond,
				}

				file, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				location, err := download.NewFileInfo(file)
				Expect(err).NotTo(HaveOccurred())

				err = downloader.Get(location, downloadLinkFetcher, GinkgoWriter)
				Expect(err).To(MatchError(ContainSubstring("file is too big to fit on this drive:")))
				Expect(err).To(MatchError(ContainSubstring("bytes required")))
				Expect(err).To(MatchError(ContainSubstring("bytes free")))
			})
		})

		Context("when content length is -1", func() {
			It("returns an error", func() {
				invalidLength := int64(-1)

				responses := []*http.Response{
					{
						Request: &http.Request{
							URL: &url.URL{
								Scheme: "https",
								Host:   "example.com",
								Path:   "some-file",
							},
						},
						ContentLength: invalidLength,
					},
				}
				errors := []error{nil, nil}

				httpClient.DoStub = func(req *http.Request) (*http.Response, error) {
					count := httpClient.DoCallCount() - 1
					return responses[count], errors[count]
				}

				downloader := download.Client{
					Logger:     &loggerfakes.FakeLogger{},
					HTTPClient: httpClient,
					Ranger:     ranger,
					Bar:        bar,
					Timeout:    5 * time.Millisecond,
				}

				file, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				location, err := download.NewFileInfo(file)
				Expect(err).NotTo(HaveOccurred())

				err = downloader.Get(location, downloadLinkFetcher, GinkgoWriter)
				Expect(err).To(MatchError(ContainSubstring("failed to find file")))
			})
		})

		Context("when the HEAD request cannot be constucted", func() {
			It("returns an error", func() {
				downloader := download.Client{
					Logger:     &loggerfakes.FakeLogger{},
					HTTPClient: nil,
					Ranger:     nil,
					Bar:        nil,
					Timeout:    5 * time.Millisecond,
				}
				downloadLinkFetcher.NewDownloadLinkStub = func() (string, error) {
					return "%%%", nil
				}

				err := downloader.Get(nil, downloadLinkFetcher, GinkgoWriter)
				Expect(err).To(MatchError(ContainSubstring("failed to construct HEAD request")))
			})
		})

		Context("when the HEAD has an error", func() {
			It("returns an error", func() {
				httpClient.DoReturns(&http.Response{}, errors.New("failed request"))

				downloader := download.Client{
					Logger:     &loggerfakes.FakeLogger{},
					HTTPClient: httpClient,
					Ranger:     nil,
					Bar:        nil,
					Timeout:    5 * time.Millisecond,
				}

				err := downloader.Get(nil, downloadLinkFetcher, GinkgoWriter)
				Expect(err).To(MatchError("failed to make HEAD request: failed request"))
			})
		})

		Context("when building a range fails", func() {
			It("returns an error", func() {
				httpClient.DoReturns(&http.Response{Request: &http.Request{
					URL: &url.URL{
						Scheme: "https",
						Host:   "example.com",
						Path:   "some-file",
					},
				},
				}, nil)

				ranger.BuildRangeReturns([]download.Range{}, errors.New("failed range build"))

				downloader := download.Client{
					Logger:     &loggerfakes.FakeLogger{},
					HTTPClient: httpClient,
					Ranger:     ranger,
					Bar:        nil,
					Timeout:    5 * time.Millisecond,
				}

				err := downloader.Get(nil, downloadLinkFetcher, GinkgoWriter)
				Expect(err).To(MatchError("failed to construct range: failed range build"))
			})
		})

		Context("when the GET fails", func() {
			It("returns an error", func() {
				responses := []*http.Response{
					{
						Request: &http.Request{
							URL: &url.URL{
								Scheme: "https",
								Host:   "example.com",
								Path:   "some-file",
							},
						},
					},
					{},
				}
				errors := []error{nil, errors.New("failed GET")}

				httpClient.DoStub = func(req *http.Request) (*http.Response, error) {
					count := httpClient.DoCallCount() - 1
					return responses[count], errors[count]
				}

				ranger.BuildRangeReturns([]download.Range{download.NewRange(0, 0, http.Header{})}, nil)

				downloader := download.Client{
					Logger:     &loggerfakes.FakeLogger{},
					HTTPClient: httpClient,
					Ranger:     ranger,
					Bar:        bar,
					Timeout:    5 * time.Millisecond,
				}

				file, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				location, err := download.NewFileInfo(file)
				Expect(err).NotTo(HaveOccurred())

				err = downloader.Get(location, downloadLinkFetcher, GinkgoWriter)
				Expect(err).To(MatchError("problem while waiting for chunks to download: failed during retryable request: download request failed: failed GET"))
			})
		})

		Context("when the GET returns a non-206", func() {
			It("returns an error", func() {
				responses := []*http.Response{
					{
						Request: &http.Request{
							URL: &url.URL{
								Scheme: "https",
								Host:   "example.com",
								Path:   "some-file",
							},
						},
					},
					{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					},
				}
				errors := []error{nil, nil}

				httpClient.DoStub = func(req *http.Request) (*http.Response, error) {
					count := httpClient.DoCallCount() - 1
					return responses[count], errors[count]
				}

				ranger.BuildRangeReturns([]download.Range{download.NewRange(0, 0, http.Header{})}, nil)

				downloader := download.Client{
					Logger:     &loggerfakes.FakeLogger{},
					HTTPClient: httpClient,
					Ranger:     ranger,
					Bar:        bar,
					Timeout:    5 * time.Millisecond,
				}

				file, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				location, err := download.NewFileInfo(file)
				Expect(err).NotTo(HaveOccurred())

				err = downloader.Get(location, downloadLinkFetcher, GinkgoWriter)
				Expect(err).To(MatchError("problem while waiting for chunks to download: failed during retryable request: during GET unexpected status code was returned: 500"))
			})
		})
	})
})
