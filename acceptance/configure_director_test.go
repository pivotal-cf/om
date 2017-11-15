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
		server                       *httptest.Server
		networkAZCallCount           int
		azCallCount                  int
		propertiesCallCount          int
		networkAZConfigurationBody   []byte
		azConfigurationBody          []byte
		directorPropertiesBody       []byte
		networkAZConfigurationMethod string
		azConfigurationMethod        string
		directorPropertiesMethod     string
	)

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
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

				var err error
				azConfigurationBody, err = ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				azConfigurationMethod = req.Method

				azCallCount++

				w.Write([]byte(`{}`))
			case "/api/v0/staged/director/network_and_az":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" && auth != "Bearer some-running-install-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				if azCallCount == 0 {
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

				propertiesCallCount++

				w.Write([]byte(`{}`))
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	It("configures the BOSH director using the API", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"configure-director",
			"--iaas-configuration",
			`{
				"project": "some-project",
				"default_deployment_tag": "my-vms",
				"auth_json": "{\"some-auth-field\": \"some-service-key\",\"some-private_key\": \"some-key\"}"
			}`,
			"--az-configuration",
			`[ {"az_property": "value"} ]`,
			"--network-assignment",
			`{
				"network": { "name": "some-network"},
				"singleton_availability_zone": {"name": "some-az"}
			}`,
			"--director-configuration",
			`{
				"ntp_servers_string": "us.example.org, time.something.com",
				"resurrector_enabled": false,
				"director_hostname": "foo.example.com",
				"max_threads": 5
			 }`,
			"--security-configuration",
			`{
				"trusted_certificates": "some-certificate",
				"generate_vm_passwords": true
			}`,
			"--syslog-configuration", `{
				"syslogconfig": "awesome"
			}`,
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "40s").Should(gexec.Exit(0))

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

		Expect(azCallCount).To(Equal(1))
		Expect(azConfigurationMethod).To(Equal("PUT"))
		Expect(azConfigurationBody).To(MatchJSON(`{
			"availability_zones": [{"az_property": "value"}]
		}`))

		Expect(propertiesCallCount).To(Equal(1))
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
	})
})
