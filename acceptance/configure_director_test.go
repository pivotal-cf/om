package acceptance

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("configure-director command", func() {
	var (
		azPutCallCount               int
		azPostCallCount              int
		azPostConfigurationBody      []byte
		azPutConfigurationBody       []byte
		azPostConfigurationMethod    string
		azPutConfigurationMethod     string
		directorPropertiesBody       []byte
		directorPropertiesCallCount  int
		directorPropertiesMethod     string
		networkAZCallCount           int
		networkAZConfigurationBody   []byte
		networkAZConfigurationMethod string
		networksConfigurationBody    []byte
		networksPutCallCount         int
		resourceConfigMethod         string
		resourceConfigBody           []byte
		verifierErrorOccured         bool

		server *httptest.Server
	)

	BeforeEach(func() {
		azPutCallCount = 0
		directorPropertiesCallCount = 0
		networkAZCallCount = 0
		networksPutCallCount = 0
		verifierErrorOccured = false

		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/api/v0/installations":
				_, err := w.Write([]byte(`{"installations": []}`))
				Expect(err).ToNot(HaveOccurred())
			case "/uaa/oauth/token":
				username := req.FormValue("username")

				if username == "some-username" {
					_, err := w.Write([]byte(`{
						"access_token": "some-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`))
					Expect(err).ToNot(HaveOccurred())
				} else {
					_, err := w.Write([]byte(`{
						"access_token": "some-running-install-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`))
					Expect(err).ToNot(HaveOccurred())
				}
			case "/api/v0/staged/director/availability_zones/existing-az-guid":
				var err error
				azPostCallCount++
				if verifierErrorOccured {
					w.WriteHeader(207)
				}
				azPutConfigurationMethod = req.Method
				azPutConfigurationBody, err = ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				_, err = w.Write([]byte(`{}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/director/availability_zones":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" && auth != "Bearer some-running-install-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				azPostConfigurationMethod = req.Method

				if req.Method == "GET" {
					_, err := w.Write([]byte(`"availability_zones": [{"guid": "existing-az-guid", "name": "some-existing-az"}]`))
					Expect(err).ToNot(HaveOccurred())
				} else if req.Method == "POST" {
					var err error
					azPostConfigurationBody, err = ioutil.ReadAll(req.Body)
					Expect(err).NotTo(HaveOccurred())

					azPutCallCount++

					if verifierErrorOccured {
						w.WriteHeader(207)
					}
					_, err = w.Write([]byte(`{}`))
					Expect(err).ToNot(HaveOccurred())
				} else {
					out, err := httputil.DumpRequest(req, true)
					Expect(err).NotTo(HaveOccurred())
					Fail(fmt.Sprintf("unexpected request: %s", out))
				}

			case "/api/v0/staged/director/networks":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" && auth != "Bearer some-running-install-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if req.Method == "GET" {
					_, err := w.Write([]byte(`"networks": [{"guid": "existing-network-guid", "name": "network-1"}]`))
					Expect(err).ToNot(HaveOccurred())
				} else if req.Method == "PUT" {
					var err error
					networksConfigurationBody, err = ioutil.ReadAll(req.Body)
					Expect(err).NotTo(HaveOccurred())

					networksPutCallCount++

					_, err = w.Write([]byte(`{}`))
					Expect(err).ToNot(HaveOccurred())
				}
			case "/api/v0/staged/director/network_and_az":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" && auth != "Bearer some-running-install-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				if azPutCallCount == 0 || networksPutCallCount == 0 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				var err error
				networkAZConfigurationBody, err = ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				networkAZConfigurationMethod = req.Method

				networkAZCallCount++

				_, err = w.Write([]byte(`{}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/director/properties":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" && auth != "Bearer some-running-install-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				var err error
				directorPropertiesBody, err = ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				directorPropertiesMethod = req.Method

				directorPropertiesCallCount++

				_, err = w.Write([]byte(`{}`))
				Expect(err).ToNot(HaveOccurred())

			case "/api/v0/staged/products":
				_, err := w.Write([]byte(`[
					{
						"installation_name": "component-type1-installation-name",
						"guid": "component-type1-guid",
						"type": "component-type1"
					},
					{
						"installation_name": "p-bosh-installation-name",
						"guid": "p-bosh-guid",
						"type": "p-bosh"
					}
				]`))
				Expect(err).ToNot(HaveOccurred())

			case "/api/v0/staged/products/p-bosh-guid/jobs":
				_, err := w.Write([]byte(`
					{
						"jobs": [
							{
								"name": "compilation",
								"guid": "compilation-guid"
							}
						]
					}
				`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/products/p-bosh-guid/jobs/compilation-guid/resource_config":
				if req.Method == "GET" {
					_, err := w.Write([]byte(`{
						"instances": 1,
						"instance_type": {
							"id": "automatic"
						},
						"persistent_disk": {
							"size_mb": "20480"
						},
						"internet_connected": true,
						"elb_names": ["my-elb"]
					}`))
					Expect(err).ToNot(HaveOccurred())
					return
				}

				var err error
				resourceConfigMethod = req.Method
				resourceConfigBody, err = ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				_, err = w.Write([]byte(`{}`))
				Expect(err).ToNot(HaveOccurred())

			case "/api/v0/deployed/director/credentials":
				w.WriteHeader(http.StatusNotFound)

			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	AfterEach(func() {
		server.Close()
	})

	When("using command line arguments", func() {
		It("displays a helpful error message when using moved director properties", func() {
			configFile, err := ioutil.TempFile("", "config.yml")
			Expect(err).ToNot(HaveOccurred())
			_, err = configFile.WriteString(`{
            "iaas-configuration": {
				"project": "some-project",
				"default_deployment_tag": "my-vms",
				"auth_json": "{\"some-auth-field\": \"some-service-key\",\"some-private_key\": \"some-key\"}"
			},
            "az-configuration": [ {"name": "some-az-1"}, {"name": "some-existing-az"} ],
            "networks-configuration": {
				"networks": [{"name": "network-1"}],
				"top-level": "the-top"
			},
            "network-assignment": {
				"network": { "name": "some-network"},
				"singleton_availability_zone": {"name": "some-az"}
			},
            "director-configuration": {
				"ntp_servers_string": "us.example.org, time.something.com",
				"resurrector_enabled": false,
				"director_hostname": "foo.example.com",
				"max_threads": 5
			},
            "security-configuration": {
				"trusted_certificates": "some-certificate",
				"generate_vm_passwords": true
			},
            "syslog-configuration": {
				"syslogconfig": "awesome"
			},
            "resource-configuration": {
				"compilation": {
					"instance_type": {
						"id": "m4.xlarge"
					}
				}
			},
            }`)
			Expect(err).NotTo(HaveOccurred())

			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"configure-director",
				"--config",
				configFile.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, "40s").Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("The following keys have recently been removed from the top level configuration: director-configuration, iaas-configuration, security-configuration, syslog-configuration"))
			Expect(session.Err).To(gbytes.Say("To fix this error, move the above keys under 'properties-configuration' and change their dashes to underscores."))

			configFile, err = ioutil.TempFile("", "config.yml")
			Expect(err).ToNot(HaveOccurred())
			_, err = configFile.WriteString(`{
           		"what is this": "key?"
            }`)
			Expect(err).NotTo(HaveOccurred())

			command = exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"configure-director",
				"--config",
				configFile.Name(),
			)

			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, "40s").Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("the config file contains unrecognized keys: \"what is this\""))
		})
		It("configures the BOSH director using the API", func() {
			configFile, err := ioutil.TempFile("", "config.yml")
			Expect(err).ToNot(HaveOccurred())
			_, err = configFile.WriteString(`{
            "properties-configuration": {
				"iaas_configuration": {
				  "project": "some-project",
				  "default_deployment_tag": "my-vms",
				  "auth_json": "{\"some-auth-field\": \"some-service-key\",\"some-private_key\": \"some-key\"}"
				},
				"director_configuration": {
					"ntp_servers_string": "us.example.org, time.something.com",
					"resurrector_enabled": false,
					"director_hostname": "foo.example.com",
					"max_threads": 5
				},
            	"security_configuration": {
					"trusted_certificates": "some-certificate",
					"generate_vm_passwords": true
				},
            	"syslog_configuration": {
					"syslogconfig": "awesome"
				},
			},
            "az-configuration": [ {"name": "some-az-1"}],
            "networks-configuration": {
				"networks": [{"name": "network-1"}],
				"top-level": "the-top"
			},
            "network-assignment": {
				"network": { "name": "some-network"},
				"singleton_availability_zone": {"name": "some-az"}
			},
            "resource-configuration": {
				"compilation": {
					"instance_type": {
						"id": "m4.xlarge"
					}
				}
			}
            }`)
			Expect(err).NotTo(HaveOccurred())

			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"configure-director",
				"--config",
				configFile.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "40s").Should(gexec.Exit(0))

			Expect(azPutCallCount).To(Equal(1))
			Expect(azPostConfigurationMethod).To(Equal("POST"))
			Expect(azPostConfigurationBody).To(MatchJSON(`{
				"availability_zone": {"name": "some-az-1"}
			}`))

			Expect(networksPutCallCount).To(Equal(1))
			Expect(networksConfigurationBody).To(MatchJSON(`{
				"networks": [{"guid": "existing-network-guid","name": "network-1"}],
				"top-level": "the-top"
			}`))

			Expect(networkAZCallCount).To(Equal(1))
			Expect(networkAZConfigurationMethod).To(Equal("PUT"))
			Expect(networkAZConfigurationBody).To(MatchJSON(`{
					"network_and_az": {
					 "network": {
						 "name": "some-network"
					 },
					 "singleton_availability_zone": {
						 "name": "some-az"
					 }
				}
			}`))

			Expect(directorPropertiesCallCount).To(Equal(1))
			Expect(directorPropertiesMethod).To(Equal("PUT"))
			Expect(directorPropertiesBody).To(MatchJSON(`{
				"iaas_configuration": {
					"project": "some-project",
					"default_deployment_tag": "my-vms",
					"auth_json": "{\"some-auth-field\": \"some-service-key\",\"some-private_key\": \"some-key\"}"
				},
				"director_configuration": {
					"ntp_servers_string": "us.example.org, time.something.com",
					"resurrector_enabled": false,
					"director_hostname": "foo.example.com",
					"max_threads": 5
				},
				"security_configuration": {
					"trusted_certificates": "some-certificate",
					"generate_vm_passwords": true
				},
				"syslog_configuration": {
					"syslogconfig": "awesome"
				}
			}`))

			Expect(resourceConfigMethod).To(Equal("PUT"))
			Expect(resourceConfigBody).To(MatchJSON(`{
				"instances": 1,
				"instance_type": {
					"id": "m4.xlarge"
				},
				"persistent_disk": {
					"size_mb": "20480"
				},
				"internet_connected": true,
				"elb_names": ["my-elb"]
	  		}`))
		})
	})

	When("specifying a config file", func() {
		It("configures the BOSH director using the API", func() {
			configYAML := []byte(`
---
az-configuration:
- name: some-az-1
networks-configuration:
  networks:
  - name: network-1
  top-level: the-top
network-assignment:
  network:
    name: some-network
  singleton_availability_zone:
    name: some-az
resource-configuration:
  compilation:
    instance_type:
      id: m4.xlarge
properties-configuration:
  syslog_configuration:
    syslogconfig: awesome
  security_configuration:
    trusted_certificates: some-certificate
    generate_vm_passwords: true
  director_configuration:
    ntp_servers_string: us.example.org, time.something.com
    resurrector_enabled: false
    director_hostname: foo.example.com
    max_threads: 5
  iaas_configuration:
    project: some-project
    default_deployment_tag: my-vms
    associated_service_account: some-service-account
    auth_json: |
      {
        "some-auth-field": "some-service-key",
        "some-private_key": "some-key"
      }
`)

			tempfile, err := ioutil.TempFile("", "config.yaml")
			Expect(err).ToNot(HaveOccurred())

			_, err = tempfile.Write(configYAML)
			Expect(err).ToNot(HaveOccurred())
			Expect(tempfile.Close()).ToNot(HaveOccurred())

			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"configure-director",
				"--config", tempfile.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "40s").Should(gexec.Exit(0))

			Expect(azPutCallCount).To(Equal(1))
			Expect(azPostConfigurationMethod).To(Equal("POST"))
			Expect(azPostConfigurationBody).To(MatchJSON(`{
			"availability_zone": {"name": "some-az-1"}
		}`))

			Expect(networksPutCallCount).To(Equal(1))
			Expect(networksConfigurationBody).To(MatchJSON(`{
			"networks": [{"guid": "existing-network-guid","name": "network-1"}],
			"top-level": "the-top"
		}`))

			Expect(networkAZCallCount).To(Equal(1))
			Expect(networkAZConfigurationMethod).To(Equal("PUT"))
			Expect(networkAZConfigurationBody).To(MatchJSON(`{
			"network_and_az": {
				 "network": {
					 "name": "some-network"
				 },
				 "singleton_availability_zone": {
					 "name": "some-az"
				 }
			}
		}`))

			Expect(directorPropertiesCallCount).To(Equal(1))
			Expect(directorPropertiesMethod).To(Equal("PUT"))
			Expect(directorPropertiesBody).To(MatchJSON(`{
			"iaas_configuration": {
				"project": "some-project",
				"default_deployment_tag": "my-vms",
				"associated_service_account": "some-service-account",
				"auth_json": "{\n  \"some-auth-field\": \"some-service-key\",\n  \"some-private_key\": \"some-key\"\n}\n"
			},
			"director_configuration": {
				"ntp_servers_string": "us.example.org, time.something.com",
				"resurrector_enabled": false,
				"director_hostname": "foo.example.com",
				"max_threads": 5
			},
			"security_configuration": {
				"trusted_certificates": "some-certificate",
				"generate_vm_passwords": true
			},
			"syslog_configuration": {
				"syslogconfig": "awesome"
			}
	  }`))

			Expect(resourceConfigMethod).To(Equal("PUT"))
			Expect(resourceConfigBody).To(MatchJSON(`{
			"instances": 1,
			"instance_type": {
				"id": "m4.xlarge"
			},
			"persistent_disk": {
				"size_mb": "20480"
			},
			"internet_connected": true,
			"elb_names": ["my-elb"]
	  }`))
		})

		Describe("--ignore-verifier-warnings flag", func() {
			It("configures the BOSH director using the API, and ignores verifier warnings", func() {
				configYAML := []byte(`
---
az-configuration:
- name: some-az-1
- name: some-existing-az
`)

				verifierErrorOccured = true
				tempfile, err := ioutil.TempFile("", "config.yaml")
				Expect(err).ToNot(HaveOccurred())

				_, err = tempfile.Write(configYAML)
				Expect(err).ToNot(HaveOccurred())
				Expect(tempfile.Close()).ToNot(HaveOccurred())

				command := exec.Command(pathToMain,
					"--target", server.URL,
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"configure-director",
					"--ignore-verifier-warnings",
					"--config", tempfile.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, "40s").Should(gexec.Exit(0))

				Expect(azPostCallCount).To(Equal(1))
				Expect(azPostConfigurationMethod).To(Equal("POST"))
				Expect(azPostConfigurationBody).To(MatchJSON(`{
					"availability_zone": {"name": "some-az-1"}
				}`))

				Expect(azPutCallCount).To(Equal(1))
				Expect(azPutConfigurationMethod).To(Equal("PUT"))
				Expect(azPutConfigurationBody).To(MatchJSON(`{
					"availability_zone": {"guid": "existing-az-guid", "name": "some-existing-az"}
				}`))
			})
		})
	})
})
