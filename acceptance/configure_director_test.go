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
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("configure-director command", func() {
	var (
		azPutCallCount               int
		azPutConfigurationBody       []byte
		azConfigurationMethod        string
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

		server *httptest.Server
	)

	BeforeEach(func() {
		azPutCallCount = 0
		directorPropertiesCallCount = 0
		networkAZCallCount = 0
		networksPutCallCount = 0

		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/api/v0/installations":
				w.Write([]byte(`{"installations": []}`))
			case "/uaa/oauth/token":
				username := req.FormValue("username")

				if username == "some-username" {
					w.Write([]byte(`{
						"access_token": "some-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`))
				} else {
					w.Write([]byte(`{
						"access_token": "some-running-install-opsman-token",
						"token_type": "bearer",
						"expires_in": 3600
					}`))
				}
			case "/api/v0/staged/director/availability_zones":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" && auth != "Bearer some-running-install-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				azConfigurationMethod = req.Method

				if req.Method == "GET" {
					w.Write([]byte(`"availability_zones": [{"guid": "existing-az-guid", "name": "some-existing-az"}]`))
				} else if req.Method == "PUT" {
					var err error
					azPutConfigurationBody, err = ioutil.ReadAll(req.Body)
					Expect(err).NotTo(HaveOccurred())

					azPutCallCount++

					w.Write([]byte(`{}`))
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
					w.Write([]byte(`"networks": [{"guid": "existing-network-guid", "name": "network-1"}]`))
				} else if req.Method == "PUT" {
					var err error
					networksConfigurationBody, err = ioutil.ReadAll(req.Body)
					Expect(err).NotTo(HaveOccurred())

					networksPutCallCount++

					w.Write([]byte(`{}`))
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

				w.Write([]byte(`{}`))
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

				w.Write([]byte(`{}`))

			case "/api/v0/staged/products":
				w.Write([]byte(`[
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

			case "/api/v0/staged/products/p-bosh-guid/jobs":
				w.Write([]byte(`
					{
						"jobs": [
							{
								"name": "compilation",
								"guid": "compilation-guid"
							}
						]
					}
				`))

			case "/api/v0/staged/products/p-bosh-guid/jobs/compilation-guid/resource_config":
				if req.Method == "GET" {
					w.Write([]byte(`{
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
					return
				}

				var err error
				resourceConfigMethod = req.Method
				resourceConfigBody, err = ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				w.Write([]byte(`{}`))

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

	Context("when using command line arguments", func() {
		It("configures the BOSH director using the API", func() {
			configFile, err := ioutil.TempFile("", "config.yml")
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
			Expect(azConfigurationMethod).To(Equal("PUT"))
			Expect(azPutConfigurationBody).To(MatchJSON(`{
			"availability_zones": [{"name": "some-az-1"}, {"guid": "existing-az-guid", "name": "some-existing-az"}]
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

	Context("when specifying a config file", func() {
		It("configures the BOSH director using the API", func() {
			configYAML := []byte(`
---
iaas-configuration:
  project: some-project
  default_deployment_tag: my-vms
  associated_service_account: some-service-account
  auth_json: |
    {
      "some-auth-field": "some-service-key",
      "some-private_key": "some-key"
    }
az-configuration:
- name: some-az-1
- name: some-existing-az
networks-configuration:
  networks:
  - name: network-1
  top-level: the-top
network-assignment:
  network:
    name: some-network
  singleton_availability_zone:
    name: some-az
director-configuration:
  ntp_servers_string: us.example.org, time.something.com
  resurrector_enabled: false
  director_hostname: foo.example.com
  max_threads: 5
security-configuration:
  trusted_certificates: some-certificate
  generate_vm_passwords: true
syslog-configuration:
  syslogconfig: awesome
resource-configuration:
  compilation:
    instance_type:
      id: m4.xlarge
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
			Expect(azConfigurationMethod).To(Equal("PUT"))
			Expect(azPutConfigurationBody).To(MatchJSON(`{
			"availability_zones": [{"name": "some-az-1"}, {"guid": "existing-az-guid", "name": "some-existing-az"}]
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
	})
})
