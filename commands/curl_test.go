package commands_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	commonfakes "github.com/pivotal-cf/om/common/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type errReader struct{}

func (er errReader) Read([]byte) (int, error) {
	return 0, errors.New("failed to read")
}

var _ = Describe("Curl", func() {
	Describe("Execute", func() {
		var (
			command        commands.Curl
			requestService *fakes.RequestService
			stdout         *commonfakes.Logger
			stderr         *commonfakes.Logger
		)

		BeforeEach(func() {
			requestService = &fakes.RequestService{}
			stdout = &commonfakes.Logger{}
			stderr = &commonfakes.Logger{}
			command = commands.NewCurl(requestService, stdout, stderr)
		})

		It("executes the API call", func() {
			requestService.InvokeReturns(api.RequestServiceInvokeOutput{
				StatusCode: http.StatusTeapot,
				Headers: http.Header{
					"Content-Length": []string{"33"},
					"Content-Type":   []string{"application/json"},
					"Accept":         []string{"text/plain"},
				},
				Body: strings.NewReader(`{"some-response-key": "some-response-value"}`),
			}, nil)

			err := command.Execute([]string{
				"--path", "/api/v0/some/path",
				"--request", "POST",
				"--data", `{"some-key": "some-value"}`,
			})
			Expect(err).NotTo(HaveOccurred())

			input := requestService.InvokeArgsForCall(0)
			Expect(input.Path).To(Equal("/api/v0/some/path"))
			Expect(input.Method).To(Equal("POST"))

			data, err := ioutil.ReadAll(input.Data)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(`{"some-key": "some-value"}`))

			Expect(stdout.Lines).To(HaveLen(1))
			Expect(stdout.Lines[0]).To(MatchJSON(`{"some-response-key": "some-response-value"}`))

			Expect(stderr.Lines).To(HaveLen(2))
			Expect(stderr.Lines[0]).To(Equal("Status: 418 I'm a teapot"))
			Expect(stderr.Lines[1]).To(Equal("Accept: text/plain\r\nContent-Length: 33\r\nContent-Type: application/json\r\n"))
		})

		Context("failure cases", func() {
			Context("when the flags cannot be parsed", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--bad-flag", "some-value"})
					Expect(err).To(MatchError("could not parse curl flags: flag provided but not defined: -bad-flag"))
				})
			})

			Context("when the request service returns an error", func() {
				It("returns an error", func() {
					requestService.InvokeReturns(api.RequestServiceInvokeOutput{}, errors.New("some request error"))
					err := command.Execute([]string{
						"--path", "/api/v0/some/path",
						"--request", "POST",
						"--data", `{"some-key": "some-value"}`,
					})
					Expect(err).To(MatchError("failed to make api request: some request error"))
				})
			})

			Context("when the response body cannot be read", func() {
				It("returns an error", func() {

					requestService.InvokeReturns(api.RequestServiceInvokeOutput{
						Body: errReader{},
					}, nil)
					err := command.Execute([]string{
						"--path", "/api/v0/some/path",
						"--request", "POST",
						"--data", `{"some-key": "some-value"}`,
					})
					Expect(err).To(MatchError("failed to read api response body: failed to read"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns the usage information for the curl command", func() {
			command := commands.NewCurl(nil, nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This command issues an authenticated API request as defined in the arguments",
				ShortDescription: "issues an authenticated API request",
				Flags:            command.Options,
			}))
		})
	})
})
