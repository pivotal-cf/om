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
		It("returns a listing of the jobs", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"jobs": [{"name":"job-1","guid":"some-guid"}]}`)),
			}, nil)

			service := api.NewJobsService(client)

			jobs, err := service.Jobs("some-product-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(jobs).To(Equal([]api.Job{
				{Name: "job-1", GUID: "some-guid"},
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
			Expect(job).To(Equal(api.JobProperties{
				Instances:         1,
				PersistentDisk:    &api.Disk{Size: "290"},
				InstanceType:      api.InstanceType{ID: "number-1"},
				InternetConnected: true,
				LBNames:           []string{"something"},
			}))
			request := client.DoArgsForCall(0)
			Expect("/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config").To(Equal(request.URL.Path))
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

			err := service.ConfigureJob("some-product-guid", "some-job-guid",
				api.JobProperties{
					Instances:         1,
					PersistentDisk:    &api.Disk{Size: "290"},
					InstanceType:      api.InstanceType{ID: "number-1"},
					InternetConnected: true,
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
				"persistent_disk": { "size_mb": "290" },
				"internet_connected": true,
				"elb_names": ["something"]
			}`))
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
