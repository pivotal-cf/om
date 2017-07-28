package api_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func parseTime(timeString interface{}) *time.Time {
	if timeString == nil {
		return nil
	}
	timeValue, err := time.Parse(time.RFC3339, timeString.(string))

	if err != nil {
		Expect(err).NotTo(HaveOccurred())
	}

	return &timeValue
}

var _ = Describe("InstallationsService", func() {
	var (
		client *fakes.HttpClient
		is     api.InstallationsService
	)
	BeforeEach(func() {
		client = &fakes.HttpClient{}
		is = api.NewInstallationsService(client)
	})

	Describe("ListInstallations", func() {
		It("lists the installations on the Ops Manager", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(strings.NewReader(`{
					"installations": [
						{
							"user_name": "admin",
							"finished_at": null,
							"started_at": "2017-05-24T23:38:37.316Z",
							"status": "running",
							"id": 3
						},
						{
							"user_name": "admin",
							"finished_at": "2017-05-24T23:55:56.106Z",
							"started_at": "2017-05-24T23:38:37.316Z",
							"status": "failed",
							"id": 5
						},
						{
							"user_name": "admin",
							"finished_at": "2017-05-24T23:55:56.106Z",
							"started_at": "2017-05-24T23:38:37.316Z",
							"status": "succeeded",
							"id": 2
						}
					]
				}`))}, nil)

			output, err := is.ListInstallations()

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal([]api.InstallationsServiceOutput{
				{
					ID:         3,
					UserName:   "admin",
					Status:     "running",
					StartedAt:  parseTime("2017-05-24T23:38:37.316Z"),
					FinishedAt: parseTime(nil),
				},
				{
					ID:         5,
					UserName:   "admin",
					Status:     "failed",
					StartedAt:  parseTime("2017-05-24T23:38:37.316Z"),
					FinishedAt: parseTime("2017-05-24T23:55:56.106Z"),
				},
				{
					ID:         2,
					UserName:   "admin",
					Status:     "succeeded",
					StartedAt:  parseTime("2017-05-24T23:38:37.316Z"),
					FinishedAt: parseTime("2017-05-24T23:55:56.106Z"),
				},
			}))

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.Path).To(Equal("/api/v0/installations"))
		})
	})

	Describe("RunningInstallation", func() {
		It("returns only the running installation on the Ops Manager", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(strings.NewReader(`{
					"installations": [
						{
							"finished_at": null,
							"status": "running",
							"id": 3
						},
						{
							"user_name": "admin",
							"finished_at": "2017-05-24T23:55:56.106Z",
							"status": "succeeded",
							"id": 2
						}
					]
				}`))}, nil)

			output, err := is.RunningInstallation()

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(api.InstallationsServiceOutput{
				ID:         3,
				Status:     "running",
				FinishedAt: parseTime(nil),
			}))

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.Path).To(Equal("/api/v0/installations"))
		})

		Context("when there are no installations", func() {
			It("returns a zero value installation", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
					"installations": []}`))}, nil)

				output, err := is.RunningInstallation()

				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(api.InstallationsServiceOutput{}))

				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("GET"))
				Expect(req.URL.Path).To(Equal("/api/v0/installations"))
			})
		})

		Context("when there is no running installation", func() {
			It("returns a zero value installation", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
					"installations": [
						{
							"user_name": "admin",
							"finished_at": "2017-05-25T00:10:00.303Z",
							"status": "succeeded",
							"id": 3
						},
						{
							"user_name": "admin",
							"finished_at": "2017-05-24T23:55:56.106Z",
							"status": "succeeded",
							"id": 2
						}
					]
				}`))}, nil)

				output, err := is.RunningInstallation()

				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(api.InstallationsServiceOutput{}))

				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("GET"))
				Expect(req.URL.Path).To(Equal("/api/v0/installations"))

			})
		})

		Context("when only an earlier installation is listed in the running state", func() {
			It("does not consider the earlier installation to be running", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
					"installations": [
						{
							"finished_at": null,
							"status": "succeeded",
							"id": 3
						},
						{
							"user_name": "admin",
							"finished_at": "2017-07-05T00:39:32.123Z",
							"status": "running",
							"id": 2
						}
					]
				}`))}, nil)

				output, err := is.RunningInstallation()

				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(api.InstallationsServiceOutput{}))

				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("GET"))
				Expect(req.URL.Path).To(Equal("/api/v0/installations"))
			})
		})

		Context("error cases", func() {
			Context("when the client has an error during the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("some error"))

					_, err := is.RunningInstallation()
					Expect(err).To(MatchError("could not make api request to installations endpoint: some error"))
				})
			})

			Context("when the client returns a non-2XX", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := is.RunningInstallation()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the json cannot be decoded", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("##################")),
					}, nil)

					_, err := is.RunningInstallation()
					Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
				})
			})
		})
	})

	Describe("Trigger", func() {
		Context("When deploying all products", func() {
			It("triggers an installation on an Ops Manager, deploying all products", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{"install":{"id":1}}`)),
				}, nil)

				output, err := is.Trigger(false, true)

				Expect(err).NotTo(HaveOccurred())
				Expect(output.ID).To(Equal(1))

				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("POST"))
				Expect(req.URL.Path).To(Equal("/api/v0/installations"))

				body, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(body)).To(Equal(`{"ignore_warnings":"false","deploy_products":"all"}`))
			})
		})

		Context("When deploying no products", func() {
			It("triggers an installation on an Ops Manager, deploying no products", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{"install":{"id":1}}`)),
				}, nil)

				output, err := is.Trigger(false, false)

				Expect(err).NotTo(HaveOccurred())
				Expect(output.ID).To(Equal(1))

				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("POST"))
				Expect(req.URL.Path).To(Equal("/api/v0/installations"))

				body, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(body)).To(Equal(`{"ignore_warnings":"false","deploy_products":"none"}`))
			})
		})

		Context("when an error occurs", func() {
			Context("when the client has an error during the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("some error"))

					_, err := is.Trigger(false, true)
					Expect(err).To(MatchError("could not make api request to installations endpoint: some error"))
				})
			})

			Context("when the client returns a non-2XX", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := is.Trigger(false, true)
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the json cannot be decoded", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("##################")),
					}, nil)

					_, err := is.Trigger(false, true)
					Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
				})
			})
		})
	})

	Describe("Status", func() {
		It("fetches the status of the installation from the Ops Manager", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"status": "running"}`)),
			}, nil)

			output, err := is.Status(3232)

			Expect(err).NotTo(HaveOccurred())
			Expect(output.Status).To(Equal("running"))

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.Path).To(Equal("/api/v0/installations/3232"))
		})

		Context("when an error occurs", func() {
			Context("when the client has an error during the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("some error"))

					_, err := is.Status(3232)
					Expect(err).To(MatchError("could not make api request to installations status endpoint: some error"))
				})
			})

			Context("when the client returns a non-2XX", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := is.Status(3232)
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the json cannot be decoded", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("##################")),
					}, nil)

					_, err := is.Status(3232)
					Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
				})
			})
		})
	})

	Describe("Logs", func() {
		It("grabs the logs from the currently running installation", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"logs": "some logs"}`)),
			}, nil)

			output, err := is.Logs(3232)

			Expect(err).NotTo(HaveOccurred())
			Expect(output.Logs).To(Equal("some logs"))

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.Path).To(Equal("/api/v0/installations/3232/logs"))
		})
	})

	Context("when an error occurs", func() {
		Context("when the client has an error during the request", func() {
			It("returns an error", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(strings.NewReader("")),
				}, errors.New("some error"))

				_, err := is.Logs(3232)
				Expect(err).To(MatchError("could not make api request to installations logs endpoint: some error"))
			})
		})

		Context("when the client returns a non-2XX", func() {
			It("returns an error", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(strings.NewReader("")),
				}, nil)

				_, err := is.Logs(3232)
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		Context("when the json cannot be decoded", func() {
			It("returns an error", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("##################")),
				}, nil)

				_, err := is.Logs(3232)
				Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
			})
		})
	})
})
