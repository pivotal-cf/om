package commands_test

import (
	"errors"
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
			Expect(err).ToNot(HaveOccurred())

			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("Getting expiring certificates...")))
			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("[✓] No certificates are expiring in 3m")))
		})

		It("sets ExpiresWithin when passed", func() {
			command := commands.NewExpiringCertificates(service, logger)
			err := command.Execute([]string{
				"--expires-within",
				"5w",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(service.ListExpiringCertificatesArgsForCall(0)).To(Equal("5w"))

			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("Getting expiring certificates...")))
			Expect(stdout).To(gbytes.Say(regexp.QuoteMeta("[✓] No certificates are expiring in 5w")))
		})
	})

	When("there are expiring certs", func() {
		It("prints a clear message of the cert expiring or expired", func() {
			omTime := "2999-01-01T01:01:01Z"
			opsManagerUntilTime, err := time.Parse(time.RFC3339, omTime)
			Expect(err).ToNot(HaveOccurred())
			credhubTime := "2999-12-12T12:12:12Z"
			credhubUntilTime, err := time.Parse(time.RFC3339, credhubTime)
			Expect(err).ToNot(HaveOccurred())
			credhubTimeAlreadyExpired := "2015-12-12T12:12:12Z"
			credhubUntilTimeAlreadyExpired, err := time.Parse(time.RFC3339, credhubTimeAlreadyExpired)
			Expect(err).ToNot(HaveOccurred())

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
						ValidUntil:        credhubUntilTimeAlreadyExpired,
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

			contentsStr := string(stdout.Contents())
			Expect(contentsStr).To(ContainSubstring("Getting expiring certificates..."))
			Expect(contentsStr).To(ContainSubstring("[X] Credhub Location"))
			Expect(contentsStr).To(ContainSubstring(fmt.Sprintf("    /opsmgr/bosh_dns/other_ca: expired on %s", credhubUntilTimeAlreadyExpired.Format(time.RFC822))))
			Expect(contentsStr).To(ContainSubstring(fmt.Sprintf("    /opsmgr/bosh_dns/tls_ca: expiring on %s", credhubUntilTime.Format(time.RFC822))))
			Expect(contentsStr).To(ContainSubstring("[X] Ops Manager"))
			Expect(contentsStr).To(ContainSubstring(fmt.Sprintf("    product-guid-1:")))
			Expect(contentsStr).To(ContainSubstring(fmt.Sprintf("        property-reference-1: expiring on %s", opsManagerUntilTime.Format(time.RFC822))))
			Expect(contentsStr).To(ContainSubstring(fmt.Sprintf("        property-reference-2: expiring on %s", opsManagerUntilTime.Format(time.RFC822))))
			Expect(contentsStr).To(ContainSubstring(fmt.Sprintf("    product-guid-2:")))
			Expect(contentsStr).To(ContainSubstring(fmt.Sprintf("        property-reference-3: expiring on %s", opsManagerUntilTime.Format(time.RFC822))))
			Expect(contentsStr).To(ContainSubstring("[X] Other Location"))
			Expect(contentsStr).To(ContainSubstring(fmt.Sprintf("    product-guid-4:")))
			Expect(contentsStr).To(ContainSubstring(fmt.Sprintf("        property-reference-4: expiring on %s", opsManagerUntilTime.Format(time.RFC822))))
			Expect(contentsStr).To(ContainSubstring(""))
		})

		It("sets ExpiresWithin to 3m as default", func() {
			command := commands.NewExpiringCertificates(service, logger)
			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(service.ListExpiringCertificatesArgsForCall(0)).To(Equal("3m"))
		})

		It("sets ExpiresWithin when passed", func() {
			command := commands.NewExpiringCertificates(service, logger)
			err := command.Execute([]string{
				"--expires-within",
				"5w",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(service.ListExpiringCertificatesArgsForCall(0)).To(Equal("5w"))
		})

		It("validates the ExpiresWithin value as d,w,m,or y when passed", func() {
			command := commands.NewExpiringCertificates(service, logger)
			err := command.Execute([]string{
				"--expires-within",
				"1s",
			})
			Expect(err).To(MatchError(ContainSubstring("only d,w,m, or y are supported. Default is \"3m\"")))

			command = commands.NewExpiringCertificates(service, logger)
			err = command.Execute([]string{
				"--expires-within",
				"0d",
			})
			Expect(err).To(MatchError(ContainSubstring("only d,w,m, or y are supported. Default is \"3m\"")))

			err = command.Execute([]string{
				"--expires-within",
				"2d",
			})
			Expect(err).ToNot(HaveOccurred())

			err = command.Execute([]string{
				"--expires-within",
				"11y",
			})
			Expect(err).ToNot(HaveOccurred())

			err = command.Execute([]string{
				"--expires-within",
				"109w",
			})
			Expect(err).ToNot(HaveOccurred())

			err = command.Execute([]string{
				"--expires-within",
				"20m",
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("certs cannot be fetched", func() {
		It("returns an error", func() {
			service.ListExpiringCertificatesReturns(nil, errors.New("an api error"))
			command := commands.NewExpiringCertificates(service, logger)

			err := command.Execute([]string{})
			Expect(err).To(MatchError(ContainSubstring("could not fetch expiring certificates: an api error")))
		})
	})
})
