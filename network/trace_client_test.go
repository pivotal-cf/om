package network_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-cf/om/network"
	"github.com/pivotal-cf/om/network/fakes"
)

var _ = Describe("Trace Client", func() {
	var (
		fakeClient  *fakes.HttpClient
		traceClient *network.TraceClient

		request  *http.Request
		response *http.Response

		out *gbytes.Buffer
	)

	BeforeEach(func() {
		var err error
		request, err = http.NewRequest("GET", "http://example.com", nil)
		Expect(err).ToNot(HaveOccurred())

		fakeClient = &fakes.HttpClient{}

		response = &http.Response{}
		fakeClient.DoReturns(response, nil)

		out = gbytes.NewBuffer()

		traceClient = network.NewTraceClient(fakeClient, out)
	})

	It("calls the underlying http client", func() {
		resp, err := traceClient.Do(request)
		Expect(err).ToNot(HaveOccurred())

		Expect(fakeClient.DoCallCount()).To(Equal(1))
		Expect(resp).To(Equal(response))
	})

	It("dumps the http request and response to the writer", func() {
		_, err := traceClient.Do(request)
		Expect(err).ToNot(HaveOccurred())

		expectedContents, err := httputil.DumpRequest(request, true)
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(gbytes.Say(string(expectedContents)))

		expectedContents, err = httputil.DumpResponse(response, true)
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(gbytes.Say(string(expectedContents)))
	})

	When("the underlying http client fails", func() {
		BeforeEach(func() {
			fakeClient.DoReturns(nil, errors.New("boom!"))
		})

		It("returns the error", func() {
			_, err := traceClient.Do(request)
			Expect(err).To(HaveOccurred())
		})
	})

	When("the request body is larger than some arbitrary value", func() {
		It("only dumps the headers", func() {
			request.Body = io.NopCloser(strings.NewReader(`{}`))
			request.ContentLength = 1024 * 1024

			_, err := traceClient.Do(request)
			Expect(err).ToNot(HaveOccurred())

			expectedContents, err := httputil.DumpRequest(request, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(gbytes.Say(string(expectedContents)))
			Expect(out).ToNot(gbytes.Say(`{}`))
		})
	})

	When("the response body is larger than some arbitrary value", func() {
		It("only dumps the headers", func() {
			responseBodySize := 1024 * 1024

			var buffer bytes.Buffer
			for i := 0; i < responseBodySize; i++ {
				buffer.WriteString("a")
			}

			response.Body = io.NopCloser(&buffer)
			response.ContentLength = int64(responseBodySize)

			_, err := traceClient.Do(request)
			Expect(err).ToNot(HaveOccurred())

			expectedContents, err := httputil.DumpResponse(response, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(gbytes.Say(string(expectedContents)))
			Expect(out).ToNot(gbytes.Say("aaaaaaaaaaaa"))
		})
	})
})
