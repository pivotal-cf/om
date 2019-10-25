package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pre_deploy_check command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("When there are products that are mis-configured", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/info"),
					ghttp.RespondWith(http.StatusOK, `{
						"info": {
							"version": "2.6.0"
						}
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/pre_deploy_check"),
					ghttp.RespondWith(http.StatusOK, misconfiguredDirectorJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[{"guid":"p-guid"}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/p-guid/pre_deploy_check"),
					ghttp.RespondWith(http.StatusOK, misconfiguredProductJSON),
				),
			)
		})

		It("exits with an error if director or products are not completely configured", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"pre-deploy-check")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))

			Expect(string(session.Out.Contents())).To(Equal(`Scanning OpsManager now ...

[X] director: p-bosh-guid
    Error: Availability Zone is not assigned

    Error: missing stemcell
    Why: Required stemcell OS: ubuntu-xenial version 250.2
    Fix: Download ubuntu-xenial version 250.2 from Pivnet and upload to OpsManager

    Error: property: .properties.iaas_configuration.project
    Why: can't be blank

    Error: resource: job-identifier
    Why: Instance : Value must be a positive integer

    Error: verifier: NetworksPingableVerifier
    Why: NetworksPingableVerifier error
    Disable: ` + "`om disable-director-verifiers --type NetworksPingableVerifier`" + `

[X] product: p-guid
    Error: Availability Zone is not assigned

    Error: missing stemcell
    Why: Required stemcell OS: ubuntu-xenial version 250.2
    Fix: Download ubuntu-xenial version 250.2 from Pivnet and upload to OpsManager

    Error: property: .properties.iaas_configuration.project
    Why: can't be blank

    Error: resource: job-identifier
    Why: Instance : Value must be a positive integer

    Error: verifier: NetworksPingableVerifier
    Why: NetworksPingableVerifier error
    Disable: ` + "`om disable-product-verifiers --product-name p --type NetworksPingableVerifier`" + `

`))
		})
	})

	Describe("all products are good", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/info"),
					ghttp.RespondWith(http.StatusOK, `{
						"info": {
							"version": "2.6.0"
						}
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/pre_deploy_check"),
					ghttp.RespondWith(http.StatusOK, correctlyConfiguredDirectorJSON),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[{"guid":"p-guid"}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/p-guid/pre_deploy_check"),
					ghttp.RespondWith(http.StatusOK, correctlyConfiguredProductJSON),
				),
			)
		})

		It("exits with an error if director or products are not completely configured", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"pre-deploy-check")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(string(session.Out.Contents())).To(Equal(`Scanning OpsManager now ...

[✓] director: p-bosh-guid
[✓] product: p-guid

The director and products are configured correctly.
`))
		})
	})
})

const misconfiguredDirectorJSON = `{
	"pre_deploy_check": {
		"identifier": "p-bosh-guid",
		"complete": false,
		"network": {
			"assigned": true
		},
		"availability_zone": {
			"assigned": false
		},
		"stemcells": [{
			"assigned": false,
			"required_stemcell_version": "250.2",
			"required_stemcell_os": "ubuntu-xenial"
		}],
		"properties": [{
			"name": ".properties.iaas_configuration.project",
			"type": null,
			"errors": [
				"can't be blank"
			]
		}],
		"resources": {
			"jobs": [{
				"identifier": "job-identifier",
				"guid": "job-guid",
				"error": [
					"Instance : Value must be a positive integer"
				]
			}]
		},
		"verifiers": [{
			"type": "NetworksPingableVerifier",
			"errors": [
				"NetworksPingableVerifier error"
			],
			"ignorable": true
		}]
	}
}`

const misconfiguredProductJSON = `{
	"pre_deploy_check": {
		"identifier": "p-guid",
		"complete": false,
		"network": {
			"assigned": true
		},
		"availability_zone": {
			"assigned": false
		},
		"stemcells": [{
			"assigned": false,
			"required_stemcell_version": "250.2",
			"required_stemcell_os": "ubuntu-xenial"
		}],
		"properties": [{
			"name": ".properties.iaas_configuration.project",
			"type": null,
			"errors": [
				"can't be blank"
			]
		}],
		"resources": {
			"jobs": [{
				"identifier": "job-identifier",
				"guid": "job-guid",
				"error": [
					"Instance : Value must be a positive integer"
				]
			}]
		},
		"verifiers": [{
			"type": "NetworksPingableVerifier",
			"errors": [
				"NetworksPingableVerifier error"
			],
			"ignorable": true
		}]
	}
}`

const correctlyConfiguredDirectorJSON = `{
	"pre_deploy_check": {
		"identifier": "p-bosh-guid",
		"complete": true,
		"network": {
			"assigned": true
		},
		"availability_zone": {
			"assigned": true
		},
		"stemcells": [{
			"assigned": true,
			"required_stemcell_version": "250.2",
			"required_stemcell_os": "ubuntu-xenial"
		}],
		"properties": [],
		"resources": {
			"jobs": []
		},
		"verifiers": []
	}
}`

const correctlyConfiguredProductJSON = `{
	"pre_deploy_check": {
		"identifier": "p-guid",
		"complete": true,
		"network": {
			"assigned": true
		},
		"availability_zone": {
			"assigned": true
		},
		"stemcells": [{
			"assigned": true,
			"required_stemcell_version": "250.2",
			"required_stemcell_os": "ubuntu-xenial"
		}],
		"properties": [],
		"resources": {
			"jobs": []
		},
		"verifiers": []
	}
}`
