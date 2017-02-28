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

const AZDocument = `
<html>
	<body>
		<form action="/some/action" method="some-method">
			<input name="_method" value="some-rails" />
			<input name="authenticity_token" value="some-authenticity" />
			<div class="content">
				<input name="availability_zones[availability_zones][][iaas_identifier]" value="some-az-name-1" \>
				<input name="availability_zones[availability_zones][][iaas_identifier]" value="some-az-name-2" \>
			</div>
			<input name="availability_zones[availability_zones][][iaas_identifier]" value="some-az-name-1" \>
			<input name="availability_zones[availability_zones][][iaas_identifier]" value="some-az-name-2" \>
			<input name="availability_zones[availability_zones][][guid]" value="also-do-not-want" \>
			<input name="availability_zones[availability_zones][][guid]" type="hidden" value="some-az-guid-1" \>
			<input name="availability_zones[availability_zones][][guid]" type="hidden" value="some-az-guid-2" \>
		</form>
	</body>
</html>`

const NoAZDocument = `
<html>
	<body>
		<form action="/some/action" method="some-method">
			<input name="_method" value="some-rails" />
			<input name="authenticity_token" value="some-authenticity" />
		</form>
	</body>
</html>`

const VsphereAZDocument = `
<html>
	<body>
		<form action="/some/action" method="some-method">
			<input name="_method" value="some-rails" />
			<input name="authenticity_token" value="some-authenticity" />
			<input name="availability_zones[availability_zones][][name]" value="some-az-name-1" \>
			<input name="availability_zones[availability_zones][][cluster]" value="some-cluster-name-1" \>
			<input name="availability_zones[availability_zones][][resource_pool]" value="some-resource-pool-name-1" \>
			<input name="availability_zones[availability_zones][][name]" value="some-az-name-2" \>
			<input name="availability_zones[availability_zones][][cluster]" value="some-cluster-name-2" \>
			<input name="availability_zones[availability_zones][][resource_pool]" value="some-resource-pool-name-2" \>
			<input name="availability_zones[availability_zones][][guid]" value="also-do-not-want" \>
			<input name="availability_zones[availability_zones][][guid]" type="hidden" value="some-az-guid-1" \>
			<input name="availability_zones[availability_zones][][guid]" type="hidden" value="some-az-guid-2" \>
		</form>
	</body>
</html>`

const AWSNetworkDocument = `
<html>
	<body>
		<form action="/some/action" method="some-method">
			<input name="_method" value="some-rails" />
			<input name="authenticity_token" value="some-authenticity" />
			<div class="controls">
			  <label for="network_collection_networks_attributes_0_subnets_0_availability_zone_references_some-guid">
				  <input id="network_collection_networks_attributes_0_subnets_0_availability_zone_references_some-guid" type="radio" value="some-guid">
					us-west-1b
				</label>
			  <label for="network_collection_networks_attributes_0_subnets_0_availability_zone_references_some-other-guid">
				  <input id="network_collection_networks_attributes_0_subnets_0_availability_zone_references_some-other-guid" type="radio" value="some-other-guid">
					us-west-1c
				</label>
			</div>
		</form>
	</body>
</html>`

const NetDocument = `
<html>
	<body>
		<form action="/some/action" method="some-method">
			<input name="_method" value="some-rails" />
			<input name="authenticity_token" value="some-authenticity" />
			<select name="bosh_product[network_reference]" id="bosh_product_network_reference">
				<option value=""></option>
				<option value="ed9b4dcf24dad744b1cf">some-net</option>
			</select>
		</form>
	</body>
</html>`

const mistmatchedForm = `
<html>
	<body>
		<form action="/some/action" method="some-method">
			<input name="_method" value="some-rails" />
			<input name="authenticity_token" value="some-authenticity" />
			<input name="availability_zones[availability_zones][][iaas_identifier]" value="some-az-name-2" \>
			<input name="availability_zones[availability_zones][][guid]" type="hidden" value="some-az-guid-1" \>
			<input name="availability_zones[availability_zones][][guid]" type="hidden" value="some-az-guid-2" \>
		</form>
	</body>
</html>`

var _ = Describe("BoshFormService", func() {
	var (
		service api.BoshFormService
		client  *fakes.HttpClient
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		service = api.NewBoshFormService(client)
	})

	Describe("GetForm", func() {
		It("returns the form details", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(AZDocument)),
			}, nil)

			form, err := service.GetForm("/some/path")
			Expect(err).NotTo(HaveOccurred())

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.Path).To(Equal("/some/path"))

			Expect(form).To(Equal(api.Form{
				Action:            "/some/action",
				AuthenticityToken: "some-authenticity",
				RailsMethod:       "some-rails",
			}))
		})

		Context("when an error occurs", func() {
			Context("when a request cannot be constructed", func() {
				It("returns an error", func() {
					_, err := service.GetForm("%%%%")
					Expect(err).To(MatchError(ContainSubstring(`invalid URL escape`)))
				})
			})

			Context("when http client fails", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("whoops"))

					_, err := service.GetForm("")
					Expect(err).To(MatchError("failed during request: whoops"))
				})
			})

			Context("when authenticity token cannot be found", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := service.GetForm("")
					Expect(err).To(MatchError("could not find the form authenticity token"))
				})
			})

			Context("when the response status is non-200", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := service.GetForm("")
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("PostForm", func() {
		It("submits the form content", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("")),
			}, nil)

			input := api.PostFormInput{
				Form: api.Form{
					Action: "/some/action",
				},
				EncodedPayload: "some-payload",
			}

			err := service.PostForm(input)
			Expect(err).NotTo(HaveOccurred())

			req := client.DoArgsForCall(0)
			Expect(req.Method).To(Equal("POST"))
			Expect(req.URL.Path).To(Equal("/some/action"))
			Expect(req.Header.Get("Content-Type")).To(Equal("application/x-www-form-urlencoded"))

			bodyBytes, err := ioutil.ReadAll(req.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(bodyBytes)).To(Equal("some-payload"))
		})

		Context("when an error occurs", func() {
			Context("when a request cannot be constructed", func() {
				It("returns an error", func() {
					input := api.PostFormInput{
						Form: api.Form{
							Action: "%%%%",
						},
					}

					err := service.PostForm(input)
					Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
				})
			})

			Context("when the client fails during the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("some error"))

					err := service.PostForm(api.PostFormInput{})
					Expect(err).To(MatchError("failed to POST form: some error"))
				})
			})

			Context("when the request responds with a non-200 status", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					err := service.PostForm(api.PostFormInput{})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("AvailabilityZones", func() {
		It("returns a map of availability zone information", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(AZDocument)),
			}, nil)

			netMap, err := service.AvailabilityZones()
			Expect(err).NotTo(HaveOccurred())

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.Path).To(Equal("/infrastructure/availability_zones/edit"))

			Expect(netMap).To(HaveKeyWithValue("some-az-name-1", "some-az-guid-1"))
			Expect(netMap).To(HaveKeyWithValue("some-az-name-2", "some-az-guid-2"))
		})

		It("returns a map of vsphere availability zone information", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(VsphereAZDocument)),
			}, nil)

			netMap, err := service.AvailabilityZones()
			Expect(err).NotTo(HaveOccurred())

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.Path).To(Equal("/infrastructure/availability_zones/edit"))

			Expect(netMap).To(HaveKeyWithValue("some-az-name-1", "some-az-guid-1"))
			Expect(netMap).To(HaveKeyWithValue("some-az-name-2", "some-az-guid-2"))
		})

		Context("when there are no AZs", func() {
			It("returns successfully", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(NoAZDocument)),
				}, nil)

				azMap, err := service.AvailabilityZones()
				Expect(err).NotTo(HaveOccurred())
				Expect(azMap).To(BeEmpty())
			})
		})

		Context("failure cases", func() {
			Context("when http client fails", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("whoops"))

					_, err := service.AvailabilityZones()
					Expect(err).To(MatchError("failed during request: whoops"))
				})
			})

			Context("when the request responds with a non-200 status", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := service.AvailabilityZones()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("Networks", func() {
		Context("when the networks are selectors", func() {
			It("returns a map of networks", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(NetDocument)),
				}, nil)

				netMap, err := service.Networks()
				Expect(err).NotTo(HaveOccurred())

				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("GET"))
				Expect(req.URL.Path).To(Equal("/infrastructure/director/az_and_network_assignment/edit"))

				Expect(netMap).To(HaveKeyWithValue("some-net", "ed9b4dcf24dad744b1cf"))
			})
		})

		Context("when the networks are radio buttons", func() {
			It("returns a map of networks", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(AWSNetworkDocument)),
				}, nil)

				netMap, err := service.Networks()
				Expect(err).NotTo(HaveOccurred())

				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("GET"))
				Expect(req.URL.Path).To(Equal("/infrastructure/director/az_and_network_assignment/edit"))

				Expect(netMap).To(HaveKeyWithValue("us-west-1b", "some-guid"))
				Expect(netMap).To(HaveKeyWithValue("us-west-1c", "some-other-guid"))
			})
		})

		Context("failure cases", func() {
			Context("when http client fails", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("whoops"))

					_, err := service.Networks()
					Expect(err).To(MatchError("failed during request: whoops"))
				})
			})

			Context("when the request responds with a non-200 status", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := service.Networks()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})
})
