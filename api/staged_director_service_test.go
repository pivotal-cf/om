package api_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"
)

var _ = Describe("StagedProducts", func() {
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

	Describe("GetDirectorProperties", func() {
		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				var resp *http.Response
				switch req.URL.Path {
				case "/api/v0/staged/director/properties":
					resp = &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
"iaas_configuration": {
  "vcenter_host": "10.10.10.0",
  "datacenter": "my-data-center",
  "ephemeral_datastores_string": "e-datastore-name",
  "persistent_datastores_string": "p-datastore-name",
  "vcenter_username": "my-user-name",
  "bosh_vm_folder": "bosh-folder",
  "bosh_template_folder": "my-bosh-template-folder",
  "bosh_disk_path": "my-disk-location",
  "ssl_verification_enabled": false,
  "nsx_networking_enabled": true,
  "nsx_mode": "nsx-v",
  "nsx_address": "10.10.10.10",
  "nsx_username": "mysterious-gremlin",
  "nsx_ca_certificate": "-----BEGIN CERTIFICATE-----\r\nMIIBsjCCARugmeow..."
},
"director_configuration": {
  "ntp_servers_string": "us.pool.ntp.org, time.google.com",
  "metrics_ip": null,
  "resurrector_enabled": false,
  "director_hostname": "hal9000.tld",
  "max_threads": 5,
  "disable_dns_release": false,
  "custom_ssh_banner": "Hello World!",
  "opentsdb_ip": "1.2.3.4",
  "director_worker_count": 5,
  "post_deploy_enabled": false,
  "bosh_recreate_on_next_deploy": false,
  "retry_bosh_deploys": false,
  "keep_unreachable_vms": false,
  "database_type": "internal",
  "hm_pager_duty_options": {"enabled": false},
  "hm_emailer_options": {"enabled": false},
  "blobstore_type": "local",
  "encryption": {
    "keys": [],
    "providers": []
  },
  "excluded_recursors": ["8.8.8.8"]
},
"security_configuration": {
  "trusted_certificates": null,
  "generate_vm_passwords": true
},
"syslog_configuration": {
  "enabled": true,
  "address": "1.2.3.4",
  "port": "514",
  "transport_protocol": "tcp",
  "tls_enabled": true,
  "permitted_peer": "*.example.com",
  "ssl_ca_certificate": "-----BEGIN CERTIFICATE-----\r\nMIIBsjCCARug..."
}}`,
						)),
					}
				}
				return resp, nil
			}
		})

		It("returns all the special properties for the Director", func() {
			config, err := service.GetStagedDirectorProperties()
			Expect(err).NotTo(HaveOccurred())

			Expect(config["iaas_configuration"]).To(Equal(map[string]interface{}{
				"vcenter_host":                 "10.10.10.0",
				"datacenter":                   "my-data-center",
				"ephemeral_datastores_string":  "e-datastore-name",
				"persistent_datastores_string": "p-datastore-name",
				"vcenter_username":             "my-user-name",
				"bosh_vm_folder":               "bosh-folder",
				"bosh_template_folder":         "my-bosh-template-folder",
				"bosh_disk_path":               "my-disk-location",
				"ssl_verification_enabled":     false,
				"nsx_networking_enabled":       true,
				"nsx_mode":                     "nsx-v",
				"nsx_address":                  "10.10.10.10",
				"nsx_username":                 "mysterious-gremlin",
				"nsx_ca_certificate":           "-----BEGIN CERTIFICATE-----\r\nMIIBsjCCARugmeow...",
			}))
			Expect(config["director_configuration"]).To(Equal(map[string]interface{}{
				"ntp_servers_string":           "us.pool.ntp.org, time.google.com",
				"metrics_ip":                   nil,
				"resurrector_enabled":          false,
				"director_hostname":            "hal9000.tld",
				"max_threads":                  5,
				"disable_dns_release":          false,
				"custom_ssh_banner":            "Hello World!",
				"opentsdb_ip":                  "1.2.3.4",
				"director_worker_count":        5,
				"post_deploy_enabled":          false,
				"bosh_recreate_on_next_deploy": false,
				"retry_bosh_deploys":           false,
				"keep_unreachable_vms":         false,
				"database_type":                "internal",
				"hm_pager_duty_options": map[interface{}]interface{}{
					"enabled": false,
				},
				"hm_emailer_options": map[interface{}]interface{}{
					"enabled": false,
				},
				"blobstore_type": "local",
				"encryption": map[interface{}]interface{}{
					"keys":      []interface{}{},
					"providers": []interface{}{},
				},
				"excluded_recursors": []interface{}{"8.8.8.8"},
			},
			))
			Expect(config["security_configuration"]).To(Equal(map[string]interface{}{
				"trusted_certificates":  nil,
				"generate_vm_passwords": true,
			},
			))
			Expect(config["syslog_configuration"]).To(Equal(map[string]interface{}{
				"enabled":            true,
				"address":            "1.2.3.4",
				"port":               "514",
				"transport_protocol": "tcp",
				"tls_enabled":        true,
				"permitted_peer":     "*.example.com",
				"ssl_ca_certificate": "-----BEGIN CERTIFICATE-----\r\nMIIBsjCCARug...",
			},
			))
		})

		Context("failure cases", func() {
			Context("when the properties request returns an error", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/director/properties":
							return &http.Response{}, errors.New("some-error")
						}
						return resp, nil
					}
				})
				It("returns an error", func() {
					_, err := service.GetStagedDirectorProperties()
					Expect(err).To(MatchError(`could not send api request to GET /api/v0/staged/director/properties: some-error`))
				})
			})

			Context("when the properties request returns a non 200 error code", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/director/properties":
							return &http.Response{
								StatusCode: http.StatusTeapot,
								Body:       ioutil.NopCloser(bytes.NewBufferString("")),
							}, nil
						}
						return resp, nil
					}
				})
				It("returns an error", func() {
					_, err := service.GetStagedDirectorProperties()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the server returns invalid json", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/director/properties":
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(bytes.NewBufferString(`{{{`)),
							}
						}
						return resp, nil
					}
				})

				It("returns an error", func() {
					_, err := service.GetStagedDirectorProperties()
					Expect(err).To(MatchError(ContainSubstring("could not parse json")))
				})
			})
		})
	})

	Describe("GetStagedDirectorAvailabilityZones", func() {
		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				var resp *http.Response
				switch req.URL.Path {
				case "/api/v0/staged/director/availability_zones":
					resp = &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`
{
  "availability_zones": [
    {
      "name": "Availability Zone 1",
      "guid": "guid-1"
    },
    {
      "name": "Availability Zone 2",
      "guid": "guid-4"
    }
  ]
}`,
						)),
					}
				}
				return resp, nil
			}
		})
		It("returns all the configured availability zones", func() {
			config, err := service.GetStagedDirectorAvailabilityZones()
			Expect(err).NotTo(HaveOccurred())

			Expect(config["availability_zones"]).To(Equal([]map[string]interface{}{
				{
					"name": "Availability Zone 1",
					"guid": "guid-1",
				},
				{
					"name": "Availability Zone 2",
					"guid": "guid-4",
				},
			}))
		})
		Describe("failure cases", func() {
			Context("when the properties request returns an error", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/director/availability_zones":
							return &http.Response{}, errors.New("some-error")
						}
						return resp, nil
					}
				})
				It("returns an error", func() {
					_, err := service.GetStagedDirectorAvailabilityZones()
					Expect(err).To(MatchError(`could not send api request to GET /api/v0/staged/director/availability_zones: some-error`))
				})
			})

			Context("when the properties request returns a non 200 error code", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/director/availability_zones":
							return &http.Response{
								StatusCode: http.StatusTeapot,
								Body:       ioutil.NopCloser(bytes.NewBufferString("")),
							}, nil
						}
						return resp, nil
					}
				})
				It("returns an error", func() {
					_, err := service.GetStagedDirectorAvailabilityZones()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the server returns invalid json", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/director/availability_zones":
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(bytes.NewBufferString(`{{{`)),
							}
						}
						return resp, nil
					}
				})

				It("returns an error", func() {
					_, err := service.GetStagedDirectorAvailabilityZones()
					Expect(err).To(MatchError(ContainSubstring("could not parse json")))
				})
			})
		})
	})

	Describe("GetStagedDirectorNetworks", func() {
		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				var resp *http.Response
				switch req.URL.Path {
				case "/api/v0/staged/director/networks":
					resp = &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`
{
  "icmp_checks_enabled": true,
  "networks": [
    {
      "guid": "0d35c70db3c592cb1ac7",
      "name": "first-network",
      "subnets": [
        {
          "guid": "433d16d727706e3be752",
          "iaas_identifier": "hinterlands-1",
          "cidr": "10.85.41.0/24",
          "dns": "10.87.8.10",
          "gateway": "10.85.41.1",
          "reserved_ip_ranges": "10.85.41.1-10.85.41.97,10.85.41.117-10.85.41.255",
          "availability_zone_names": [
            "first-az",
            "second-az"
          ]
        }
      ]
    }
  ]
}`,
						)),
					}
				}
				return resp, nil
			}
		})

		It("returns all the configured networks", func() {
			config, err := service.GetStagedDirectorNetworks()
			Expect(err).NotTo(HaveOccurred())

			Expect(config["icmp_checks_enabled"]).To(Equal(true))

			Expect(config["networks"]).To(ContainElement(Equal(
				map[interface{}]interface{} {
					"name": "first-network",
					"guid": "0d35c70db3c592cb1ac7",
					"subnets": []interface{}{
						map[interface{}]interface{} {"guid": "433d16d727706e3be752",
							"iaas_identifier":    "hinterlands-1",
							"cidr":               "10.85.41.0/24",
							"dns":                "10.87.8.10",
							"gateway":            "10.85.41.1",
							"reserved_ip_ranges": "10.85.41.1-10.85.41.97,10.85.41.117-10.85.41.255",
							"availability_zone_names": []interface{}{
								"first-az",
								"second-az",
							}}},
				},
			)))
		})

		Context("failure cases", func() {
			Context("when the properties request returns an error", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/director/networks":
							return &http.Response{}, errors.New("some-error")
						}
						return resp, nil
					}
				})
				It("returns an error", func() {
					_, err := service.GetStagedDirectorNetworks()
					Expect(err).To(MatchError(`could not send api request to GET /api/v0/staged/director/networks: some-error`))
				})
			})

			Context("when the properties request returns a non 200 error code", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/director/networks":
							return &http.Response{
								StatusCode: http.StatusTeapot,
								Body:       ioutil.NopCloser(bytes.NewBufferString("")),
							}, nil
						}
						return resp, nil
					}
				})
				It("returns an error", func() {
					_, err := service.GetStagedDirectorNetworks()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the server returns invalid json", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/director/networks":
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(bytes.NewBufferString(`{{{`)),
							}
						}
						return resp, nil
					}
				})

				It("returns an error", func() {
					_, err := service.GetStagedDirectorNetworks()
					Expect(err).To(MatchError(ContainSubstring("could not parse json")))
				})
			})
		})
	})
})
