package commands_test

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = FDescribe("ExpiringCertificates", func() {
	var (
		service *fakes.ExpiringCertsService
		stdout  *gbytes.Buffer
		logger  *log.Logger
	)

	BeforeEach(func() {
		service = &fakes.ExpiringCertsService{}
		stdout = gbytes.NewBuffer()
		logger = log.New(stdout, "", 0)
	})

	When("there are no expiring certificates in the time range", func() {
		It("displays a helpful message", func() {
			command := commands.NewExpiringCertificates(service, logger)
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("Getting expiring certificates...")))
			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("[✓] Ops Manager")))
			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("[✓] Credhub")))
		})
	})

	When("there are expiringing certs", func() {
		/*
				{
					"issuer": "/C=US/O=Pivotal",
					"valid_from": "2017-02-23T19:31:00Z",
					"valid_until": "2019-02-23T19:31:00Z",
					"configurable": false,
					"property_reference": ".properties.director_ssl",
					"property_type": "rsa_cert_credentials",
					"product_guid": "p-bosh-47f3d0d7ef2f573fbc95",
					"location": "ops_manager",
					"variable_path": null
					},
			    {
					"issuer": "/CN=opsmgr-bosh-dns-tls-ca",
					"valid_from": "2018-08-10T21:07:37Z",
					"valid_until": "2022-08-09T21:07:37Z",
					"configurable": false,
					"property_reference": null,
					"property_type": null,
					"product_guid": null,
					"location": "credhub",
					"variable_path": "/opsmgr/bosh_dns/tls_ca"
					}
		*/
		It("prints a clear message of the cert expiring", func() {
			omTime := "2019-01-01T01:01:01Z"
			opsManagerUntilTime, err := time.Parse(time.RFC3339, omTime)
			Expect(err).NotTo(HaveOccurred())
			credhubTime := "2019-12-12T12:12:12Z"
			credhubUntilTime, err := time.Parse(time.RFC3339, credhubTime)
			Expect(err).NotTo(HaveOccurred())

			service.ListExpiringCertificatesStub = func(duration string) ([]api.ExpiringCertificate, error) {
				return []api.ExpiringCertificate{
					{
						Issuer:            "",
						ValidFrom:         time.Time{},
						ValidUntil:        opsManagerUntilTime,
						Configurable:      false,
						PropertyReference: "",
						PropertyType:      "",
						ProductGUID:       "",
						Location:          "ops_manager",
						VariablePath:      "",
					},
					{
						Issuer:            "",
						ValidFrom:         time.Time{},
						ValidUntil:        credhubUntilTime,
						Configurable:      false,
						PropertyReference: "",
						PropertyType:      "",
						ProductGUID:       "",
						Location:          "credhub",
						VariablePath:      "",
					},
				}, nil
			}
			command := commands.NewExpiringCertificates(service, logger)
			err = command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("Getting expiring certificates...")))
			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("[X] Ops Manager")))
			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta(fmt.Sprintf("    Cert expiring on: %s", opsManagerUntilTime))))
			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("[X] Credhub")))
			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta(fmt.Sprintf("    Cert expiring on: %s", credhubUntilTime))))
		})
	})
})
