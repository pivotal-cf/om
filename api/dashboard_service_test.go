package api_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const dashboardForms = `
<html>
	<body>
		<form action="/product_adds" method="some-method">
			<input name="_method" value="some-rails1" />
			<input name="authenticity_token" value="some-authenticity1" />
		</form>
		<form action="/installation" method="some-method">
			<input name="_method" value="delete" />
			<input name="authenticity_token" value="revert-authenticity-token" />
		</form>
		<form action="/install" method="some-method">
			<input name="_method" value="some-rails2" />
			<input name="authenticity_token" value="some-authenticity2" />
		</form>
	</body>
</html>`

var _ = Describe("DashboardService", func() {
	var (
		service api.DashboardService
		client  *fakes.HttpClient
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		service = api.NewDashboardService(client)
	})

	Describe("GetRevertForm", func() {
		It("returns the form details", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(dashboardForms)),
			}, nil)

			form, err := service.GetRevertForm()
			Expect(err).NotTo(HaveOccurred())

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.Path).To(Equal("/"))

			Expect(form).To(Equal(api.Form{
				Action:            "/installation",
				AuthenticityToken: "revert-authenticity-token",
				RailsMethod:       "delete",
			}))
		})

		Context("when the form does not exist", func() {
			It("returns an empty form", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("")),
				}, nil)
				form, err := service.GetRevertForm()
				Expect(err).To(Not(HaveOccurred()))
				Expect(form).To(Equal(api.Form{}))
			})
		})

		Context("when an error occurs", func() {
			Context("when http client fails", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("whoops"))

					_, err := service.GetRevertForm()
					Expect(err).To(MatchError("failed during request: whoops"))
				})
			})

			Context("when the authenticity token cannot be found", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`<form action="/installation"><input name="_method" value="post"/></form>`)),
					}, nil)

					_, err := service.GetRevertForm()
					Expect(err).To(MatchError("could not find the form authenticity token"))
				})
			})

			Context("when the form method cannot be found", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`<form action="/installation"><input name="authenticity_token" /></form>`)),
					}, nil)

					_, err := service.GetRevertForm()
					Expect(err).To(MatchError("could not find the form method"))
				})
			})

			Context("when the response status is non-200", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := service.GetRevertForm()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("GetInstallForm", func() {
		It("returns the form details", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(dashboardForms)),
			}, nil)

			form, err := service.GetInstallForm()
			Expect(err).NotTo(HaveOccurred())

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.Path).To(Equal("/"))

			Expect(form).To(Equal(api.Form{
				Action:            "/install",
				AuthenticityToken: "some-authenticity2",
				RailsMethod:       "some-rails2",
			}))
		})

		Context("when an error occurs", func() {
			Context("when the form does not exist", func() {
				It("returns an empty form", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)
					_, err := service.GetInstallForm()
					Expect(err).To(MatchError(ContainSubstring("could not find the install form")))
				})
			})

			Context("when http client fails", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("whoops"))

					_, err := service.GetInstallForm()
					Expect(err).To(MatchError("failed during request: whoops"))
				})
			})

			Context("when authenticity token cannot be found", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`<form action="/install"><input name="_method" value="post"/></form>`)),
					}, nil)

					_, err := service.GetInstallForm()
					Expect(err).To(MatchError("could not find the form authenticity token"))
				})
			})

			Context("when the response status is non-200", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := service.GetInstallForm()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("PostInstallForm", func() {
		It("submits the form content", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("")),
			}, nil)

			input := api.PostFormInput{
				Form:           api.Form{},
				EncodedPayload: "some-payload",
			}

			err := service.PostInstallForm(input)
			Expect(err).NotTo(HaveOccurred())

			req := client.DoArgsForCall(0)
			Expect(req.Method).To(Equal("POST"))
			Expect(req.URL.Path).To(Equal("/installation"))
			Expect(req.Header.Get("Content-Type")).To(Equal("application/x-www-form-urlencoded"))

			bodyBytes, err := ioutil.ReadAll(req.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(bodyBytes)).To(Equal("some-payload"))
		})

		Context("when an error occurs", func() {
			Context("when the client fails during the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("some error"))

					err := service.PostInstallForm(api.PostFormInput{})
					Expect(err).To(MatchError("failed to POST form: some error"))
				})
			})

			Context("when the request responds with a non-200 status", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					err := service.PostInstallForm(api.PostFormInput{})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})
})
