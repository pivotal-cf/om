package commands_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

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
			stdout         *fakes.Logger
			stderr         *fakes.Logger
		)

		BeforeEach(func() {
			requestService = &fakes.RequestService{}
			stdout = &fakes.Logger{}
			stderr = &fakes.Logger{}
			command = commands.NewCurl(requestService, stdout, stderr)
		})

		It("executes the API call", func() {
			requestService.InvokeReturns(api.RequestServiceInvokeOutput{
				StatusCode: http.StatusOK,
				Headers: http.Header{
					"Content-Length": []string{"33"},
					"Content-Type":   []string{"application/json"},
					"Accept":         []string{"text/plain"},
				},
				Body: strings.NewReader(`{"some-response-key": "%some-response-value"}`),
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
			Expect(input.Headers).To(HaveKeyWithValue("Content-Type", []string{"application/json"}))

			data, err := ioutil.ReadAll(input.Data)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(`{"some-key": "some-value"}`))
			content := stdout.PrintlnArgsForCall(0)
			Expect(fmt.Sprint(content...)).To(MatchJSON(`{"some-response-key": "%some-response-value"}`))

			format, content := stderr.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("Status: 200 OK"))

			format, content = stderr.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("Accept: text/plain\r\nContent-Length: 33\r\nContent-Type: application/json\r\n"))
		})

		Context("when --silent is specified", func() {
			It("does not write anything to stderr if the status is 200", func() {
				requestService.InvokeReturns(api.RequestServiceInvokeOutput{
					StatusCode: http.StatusOK,
					Headers: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: strings.NewReader("{}"),
				}, nil)

				err := command.Execute([]string{
					"--path", "/api/v0/some/path",
					"--request", "GET",
					"--silent",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stderr.Invocations()).To(BeEmpty())
			})

			It("does not write anything to stderr if the status is 201", func() {
				requestService.InvokeReturns(api.RequestServiceInvokeOutput{
					StatusCode: http.StatusCreated,
					Headers: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: strings.NewReader("{}"),
				}, nil)

				err := command.Execute([]string{
					"--path", "/api/v0/some/path",
					"--request", "POST",
					"--silent",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stderr.Invocations()).To(HaveLen(0))
			})

			It("still writes response headers to stderr if the status is 404", func() {
				requestService.InvokeReturns(api.RequestServiceInvokeOutput{
					StatusCode: http.StatusNotFound,
					Headers: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: strings.NewReader("{}"),
				}, nil)

				err := command.Execute([]string{
					"--path", "/api/v0/some/path",
					"--request", "GET",
					"--silent",
				})
				Expect(err).To(MatchError("server responded with an error"))

				format, content := stderr.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("Status: 404 Not Found"))

				format, content = stderr.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("Content-Type: application/json\r\n"))
			})
		})

		Context("When a custom content-type is passed in", func() {
			It("executes the API call with the given content type", func() {
				requestService.InvokeReturns(api.RequestServiceInvokeOutput{
					Headers: http.Header{
						"Content-Length": []string{"33"},
						"Content-Type":   []string{"application/json"},
						"Accept":         []string{"text/plain"},
					},
					Body: strings.NewReader(`{"some-response-key": "%some-response-value"}`),
				}, nil)

				err := command.Execute([]string{
					"--path", "/api/v0/some/path",
					"--request", "POST",
					"--data", `some_key=some_value`,
					"--header", "Content-Type: application/x-www-form-urlencoded",
				})
				Expect(err).NotTo(HaveOccurred())

				input := requestService.InvokeArgsForCall(0)
				Expect(input.Path).To(Equal("/api/v0/some/path"))
				Expect(input.Method).To(Equal("POST"))
				Expect(input.Headers).To(HaveKeyWithValue("Content-Type", []string{"application/x-www-form-urlencoded"}))
			})
		})

		Describe("pretty printing", func() {
			Context("when the response type is JSON", func() {
				It("pretty prints the response body", func() {
					requestService.InvokeReturns(api.RequestServiceInvokeOutput{
						Headers: http.Header{
							"Content-Length": []string{"33"},
							"Content-Type":   []string{"application/json; charset=utf-8"},
						},
						Body: strings.NewReader(`{"some-response-key": "some-response-value"}`),
					}, nil)

					err := command.Execute([]string{
						"--path", "/api/v0/some/path",
						"--request", "POST",
						"--data", `{"some-key": "some-value"}`,
					})
					Expect(err).NotTo(HaveOccurred())

					content := stdout.PrintlnArgsForCall(0)
					Expect(fmt.Sprint(content...)).To(Equal("{\n  \"some-response-key\": \"some-response-value\"\n}"))
				})
			})

			Context("when the response type is not JSON", func() {
				It("doesn't format the response body", func() {
					requestService.InvokeReturns(api.RequestServiceInvokeOutput{
						Headers: http.Header{
							"Content-Length": []string{"33"},
							"Content-Type":   []string{"text/plain; charset=utf-8"},
						},
						Body: strings.NewReader(`{"some-response-key": "some-response-value"}`),
					}, nil)

					err := command.Execute([]string{
						"--path", "/api/v0/some/path",
						"--request", "POST",
						"--data", `{"some-key": "some-value"}`,
					})
					Expect(err).NotTo(HaveOccurred())

					content := stdout.PrintlnArgsForCall(0)
					Expect(fmt.Sprint(content...)).To(Equal(`{"some-response-key": "some-response-value"}`))
				})
			})
		})

		Context("failure cases", func() {
			Context("when the flags cannot be parsed", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--bad-flag", "some-value"})
					Expect(err).To(MatchError("could not parse curl flags: flag provided but not defined: -bad-flag"))
				})
			})

			Context("when the request path is not provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{
						"--request", "GET",
						"--data", `{"some-key": "some-value"}`,
					})
					Expect(err).To(MatchError("could not parse curl flags: missing required flag \"--path\""))
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

			Context("when the response code is 400 or higher", func() {
				It("returns an error", func() {
					requestService.InvokeReturns(api.RequestServiceInvokeOutput{
						StatusCode: 401,
						Body:       strings.NewReader(`{"some-response-key": "some-response-value"}`),
					}, nil)

					err := command.Execute([]string{
						"--path", "/api/v0/some/path",
						"--request", "POST",
						"--data", `{"some-key": "some-value"}`,
					})
					Expect(err).To(MatchError("server responded with an error"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns the usage information for the curl command", func() {
			command := commands.NewCurl(nil, nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command issues an authenticated API request as defined in the arguments",
				ShortDescription: "issues an authenticated API request",
				Flags:            command.Options,
			}))
		})
	})
})
