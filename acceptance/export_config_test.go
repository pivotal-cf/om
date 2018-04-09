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

var _ = Describe("export-config command", func() {
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
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	It("outputs a configuration template based on the staged product", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"export-config",
			"--product-name", "some-product",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "10s").Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(MatchYAML(`---
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
`))
	})
})
