package api_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf/om/api"
)

var _ = Describe("PreDeployCheckService", func() {
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

	Describe("ListPendingDirectorChanges", func() {
		It("lists pending director changes", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/pre_deploy_check"),
					ghttp.RespondWith(http.StatusOK, `{
						"pre_deploy_check": {
							"identifier": "p-bosh-guid",
							"complete": false,
							"network": {
								"assigned": true
							},
							"availability_zone": {
								"assigned": false
							},
							"stemcells": [{
								"assigned": false,
								"required_stemcell_version": "250.2",
								"required_stemcell_os": "ubuntu-xenial"
							}],
							"properties": [{
								"name": ".properties.iaas_configuration.project",
								"type": null,
								"errors": [
									"can't be blank"
								]
							}],
							"resources": {
								"jobs": [{
									"identifier": "job-identifier",
									"guid": "job-guid",
									"error": [
										"Instance : Value must be a positive integer"
									]
								}]
							},
							"verifiers": [{
								"type": "NetworksPingableVerifier",
								"errors": [
									"NetworksPingableVerifier error"
								],
								"ignorable": true
							}]
						}
					}`),
				),
			)

			output, err := service.ListPendingDirectorChanges()
			Expect(err).ToNot(HaveOccurred())

			Expect(output.EndpointResults).To(Equal(api.PreDeployCheck{
				Identifier: "p-bosh-guid",
				Complete:   false,
				Network: api.PreDeployNetwork{
					Assigned: true,
				},
				AvailabilityZone: api.PreDeployAvailabilityZone{
					Assigned: false,
				},
				Stemcells: []api.PreDeployStemcells{
					{
						Assigned:                false,
						RequiredStemcellVersion: "250.2",
						RequiredStemcellOS:      "ubuntu-xenial",
					},
				},
				Properties: []api.PreDeployProperty{
					{
						Name: ".properties.iaas_configuration.project",
						Type: "",
						Errors: []string{
							"can't be blank",
						},
					},
				},
				Resources: api.PreDeployResources{
					Jobs: []api.PreDeployJob{
						{
							Identifier: "job-identifier",
							GUID:       "job-guid",
							Errors: []string{
								"Instance : Value must be a positive integer",
							},
						},
					},
				},
				Verifiers: []api.PreDeployVerifier{
					{
						Type:      "NetworksPingableVerifier",
						Errors:    []string{"NetworksPingableVerifier error"},
						Ignorable: true,
					},
				},
			}))
		})

		It("returns an error when not 200-OK", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/pre_deploy_check"),
					ghttp.RespondWith(http.StatusTeapot, `{}`),
				),
			)

			_, err := service.ListPendingDirectorChanges()
			Expect(err).To(MatchError(ContainSubstring("unexpected response")))
		})

		It("returns an error when the http request could not be made", func() {
			client.Close()

			_, err := service.ListPendingDirectorChanges()
			Expect(err).To(MatchError(ContainSubstring("could not make api request to pre_deploy_check endpoint")))
		})

		It("returns an error when the response is not JSON", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/pre_deploy_check"),
					ghttp.RespondWith(http.StatusOK, `invalid-json`),
				),
			)

			_, err := service.ListPendingDirectorChanges()
			Expect(err).To(MatchError(ContainSubstring("could not unmarshal pre_deploy_check response")))
		})
	})

	Describe("ListPendingProductChanges", func() {
		It("lists pending product changes for specific product", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[{"guid":"p-guid"}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/p-guid/pre_deploy_check"),
					ghttp.RespondWith(http.StatusOK, `{
						"pre_deploy_check": {
							"identifier": "p-guid",
							"complete": false,
							"network": {
								"assigned": true
							},
							"availability_zone": {
								"assigned": false
							},
							"stemcells": [{
								"assigned": false,
								"required_stemcell_version": "250.2",
								"required_stemcell_os": "ubuntu-xenial"
							}],
							"properties": [{
								"name": ".properties.iaas_configuration.project",
								"type": "string",
								"errors": [
									"can't be blank"
								]
							}, {
								"name": ".my_job.example_collection",
								"type": "collection",
								"errors": [
									"String Property can't be blank"
								],
								"records": [{
									"index": 3,
									"errors": [{
										"name": "string-property",
										"type": "string",
										"errors": [
											"can't be blank"
										]
									}]
								}]
							}],
							"resources": {
								"jobs": [{
									"identifier": "job-identifier",
									"guid": "job-guid",
									"error": [
										"Instance : Value must be a positive integer"
									]
								}]
							},
							"verifiers": [{
								"type": "NetworksPingableVerifier",
								"errors": [ 
									"NetworksPingableVerifier error"
								],
								"ignorable": true
							}]
						}
					}`),
				),
			)

			output, err := service.ListAllPendingProductChanges()
			Expect(err).ToNot(HaveOccurred())

			Expect(output).To(Equal([]api.PendingProductChangesOutput{
				{
					EndpointResults: api.PreDeployCheck{
						Identifier: "p-guid",
						Complete:   false,
						Network: api.PreDeployNetwork{
							Assigned: true,
						},
						AvailabilityZone: api.PreDeployAvailabilityZone{
							Assigned: false,
						},
						Stemcells: []api.PreDeployStemcells{
							{
								Assigned:                false,
								RequiredStemcellVersion: "250.2",
								RequiredStemcellOS:      "ubuntu-xenial",
							},
						},
						Properties: []api.PreDeployProperty{
							{
								Name: ".properties.iaas_configuration.project",
								Type: "string",
								Errors: []string{
									"can't be blank",
								},
							},
							{
								Name: ".my_job.example_collection",
								Type: "collection",
								Errors: []string{
									"String Property can't be blank",
								},
								Records: []api.PreDeployRecord{
									{
										Index: 3,
										Errors: []api.PreDeployProperty{
											{
												Name: "string-property",
												Type: "string",
												Errors: []string{
													"can't be blank",
												},
											},
										},
									},
								},
							},
						},
						Resources: api.PreDeployResources{
							Jobs: []api.PreDeployJob{
								{
									Identifier: "job-identifier",
									GUID:       "job-guid",
									Errors: []string{
										"Instance : Value must be a positive integer",
									},
								},
							},
						},
						Verifiers: []api.PreDeployVerifier{
							{
								Type:      "NetworksPingableVerifier",
								Errors:    []string{"NetworksPingableVerifier error"},
								Ignorable: true,
							},
						},
					},
				},
			}))
		})

		When("hitting the endpoint /api/v0/staged/products errors", func() {
			It("returns an error when not 200-OK", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.ListAllPendingProductChanges()
				Expect(err).To(MatchError(ContainSubstring("unexpected response")))
			})

			It("returns an error when the http request could not be made", func() {
				client.Close()

				_, err := service.ListAllPendingProductChanges()
				Expect(err).To(MatchError(ContainSubstring("could not make api request to pre_deploy_check endpoint")))
			})

			It("returns an error when the response is not JSON", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.ListAllPendingProductChanges()
				Expect(err).To(MatchError(ContainSubstring("could not unmarshal pre_deploy_check response")))
			})
		})

		When("hitting the endpoint /api/v0/staged/products/p-guid/pre_deploy_checks errors", func() {
			When("hitting the endpoint /api/v0/staged/products errors", func() {
				It("returns an error when not 200-OK", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
							ghttp.RespondWith(http.StatusOK, `[{"guid":"p-guid"}]`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products/p-guid/pre_deploy_check"),
							ghttp.RespondWith(http.StatusTeapot, `{}}`),
						),
					)

					_, err := service.ListAllPendingProductChanges()
					Expect(err).To(MatchError(ContainSubstring("unexpected response")))
				})

				It("returns an error when the response is not JSON", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
							ghttp.RespondWith(http.StatusOK, `[{"guid":"p-guid"}]`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products/p-guid/pre_deploy_check"),
							ghttp.RespondWith(http.StatusOK, `invalid-json`),
						),
					)

					_, err := service.ListAllPendingProductChanges()
					Expect(err).To(MatchError(ContainSubstring("could not unmarshal pre_deploy_check response")))
				})
			})
		})
	})
})
