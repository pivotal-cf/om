package network_test

import (
	"errors"
	"net/http"
	"net/http/httputil"

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
		Expect(err).NotTo(HaveOccurred())

		fakeClient = &fakes.HttpClient{}

		response = &http.Response{}
		fakeClient.DoReturns(response, nil)

		out = gbytes.NewBuffer()

		traceClient = network.NewTraceClient(fakeClient, out)
	})

	It("calls the underlying http client", func() {
		resp, err := traceClient.Do(request)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeClient.DoCallCount()).To(Equal(1))
		Expect(resp).To(Equal(response))
	})

	It("dumps the http request and response to the writer", func() {
		_, err := traceClient.Do(request)
		Expect(err).NotTo(HaveOccurred())

		expectedContents, err := httputil.DumpRequest(request, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(gbytes.Say(string(expectedContents)))

		expectedContents, err = httputil.DumpResponse(response, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(gbytes.Say(string(expectedContents)))
	})

	Context("when the underlying http client fails", func() {
		BeforeEach(func() {
			fakeClient.DoReturns(nil, errors.New("boom!"))
		})

		It("returns the error", func() {
			_, err := traceClient.Do(request)
			Expect(err).To(HaveOccurred())
		})
	})
})
