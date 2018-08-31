package network_test

import (
	"errors"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/network"
	"github.com/pivotal-cf/om/network/fakes"
)

var _ = Describe("RetryClient", func() {
	var (
		fakeClient *fakes.HttpClient
		retryCount = 2
	)

	BeforeEach(func() {
		fakeClient = &fakes.HttpClient{}
	})

	Describe("Do", func() {
		Context("when the response is successful", func() {
			BeforeEach(func() {
				fakeClient.DoReturns(&http.Response{StatusCode: http.StatusTeapot}, nil)
			})

			It("returns the response", func() {
				retryClient := network.NewRetryClient(fakeClient, retryCount, time.Millisecond)

				req := http.Request{Method: "some-method"}
				resp, err := retryClient.Do(&req)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeClient.DoCallCount()).To(Equal(1))
				Expect(fakeClient.DoArgsForCall(0).Method).To(Equal("some-method"))
				Expect(resp.StatusCode).To(Equal(http.StatusTeapot))
			})
		})

		Context("when the request errors", func() {
			BeforeEach(func() {
				fakeClient.DoReturns(nil, errors.New("some-error"))
			})

			It("retries the request the given number of times then returns the error", func() {
				retryClient := network.NewRetryClient(fakeClient, retryCount, time.Millisecond)

				req := http.Request{Method: "some-method"}
				_, err := retryClient.Do(&req)
				Expect(err).To(MatchError("some-error"))

				Expect(fakeClient.DoCallCount()).To(Equal(3))
				Expect(fakeClient.DoArgsForCall(0).Method).To(Equal("some-method"))
				Expect(fakeClient.DoArgsForCall(1).Method).To(Equal("some-method"))
				Expect(fakeClient.DoArgsForCall(2).Method).To(Equal("some-method"))
			})
		})

		Context("when the response has a 5XX status code", func() {
			BeforeEach(func() {
				fakeClient.DoReturns(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
			})

			It("retries the request the given number of times then returns the response", func() {
				retryClient := network.NewRetryClient(fakeClient, retryCount, time.Millisecond)

				req := http.Request{Method: "some-method"}
				resp, err := retryClient.Do(&req)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeClient.DoCallCount()).To(Equal(3))
				Expect(fakeClient.DoArgsForCall(0).Method).To(Equal("some-method"))
				Expect(fakeClient.DoArgsForCall(1).Method).To(Equal("some-method"))
				Expect(fakeClient.DoArgsForCall(2).Method).To(Equal("some-method"))
				Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when the request errors initially then succeeds", func() {
			BeforeEach(func() {
				callCount := 0
				fakeClient.DoStub = func(*http.Request) (*http.Response, error) {
					if callCount == 0 {
						callCount++
						return nil, errors.New("some-error")
					} else {
						callCount++
						return &http.Response{StatusCode: http.StatusTeapot}, nil
					}
				}
			})

			It("retries the request the given number of times then returns the error", func() {
				retryClient := network.NewRetryClient(fakeClient, retryCount, time.Millisecond)

				req := http.Request{Method: "some-method"}
				resp, err := retryClient.Do(&req)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeClient.DoCallCount()).To(Equal(2))
				Expect(fakeClient.DoArgsForCall(0).Method).To(Equal("some-method"))
				Expect(fakeClient.DoArgsForCall(1).Method).To(Equal("some-method"))
				Expect(resp.StatusCode).To(Equal(http.StatusTeapot))
			})
		})
	})
})
