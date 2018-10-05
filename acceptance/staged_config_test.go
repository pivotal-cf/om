package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("staged-config command", func() {
	var (
		server *httptest.Server
	)

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
			case "/api/v0/staged/products":
				w.Write([]byte(`[
					{"installation_name":"p-bosh","guid":"p-bosh-guid","type":"p-bosh","product_version":"1.10.0.0"},
					{"installation_name":"some-product","guid":"some-product-guid","type":"some-product","product_version":"1.0.0"}
				]`))
			case "/api/v0/deployed/products":
				w.Write([]byte(`[
					{"guid":"p-bosh-guid","type":"p-bosh"},
					{"guid":"some-product-guid","type":"some-product"}
				]`))
			case "/api/v0/staged/products/some-product-guid/properties":
				w.Write([]byte(`{
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
        }`))
			case "/api/v0/staged/products/some-product-guid/networks_and_azs":
				w.Write([]byte(`{
          "networks_and_azs": {
            "singleton_availability_zone": {
              "name": "az-one"
            },
            "other_availability_zones": [
              {
                "name": "az-two"
              },
              {
                "name": "az-three"
              }
            ],
            "network": {
              "name": "network-one"
            }
          }
        }`))
			case "/api/v0/staged/products/some-product-guid/jobs":
				w.Write([]byte(`{
					"jobs": [
					  {
							"name": "some-job",
							"guid": "some-guid"
						}
					]
				}`))
			case "/api/v0/staged/products/some-product-guid/jobs/some-guid/resource_config":
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
			case "/api/v0/deployed/products/some-product-guid/credentials/.properties.some-secret-property":
				w.Write([]byte(`{
						"credential": {
							"type": "some-secret-type",
							"value": {
								"some-secret-key": "some-secret-value"
							}
						}
					}`))
			case "/api/v0/staged/products/some-product-guid/errands":
				w.Write([]byte(`{ "errands": [
                           {
                             "name": "errand-1",
                             "post_deploy": false,
                             "label": "Errand 1 Label"
                           },
                           {
                             "name": "errand-2",
                             "pre_delete": true,
                             "label": "Errand 2 Label"
                           }]}`))
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

	It("outputs a configuration template based on the staged product", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"staged-config",
			"--product-name", "some-product",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

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
errand-config:
  errand-1:
    post-deploy-state: false
  errand-2:
    pre-delete-state: true
`))
	})

	Context("when --include-credentials is used", func() {
		It("outputs the secret values in the template", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"staged-config",
				"--product-name", "some-product",
				"--include-credentials",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

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
errand-config:
  errand-1:
    post-deploy-state: false
  errand-2:
    pre-delete-state: true
`))
		})
	})
})
