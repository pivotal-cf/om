package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pre_deploy_check command", func() {
	var (
		server *httptest.Server
	)

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch req.URL.Path {
			case "/uaa/oauth/token":
				_, err := w.Write([]byte(`{
					"access_token": "some-opsman-token",
					"token_type": "bearer",
					"expires_in": 3600
				}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/info":
				_, err := w.Write([]byte(`{
						"info": {
							"version": "2.6.0"
						}
					}`))

				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/director/pre_deploy_check":
				_, err := w.Write([]byte(`{
				  "pre_deploy_check": {
					"identifier": "p-bosh-guid",
					"complete": false,
					"network": {
					  "assigned": true
					},
					"availability_zone": {
					  "assigned": false
					},
					"stemcells": [
					  {
						"assigned": false,
						"required_stemcell_version": "250.2",
						"required_stemcell_os": "ubuntu-xenial"
					  }
					],
					"properties": [
						{
							"name": ".properties.iaas_configuration.project",
							"type": null,
							"errors": [
								"can't be blank"
							]
						}
					],
					"resources": {
					  "jobs": [{
						"identifier": "job-identifier",
						"guid": "job-guid",
						"error": [
						  "Instance : Value must be a positive integer"
						]
					  }]
					},
					"verifiers": [
					  {
						"type": "NetworksPingableVerifier",
						"errors": [ 
						  "NetworksPingableVerifier error"
						],
						"ignorable": true
					  }
					]
				  }
				}`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/products":
				_, err := w.Write([]byte(`[{"guid":"p-guid"}]`))
				Expect(err).ToNot(HaveOccurred())
			case "/api/v0/staged/products/p-guid/pre_deploy_check":
				_, err := w.Write([]byte(`{
				  "pre_deploy_check": {
					"identifier": "p-guid",
					"complete": false,
					"network": {
					  "assigned": true
					},
					"availability_zone": {
					  "assigned": false
					},
					"stemcells": [
					  {
						"assigned": false,
						"required_stemcell_version": "250.2",
						"required_stemcell_os": "ubuntu-xenial"
					  }
					],
					"properties": [
						{
							"name": ".properties.iaas_configuration.project",
							"type": null,
							"errors": [
								"can't be blank"
							]
						}
					],
					"resources": {
					  "jobs": [{
						"identifier": "job-identifier",
						"guid": "job-guid",
						"error": [
						  "Instance : Value must be a positive integer"
						]
					  }]
					},
					"verifiers": [
					  {
						"type": "NetworksPingableVerifier",
						"errors": [ 
						  "NetworksPingableVerifier error"
						],
						"ignorable": true
					  }
					]
				  }
				}`))
				Expect(err).ToNot(HaveOccurred())
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

	It("exits with an error if director or products are not completely configured", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"pre-deploy-check")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(1))

		Expect(string(session.Out.Contents())).To(ContainSubstring("[X] director: p-bosh-guid"))
	})
})
