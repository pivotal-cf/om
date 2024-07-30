package api_test

import (
	"net/http"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/api"
)

var _ = Describe("StagedProducts", func() {
	var (
		server  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		service = api.New(api.ApiInput{
			Client: httpClient{
				server.URL(),
			},
		})
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("GetDirectorProperties", func() {
		It("returns all the special properties for the Director", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/properties", "redact=true"),
					ghttp.RespondWith(http.StatusOK, `{
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
						}
					}`),
				),
			)

			config, err := service.GetStagedDirectorProperties(true)
			Expect(err).ToNot(HaveOccurred())

			Expect(config["iaas_configuration"]).To(Equal(map[interface{}]interface{}{
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
			Expect(config["director_configuration"]).To(Equal(map[interface{}]interface{}{
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
			Expect(config["security_configuration"]).To(Equal(map[interface{}]interface{}{
				"trusted_certificates":  nil,
				"generate_vm_passwords": true,
			},
			))
			Expect(config["syslog_configuration"]).To(Equal(map[interface{}]interface{}{
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

		It("disables redaction", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/properties", "redact=false"),
					ghttp.RespondWith(http.StatusOK, `{
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
						}
					}`),
				),
			)

			_, err := service.GetStagedDirectorProperties(false)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("failure cases", func() {
			When("the properties request returns an error", func() {
				It("returns an error", func() {
					server.Close()

					_, err := service.GetStagedDirectorProperties(false)
					Expect(err).To(MatchError(ContainSubstring(`could not send api request to GET /api/v0/staged/director/properties?redact=false`)))
				})
			})

			When("the properties request returns a non 200 error code", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/director/properties"),
							ghttp.RespondWith(http.StatusTeapot, ""),
						),
					)

					_, err := service.GetStagedDirectorProperties(false)
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response from /api/v0/staged/director/properties")))
				})
			})

			When("the server returns invalid json", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/director/properties"),
							ghttp.RespondWith(http.StatusOK, "{{{"),
						),
					)

					_, err := service.GetStagedDirectorProperties(false)
					Expect(err).To(MatchError(ContainSubstring("could not parse json")))
				})
			})
		})
	})

	Describe("GetStagedDirectorIaasConfigurations", func() {
		It("returns all the special properties for the Director", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations", "redact=true"),
					ghttp.RespondWith(http.StatusOK, `{
						"iaas_configurations": [{
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
						},{
							"guid": "some-guid",
							"name": "default",
							"project": "my-google-project",
							"associated_service_account": "my-google-service-account",
							"auth_json": "****"
						}]
					}`),
				),
			)

			config, err := service.GetStagedDirectorIaasConfigurations(true)
			Expect(err).ToNot(HaveOccurred())

			Expect(config["iaas_configurations"]).Should(Equal([]map[string]interface{}{
				{
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
				},
				{
					"guid":                       "some-guid",
					"name":                       "default",
					"project":                    "my-google-project",
					"associated_service_account": "my-google-service-account",
					"auth_json":                  "****",
				},
			}))
		})

		It("disables redaction", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations", "redact=false"),
					ghttp.RespondWith(http.StatusOK, `{
						"iaas_configurations": [{
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
						},{
							"guid": "some-guid",
							"name": "default",
							"project": "my-google-project",
							"associated_service_account": "my-google-service-account",
							"auth_json": "****"
						}]
					}`),
				),
			)

			_, err := service.GetStagedDirectorIaasConfigurations(false)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("with nested json objects in response", func() {
			It("returns the fields", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
						ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "some-guid",
								"name": "default",
								"subscription_id": "my-subscription",
								"tenant_id": "my-tenant",
								"client_id": "my-client",
								"resource_group_name": "my-resource-group",
								"cloud_storage_type": "managed_disks",
								"bosh_storage_account_name": "storage-account-bosh",
								"storage_account_type": "Standard_LRS",
								"deployments_storage_account_name": null,
								"default_security_group": "my-security-group",
								"ssh_public_key": "ssh-rsa ...",
								"environment": "AzureStack",
								"azure_stack": {
									"resource": "https://management.somedomain.onmicrosoft.com/some-guid",
									"domain": "subdomain.somedomain.onmicrosoft.com",
									"authentication": "AzureAD",
									"endpoint_prefix": "management",
									"ca_cert": "-----BEGIN CERTIFICATE-----\nMIIJKgIBAAKCAgE..."
								}
							}]
						}`),
					),
				)

				config, err := service.GetStagedDirectorIaasConfigurations(true)
				Expect(err).ToNot(HaveOccurred())

				Expect(config["iaas_configurations"]).Should(Equal([]map[string]interface{}{
					{
						"guid":                             "some-guid",
						"name":                             "default",
						"subscription_id":                  "my-subscription",
						"tenant_id":                        "my-tenant",
						"client_id":                        "my-client",
						"resource_group_name":              "my-resource-group",
						"cloud_storage_type":               "managed_disks",
						"bosh_storage_account_name":        "storage-account-bosh",
						"storage_account_type":             "Standard_LRS",
						"deployments_storage_account_name": nil,
						"default_security_group":           "my-security-group",
						"ssh_public_key":                   "ssh-rsa ...",
						"environment":                      "AzureStack",
						"azure_stack": map[interface{}]interface{}{
							"resource":        "https://management.somedomain.onmicrosoft.com/some-guid",
							"domain":          "subdomain.somedomain.onmicrosoft.com",
							"authentication":  "AzureAD",
							"endpoint_prefix": "management",
							"ca_cert":         "-----BEGIN CERTIFICATE-----\nMIIJKgIBAAKCAgE...",
						},
					},
				}))
			})
		})

		Context("not found", func() {
			It("returns an empty response", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
						ghttp.RespondWith(http.StatusNotFound, "{not-found}"),
					),
				)

				resp, err := service.GetStagedDirectorIaasConfigurations(false)
				Expect(resp).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("failure cases", func() {
			When("the properties request returns a non 200 error code", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
							ghttp.RespondWith(http.StatusTeapot, ""),
						),
					)

					_, err := service.GetStagedDirectorIaasConfigurations(false)
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response from /api/v0/staged/director/iaas_configurations")))
				})
			})

			When("the server returns invalid json", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
							ghttp.RespondWith(http.StatusOK, "{{{"),
						),
					)

					_, err := service.GetStagedDirectorIaasConfigurations(false)
					Expect(err).To(MatchError(ContainSubstring("could not parse json")))
				})
			})
		})
	})

	Describe("GetStagedDirectorAvailabilityZones", func() {
		It("returns all the configured availability zones", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
					ghttp.RespondWith(http.StatusOK, `{
					"availability_zones": [
						{
							"name": "Availability Zone 1",
							"guid": "guid-1",
							"iaas_configuration_guid": "iaas-configuration-guid-one"
						}, {
							"name": "Availability Zone 2",
							"guid": "guid-4",
							"iaas_configuration_guid": "iaas-configuration-guid-two"
						}
					]}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
					ghttp.RespondWith(http.StatusOK, `{
						"iaas_configurations": [{
							"guid": "iaas-configuration-guid-one",
							"name": "iaas-configuration-one"
						},{
							"guid": "iaas-configuration-guid-two",
							"name": "iaas-configuration-two"
						}]
					}`),
				),
			)

			config, err := service.GetStagedDirectorAvailabilityZones()
			Expect(err).ToNot(HaveOccurred())

			Expect(config.AvailabilityZones).To(Equal([]api.AvailabilityZoneOutput{
				{
					Name:                  "Availability Zone 1",
					IAASConfigurationName: "iaas-configuration-one",
					Fields:                map[string]interface{}{"guid": "guid-1"},
				},
				{
					Name:                  "Availability Zone 2",
					IAASConfigurationName: "iaas-configuration-two",
					Fields:                map[string]interface{}{"guid": "guid-4"},
				},
			}))
		})

		It("returns an empty list when status code is 405", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
					ghttp.RespondWith(http.StatusMethodNotAllowed, ``),
				),
			)

			config, err := service.GetStagedDirectorAvailabilityZones()
			Expect(err).ToNot(HaveOccurred())

			Expect(config).To(Equal(api.AvailabilityZonesOutput{}))

		})
		Describe("failure cases", func() {
			When("the properties request returns an error", func() {
				It("returns an error", func() {
					server.Close()

					_, err := service.GetStagedDirectorAvailabilityZones()
					Expect(err).To(MatchError(ContainSubstring(`could not send api request to GET /api/v0/staged/director/availability_zones`)))
				})
			})

			When("the properties request returns a non 200 error code", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
							ghttp.RespondWith(http.StatusTeapot, ""),
						),
					)

					_, err := service.GetStagedDirectorAvailabilityZones()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response from /api/v0/staged/director/availability_zones")))
				})
			})

			When("the server returns invalid json", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
							ghttp.RespondWith(http.StatusOK, "{{{"),
						),
					)
					_, err := service.GetStagedDirectorAvailabilityZones()
					Expect(err).To(MatchError(ContainSubstring("could not parse json")))
				})
			})

			When("iaas configurations request returns a non 200 error code", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
							ghttp.RespondWith(http.StatusTeapot, `{}`),
						),
					)

					_, err := service.GetStagedDirectorAvailabilityZones()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response from /api/v0/staged/director/iaas_configurations")))
				})
			})

			When("the server returns invalid json", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
							ghttp.RespondWith(http.StatusOK, "{{{"),
						),
					)
					_, err := service.GetStagedDirectorAvailabilityZones()
					Expect(err).To(MatchError(ContainSubstring("could not parse json")))
				})
			})
		})
	})

	Describe("GetStagedDirectorNetworks", func() {
		It("returns all the configured networks", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/networks"),
					ghttp.RespondWith(http.StatusOK, `{
						"icmp_checks_enabled": true,
						"networks": [{
							"guid": "0d35c70db3c592cb1ac7",
							"name": "first-network",
							"subnets": [{
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
							}]
						}]
					}`),
				),
			)

			config, err := service.GetStagedDirectorNetworks()
			Expect(err).ToNot(HaveOccurred())

			Expect(config.ICMP).To(Equal(true))

			Expect(config.Networks).To(ContainElement(Equal(
				api.NetworkConfigurationOutput{
					Name: "first-network",
					Subnets: []api.SubnetOutput{
						api.SubnetOutput{
							IAASIdentifier:   "hinterlands-1",
							CIDR:             "10.85.41.0/24",
							DNS:              "10.87.8.10",
							Gateway:          "10.85.41.1",
							ReservedIPRanges: "10.85.41.1-10.85.41.97,10.85.41.117-10.85.41.255",
							AvailabilityZones: []string{
								"first-az",
								"second-az",
							},
						},
					},
				},
			)))
		})

		Context("failure cases", func() {
			When("the properties request returns an error", func() {
				It("returns an error", func() {
					server.Close()

					_, err := service.GetStagedDirectorNetworks()
					Expect(err).To(MatchError(ContainSubstring(`could not send api request to GET /api/v0/staged/director/networks`)))
				})
			})

			When("the properties request returns a non 200 error code", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/director/networks"),
							ghttp.RespondWith(http.StatusTeapot, ""),
						),
					)

					_, err := service.GetStagedDirectorNetworks()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response from /api/v0/staged/director/networks")))
				})
			})

			When("the server returns invalid json", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/director/networks"),
							ghttp.RespondWith(http.StatusOK, "{{{"),
						),
					)

					_, err := service.GetStagedDirectorNetworks()
					Expect(err).To(MatchError(ContainSubstring("could not parse json")))
				})
			})
		})
	})
})
