package api_test

import (
	"fmt"
	"net/http"

	"github.com/onsi/gomega/ghttp"
	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/om/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JobsService", func() {
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

	Describe("ListStagedProductJobs", func() {
		It("returns a map of the jobs", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs"),
					ghttp.RespondWith(http.StatusOK, `{
						"jobs": [{
							"name": "job-1",
							"guid": "some-guid-1"
						}, {
							"name": "job-2",
							"guid": "some-guid-2"
						}]
					}`),
				),
			)

			jobs, err := service.ListStagedProductJobs("some-product-guid")
			Expect(err).ToNot(HaveOccurred())
			Expect(jobs).To(Equal(map[string]string{
				"job-1": "some-guid-1",
				"job-2": "some-guid-2",
			}))
		})

		When("an error occurs", func() {
			When("the client errors before the request", func() {
				It("returns an error", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs"),
							http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
								client.CloseClientConnections()
							}),
						),
					)

					_, err := service.ListStagedProductJobs("some-product-guid")
					Expect(err).To(MatchError(ContainSubstring("could not make api request to jobs endpoint: could not send api request to GET /api/v0/staged/products/some-product-guid/jobs")))
				})
			})

			When("the jobs endpoint returns a non-200 status code", func() {
				It("returns an error", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs"),
							ghttp.RespondWith(http.StatusNotFound, `{}`),
						),
					)

					_, err := service.ListStagedProductJobs("some-product-guid")
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			When("decoding the json fails", func() {
				It("returns an error", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs"),
							ghttp.RespondWith(http.StatusOK, `bad-json`),
						),
					)

					_, err := service.ListStagedProductJobs("some-product-guid")
					Expect(err).To(MatchError(ContainSubstring("failed to decode jobs json response:")))
				})
			})
		})
	})

	Describe("GetStagedProductJobResourceConfig", func() {
		It("fetches the resource config for a given job", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config"),
					ghttp.RespondWith(http.StatusOK, `{
						"instances": 1,
						"instance_type": { "id": "number-1" },
						"persistent_disk": { "size_mb": "290" },
						"internet_connected": true,
						"elb_names": ["something"],
						"additional_vm_extensions": ["some-vm-extension","some-other-vm-extension"]
					}`),
				),
			)

			job, err := service.GetStagedProductJobResourceConfig("some-product-guid", "some-guid")

			Expect(err).ToNot(HaveOccurred())

			jobProperties := api.JobProperties{
				"instances":          1.0,
				"instance_type":      map[string]interface{}{"id": "number-1"},
				"persistent_disk":    map[string]interface{}{"size_mb": "290"},
				"internet_connected": true,
				"elb_names":          []interface{}{"something"},
				"additional_vm_extensions": []interface{}{
					"some-vm-extension",
					"some-other-vm-extension",
				},
			}

			Expect(job).To(BeEquivalentTo(jobProperties))
		})

		Context("failure cases", func() {
			When("the resource config endpoint returns an error", func() {
				It("returns an error", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config"),
							http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
								client.CloseClientConnections()
							}),
						),
					)

					_, err := service.GetStagedProductJobResourceConfig("some-product-guid", "some-guid")
					Expect(err).To(MatchError(ContainSubstring("could not make api request to resource_config endpoint: could not send api request to GET /api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config")))
				})
			})

			When("the resource config endpoint returns a non-200 status code", func() {
				It("returns an error", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config"),
							ghttp.RespondWith(http.StatusNotFound, `{}`),
						),
					)

					_, err := service.GetStagedProductJobResourceConfig("some-product-guid", "some-guid")
					Expect(err).To(MatchError(ContainSubstring("unexpected response")))
				})
			})

			When("the resource config returns invalid JSON", func() {
				It("returns an error", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config"),
							ghttp.RespondWith(http.StatusOK, `bad-json`),
						),
					)

					_, err := service.GetStagedProductJobResourceConfig("some-product-guid", "some-guid")
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})
	})

	Describe("ConfigureJobResourceConfig", func() {
		It("configures job resources", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs"),
					ghttp.RespondWith(http.StatusOK, `{
						"jobs": [{
							"name": "some-job",
							"guid": "some-guid"
						}, {
							"name": "another-job",
							"guid": "another-guid"
						}, {
							"name": "third-job",
							"guid": "third-guid"
						}]
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/another-guid/resource_config"),
					ghttp.RespondWith(http.StatusOK, `{
						"instance_type": {
							"id": "automatic"
						},
						"persistent_disk": {
							"size_mb": "20480"
						}
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/jobs/another-guid/resource_config"),
					ghttp.VerifyJSON(`{
						"instance_type": {
							"id": "automatic"
						},
						"persistent_disk": {
							"size_mb": "20480"
						},
				  		"additional_vm_extensions": [],
				  		"instance_type": {
							"id": "automatic"
				  		},
				  		"instances": 2
        			}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config"),
					ghttp.RespondWith(http.StatusOK, `{
						"instance_type": {
							"id": "automatic"
						},
						"persistent_disk": {
							"size_mb": "20480"
						}
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config"),
					ghttp.VerifyJSON(`{
				  		"additional_vm_extensions": [
							"some-vm-extension",
							"some-other-vm-extension"
				  		],
				  		"instance_type": {
							"id": "number-1"
				  		},
				  		"instances": 1,
				  		"internet_connected": true,
				  		"lb_names": [
							"something"
				  		],
				  		"persistent_disk": {
							"size": 290
				  		}
        			}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/third-guid/resource_config"),
					ghttp.RespondWith(http.StatusOK, `{
						"instance_type": {
							"id": "automatic"
						},
						"persistent_disk": {
							"size_mb": "20480"
						},
				  		"additional_vm_extensions": ["test-extension"]
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/jobs/third-guid/resource_config"),
					ghttp.VerifyJSON(`{
				  		"additional_vm_extensions": ["test-extension"],
				  		"instance_type": {
							"id": "number-2"
				  		},
						"persistent_disk": {
							"size_mb": "20480"
						}
        			}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			configContents := `
some-job:
  instances: 1
  persistent_disk:
    size: 290
  instance_type:
    id: number-1
  internet_connected: true
  lb_names: ["something"]
  additional_vm_extensions: ["some-vm-extension", "some-other-vm-extension"]
another-job:
  instances: 2
  additional_vm_extensions: []
third-job:
  instance_type:
    id: number-2
`

			var config map[string]interface{}
			err := yaml.UnmarshalStrict([]byte(configContents), &config)
			Expect(err).ToNot(HaveOccurred())

			err = service.ConfigureJobResourceConfig("some-product-guid", config)
			Expect(err).ToNot(HaveOccurred())
		})

		DescribeTable("additional_vm_extensions", func(serverExtensions string, configExtensions string, expectatedExtensions string) {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs"),
					ghttp.RespondWith(http.StatusOK, `{
						"jobs": [{
							"name": "some-job",
							"guid": "some-guid"
						}]
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{
						"instance_type": {
							"id": "automatic"
						},
						"additional_vm_extensions": [%s],
						"persistent_disk": {
							"size_mb": "20480"
						}
					}`, serverExtensions)),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config"),
					ghttp.VerifyJSON(fmt.Sprintf(`{
				  		"additional_vm_extensions": [
							%s
				  		],
						"instances": 2,
				  		"instance_type": {
							"id": "automatic"
				  		},
				  		"persistent_disk": {
							"size_mb": "20480"
				  		}
        			}`, expectatedExtensions)),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			configContents := fmt.Sprintf(`
some-job:
  instances: 2
  %s
`, configExtensions)

			var config map[string]interface{}
			err := yaml.UnmarshalStrict([]byte(configContents), &config)
			Expect(err).ToNot(HaveOccurred())

			err = service.ConfigureJobResourceConfig("some-product-guid", config)
			Expect(err).ToNot(HaveOccurred())
		},
			Entry("empty-missing-empty", ``, ``, ``),
			Entry("empty-empty-empty", ``, `additional_vm_extensions: []`, ``),
			Entry("empty-filled-filled", ``, `additional_vm_extensions: ["a"]`, `"a"`),
			Entry("filled-missing-filled", `"a"`, ``, `"a"`),
			Entry("filled-empty-filled", `"a"`, `additional_vm_extensions: []`, ``),
			Entry("filled-filled-filled", `"b"`, `additional_vm_extensions: ["a"]`, `"a"`),
		)

		When("an error occurs", func() {
			BeforeEach(func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs"),
						ghttp.RespondWith(http.StatusOK, `{
								"jobs": [{
									"name": "some-job",
									"guid": "some-guid"
								}]
							}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config"),
						ghttp.RespondWith(http.StatusOK, `{
								"instance_type": {
									"id": "automatic"
								},
								"persistent_disk": {
									"size_mb": "20480"
								}
							}`),
					),
				)
			})

			When("the client errors before the request", func() {
				It("returns an error when updating the resource config", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config"),
							http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
								client.CloseClientConnections()
							}),
						),
					)

					configContents := `
some-job:
  instances: 1
`

					var config map[string]interface{}
					err := yaml.UnmarshalStrict([]byte(configContents), &config)
					Expect(err).ToNot(HaveOccurred())

					err = service.ConfigureJobResourceConfig("some-product-guid", config)

					Expect(err).To(MatchError(ContainSubstring("could not make api request to jobs resource_config endpoint")))
				})
			})

			When("the client errors before the request", func() {
				It("returns an error when updating the resource config", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config"),
							ghttp.RespondWith(http.StatusNotFound, `{}`),
						),
					)

					configContents := `
some-job:
  instances: 1
`

					var config map[string]interface{}
					err := yaml.UnmarshalStrict([]byte(configContents), &config)
					Expect(err).ToNot(HaveOccurred())

					err = service.ConfigureJobResourceConfig("some-product-guid", config)

					Expect(err).To(MatchError(ContainSubstring("failed to configure resources for some-job")))
				})
			})
		})
	})
})
