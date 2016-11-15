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

			output, err := service.Jobs("some-product-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(api.JobsOutput{
				Jobs: []api.Job{
					{Name: "job-1", GUID: "some-guid"},
				}}))

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

	Describe("Configure", func() {
		It("configures job resources", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
			}, nil)

			service := api.NewJobsService(client)

			err := service.Configure(api.JobConfigurationInput{
				ProductGUID: "some-product-guid",
				Jobs: api.JobConfig{
					"some-job-guid": {
						Instances:         1,
						PersistentDisk:    &api.Disk{Size: "290"},
						InstanceType:      api.InstanceType{ID: "number-1"},
						InternetConnected: true,
						LBNames:           []string{"something"},
					},
					"some-other-guid": {
						Instances:      2,
						PersistentDisk: &api.Disk{Size: "000"},
						InstanceType:   api.InstanceType{ID: "number-2"},
					},
					"no-persistent-disk": {
						Instances:    2,
						InstanceType: api.InstanceType{ID: "number-2"},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())

			firstRequest := client.DoArgsForCall(0)
			secondRequest := client.DoArgsForCall(1)
			thirdRequest := client.DoArgsForCall(2)

			Expect("application/json").To(And(Equal(firstRequest.Header.Get("Content-Type")), Equal(secondRequest.Header.Get("Content-Type"))))
			Expect("PUT").To(And(Equal(firstRequest.Method), Equal(secondRequest.Method)))
			Expect("/api/v0/staged/products/some-product-guid/jobs/some-job-guid/resource_config").To(Or(Equal(firstRequest.URL.Path), Equal(secondRequest.URL.Path), Equal(thirdRequest.URL.Path)))
			Expect("/api/v0/staged/products/some-product-guid/jobs/some-other-guid/resource_config").To(Or(Equal(firstRequest.URL.Path), Equal(secondRequest.URL.Path), Equal(thirdRequest.URL.Path)))
			Expect("/api/v0/staged/products/some-product-guid/jobs/no-persistent-disk/resource_config").To(Or(Equal(firstRequest.URL.Path), Equal(secondRequest.URL.Path), Equal(thirdRequest.URL.Path)))

			responseBodyOne := `
			{
				"instances": 1,
				"instance_type": { "id": "number-1" },
				"persistent_disk": { "size_mb": "290" },
				"internet_connected": true,
				"elb_names": ["something"]
			}`

			responseBodyTwo := `
			{
				"instances": 2,
				"instance_type": { "id": "number-2" },
				"persistent_disk": { "size_mb": "000" },
				"internet_connected": false,
				"elb_names": null
			}`

			responseBodyThree := `
			{
				"instances": 2,
				"instance_type": { "id": "number-2" },
				"internet_connected": false,
				"elb_names": null
			}`

			respBytes, err := ioutil.ReadAll(firstRequest.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(respBytes)).To(Or(MatchJSON(responseBodyOne), MatchJSON(responseBodyTwo), MatchJSON(responseBodyThree)))

			respBytes, err = ioutil.ReadAll(secondRequest.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(respBytes)).To(Or(MatchJSON(responseBodyOne), MatchJSON(responseBodyTwo), MatchJSON(responseBodyThree)))

			respBytes, err = ioutil.ReadAll(thirdRequest.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(respBytes)).To(Or(MatchJSON(responseBodyOne), MatchJSON(responseBodyTwo), MatchJSON(responseBodyThree)))
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, errors.New("bad things"))

					service := api.NewJobsService(client)

					err := service.Configure(api.JobConfigurationInput{
						ProductGUID: "some-product-guid",
						Jobs: api.JobConfig{
							"some-other-guid": {
								Instances:      2,
								PersistentDisk: &api.Disk{Size: "000"},
								InstanceType:   api.InstanceType{ID: "number-2"},
							},
						},
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

					err := service.Configure(api.JobConfigurationInput{
						ProductGUID: "some-product-guid",
						Jobs: api.JobConfig{
							"some-other-guid": {
								Instances:      2,
								PersistentDisk: &api.Disk{Size: "000"},
								InstanceType:   api.InstanceType{ID: "number-2"},
							},
						},
					})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response:")))
				})
			})
		})
	})
})
