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

var _ = Describe("JobsService", func() {
	var (
		client *fakes.HttpClient
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
	})

	Describe("Jobs", func() {
		It("returns a map of the jobs", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(strings.NewReader(`{"jobs": [{"name":"job-1","guid":"some-guid-1"},
				{"name":"job-2","guid":"some-guid-2"}]}`)),
			}, nil)

			service := api.NewJobsService(client)

			jobs, err := service.Jobs("some-product-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(jobs).To(Equal(map[string]string{
				"job-1": "some-guid-1",
				"job-2": "some-guid-2",
			}))

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/api/v0/staged/products/some-product-guid/jobs"))
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("bad"))

					service := api.NewJobsService(client)

					_, err := service.Jobs("some-product-guid")
					Expect(err).To(MatchError("could not make api request to jobs endpoint: bad"))
				})
			})

			Context("when the jobs endpoint returns a non-200 status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader(``)),
					}, nil)

					service := api.NewJobsService(client)

					_, err := service.Jobs("some-product-guid")
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response:")))
				})
			})

			Context("when decoding the json fails", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(``)),
					}, nil)

					service := api.NewJobsService(client)

					_, err := service.Jobs("some-product-guid")
					Expect(err).To(MatchError(ContainSubstring("failed to decode jobs json response:")))
				})
			})
		})
	})

	Describe("GetExistingJobConfig", func() {
		It("fetches the resource config for a given job", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(strings.NewReader(`{
					"instances": 1,
					"instance_type": { "id": "number-1" },
					"persistent_disk": { "size_mb": "290" },
					"internet_connected": true,
					"elb_names": ["something"]
				}`)),
			}, nil)

			service := api.NewJobsService(client)
			job, err := service.GetExistingJobConfig("some-product-guid", "some-guid")

			Expect(err).NotTo(HaveOccurred())
			Expect(client.DoCallCount()).To(Equal(1))
			jobProperties := api.JobProperties{
				Instances:         float64(1),
				PersistentDisk:    &api.Disk{Size: "290"},
				InstanceType:      api.InstanceType{ID: "number-1"},
				InternetConnected: new(bool),
				LBNames:           []string{"something"},
			}
			*jobProperties.InternetConnected = true
			Expect(job).To(Equal(jobProperties))
			request := client.DoArgsForCall(0)
			Expect("/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config").To(Equal(request.URL.Path))
		})

		Context("with floating ips", func() {
			It("fetches the resource config for a given job including floating ips", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
						"instances": 1,
						"instance_type": { "id": "number-1" },
						"persistent_disk": { "size_mb": "290" },
						"internet_connected": true,
						"floating_ips": "some-floating-ip"
					}`)),
				}, nil)

				service := api.NewJobsService(client)
				job, err := service.GetExistingJobConfig("some-product-guid", "some-guid")

				Expect(err).NotTo(HaveOccurred())
				Expect(client.DoCallCount()).To(Equal(1))
				jobProperties := api.JobProperties{
					Instances:         float64(1),
					PersistentDisk:    &api.Disk{Size: "290"},
					InstanceType:      api.InstanceType{ID: "number-1"},
					InternetConnected: new(bool),
					FloatingIPs:       "some-floating-ip",
				}
				*jobProperties.InternetConnected = true
				Expect(job).To(Equal(jobProperties))
				request := client.DoArgsForCall(0)
				Expect("/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config").To(Equal(request.URL.Path))
			})
		})

		Context("with nsx", func() {
			It("fetches the resource config for a given job including nsx properties", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
						"instances": 1,
						"instance_type": { "id": "number-1" },
						"persistent_disk": { "size_mb": "290" },
						"internet_connected": true,
						"elb_names": ["something"],
						"nsx_security_groups":["sg1", "sg2"],
						"nsx_lbs": [
							{
								"edge_name": "edge-1",
								"pool_name": "pool-1",
								"security_group": "sg-1",
								"port": "5000"
							},
							{
								"edge_name": "edge-2",
								"pool_name": "pool-2",
								"security_group": "sg-2",
								"port": "5000"
							}
						]
					}`)),
				}, nil)

				service := api.NewJobsService(client)
				job, err := service.GetExistingJobConfig("some-product-guid", "some-guid")

				Expect(err).NotTo(HaveOccurred())
				Expect(client.DoCallCount()).To(Equal(1))
				jobProperties := api.JobProperties{
					Instances:         float64(1),
					PersistentDisk:    &api.Disk{Size: "290"},
					InstanceType:      api.InstanceType{ID: "number-1"},
					InternetConnected: new(bool),
					LBNames:           []string{"something"},
					NSXSecurityGroups: []string{"sg1", "sg2"},
					NSXLBS: []api.NSXLB{
						api.NSXLB{
							EdgeName:      "edge-1",
							PoolName:      "pool-1",
							SecurityGroup: "sg-1",
							Port:          "5000",
						},
						api.NSXLB{
							EdgeName:      "edge-2",
							PoolName:      "pool-2",
							SecurityGroup: "sg-2",
							Port:          "5000",
						},
					},
				}
				*jobProperties.InternetConnected = true
				Expect(job).To(Equal(jobProperties))
				request := client.DoArgsForCall(0)
				Expect("/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config").To(Equal(request.URL.Path))
			})
		})

		Context("failure cases", func() {
			Context("when the resource config endpoint returns an error", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, errors.New("some client error"))

					service := api.NewJobsService(client)
					_, err := service.GetExistingJobConfig("some-product-guid", "some-guid")

					Expect(err).To(MatchError("could not make api request to resource_config endpoint: some client error"))
				})
			})

			Context("when the resource config endpoint returns a non-200 status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusTeapot,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, nil)

					service := api.NewJobsService(client)
					_, err := service.GetExistingJobConfig("some-product-guid", "some-guid")

					Expect(err).To(MatchError(ContainSubstring("unexpected response")))
				})
			})

			Context("when the resource config returns invalid JSON", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`%%%`)),
					}, nil)

					service := api.NewJobsService(client)
					_, err := service.GetExistingJobConfig("some-product-guid", "some-guid")

					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})
	})

	Describe("ConfigureJob", func() {
		It("configures job resources", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
			}, nil)

			service := api.NewJobsService(client)
			jobProperties := api.JobProperties{
				Instances:         1,
				PersistentDisk:    &api.Disk{Size: "290"},
				InstanceType:      api.InstanceType{ID: "number-1"},
				InternetConnected: new(bool),
				LBNames:           []string{"something"},
			}
			*jobProperties.InternetConnected = true

			err := service.ConfigureJob("some-product-guid", "some-job-guid", jobProperties)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(1))
			request := client.DoArgsForCall(0)

			Expect("application/json").To(Equal(request.Header.Get("Content-Type")))
			Expect("PUT").To(Equal(request.Method))
			Expect("/api/v0/staged/products/some-product-guid/jobs/some-job-guid/resource_config").To(Equal(request.URL.Path))
			reqBytes, err := ioutil.ReadAll(request.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(reqBytes).To(MatchJSON(`{
				"instances": 1,
				"instance_type": { "id": "number-1" },
				"persistent_disk": { "size_mb": "290" },
				"internet_connected": true,
				"elb_names": ["something"]
			}`))
		})

		Context("when internet_connected property is false", func() {
			It("passes the value to the flag in the JSON request", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
				}, nil)

				service := api.NewJobsService(client)

				err := service.ConfigureJob("some-product-guid", "some-job-guid",
					api.JobProperties{
						Instances:         1,
						PersistentDisk:    &api.Disk{Size: "290"},
						InstanceType:      api.InstanceType{ID: "number-1"},
						InternetConnected: new(bool),
						LBNames:           []string{"something"},
					})
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DoCallCount()).To(Equal(1))
				request := client.DoArgsForCall(0)

				Expect("application/json").To(Equal(request.Header.Get("Content-Type")))
				Expect("PUT").To(Equal(request.Method))
				Expect("/api/v0/staged/products/some-product-guid/jobs/some-job-guid/resource_config").To(Equal(request.URL.Path))
				reqBytes, err := ioutil.ReadAll(request.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(reqBytes).To(MatchJSON(`{
				"instances": 1,
				"instance_type": { "id": "number-1" },
				"internet_connected": false,
				"persistent_disk": { "size_mb": "290" },
				"elb_names": ["something"]
			}`))
			})
		})

		Context("when floating_ips is specified", func() {
			It("passes the value to the flag in the JSON request", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
				}, nil)

				service := api.NewJobsService(client)

				err := service.ConfigureJob("some-product-guid", "some-job-guid",
					api.JobProperties{
						Instances:         1,
						PersistentDisk:    &api.Disk{Size: "290"},
						InstanceType:      api.InstanceType{ID: "number-1"},
						InternetConnected: new(bool),
						LBNames:           []string{"something"},
						FloatingIPs:       "fl.oa.ting.ip",
					})
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DoCallCount()).To(Equal(1))
				request := client.DoArgsForCall(0)

				Expect("application/json").To(Equal(request.Header.Get("Content-Type")))
				Expect("PUT").To(Equal(request.Method))
				Expect("/api/v0/staged/products/some-product-guid/jobs/some-job-guid/resource_config").To(Equal(request.URL.Path))
				reqBytes, err := ioutil.ReadAll(request.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(reqBytes).To(MatchJSON(`{
				"instances": 1,
				"instance_type": { "id": "number-1" },
				"internet_connected": false,
				"persistent_disk": { "size_mb": "290" },
				"elb_names": ["something"],
				"floating_ips": "fl.oa.ting.ip"
			}`))
			})
		})

		Context("when the internet_connected property is not passed", func() {
			It("does not pass the flag to the JSON request", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
				}, nil)

				service := api.NewJobsService(client)

				err := service.ConfigureJob("some-product-guid", "some-job-guid",
					api.JobProperties{
						Instances:      1,
						PersistentDisk: &api.Disk{Size: "290"},
						InstanceType:   api.InstanceType{ID: "number-1"},
						LBNames:        []string{"something"},
					})
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DoCallCount()).To(Equal(1))
				request := client.DoArgsForCall(0)

				Expect("application/json").To(Equal(request.Header.Get("Content-Type")))
				Expect("PUT").To(Equal(request.Method))
				Expect("/api/v0/staged/products/some-product-guid/jobs/some-job-guid/resource_config").To(Equal(request.URL.Path))
				reqBytes, err := ioutil.ReadAll(request.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(reqBytes).To(MatchJSON(`{
				"instances": 1,
				"instance_type": { "id": "number-1" },
				"persistent_disk": { "size_mb": "290" },
				"elb_names": ["something"]
			}`))
			})
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, errors.New("bad things"))

					service := api.NewJobsService(client)

					err := service.ConfigureJob("some-product-guid", "some-other-guid", api.JobProperties{
						Instances:      2,
						PersistentDisk: &api.Disk{Size: "000"},
						InstanceType:   api.InstanceType{ID: "number-2"},
					})
					Expect(err).To(MatchError("could not make api request to jobs resource_config endpoint: bad things"))
				})
			})

			Context("when the server returns a non-200 status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, nil)

					service := api.NewJobsService(client)

					err := service.ConfigureJob("some-product-guid", "some-other-guid", api.JobProperties{
						Instances:      2,
						PersistentDisk: &api.Disk{Size: "000"},
						InstanceType:   api.InstanceType{ID: "number-2"},
					})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response:")))
				})
			})
		})
	})
})
