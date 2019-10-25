package api_test

import (
	"errors"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PreDeployCheckService", func() {
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

	Describe("ListPendingDirectorChanges", func() {
		It("lists pending director changes", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
					  "pre_deploy_check": {
						"identifier": "p-bosh-guid",
						"complete": false,
						"network": {
						  "assigned": true
						},
						"availability_zone": {
						  "assigned": false
						},
						"stemcells": [
						  {
							"assigned": false,
							"required_stemcell_version": "250.2",
							"required_stemcell_os": "ubuntu-xenial"
						  }
						],
						"properties": [
							{
								"name": ".properties.iaas_configuration.project",
								"type": null,
								"errors": [
									"can't be blank"
								]
							}
						],
						"resources": {
						  "jobs": [{
                            "identifier": "job-identifier",
                            "guid": "job-guid",
                            "error": [
                              "Instance : Value must be a positive integer"
                            ]
                          }]
						},
						"verifiers": [
						  {
							"type": "NetworksPingableVerifier",
							"errors": [ 
							  "NetworksPingableVerifier error"
							],
							"ignorable": true
						  }
						]
					  }
					}`)),
				}, nil
			}

			output, err := service.ListPendingDirectorChanges()
			Expect(err).NotTo(HaveOccurred())

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

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/api/v0/staged/director/pre_deploy_check"))
		})

		Context("failure cases", func() {
			It("returns an error when not 200-OK", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusInternalServerError,
						Body: ioutil.NopCloser(strings.NewReader(`{}`))}, nil
				}

				_, err := service.ListPendingDirectorChanges()
				Expect(err).To(MatchError(ContainSubstring("unexpected response")))
			})

			It("returns an error when the http request could not be made", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("something happened")
				}

				_, err := service.ListPendingDirectorChanges()
				Expect(err).To(MatchError(ContainSubstring("could not make api request to pre_deploy_check endpoint")))
			})

			It("returns an error when the response is not JSON", func() {
				client.DoStub = func(req *http.Request) (response *http.Response, e error) {
					return &http.Response{
						// staged products list
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`invalid JSON`))}, nil
				}

				_, err := service.ListPendingDirectorChanges()
				Expect(err).To(MatchError(ContainSubstring("could not unmarshal pre_deploy_check response")))
			})
		})
	})

	Describe("ListPendingProductChanges", func() {
		It("lists pending product changes for specific product", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/api/v0/staged/products" {
					return &http.Response{
						// staged products list
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`[{"guid":"p-guid"}]`))}, nil
				} else {
					// product/guid/pre_deploy_check response

					return &http.Response{StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(strings.NewReader(`{
					  "pre_deploy_check": {
						"identifier": "p-guid",
						"complete": false,
						"network": {
						  "assigned": true
						},
						"availability_zone": {
						  "assigned": false
						},
						"stemcells": [
						  {
							"assigned": false,
							"required_stemcell_version": "250.2",
							"required_stemcell_os": "ubuntu-xenial"
						  }
						],
						"properties": [
							{
								"name": ".properties.iaas_configuration.project",
								"type": "string",
								"errors": [
									"can't be blank"
								]
							},
							{
								"name": ".my_job.example_collection",
								"type": "collection",
								"errors": [
								  "String Property can't be blank"
								],
								"records": [
								  {
									"index": 3,
									"errors": [
									  {
										"name": "string-property",
										"type": "string",
										"errors": [
										  "can't be blank"
										]
									  }
									]
								  }
								]
							}
						],
						"resources": {
						  "jobs": [{
                            "identifier": "job-identifier",
                            "guid": "job-guid",
                            "error": [
                              "Instance : Value must be a positive integer"
                            ]
                          }]
						},
						"verifiers": [
						  {
							"type": "NetworksPingableVerifier",
							"errors": [ 
							  "NetworksPingableVerifier error"
							],
							"ignorable": true
						  }
						]
					  }
					}`)),
					}, nil
				}
			}

			output, err := service.ListAllPendingProductChanges()
			Expect(err).NotTo(HaveOccurred())

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

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/api/v0/staged/products"))

			request = client.DoArgsForCall(1)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/api/v0/staged/products/p-guid/pre_deploy_check"))
		})

		Context("failure cases", func() {
			When("hitting the endpoint /api/v0/staged/products errors", func() {
				It("returns an error when not 200-OK", func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						return &http.Response{StatusCode: http.StatusInternalServerError,
							Body: ioutil.NopCloser(strings.NewReader(`{}`))}, nil
					}

					_, err := service.ListAllPendingProductChanges()
					Expect(err).To(MatchError(ContainSubstring("unexpected response")))
				})

				It("returns an error when the http request could not be made", func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("something happened")
					}

					_, err := service.ListAllPendingProductChanges()
					Expect(err).To(MatchError(ContainSubstring("could not make api request to pre_deploy_check endpoint")))
				})

				It("returns an error when the response is not JSON", func() {
					client.DoStub = func(req *http.Request) (response *http.Response, e error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(`invalid JSON`))}, nil
					}

					_, err := service.ListAllPendingProductChanges()
					Expect(err).To(MatchError(ContainSubstring("could not unmarshal pre_deploy_check response")))
				})
			})
		})

		When("hitting the endpoint /api/v0/staged/products/p-guid/pre_deploy_checks errors", func() {
			When("hitting the endpoint /api/v0/staged/products errors", func() {
				It("returns an error when not 200-OK", func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						if req.URL.Path == "/api/v0/staged/products" {
							return &http.Response{
								// staged products list
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(strings.NewReader(`[{"guid":"p-guid"}]`))}, nil
						} else {
							return &http.Response{StatusCode: http.StatusInternalServerError,
								Body: ioutil.NopCloser(strings.NewReader(`{}`))}, nil
						}
					}

					_, err := service.ListAllPendingProductChanges()
					Expect(err).To(MatchError(ContainSubstring("unexpected response")))
				})

				It("returns an error when the http request could not be made", func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						if req.URL.Path == "/api/v0/staged/products" {
							return &http.Response{
								// staged products list
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(strings.NewReader(`[{"guid":"p-guid"}]`))}, nil
						} else {
							return nil, errors.New("something happened")
						}
					}

					_, err := service.ListAllPendingProductChanges()
					Expect(err).To(MatchError(ContainSubstring("could not make api request to pre_deploy_check endpoint")))
				})

				It("returns an error when the response is not JSON", func() {
					client.DoStub = func(req *http.Request) (response *http.Response, e error) {
						if req.URL.Path == "/api/v0/staged/products" {
							return &http.Response{
								// staged products list
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(strings.NewReader(`[{"guid":"p-guid"}]`))}, nil
						} else {
							return &http.Response{
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(strings.NewReader(`invalid JSON`))}, nil
						}
					}

					_, err := service.ListAllPendingProductChanges()
					Expect(err).To(MatchError(ContainSubstring("could not unmarshal pre_deploy_check response")))
				})
			})
		})
	})
})
