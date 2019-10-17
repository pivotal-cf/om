package acceptance

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"
	"time"
)

var _ = Describe("diagnostic-report command", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
	})

	AfterEach(func() {
		server.Close()
	})

	When("The Operations Manager is version 2.5-", func() {
		It("successfully returns the entire diagnostic report json", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
					ghttp.RespondWith(http.StatusOK, fakeReport25),
				),
			)

			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "pass",
				"--skip-ssl-validation",
				"diagnostic-report",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 10*time.Second).Should(gexec.Exit(0))
			Eventually(session.Out.Contents()).Should(MatchJSON(fakeReport25))
		})
	})

	When("The Operations Manager is version 2.6+", func() {
		It("successfully returns the entire diagnostic report json", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
					ghttp.RespondWith(http.StatusOK, fakeReport26),
				),
			)

			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "pass",
				"--skip-ssl-validation",
				"diagnostic-report",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 10*time.Second).Should(gexec.Exit(0))
			Eventually(session.Out.Contents()).Should(MatchJSON(fakeReport26))
		})
	})
})

const fakeReport26 = `{
  "versions": {
    "installation_schema_version": "2.6",
    "metadata_version": "2.6",
    "release_version": "2.6.0-build.77",
    "javascript_migrations_version": "v1"
  },
  "generation_time": "2019-05-01T19:07:09Z",
  "infrastructure_type": "vsphere",
  "director_configuration": {
    "bosh_recreate_on_next_deploy": false,
    "resurrector_enabled": false,
    "blobstore_type": "local",
    "max_threads": null,
    "database_type": "internal",
    "ntp_servers": [
      "ntp.ubuntu.com"
    ],
    "hm_pager_duty_enabled": false,
    "hm_emailer_enabled": false,
    "vm_password_type": "generate"
  },
  "releases": [

  ],
  "available_stemcells": [
    {
      "filename": "bosh-stemcell-250.29-vsphere-esxi-ubuntu-xenial-go_agent.tgz",
      "os": "ubuntu-xenial",
      "version": "250.29"
    }
  ],
  "product_templates": [
    "product-template20190501-807-lcs5rm.yml",
    "product-template20190501-807-1jzgxdc.yml",
    "product-template20190501-807-iocy0f.yml"
  ],
  "added_products": {
    "deployed": [
      {
        "name": "cf",
        "version": "2.5.0",
        "stemcells": [
          {
            "filename": "bosh-stemcell-170.45-vsphere-esxi-ubuntu-xenial-go_agent.tgz",
            "os": "ubuntu-xenial",
            "version": "170.45"
          }
        ]
      },
      {
        "name": "pas-windows",
        "version": "2.5.0",
        "stemcells": [
          {
            "filename": "bosh-stemcell-1803.6-vsphere-esxi-windows1803-go_agent.tgz",
            "os": "windows1803",
            "version": "1803.6"
          }
        ]
      },
      {
        "name": "p-healthwatch",
        "version": "1.4.4-build.1",
        "stemcells": [
          {
            "filename": "bosh-stemcell-97.74-vsphere-esxi-ubuntu-xenial-go_agent.tgz",
            "os": "ubuntu-xenial",
            "version": "97.74"
          }
        ]
      },
      {
        "name": "p-bosh",
        "version": "2.6.0-build.77",
        "stemcells": [
          {
            "filename": "bosh-stemcell-250.29-vsphere-esxi-ubuntu-xenial-go_agent.tgz",
            "os": "ubuntu-xenial",
            "version": "250.29"
          }
        ]
      }
    ],
    "staged": [
      {
        "name": "cf",
        "version": "2.5.0",
        "stemcells": [
          {
            "filename": "bosh-stemcell-170.45-vsphere-esxi-ubuntu-xenial-go_agent.tgz",
            "os": "ubuntu-xenial",
            "version": "170.45"
          }
        ]
      },
      {
        "name": "pas-windows",
        "version": "2.5.0",
        "stemcells": [
          {
            "filename": "bosh-stemcell-1803.6-vsphere-esxi-windows1803-go_agent.tgz",
            "os": "windows1803",
            "version": "1803.6"
          }
        ]
      },
      {
        "name": "p-healthwatch",
        "version": "1.4.4-build.1",
        "stemcells": [
          {
            "filename": "bosh-stemcell-97.74-vsphere-esxi-ubuntu-xenial-go_agent.tgz",
            "os": "ubuntu-xenial",
            "version": "97.74"
          }
        ]
      },
      {
        "name": "p-bosh",
        "version": "2.6.0-build.77",
        "stemcells": [
          {
            "filename": "bosh-stemcell-250.29-vsphere-esxi-ubuntu-xenial-go_agent.tgz",
            "os": "ubuntu-xenial",
            "version": "250.29"
          }
        ]
      }
    ]
  }
}`

const fakeReport25 = `{
  "versions": {
    "installation_schema_version": "2.5",
    "metadata_version": "2.5",
    "release_version": "2.5.0-build.8",
    "javascript_migrations_version": "v1"
  },
  "generation_time": "2016-04-22T18:06:46Z",
  "infrastructure_type": "vsphere",
  "director_configuration": {
    "bosh_recreate_on_next_deploy": false,
    "resurrector_enabled": false,
    "blobstore_type": "local",
    "max_threads": null,
    "database_type": "internal",
    "ntp_servers": [],
    "hm_pager_duty_enabled": false,
    "hm_emailer_enabled": false,
    "vm_password_type": "generate"
  },
  "releases": [
    "example-release-14.tgz"
  ],
  "stemcells": [
    "bosh-stemcell-3215-vsphere-esxi-ubuntu-trusty-go_agent.tgz"
  ],
  "product_templates": [
    "e08002f028a5.yml"
  ],
  "added_products": {
    "deployed": [],
    "staged": [
      {
        "name": "p-bosh",
        "version": "2.5.0-build.8",
        "stemcell": "bosh-stemcell-3215-vsphere-esxi-ubuntu-trusty-go_agent.tgz"
      },
      {
        "name": "example-product",
        "version": "1.0.0.0-alpha",
        "stemcell": "bosh-stemcell-3215-vsphere-esxi-ubuntu-trusty-go_agent.tgz"
      }
    ]
  }
}`
