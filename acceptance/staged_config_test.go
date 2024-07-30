package acceptance

import (
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("staged-config command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/info"),
				ghttp.RespondWith(http.StatusOK, `{
					"info": {
						"version": "2.4-build.79"
					}
				}`),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	When("--include-credentials is not used", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, stagedProductsJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/properties", "redact=true"),
					ghttp.RespondWith(http.StatusOK, stagedPropertiesJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/networks_and_azs"),
					ghttp.RespondWith(http.StatusOK, stagedNetworksAndAzsJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs"),
					ghttp.RespondWith(http.StatusOK, stagedJobsJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/max_in_flight"),
					ghttp.RespondWith(http.StatusOK, `{"max_in_flight": {"some-guid": "20%"}}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/syslog_configuration"),
					ghttp.RespondWith(http.StatusOK, stagedSyslogConfigurationJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config"),
					ghttp.RespondWith(http.StatusOK, stagedResourceConfigJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/errands"),
					ghttp.RespondWith(http.StatusOK, stagedErrandsJSON),
				),
			)
		})

		It("outputs a configuration template based on the staged product", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"staged-config",
				"--product-name", "some-product",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, "10s").Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(MatchYAML(`---
product-name: some-product
product-properties:
  .properties.some-configurable-property:
    value: some-configurable-value
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
    instance_type: { id: automatic }
    elb_names: ["my-elb"]
    internet_connected: true
    max_in_flight: 20%
errand-config:
  errand-1:
    post-deploy-state: false
  errand-2:
    pre-delete-state: true
syslog-properties:
  enabled: true
  address: example.com
  port: 514
  transport_protocol: tcp
  queue_size: null
  tls_enabled: true
  permitted_peer: "*.example.com"
  ssl_ca_certificate: "-----BEGIN CERTIFICATE-----\r\nMIIBsjCCARug..."
`))
		})
	})

	When("--include-credentials is used", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
					ghttp.RespondWith(http.StatusOK, `[
						{"guid":"p-bosh-guid","type":"p-bosh"},
						{"guid":"some-product-guid","type":"some-product"}
					]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, stagedProductsJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/properties", "redact=false"),
					ghttp.RespondWith(http.StatusOK, stagedPropertiesWithSecretsNotRedactedJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/networks_and_azs"),
					ghttp.RespondWith(http.StatusOK, stagedNetworksAndAzsJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs"),
					ghttp.RespondWith(http.StatusOK, stagedJobsJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/max_in_flight"),
					ghttp.RespondWith(http.StatusOK, `{
						"max_in_flight": {"some-guid": "20%"}
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/syslog_configuration"),
					ghttp.RespondWith(http.StatusOK, stagedSyslogConfigurationJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config"),
					ghttp.RespondWith(http.StatusOK, stagedResourceConfigJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/errands"),
					ghttp.RespondWith(http.StatusOK, stagedErrandsJSON),
				),
			)
		})

		It("outputs the secret values in the template", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"staged-config",
				"--product-name", "some-product",
				"--include-credentials",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, "10s").Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(MatchYAML(`---
product-name: some-product
product-properties:
  .properties.some-configurable-property:
    value: some-configurable-value
  .properties.some-secret-property:
    value:
      some-secret-key: some-secret-value
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
    instance_type: { id: automatic }
    elb_names: ["my-elb"]
    internet_connected: true
    max_in_flight: 20%
errand-config:
  errand-1:
    post-deploy-state: false
  errand-2:
    pre-delete-state: true
syslog-properties:
  enabled: true
  address: example.com
  port: 514
  transport_protocol: tcp
  queue_size: null
  tls_enabled: true
  permitted_peer: "*.example.com"
  ssl_ca_certificate: "-----BEGIN CERTIFICATE-----\r\nMIIBsjCCARug..."
`))
		})
	})
})

const stagedProductsJSON = `[
	{"installation_name":"p-bosh","guid":"p-bosh-guid","type":"p-bosh","product_version":"1.10.0.0"},
	{"installation_name":"some-product","guid":"some-product-guid","type":"some-product","product_version":"1.0.0"}
]`

const stagedPropertiesWithSecretsNotRedactedJSON = `{
	"properties": {
		".properties.some-configurable-property": {
			"type": "string",
			"configurable": true,
			"credential": false,
			"value": "some-configurable-value",
			"optional": true
		},
		".properties.some-non-configurable-property": {
			"type": "string",
			"configurable": false,
			"credential": false,
			"value": "some-non-configurable-value",
			"optional": false
		},
		".properties.some-secret-property": {
			"type": "string",
			"configurable": true,
			"credential": true,
			"value": {
				"some-secret-key": "some-secret-value"
			},
			"optional": true
		}
	}
}`

const stagedPropertiesJSON = `{
	"properties": {
		".properties.some-configurable-property": {
			"type": "string",
			"configurable": true,
			"credential": false,
			"value": "some-configurable-value",
			"optional": true
		},
		".properties.some-non-configurable-property": {
			"type": "string",
			"configurable": false,
			"credential": false,
			"value": "some-non-configurable-value",
			"optional": false
		},
		".properties.some-secret-property": {
			"type": "string",
			"configurable": true,
			"credential": true,
			"value": {
				"some-secret-key": "***"
			},
			"optional": true
		}
	}
}`

const stagedNetworksAndAzsJSON = `{
	"networks_and_azs": {
		"singleton_availability_zone": {
			"name": "az-one"
		},
		"other_availability_zones": [{
			"name": "az-two"
		}, {
			"name": "az-three"
		}],
		"network": {
			"name": "network-one"
		}
	}
}`

const stagedSyslogConfigurationJSON = `{
	"syslog_configuration": {
		"enabled": true,
		"address": "example.com",
		"port": 514,
		"transport_protocol": "tcp",
		"queue_size": null,
		"tls_enabled": true,
		"permitted_peer": "*.example.com",
		"ssl_ca_certificate": "-----BEGIN CERTIFICATE-----\r\nMIIBsjCCARug..."
	}
}`

const stagedResourceConfigJSON = `{
	"instances": 1,
	"instance_type": {
		"id": "automatic"
	},
	"persistent_disk": {
		"size_mb": "20480"
	},
	"internet_connected": true,
	"elb_names": ["my-elb"]
}`

const stagedErrandsJSON = `{
	"errands": [{
		"name": "errand-1",
		"post_deploy": false,
		"label": "Errand 1 Label"
	}, {
		"name": "errand-2",
		"pre_delete": true,
		"label": "Errand 2 Label"
	}]
}`

const stagedJobsJSON = `{
	"jobs": [{
		"name": "some-job",
		"guid": "some-guid"
	}]
}`
