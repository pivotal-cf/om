package api_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf/om/api"
)

var _ = Describe("PendingChangesService", func() {
	var (
		client  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		client = ghttp.NewServer()

		service = api.New(api.ApiInput{
			Client: httpClient{client.URL()},
		})
	})

	AfterEach(func() {
		client.Close()
	})

	Describe("ListStagedPendingChanges", func() {
		It("lists pending changes", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/pending_changes"),
					ghttp.RespondWith(http.StatusOK, `{
						"product_changes": [{
							"guid":"product-123",
							"errands":[
								{ "name":"errand-1", "post_deploy":"true" }
							],
							"action":"install",
                            "extra_field": "needs to be preserved"
						},
						{
							"guid":"product-234",
							"errands":[
								{ "name":"errand-3", "post_deploy":"true" }
							],
							"action":"update",
                            "completeness_checks": {"configuration_complete": true, "stemcell_present": false, "configurable_properties_valid": true}
						}]
				  }`),
				),
			)

			output, err := service.ListStagedPendingChanges()
			Expect(err).ToNot(HaveOccurred())

			Expect(output.ChangeList).To(ConsistOf([]api.ProductChange{{
				GUID: "product-123",
				Errands: []api.Errand{
					{Name: "errand-1", PostDeploy: "true"},
				},
				Action: "install",
			}, {
				GUID:   "product-234",
				Action: "update",
				Errands: []api.Errand{
					{Name: "errand-3", PostDeploy: "true"},
				},
				CompletenessChecks: &api.CompletenessChecks{
					ConfigurationComplete:       true,
					StemcellPresent:             false,
					ConfigurablePropertiesValid: true,
				},
			}}))

			Expect(output.FullReport).To(MatchJSON(`[{
				"guid":"product-123",
				"errands":[
					{ "name":"errand-1", "post_deploy":"true" }
				],
				"action":"install",
				"extra_field": "needs to be preserved"
			}, {
				"guid":"product-234",
				"errands":[
					{ "name":"errand-3", "post_deploy":"true" }
				],
				"action":"update",
				"completeness_checks": {"configuration_complete": true, "stemcell_present": false, "configurable_properties_valid": true}
			}]`))
		})

		Context("the client can't connect to the server", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.ListStagedPendingChanges()
				Expect(err).To(MatchError(ContainSubstring("could not send api request")))
			})
		})

		When("the server won't fetch pending changes", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/pending_changes"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.ListStagedPendingChanges()
				Expect(err).To(MatchError(ContainSubstring("request failed")))
			})
		})

		When("the response is not JSON", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/pending_changes"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.ListStagedPendingChanges()
				Expect(err).To(MatchError(ContainSubstring("could not unmarshal")))
			})
		})
	})
})
