package commands_test

import (
	"errors"
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("ExpiringCertificates", func() {
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
			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("[✓] No certificates are expiring in 3m")))
		})

		It("sets ExpiresWithin when passed", func() {
			command := commands.NewExpiringCertificates(service, logger)
			err := command.Execute([]string{
				"--expires-within",
				"5w",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.ListExpiringCertificatesArgsForCall(0)).To(Equal("5w"))

			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("Getting expiring certificates...")))
			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("[✓] No certificates are expiring in 5w")))
		})
	})

	When("there are expiring certs", func() {
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
						PropertyReference: "property-reference-1",
						PropertyType:      "",
						ProductGUID:       "product-guid-1",
						Location:          "ops_manager",
						VariablePath:      "",
					},
					{
						Issuer:            "",
						ValidFrom:         time.Time{},
						ValidUntil:        opsManagerUntilTime,
						Configurable:      false,
						PropertyReference: "property-reference-2",
						PropertyType:      "",
						ProductGUID:       "product-guid-1",
						Location:          "ops_manager",
						VariablePath:      "",
					},
					{
						Issuer:            "",
						ValidFrom:         time.Time{},
						ValidUntil:        opsManagerUntilTime,
						Configurable:      false,
						PropertyReference: "property-reference-3",
						PropertyType:      "",
						ProductGUID:       "product-guid-2",
						Location:          "ops_manager",
						VariablePath:      "",
					},
					{
						Issuer:            "",
						ValidFrom:         time.Time{},
						ValidUntil:        opsManagerUntilTime,
						Configurable:      false,
						PropertyReference: "property-reference-4",
						PropertyType:      "",
						ProductGUID:       "product-guid-4",
						Location:          "other_location",
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
						Location:          "credhub_location",
						VariablePath:      "/opsmgr/bosh_dns/other_ca",
					},
					{
						Issuer:            "",
						ValidFrom:         time.Time{},
						ValidUntil:        credhubUntilTime,
						Configurable:      false,
						PropertyReference: "",
						PropertyType:      "",
						ProductGUID:       "",
						Location:          "credhub_location",
						VariablePath:      "/opsmgr/bosh_dns/tls_ca",
					},
				}, nil
			}
			command := commands.NewExpiringCertificates(service, logger)
			err = command.Execute([]string{})
			Expect(err).To(HaveOccurred())

			contents := strings.Split(string(stdout.Contents()), "\n")
			Expect(contents).To(ConsistOf(
				"Getting expiring certificates...",
				"[X] Credhub Location",
				fmt.Sprintf("    /opsmgr/bosh_dns/other_ca: expiring on %s", credhubUntilTime.Format(time.RFC822)),
				fmt.Sprintf("    /opsmgr/bosh_dns/tls_ca: expiring on %s", credhubUntilTime.Format(time.RFC822)),
				"[X] Ops Manager",
				fmt.Sprintf("    product-guid-1:"),
				fmt.Sprintf("        property-reference-1: expiring on %s", opsManagerUntilTime.Format(time.RFC822)),
				fmt.Sprintf("        property-reference-2: expiring on %s", opsManagerUntilTime.Format(time.RFC822)),
				fmt.Sprintf("    product-guid-2:"),
				fmt.Sprintf("        property-reference-3: expiring on %s", opsManagerUntilTime.Format(time.RFC822)),
				"[X] Other Location",
				fmt.Sprintf("    product-guid-4:"),
				fmt.Sprintf("        property-reference-4: expiring on %s", opsManagerUntilTime.Format(time.RFC822)),
				"",
			))
		})

		It("sets ExpiresWithin to 3m as default", func() {
			command := commands.NewExpiringCertificates(service, logger)
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.ListExpiringCertificatesArgsForCall(0)).To(Equal("3m"))
		})

		It("sets ExpiresWithin when passed", func() {
			command := commands.NewExpiringCertificates(service, logger)
			err := command.Execute([]string{
				"--expires-within",
				"5w",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.ListExpiringCertificatesArgsForCall(0)).To(Equal("5w"))
		})

		It("validates the ExpiresWithin value as d,w,m,or y when passed", func() {
			command := commands.NewExpiringCertificates(service, logger)
			err := command.Execute([]string{
				"--expires-within",
				"1s",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("only d,w,m, or y are supported. Default is \"3m\""))

			command.Options.ExpiresWithin = "2d"
			err = command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())

			command.Options.ExpiresWithin = "11y"
			err = command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("certs cannot be fetched", func() {
		It("returns an error", func() {
			service.ListExpiringCertificatesReturns(nil, errors.New("an api error"))
			command := commands.NewExpiringCertificates(service, logger)

			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("could not fetch expiring certificates: an api error"))
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewStagedConfig(nil, nil)

			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command generates a config from a staged product that can be passed in to om configure-product (Note: credentials are not available and will appear as '***')",
				ShortDescription: "**EXPERIMENTAL** generates a config from a staged product",
				Flags:            command.Options,
			}))
		})
	})
})
