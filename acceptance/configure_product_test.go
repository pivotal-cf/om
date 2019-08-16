package acceptance

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("configure-product command", func() {
	var (
		server                    *httptest.Server
		productPropertiesMethod   string
		productPropertiesBody     []byte
		productNetworkMethod      string
		productNetworkBody        []byte
		resourceConfigMethod      []string
		resourceConfigBody        [][]byte
		syslogConfigurationMethod string
		syslogConfigurationBody   []byte
	)

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/api/v0/staged/pending_changes":
				_, err := w.Write([]byte(`{}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/installations":
				_, err := w.Write([]byte(`{"installations": []}`))
				Expect(err).ToNot(HaveOccurred())
			case "/uaa/oauth/token":
				_, err := w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/products":
				_, err := w.Write([]byte(`[
					{
						"installation_name": "some-product-guid",
						"guid": "some-product-guid",
						"type": "cf"
					},
					{
						"installation_name": "p-bosh-installation-name",
						"guid": "p-bosh-guid",
						"type": "p-bosh"
					}
				]`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/products/some-product-guid/jobs":
				_, err := w.Write([]byte(`{
					"jobs": [
					  {
							"name": "not-the-job",
							"guid": "bad-guid"
						},
					  {
							"name": "some-job",
							"guid": "the-right-guid"
						},
					  {
							"name": "some-other-job",
							"guid": "just-a-guid"
						}
					]
				}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/products/some-product-guid/properties":
				var err error
				productPropertiesMethod = req.Method
				productPropertiesBody, err = ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				_, err = w.Write([]byte(`{}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/products/some-product-guid/networks_and_azs":
				var err error
				productNetworkMethod = req.Method
				productNetworkBody, err = ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				_, err = w.Write([]byte(`{}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/products/some-product-guid/jobs/just-a-guid/resource_config":
				fallthrough
			case "/api/v0/staged/products/some-product-guid/jobs/the-right-guid/resource_config":
				resourceConfigMethod = append(resourceConfigMethod, req.Method)
				body, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				resourceConfigBody = append(resourceConfigBody, body)

				_, err = w.Write([]byte(`{}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/products/some-product-guid/max_in_flight":
				resourceConfigMethod = append(resourceConfigMethod, req.Method)
				body, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				resourceConfigBody = append(resourceConfigBody, body)

				_, err = w.Write([]byte(`{}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/products/some-product-guid/syslog_configuration":
				var err error
				syslogConfigurationMethod = req.Method
				syslogConfigurationBody, err = ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				_, err = w.Write([]byte(`{}`))
				Expect(err).ToNot(HaveOccurred())
			default:
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-opsman-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	AfterEach(func() {
		resourceConfigMethod = []string{}
		resourceConfigBody = [][]byte{}
		server.Close()
	})

	It("successfully configures any product", func() {
		configFileContents := fmt.Sprintf(`{
		"product-name": "cf",
		"product-properties": %s,
		"network-properties": %s,
		"resource-config": %s
		}`, propertiesJSON, productNetworkJSON, resourceConfigJSON)
		configFile, err := ioutil.TempFile("", "")
		Expect(err).ToNot(HaveOccurred())
		_, err = configFile.WriteString(configFileContents)
		Expect(err).ToNot(HaveOccurred())

		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"configure-product",
			"--config", configFile.Name(),
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say("setting properties"))
		Expect(session.Out).To(gbytes.Say("finished setting properties"))

		Expect(productPropertiesMethod).To(Equal("PUT"))
		Expect(productPropertiesBody).To(MatchJSON(fmt.Sprintf(`{"properties": %s}`, propertiesJSON)))

		Expect(productNetworkMethod).To(Equal("PUT"))
		Expect(productNetworkBody).To(MatchJSON(fmt.Sprintf(`{"networks_and_azs": %s}`, productNetworkJSON)))

		Expect(resourceConfigMethod[1]).To(Equal("PUT"))
		Expect(resourceConfigBody[1]).To(MatchJSON(`{
        "instances": 1,
        "persistent_disk": {
          "size_mb": "20480"
        },
        "instance_type": {
          "id": "m1.medium"
        },
        "additional_vm_extensions": ["some-vm-extension","some-other-vm-extension"]
      }`))

		Expect(resourceConfigMethod[3]).To(Equal("PUT"))
		Expect(resourceConfigBody[3]).To(MatchJSON(`{
        "instances": "automatic",
        "persistent_disk": {
          "size_mb": "20480"
        },
        "instance_type": {
          "id": "m1.medium"
        }
      }`))
	})

	Context("when a config file is provided", func() {
		var (
			configFile *os.File
			err        error
		)

		AfterEach(func() {
			os.RemoveAll(configFile.Name())
		})

		It("successfully configures any product", func() {
			configFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = configFile.WriteString(configFileContents)
			Expect(err).NotTo(HaveOccurred())

			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"configure-product",
				"--config", configFile.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(session.Out).To(gbytes.Say("setting properties"))
			Expect(session.Out).To(gbytes.Say("finished setting properties"))

			Expect(productPropertiesMethod).To(Equal("PUT"))
			Expect(productPropertiesBody).To(MatchJSON(fmt.Sprintf(`{"properties": %s}`, propertiesJSON)))

			Expect(productNetworkMethod).To(Equal("PUT"))
			Expect(productNetworkBody).To(MatchJSON(fmt.Sprintf(`{"networks_and_azs": %s}`, productNetworkJSON)))

			Expect(resourceConfigMethod[1]).To(Equal("PUT"))
			Expect(resourceConfigBody[1]).To(MatchJSON(`{
        "instances": 1,
        "persistent_disk": {
          "size_mb": "20480"
        },
        "instance_type": {
          "id": "m1.medium"
        },
        "additional_vm_extensions": ["some-vm-extension","some-other-vm-extension"]
      }`))

			Expect(resourceConfigMethod[3]).To(Equal("PUT"))
			Expect(resourceConfigBody[3]).To(MatchJSON(`{
        "instances": "automatic",
        "persistent_disk": {
          "size_mb": "20480"
        },
        "instance_type": {
          "id": "m1.medium"
        }
      }`))

			Expect(syslogConfigurationMethod).To(Equal("PUT"))
			Expect(syslogConfigurationBody).To(MatchJSON(`{"syslog_configuration": {
				"enabled": true,
				"address": "example.com",
				"port": 514,
				"transport_protocol": "udp",
				"queue_size": null,
				"tls_enabled": false
  			}}`))
		})
	})

	It("successfully configures a product on nsx", func() {
		configFileContents := fmt.Sprintf(`{
		"product-name": "cf",
		"product-properties": %s,
		"network-properties": %s,
		"resource-config": %s
		}`, propertiesJSON, productNetworkJSON, nsxResourceConfigJSON)
		configFile, err := ioutil.TempFile("", "")
		Expect(err).ToNot(HaveOccurred())
		_, err = configFile.WriteString(configFileContents)
		Expect(err).ToNot(HaveOccurred())

		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"configure-product",
			"--config", configFile.Name(),
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(resourceConfigMethod[1]).To(Equal("PUT"))
		Expect(resourceConfigBody[1]).To(MatchJSON(`{
			"instances": 1,
			"persistent_disk": {
				"size_mb": "20480"
			},
			"instance_type": {
				"id": "m1.medium"
			},
			"nsx_security_groups":["sg1", "sg2"],
			"nsx_lbs": [
				{
					"edge_name": "edge-1",
					"pool_name": "pool-1",
					"security_group": "sg-1",
					"port": 5000
				},
				{
					"edge_name": "edge-2",
					"pool_name": "pool-2",
					"security_group": "sg-2",
					"port": 5000
				}
			]
		}`))
	})
})

const propertiesJSON = `{
	".properties.something": {"value": "configure-me"},
	".a-job.job-property": {"value": {"identity": "username", "password": "example-new-password"} },
	".top-level-property": { "value": [ { "guid": "some-guid", "name": "max", "my-secret": {"secret": "headroom"} } ] }
}`

const productNetworkJSON = `{
  "singleton_availability_zone": {"name": "az-one"},
  "other_availability_zones": [{"name": "az-two" }, {"name": "az-three"}],
  "network": {"name": "network-one"}
}`

const nsxResourceConfigJSON = `
{
  "some-job": {
    "instances": 1,
    "persistent_disk": { "size_mb": "20480" },
    "instance_type": { "id": "m1.medium" },
    "nsx_security_groups":["sg1", "sg2"],
    "nsx_lbs": [
    {
      "edge_name": "edge-1",
      "pool_name": "pool-1",
      "security_group": "sg-1",
      "port": 5000
    },
    {
      "edge_name": "edge-2",
      "pool_name": "pool-2",
      "security_group": "sg-2",
      "port": 5000
    }]
  }
}`

const resourceConfigJSON = `
{
  "some-job": {
    "instances": 1,
    "persistent_disk": { "size_mb": "20480" },
    "instance_type": { "id": "m1.medium" },
    "additional_vm_extensions": ["some-vm-extension", "some-other-vm-extension"]
  },
  "some-other-job": {
	  "instances": "automatic",
		"persistent_disk": { "size_mb": "20480" },
    "instance_type": { "id": "m1.medium" }
  }
}`

const configFileContents = `---
product-name: cf
product-properties:
  .properties.something:
    value: configure-me
  .a-job.job-property:
    value:
      identity: username
      password: example-new-password
  .top-level-property:
    value: [ { guid: some-guid, name: max, my-secret: {secret: headroom} } ]
network-properties:
  singleton_availability_zone:
    name: az-one
  other_availability_zones:
    - name: az-two
    - name: az-three
  network:
    name: network-one
resource-config:
  some-job:
    instances: 1
    persistent_disk: { size_mb: "20480" }
    instance_type: { id: m1.medium }
    additional_vm_extensions: [some-vm-extension, some-other-vm-extension]
  some-other-job:
    instances: automatic
    persistent_disk: { size_mb: "20480" }
    instance_type: { id: m1.medium }
syslog-properties:
  enabled: true
  address: "example.com"
  port: 514
  transport_protocol: "udp"
  queue_size: null
  tls_enabled: false
`
