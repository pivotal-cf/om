package api_test

import (
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
	"log"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
)

var _ = Describe("InstallationsService", func() {
	var (
		client  *ghttp.Server
		service api.Api
		stdout  *gbytes.Buffer
		logger  *log.Logger
	)

	BeforeEach(func() {
		stdout = gbytes.NewBuffer()
		logger = log.New(stdout, "", 0)

		client = ghttp.NewServer()

		service = api.New(api.ApiInput{
			Client: httpClient{serverURI: client.URL()},
			Logger: logger,
		})
	})

	AfterEach(func() {
		client.Close()
	})

	Describe("ListInstallations", func() {
		It("lists the installations on the Ops Manager", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/installations"),
					ghttp.RespondWith(http.StatusOK, `{
					"installations": [{
							"user_name": "admin",
							"finished_at": null,
							"started_at": "2017-05-24T23:38:37.316Z",
							"status": "running",
							"id": 3
						}, {
							"user_name": "admin",
							"finished_at": "2017-05-24T23:55:56.106Z",
							"started_at": "2017-05-24T23:38:37.316Z",
							"status": "failed",
							"id": 5
						}, {
							"user_name": "admin",
							"finished_at": "2017-05-24T23:55:56.106Z",
							"started_at": "2017-05-24T23:38:37.316Z",
							"status": "succeeded",
							"id": 2
						}]
					}`),
				),
			)

			output, err := service.ListInstallations()

			Expect(err).ToNot(HaveOccurred())
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
		})
	})

	Describe("CreateInstallation", func() {
		When("deploying all products", func() {
			It("triggers an installation on an Ops Manager, deploying all products", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/installations"),
						ghttp.VerifyJSON(`{"ignore_warnings":"false", "deploy_products":"all"}`),
						ghttp.RespondWith(http.StatusOK, `{"install": {"id":1}}`),
					),
				)

				output, err := service.CreateInstallation(false, true, nil, api.ApplyErrandChanges{})

				Expect(err).ToNot(HaveOccurred())
				Expect(output.ID).To(Equal(1))
			})
		})

		When("deploying no products", func() {
			It("triggers an installation on an Ops Manager, deploying no products", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/installations"),
						ghttp.VerifyJSON(`{"ignore_warnings":"false", "deploy_products":"none"}`),
						ghttp.RespondWith(http.StatusOK, `{"install": {"id":1}}`),
					),
				)

				output, err := service.CreateInstallation(false, false, nil, api.ApplyErrandChanges{})

				Expect(err).ToNot(HaveOccurred())
				Expect(output.ID).To(Equal(1))
			})
		})

		When("deploying some products", func() {
			It("triggers an installation on an Ops Manager, deploying some products", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/installations"),
						ghttp.VerifyJSON(`{"ignore_warnings":"false","deploy_products":["guid2"]}`),
						ghttp.RespondWith(http.StatusOK, `{"install": {"id":1}}`),
					),
				)

				output, err := service.CreateInstallation(false, true, []string{"product2"}, api.ApplyErrandChanges{})
				Expect(err).ToNot(HaveOccurred())
				Expect(output.ID).To(Equal(1))
			})

			It("errors when the product does not exist", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}]`),
					),
				)

				_, err := service.CreateInstallation(false, true, []string{"product2"}, api.ApplyErrandChanges{})
				Expect(err).To(HaveOccurred())
			})
		})

		When("given the errands", func() {
			When("product names are passed", func() {
				It("sends the errands as a json parameter", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
							ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
							ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v0/installations"),
							ghttp.VerifyJSON(`{"ignore_warnings": "false", "deploy_products": ["guid1"], "errands": {"guid1": {"run_post_deploy": {"errand1": "default"}}}}`),
							ghttp.RespondWith(http.StatusOK, `{"install": {"id":1}}`),
						),
					)

					output, err := service.CreateInstallation(false, true, []string{"product1"}, api.ApplyErrandChanges{
						Errands: map[string]api.ProductErrand{
							"product1": {
								RunPostDeploy: map[string]interface{}{
									"errand1": "default",
								},
							},
						},
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(output.ID).To(Equal(1))
				})

				It("returns an error if the product name is not found", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
							ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}, {"guid": "guid3", "type": "product3"}]`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
							ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}, {"guid": "guid3", "type": "product3"}]`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v0/installations"),
							ghttp.VerifyJSON(`{"ignore_warnings": "false", "deploy_products": ["guid2"]}`),
							ghttp.RespondWith(http.StatusOK, `{"install": {"id":1}}`),
						),
					)

					_, err := service.CreateInstallation(false, true, []string{"product2"}, api.ApplyErrandChanges{
						Errands: map[string]api.ProductErrand{
							"product3": {
								RunPostDeploy: map[string]interface{}{
									"errand1": "default",
								},
							},
						},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(stdout).To(gbytes.Say("skipping errand configuration for 'product3' since it was not provided as a productName flag"))
				})
			})
			When("product names are not passed", func() {
				It("sends the errands as a json parameter if the product exists on the ops manager", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
							ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
							ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v0/installations"),
							ghttp.VerifyJSON(`{"ignore_warnings": "false", "deploy_products": "all", "errands": {"guid1": {"run_post_deploy": {"errand1": "default"}}}}`),
							ghttp.RespondWith(http.StatusOK, `{"install": {"id":1}}`),
						),
					)

					output, err := service.CreateInstallation(false, true, []string{}, api.ApplyErrandChanges{
						Errands: map[string]api.ProductErrand{
							"product1": {
								RunPostDeploy: map[string]interface{}{
									"errand1": "default",
								},
							},
						},
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(output.ID).To(Equal(1))
				})

				It("errors if the product does not exist on the ops manager", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
							ghttp.RespondWith(http.StatusOK, `[]`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
							ghttp.RespondWith(http.StatusOK, `[]`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v0/installations"),
							ghttp.VerifyJSON(`{"ignore_warnings": "false", "deploy_products": ["guid2"], "errands": {"guid1": {"run_post_deploy": {"errand1": "default"}}}}`),
							ghttp.RespondWith(http.StatusOK, `{"install": {"id":1}}`),
						),
					)

					_, err := service.CreateInstallation(false, true, []string{}, api.ApplyErrandChanges{
						Errands: map[string]api.ProductErrand{
							"product1": {
								RunPostDeploy: map[string]interface{}{
									"errand1": "default",
								},
							},
						},
					})
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(ContainSubstring("failed to configure errands for product 'product1': could not find product on Ops Manager")))
				})
			})
		})

		When("the client has an error during the request", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/installations"),
						http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
							client.CloseClientConnections()
						}),
					),
				)

				_, err := service.CreateInstallation(false, true, nil, api.ApplyErrandChanges{})
				Expect(err).To(MatchError(ContainSubstring("could not make api request to installations endpoint: could not send api request to POST /api/v0/installations")))
			})
		})

		When("the client returns a 422", func() {
			It("returns an explanation of how to run pre-deploy-check", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/installations"),
						ghttp.VerifyJSON(`{"ignore_warnings":"false","deploy_products":["guid2"]}`),
						ghttp.RespondWith(http.StatusUnprocessableEntity, `{
							"errors": ["'Some IAAS Error', type: SomeVerifier""],
							"deployment_errors": {
								"products": [{
							  		"identifier": "guid2",
							  		"verifiers": [{
								 	 	"type": "SomeVerifier",
										"errors": ["'Some IAAS Error', type: SomeVerifier"],
								 	 	"ignorable": true
									}]
								}]
							}
						}`),
					),
				)

				_, err := service.CreateInstallation(false, true, []string{"product2"}, api.ApplyErrandChanges{})
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				Expect(err).To(MatchError(ContainSubstring("Tip: In Ops Manager 2.6 or newer, you can use `om pre-deploy-check` to get a complete list of failed verifiers and om commands to disable them.")))
			})
		})

		When("the client returns a non-2XX or 422", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.CreateInstallation(false, true, nil, api.ApplyErrandChanges{})
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the json cannot be decoded", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusOK, `[{"guid": "guid1", "type": "product1"}, {"guid": "guid2", "type": "product2"}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/installations"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.CreateInstallation(false, true, nil, api.ApplyErrandChanges{})
				Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
			})
		})
	})

	Describe("GetInstallation", func() {
		It("fetches the status of the installation from the Ops Manager", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/installations/3232"),
					ghttp.RespondWith(http.StatusOK, `{"status": "running"}`),
				),
			)

			output, err := service.GetInstallation(3232)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.Status).To(Equal("running"))
		})

		When("the client has an error during the request", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.GetInstallation(3232)
				Expect(err).To(MatchError(ContainSubstring("could not make api request to installations status endpoint: could not send api request to GET /api/v0/installations/3232")))
			})
		})

		When("the client returns a non-2XX", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/installations/3232"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.GetInstallation(3232)
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the json cannot be decoded", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/installations/3232"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.GetInstallation(3232)
				Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
			})
		})
	})

	Describe("GetInstallationLogs", func() {
		It("grabs the logs from the currently running installation", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/installations/3232/logs"),
					ghttp.RespondWith(http.StatusOK, `{"logs": "some logs"}`),
				),
			)

			output, err := service.GetInstallationLogs(3232)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.Logs).To(Equal("some logs"))
		})

		When("the client returns a non-2XX", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/installations/3232/logs"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.GetInstallationLogs(3232)
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the client has an error during the request", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.GetInstallationLogs(3232)
				Expect(err).To(MatchError(ContainSubstring("could not make api request to installations logs endpoint: could not send api request to GET /api/v0/installations/3232/logs")))
			})
		})

		When("the json cannot be decoded", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/installations/3232/logs"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.GetInstallationLogs(3232)
				Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
			})
		})
	})
})

func parseTime(timeString interface{}) *time.Time {
	if timeString == nil {
		return nil
	}
	timeValue, err := time.Parse(time.RFC3339, timeString.(string))

	if err != nil {
		Expect(err).ToNot(HaveOccurred())
	}

	return &timeValue
}
