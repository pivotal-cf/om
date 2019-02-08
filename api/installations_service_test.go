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
		client  *fakes.HttpClient
		service api.Api
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		service = api.New(api.ApiInput{
			Client: client,
		})
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

			output, err := service.ListInstallations()

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

	Describe("CreateInstallation", func() {
		Context("When deploying all products", func() {
			It("triggers an installation on an Ops Manager, deploying all products", func() {
				client.DoReturnsOnCall(0, &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
				}, nil)
				client.DoReturnsOnCall(1, &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
				}, nil)
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{"install":{"id":1}}`)),
				}, nil)

				output, err := service.CreateInstallation(false, true, nil, api.ApplyErrandChanges{})

				Expect(err).NotTo(HaveOccurred())
				Expect(output.ID).To(Equal(1))

				req := client.DoArgsForCall(2)

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
				client.DoReturnsOnCall(0, &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
				}, nil)
				client.DoReturnsOnCall(1, &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
				}, nil)
				output, err := service.CreateInstallation(false, false, nil, api.ApplyErrandChanges{})

				Expect(err).NotTo(HaveOccurred())
				Expect(output.ID).To(Equal(1))

				req := client.DoArgsForCall(2)

				Expect(req.Method).To(Equal("POST"))
				Expect(req.URL.Path).To(Equal("/api/v0/installations"))

				body, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(body)).To(Equal(`{"ignore_warnings":"false","deploy_products":"none"}`))
			})
		})

		Context("When deploying some products", func() {
			It("triggers an installation on an Ops Manager, deploying some products", func() {
				client.DoReturnsOnCall(0, &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
				}, nil)
				client.DoReturnsOnCall(1, &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
				}, nil)
				client.DoReturnsOnCall(2, &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{"install":{"id":1}}`)),
				}, nil)

				output, err := service.CreateInstallation(false, true, []string{"product2"}, api.ApplyErrandChanges{})
				Expect(err).NotTo(HaveOccurred())
				Expect(output.ID).To(Equal(1))

				req := client.DoArgsForCall(2)
				Expect(req.Method).To(Equal("POST"))
				Expect(req.URL.Path).To(Equal("/api/v0/installations"))

				body, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(body)).To(Equal(`{"ignore_warnings":"false","deploy_products":["guid2"]}`))
			})
		})

		Context("when given the errands", func() {
			It("sends the errands as a json parameter", func() {
				client.DoReturnsOnCall(0, &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
				}, nil)
				client.DoReturnsOnCall(1, &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
				}, nil)
				client.DoReturnsOnCall(2, &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{"install":{"id":1}}`)),
				}, nil)

				output, err := service.CreateInstallation(false, true, []string{"product2"}, api.ApplyErrandChanges{
					Errands: map[string]api.ProductErrand{
						"product1": {
							RunPostDeploy: map[string]interface{}{
								"errand1": "default",
							},
						},
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(output.ID).To(Equal(1))

				req := client.DoArgsForCall(2)
				Expect(req.Method).To(Equal("POST"))
				Expect(req.URL.Path).To(Equal("/api/v0/installations"))

				body, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(body)).To(MatchJSON(`{"ignore_warnings":"false","deploy_products":["guid2"],"errands":{"guid1":{"run_post_deploy":{"errand1":"default"}}}}`)) // TODO product1-guid
			})

			It("returns an error if the product guid is not found", func() {
				client.DoReturnsOnCall(0, &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
				}, nil)
				client.DoReturnsOnCall(1, &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
				}, nil)
				client.DoReturnsOnCall(2, &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{"install":{"id":1}}`)),
				}, nil)

				_, err := service.CreateInstallation(false, true, []string{"product2"}, api.ApplyErrandChanges{
					Errands: map[string]api.ProductErrand{
						"product3": {
							RunPostDeploy: map[string]interface{}{
								"errand1": "default",
							},
						},
					},
				})
				Expect(err).To(MatchError("failed to fetch product GUID for product: product3"))
			})
		})

		Context("when an error occurs", func() {
			Context("when the client has an error during the request", func() {
				It("returns an error", func() {
					client.DoReturnsOnCall(0, &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
					}, nil)
					client.DoReturnsOnCall(1, &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
					}, nil)
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("some error"))

					_, err := service.CreateInstallation(false, true, nil, api.ApplyErrandChanges{})
					Expect(err).To(MatchError("could not make api request to installations endpoint: could not send api request to POST /api/v0/installations: some error"))
				})
			})

			Context("when the client returns a non-2XX", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := service.CreateInstallation(false, true, nil, api.ApplyErrandChanges{})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the json cannot be decoded", func() {
				It("returns an error", func() {
					client.DoReturnsOnCall(0, &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
					}, nil)
					client.DoReturnsOnCall(1, &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`[ { "guid": "guid1", "type": "product1"}, { "guid": "guid2", "type": "product2"} ]`)),
					}, nil)
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("##################")),
					}, nil)

					_, err := service.CreateInstallation(false, true, nil, api.ApplyErrandChanges{})
					Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
				})
			})
		})
	})

	Describe("GetInstallation", func() {
		It("fetches the status of the installation from the Ops Manager", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"status": "running"}`)),
			}, nil)

			output, err := service.GetInstallation(3232)

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

					_, err := service.GetInstallation(3232)
					Expect(err).To(MatchError("could not make api request to installations status endpoint: could not send api request to GET /api/v0/installations/3232: some error"))
				})
			})

			Context("when the client returns a non-2XX", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := service.GetInstallation(3232)
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the json cannot be decoded", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("##################")),
					}, nil)

					_, err := service.GetInstallation(3232)
					Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
				})
			})
		})
	})

	Describe("GetInstallationLogs", func() {
		It("grabs the logs from the currently running installation", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"logs": "some logs"}`)),
			}, nil)

			output, err := service.GetInstallationLogs(3232)

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

				_, err := service.GetInstallationLogs(3232)
				Expect(err).To(MatchError("could not make api request to installations logs endpoint: could not send api request to GET /api/v0/installations/3232/logs: some error"))
			})
		})

		Context("when the client returns a non-2XX", func() {
			It("returns an error", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(strings.NewReader("")),
				}, nil)

				_, err := service.GetInstallationLogs(3232)
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		Context("when the json cannot be decoded", func() {
			It("returns an error", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("##################")),
				}, nil)

				_, err := service.GetInstallationLogs(3232)
				Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
			})
		})
	})
})
