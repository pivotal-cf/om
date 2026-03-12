package api_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf/om/api"
)

var _ = Describe("KubernetesDistributionService", func() {
	var (
		server  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		client := httpClient{
			server.URL(),
		}

		service = api.New(api.ApiInput{
			Client:         client,
			UnauthedClient: client,
		})
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ListKubernetesDistributions", func() {
		It("deserializes products and the kubernetes distribution library", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/kubernetes_distribution_associations"),
					ghttp.RespondWith(http.StatusOK, `{
						"products": [{
							"guid": "redis-def456",
							"identifier": "redis",
							"is_staged_for_deletion": false,
							"staged_kubernetes_distribution": {
								"identifier": "managed-k8s",
								"version": "0.2.0"
							},
							"deployed_kubernetes_distribution": {
								"identifier": "managed-k8s",
								"version": "0.1.0"
							},
							"available_kubernetes_distributions": [
								{"identifier": "managed-k8s", "version": "0.1.0"},
								{"identifier": "managed-k8s", "version": "0.2.0"}
							]
						}],
						"kubernetes_distribution_library": [
							{"identifier": "managed-k8s", "version": "0.1.0", "rank": 1, "label": "Managed Kubernetes"},
							{"identifier": "managed-k8s", "version": "0.2.0", "rank": 1, "label": "Managed Kubernetes"},
							{"identifier": "unmanaged-k8s", "version": "0.1.0", "rank": 50, "label": "Unmanaged Kubernetes"}
						]
					}`),
				),
			)

			output, err := service.ListKubernetesDistributions()
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(api.KubernetesDistributionAssociationsResponse{
				Products: []api.KubernetesProductDistributionEntry{
					{
						GUID:              "redis-def456",
						ProductName:       "redis",
						StagedForDeletion: false,
						StagedKubernetesDistribution: &api.KubernetesDistribution{
							Identifier: "managed-k8s",
							Version:    "0.2.0",
						},
						DeployedKubernetesDistribution: &api.KubernetesDistribution{
							Identifier: "managed-k8s",
							Version:    "0.1.0",
						},
						AvailableKubernetesDistributions: []api.KubernetesDistribution{
							{Identifier: "managed-k8s", Version: "0.1.0"},
							{Identifier: "managed-k8s", Version: "0.2.0"},
						},
					},
				},
				Library: []api.KubernetesDistributionLibraryEntry{
					{Identifier: "managed-k8s", Version: "0.1.0", Rank: 1, Label: "Managed Kubernetes"},
					{Identifier: "managed-k8s", Version: "0.2.0", Rank: 1, Label: "Managed Kubernetes"},
					{Identifier: "unmanaged-k8s", Version: "0.1.0", Rank: 50, Label: "Unmanaged Kubernetes"},
				},
			}))
		})

		When("invalid JSON is returned", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/kubernetes_distribution_associations"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.ListKubernetesDistributions()
				Expect(err).To(MatchError(ContainSubstring("invalid JSON: invalid character 'i' looking for beginning of value")))
			})
		})

		When("the server errors before the request", func() {
			It("returns an error", func() {
				server.Close()

				_, err := service.ListKubernetesDistributions()
				Expect(err).To(MatchError(ContainSubstring("could not make api request to list kubernetes distributions: could not send api request to GET /api/v0/kubernetes_distribution_associations")))
			})
		})

		When("the api returns a non-200 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/kubernetes_distribution_associations"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.ListKubernetesDistributions()
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})

	Describe("AssignKubernetesDistribution", func() {
		It("makes a request to assign the kubernetes distribution", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", "/api/v0/kubernetes_distribution_associations"),
					ghttp.VerifyJSON(`{
						"products": [{
							"guid": "postgres-ghi789",
							"kubernetes_distribution": {
								"identifier": "managed-k8s",
								"version": "1.28.0"
							}
						}]
					}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.AssignKubernetesDistribution(api.AssignKubernetesDistributionInput{
				Products: []api.AssignKubernetesDistributionProduct{{
					GUID: "postgres-ghi789",
					KubernetesDistribution: api.KubernetesDistribution{
						Identifier: "managed-k8s",
						Version:    "1.28.0",
					},
				}},
			})
			Expect(err).ToNot(HaveOccurred())
		})

		When("the server errors before the request", func() {
			It("returns an error", func() {
				server.Close()

				err := service.AssignKubernetesDistribution(api.AssignKubernetesDistributionInput{
					Products: []api.AssignKubernetesDistributionProduct{{
						GUID: "postgres-ghi789",
						KubernetesDistribution: api.KubernetesDistribution{
							Identifier: "managed-k8s",
							Version:    "1.28.0",
						},
					}},
				})
				Expect(err).To(MatchError(ContainSubstring("could not send api request to PATCH /api/v0/kubernetes_distribution_associations")))
			})
		})

		When("the api returns a non-200 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", "/api/v0/kubernetes_distribution_associations"),
						ghttp.VerifyJSON(`{
							"products": [{
								"guid": "postgres-ghi789",
								"kubernetes_distribution": {
									"identifier": "managed-k8s",
									"version": "1.28.0"
								}
							}]
						}`),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				err := service.AssignKubernetesDistribution(api.AssignKubernetesDistributionInput{
					Products: []api.AssignKubernetesDistributionProduct{{
						GUID: "postgres-ghi789",
						KubernetesDistribution: api.KubernetesDistribution{
							Identifier: "managed-k8s",
							Version:    "1.28.0",
						},
					}},
				})
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})
})
