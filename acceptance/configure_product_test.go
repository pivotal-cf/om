package acceptance

import (
	"fmt"
	"github.com/onsi/gomega/ghttp"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("configure-product command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	BeforeEach(func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/installations"),
				ghttp.RespondWith(http.StatusOK, `{"installations": []}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
				ghttp.RespondWith(http.StatusOK, `[{
					"installation_name": "some-product-guid",
					"guid": "some-product-guid",
					"type": "cf"
				}, {
					"installation_name": "p-bosh-installation-name",
					"guid": "p-bosh-guid",
					"type": "p-bosh"
				}]`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/networks_and_azs"),
				ghttp.VerifyJSON(fmt.Sprintf(`{"networks_and_azs": %s}`, productNetworkJSON)),
				ghttp.RespondWith(http.StatusOK, `{}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/properties"),
				ghttp.RespondWith(http.StatusOK, `{}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/properties"),
				ghttp.VerifyJSON(fmt.Sprintf(`{"properties": %s}`, propertiesJSON)),
				ghttp.RespondWith(http.StatusOK, "{}"),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs"),
				ghttp.RespondWith(http.StatusOK, `{
					"jobs": [{
						"name": "not-the-job",
						"guid": "bad-guid"
					}, {
						"name": "some-job",
						"guid": "the-right-guid"
					}, {
						"name": "some-other-job",
						"guid": "just-a-guid"
					}]
				}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/the-right-guid/resource_config"),
				ghttp.RespondWith(http.StatusOK, `{}`),
			),
		)
	})

	It("successfully configures any product", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/jobs/the-right-guid/resource_config"),
				ghttp.VerifyJSON(`{
					"instances": 1,
					"persistent_disk": {
						"size_mb": "20480"
					},
					"instance_type": {
						"id": "m1.medium"
					},
					"additional_vm_extensions": [
						"some-vm-extension",
						"some-other-vm-extension"
					]
				}`),
				ghttp.RespondWith(http.StatusOK, `{}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/just-a-guid/resource_config"),
				ghttp.RespondWith(http.StatusOK, `{}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/jobs/just-a-guid/resource_config"),
				ghttp.VerifyJSON(`{
        			"instances": "automatic",
        			"persistent_disk": {
          				"size_mb": "20480"
        			},
        			"instance_type": {
          				"id": "m1.medium"
        			}
      			}`),
				ghttp.RespondWith(http.StatusOK, `{}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs"),
				ghttp.RespondWith(http.StatusOK, `{
					"jobs": [{
						"name": "not-the-job",
						"guid": "bad-guid"
					}, {
						"name": "some-job",
						"guid": "the-right-guid"
					}, {
						"name": "some-other-job",
						"guid": "just-a-guid"
					}]
				}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/staged/pending_changes"),
				ghttp.RespondWith(http.StatusOK, `{}`),
			),
		)
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
			"--target", server.URL(),
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
	})

	When("a config file is provided", func() {
		var (
			configFile *os.File
			err        error
		)

		AfterEach(func() {
			os.RemoveAll(configFile.Name())
		})

		It("successfully configures any product", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/jobs/the-right-guid/resource_config"),
					ghttp.VerifyJSON(`{
						"instances": 1,
						"persistent_disk": {
							"size_mb": "20480"
						},
						"instance_type": {
							"id": "m1.medium"
						},
						"additional_vm_extensions": [
							"some-vm-extension",
							"some-other-vm-extension"
						]
					}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/just-a-guid/resource_config"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/jobs/just-a-guid/resource_config"),
					ghttp.VerifyJSON(`{
        				"instances": "automatic",
        				"persistent_disk": {
          					"size_mb": "20480"
        				},
        				"instance_type": {
          					"id": "m1.medium"
        				}
      				}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs"),
					ghttp.RespondWith(http.StatusOK, `{
						"jobs": [{
							"name": "not-the-job",
							"guid": "bad-guid"
						}, {
							"name": "some-job",
							"guid": "the-right-guid"
						}, {
							"name": "some-other-job",
							"guid": "just-a-guid"
						}]
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/syslog_configuration"),
					ghttp.VerifyJSON(`{
				  		"syslog_configuration": {
							"address": "example.com",
							"enabled": true,
							"port": 514,
							"queue_size": null,
							"tls_enabled": false,
							"transport_protocol": "udp"
				  		}
        			}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/pending_changes"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			configFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = configFile.WriteString(configFileContents)
			Expect(err).NotTo(HaveOccurred())

			command := exec.Command(pathToMain,
				"--target", server.URL(),
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
		})

		It("successfully configures a product on nsx", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/jobs/the-right-guid/resource_config"),
					ghttp.VerifyJSON(`{
          				"instances": 1,
          				"persistent_disk": {
					  		"size_mb": "20480"
          				},
          				"instance_type": {
					  		"id": "m1.medium"
          				},
          				"nsx_security_groups": [
							"sg1",
					  		"sg2"
          				],
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
        			}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs"),
					ghttp.RespondWith(http.StatusOK, `{
						"jobs": [{
							"name": "not-the-job",
							"guid": "bad-guid"
						}, {
							"name": "some-job",
							"guid": "the-right-guid"
						}, {
							"name": "some-other-job",
							"guid": "just-a-guid"
						}]
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/pending_changes"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			nsxConfigFileContents := fmt.Sprintf(`{
				"product-name": "cf",
				"product-properties": %s,
				"network-properties": %s,
				"resource-config": %s
			}`, propertiesJSON, productNetworkJSON, nsxResourceConfigJSON)

			configFile, err = ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			_, err = configFile.WriteString(nsxConfigFileContents)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"configure-product",
				"--config", configFile.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
		})
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
