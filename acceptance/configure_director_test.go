package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"io/ioutil"
	"net/http"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = FDescribe("configure-director command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.RouteToHandler("GET", "/api/v0/installations",
			ghttp.RespondWith(http.StatusOK, `{"installations": []}`, map[string][]string{
				"Content-Type": []string{"application/json"},
			}),
		)
		server.RouteToHandler("GET", "/api/v0/staged/director/availability_zones",
			ghttp.RespondWith(http.StatusOK, "{}"),
		)
	})

	AfterEach(func() {
		server.Close()
	})

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
			"--target", server.URL(),
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
			"--target", server.URL(),
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
		server.RouteToHandler("PUT", "/api/v0/staged/director/properties",
			ghttp.VerifyJSON(`{
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
	  			}`),
		)
		server.RouteToHandler("POST", "/api/v0/staged/director/availability_zones",
			ghttp.VerifyJSON(`{"availability_zone":{"name":"some-az-1"}}`),
		)
		server.RouteToHandler("GET", "/api/v0/staged/director/networks",
			ghttp.RespondWith(http.StatusOK, ""),
		)
		server.RouteToHandler("PUT", "/api/v0/staged/director/networks",
			ghttp.VerifyJSON(`{"networks":[{"name":"network-1"}],"top-level":"the-top"}`),
		)
		server.RouteToHandler("GET", "/api/v0/deployed/director/credentials",
			ghttp.RespondWith(http.StatusNotFound, ""),
		)
		server.RouteToHandler("PUT", "/api/v0/staged/director/network_and_az",
			ghttp.VerifyJSON(`{"network_and_az":{"network":{"name":"some-network"},"singleton_availability_zone":{"name":"some-az"}}}`),
		)
		server.RouteToHandler("GET", "/api/v0/staged/products",
			ghttp.RespondWith(http.StatusOK, `[
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
				]`),
		)
		server.RouteToHandler("GET", "/api/v0/staged/products/p-bosh-guid/jobs",
			ghttp.RespondWith(http.StatusOK, `
					{
						"jobs": [
							{
								"name": "compilation",
								"guid": "compilation-guid"
							}
						]
					}
				`),
		)
		server.RouteToHandler("GET", "/api/v0/staged/products/p-bosh-guid/jobs/compilation-guid/resource_config",
			ghttp.RespondWith(http.StatusOK, `{
					"instances": 1,
					"instance_type": {
						"id": "automatic"
					},
					"persistent_disk": {
						"size_mb": "20480"
					},
					"internet_connected": true,
					"elb_names": ["my-elb"]
				}`),
		)
		server.RouteToHandler("PUT", "/api/v0/staged/products/p-bosh-guid/jobs/compilation-guid/resource_config",
			ghttp.VerifyJSON(`{"instances":1,"persistent_disk":{"size_mb":"20480"},"instance_type":{"id":"m4.xlarge"},"internet_connected":true,"elb_names":["my-elb"]}`),
		)
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
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"configure-director",
			"--config", tempfile.Name(),
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "40s").Should(gexec.Exit(0))
	})

	Describe("--ignore-verifier-warnings flag", func() {
		It("configures the BOSH director using the API, and ignores verifier warnings", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/staged/director/availability_zones"),
					ghttp.VerifyJSON(`{"availability_zone":{"name":"some-az-1"}}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/staged/director/availability_zones"),
					ghttp.VerifyJSON(`{"availability_zone":{"name":"some-existing-az"}}`),
				),
			)

			configYAML := []byte(`
---
az-configuration:
- name: some-az-1
- name: some-existing-az
`)

			//verifierErrorOccured = true
			tempfile, err := ioutil.TempFile("", "config.yaml")
			Expect(err).ToNot(HaveOccurred())

			_, err = tempfile.Write(configYAML)
			Expect(err).ToNot(HaveOccurred())
			Expect(tempfile.Close()).ToNot(HaveOccurred())

			command := exec.Command(pathToMain,
				"--target", server.URL(),
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
		})
	})
})
