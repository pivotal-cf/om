package commands_test

import (
	"errors"
	"log"
	"time"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("GetCertificates", func() {
	var (
		service *fakes.GetCertificatesService
		stdout  *gbytes.Buffer
		logger  *log.Logger
		command *commands.GetCertificates
	)

	BeforeEach(func() {
		service = &fakes.GetCertificatesService{}
		stdout = gbytes.NewBuffer()
		logger = log.New(stdout, "", 0)
		command = commands.NewGetCertificates(service, logger)
	})

	Describe("Execute", func() {
		Context("when no certificates are found for the product", func() {
			It("displays a helpful message", func() {
				service.ListDeployedCertificatesReturns([]api.ExpiringCertificate{}, nil)
				service.ListDeployedProductsReturns([]api.DeployedProductOutput{}, nil)

				command.Options.Product = "cf"
				err := command.Execute([]string{})

				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).To(gbytes.Say("Getting certificates for cf..."))
				Expect(stdout).To(gbytes.Say("No certificates found for product 'cf'"))
			})
		})

		Context("when certificates are found for the product", func() {
			var expiryTime time.Time

			BeforeEach(func() {
				expiryTime = time.Now().Add(30 * 24 * time.Hour) // 30 days from now

				service.ListDeployedProductsReturns([]api.DeployedProductOutput{
					{
						GUID: "cf-guid-123",
						Type: "cf",
					},
					{
						GUID: "p-bosh-guid-456",
						Type: "p-bosh",
					},
				}, nil)

				service.ListDeployedCertificatesReturns([]api.ExpiringCertificate{
					{
						ProductGUID:       "cf-guid-123",
						PropertyReference: "cert-property-1",
						ValidUntil:        expiryTime,
						Location:          "deployed_products",
					},
					{
						ProductGUID:       "p-bosh-guid-456",
						PropertyReference: "cert-property-2",
						ValidUntil:        expiryTime,
						Location:          "deployed_products",
					},
				}, nil)
			})

			It("processes certificates for the specified product", func() {
				validPEM := `-----BEGIN CERTIFICATE-----
MIIDaTCCAlGgAwIBAgIUXjSvpwI5RX6NQdS7iwAAAAAAAMwwDQYJKoZIhvcNAQEL
BQAwRzELMAkGA1UEBhMCVVMxDjAMBgNVBAoMBUR1bW15MSgwJgYDVQQDDB9sb2Mt
Y2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMB4XDTI1MDcyOTE0NTY0NVoXDTI3
MDcyOTE0NTY0NVowRzELMAkGA1UEBhMCVVMxDjAMBgNVBAoMBUR1bW15MSgwJgYD
VQQDDB9sb2ctY2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxkrPEu1uD1QcvVmHgql60r1u0fl4BbmB+5pQ
9J8wnbSOpefbq6YiTb8auHf/ChpwrQnIVv4NiFYOy4s73CutY1vXSalfhrbzMdug
GGOtZB0LVtVHZi1GGxysx9DteDFHsKuPCCa+LrSLR89b2doP6jZ/031L7rn7J+k2
wevEStmtjLAiekiMx+b4caWKYmhHjmHfgw9r5obEFN3JSKfLNPBqAbEjyTPjx4U6
BxaxJwaABeY6t8iJKXs+pmbDoh5BrXgOriLzyFy3ws8oP+gK9aSLBZk9/37LMql+
fE/JB4EWTz+9LDeKfRANlWGgcBqqlWb6E9kyV8TqC4ncvIHIcwIDAQABo00wSzAq
BgNVHREEIzAhgh9sb2ctY2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMB0GA1Ud
DgQWBBSBl0c4tqKJnWkK3+ac0edqmXJRSTANBgkqhkiG9w0BAQsFAAOCAQEAVBMv
/+yW4XkZxqzjTm1ZCryXT8+mtD7tYLSuuvHFWKWsAnVjf1Ve2V8OH5caZEoAeT7b
Z4XHCs7sLlg81HLI9tXMOhlrRATgC+ccnuGom2Ts1e4mqworhz5/uhF35ci6+qGv
Zv8a+d9NK1mm5vIPv4y2jE2bE3+tR8ggtrkxPRoRnvlzFx8C7XK9xVHZH88UOJSl
1lVnJAeOd3VTyy3ADqHRmoh3gTq6lL5CfDn4gfdxAxfaLsuwmusu9I4Zt1uKRA2A
kPg01BLHtNI/U4oTDLku7wLJGFhXBAnLVpomuV5m6xwHZ4eOPvsp6XN1laUBwaAu
svTEb6EMuB8T2B9Rtg==
-----END CERTIFICATE-----`

				service.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
					Credential: api.Credential{
						Value: map[string]string{
							"cert_pem": validPEM,
						},
					},
				}, nil)

				command.Options.Product = "cf"
				err := command.Execute([]string{})

				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).To(gbytes.Say("Getting certificates for cf..."))
				Expect(stdout).To(gbytes.Say("Processing 1 certificates"))

				// Verify API calls
				Expect(service.ListDeployedCertificatesCallCount()).To(Equal(1))
				Expect(service.ListDeployedProductsCallCount()).To(Equal(1))
				Expect(service.GetDeployedProductCredentialCallCount()).To(Equal(1))

				input := service.GetDeployedProductCredentialArgsForCall(0)
				Expect(input.DeployedGUID).To(Equal("cf-guid-123"))
				Expect(input.CredentialReference).To(Equal("cert-property-1"))
			})
		})

		Context("API failure scenarios", func() {
			It("returns an error when ListDeployedCertificates fails", func() {
				service.ListDeployedCertificatesReturns(nil, errors.New("network error"))

				command.Options.Product = "cf"
				err := command.Execute([]string{})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to fetch deployed certificates: network error"))
			})

			It("returns an error when ListDeployedProducts fails", func() {
				service.ListDeployedCertificatesReturns([]api.ExpiringCertificate{}, nil)
				service.ListDeployedProductsReturns(nil, errors.New("authentication failed"))

				command.Options.Product = "cf"
				err := command.Execute([]string{})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to fetch deployed products: authentication failed"))
			})

			It("handles GetDeployedProductCredential API failures gracefully", func() {
				expiryTime := time.Now().Add(30 * 24 * time.Hour)

				service.ListDeployedProductsReturns([]api.DeployedProductOutput{
					{GUID: "cf-guid-123", Type: "cf"},
				}, nil)

				service.ListDeployedCertificatesReturns([]api.ExpiringCertificate{
					{
						ProductGUID:       "cf-guid-123",
						PropertyReference: "cert-property-1",
						ValidUntil:        expiryTime,
					},
				}, nil)

				service.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{}, errors.New("credential fetch failed"))

				command.Options.Product = "cf"
				err := command.Execute([]string{})

				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).To(gbytes.Say("Getting certificates for cf..."))
				Expect(stdout).To(gbytes.Say("credential fetch failed"))
			})

			It("handles 404 errors from GetDeployedProductCredential", func() {
				expiryTime := time.Now().Add(30 * 24 * time.Hour)

				service.ListDeployedProductsReturns([]api.DeployedProductOutput{
					{GUID: "cf-guid-123", Type: "cf"},
				}, nil)

				service.ListDeployedCertificatesReturns([]api.ExpiringCertificate{
					{
						ProductGUID:       "cf-guid-123",
						PropertyReference: "non-existent-cert",
						ValidUntil:        expiryTime,
					},
				}, nil)

				service.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{}, errors.New("request failed: unexpected response 404"))

				command.Options.Product = "cf"
				err := command.Execute([]string{})

				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).To(gbytes.Say("credential not found for reference 'non-existent-cert'"))
			})
		})

		Context("edge cases and data validation", func() {
			BeforeEach(func() {
				service.ListDeployedProductsReturns([]api.DeployedProductOutput{
					{GUID: "cf-guid-123", Type: "cf"},
				}, nil)
			})

			It("handles certificates with missing ProductGUID", func() {
				service.ListDeployedCertificatesReturns([]api.ExpiringCertificate{
					{
						ProductGUID:       "", // Missing GUID
						PropertyReference: "cert-property-1",
						ValidUntil:        time.Now().Add(30 * 24 * time.Hour),
					},
				}, nil)

				command.Options.Product = "cf"
				err := command.Execute([]string{})

				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).To(gbytes.Say("No certificates found for product 'cf'"))
			})

			It("handles certificates with missing PropertyReference", func() {
				service.ListDeployedCertificatesReturns([]api.ExpiringCertificate{
					{
						ProductGUID:       "cf-guid-123",
						PropertyReference: "", // Missing reference
						ValidUntil:        time.Now().Add(30 * 24 * time.Hour),
					},
				}, nil)

				command.Options.Product = "cf"
				err := command.Execute([]string{})

				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).To(gbytes.Say("missing product_guid or property_reference"))
			})

			It("handles credentials missing cert_pem field", func() {
				expiryTime := time.Now().Add(30 * 24 * time.Hour)

				service.ListDeployedCertificatesReturns([]api.ExpiringCertificate{
					{
						ProductGUID:       "cf-guid-123",
						PropertyReference: "cert-property-1",
						ValidUntil:        expiryTime,
					},
				}, nil)

				service.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
					Credential: api.Credential{
						Value: map[string]string{
							"private_key": "some-key", // Missing cert_pem
						},
					},
				}, nil)

				command.Options.Product = "cf"
				err := command.Execute([]string{})

				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).To(gbytes.Say("cert_pem not found in credential"))
			})

			It("handles credentials with empty cert_pem", func() {
				expiryTime := time.Now().Add(30 * 24 * time.Hour)

				service.ListDeployedCertificatesReturns([]api.ExpiringCertificate{
					{
						ProductGUID:       "cf-guid-123",
						PropertyReference: "cert-property-1",
						ValidUntil:        expiryTime,
					},
				}, nil)

				service.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
					Credential: api.Credential{
						Value: map[string]string{
							"cert_pem": "", // Empty cert_pem
						},
					},
				}, nil)

				command.Options.Product = "cf"
				err := command.Execute([]string{})

				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).To(gbytes.Say("cert_pem not found in credential"))
			})

			It("handles invalid PEM data", func() {
				expiryTime := time.Now().Add(30 * 24 * time.Hour)

				service.ListDeployedCertificatesReturns([]api.ExpiringCertificate{
					{
						ProductGUID:       "cf-guid-123",
						PropertyReference: "cert-property-1",
						ValidUntil:        expiryTime,
					},
				}, nil)

				service.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
					Credential: api.Credential{
						Value: map[string]string{
							"cert_pem": "invalid-pem-data",
						},
					},
				}, nil)

				command.Options.Product = "cf"
				err := command.Execute([]string{})

				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).To(gbytes.Say("failed to extract serial number"))
			})
		})

		Context("filtering and processing", func() {
			It("filters certificates by the correct product type", func() {
				expiryTime := time.Now().Add(30 * 24 * time.Hour)

				service.ListDeployedProductsReturns([]api.DeployedProductOutput{
					{GUID: "cf-guid-123", Type: "cf"},
					{GUID: "p-bosh-guid-456", Type: "p-bosh"},
					{GUID: "mysql-guid-789", Type: "pivotal-mysql"},
				}, nil)

				service.ListDeployedCertificatesReturns([]api.ExpiringCertificate{
					{
						ProductGUID:       "cf-guid-123",
						PropertyReference: "cf-cert",
						ValidUntil:        expiryTime,
					},
					{
						ProductGUID:       "p-bosh-guid-456",
						PropertyReference: "bosh-cert",
						ValidUntil:        expiryTime,
					},
					{
						ProductGUID:       "mysql-guid-789",
						PropertyReference: "mysql-cert",
						ValidUntil:        expiryTime,
					},
				}, nil)

				validPEM := `-----BEGIN CERTIFICATE-----
MIIDaTCCAlGgAwIBAgIUXjSvpwI5RX6NQdS7iwAAAAAAAMwwDQYJKoZIhvcNAQEL
BQAwRzELMAkGA1UEBhMCVVMxDjAMBgNVBAoMBUR1bW15MSgwJgYDVQQDDB9sb2Mt
Y2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMB4XDTI1MDcyOTE0NTY0NVoXDTI3
MDcyOTE0NTY0NVowRzELMAkGA1UEBhMCVVMxDjAMBgNVBAoMBUR1bW15MSgwJgYD
VQQDDB9sb2ctY2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxkrPEu1uD1QcvVmHgql60r1u0fl4BbmB+5pQ
9J8wnbSOpefbq6YiTb8auHf/ChpwrQnIVv4NiFYOy4s73CutY1vXSalfhrbzMdug
GGOtZB0LVtVHZi1GGxysx9DteDFHsKuPCCa+LrSLR89b2doP6jZ/031L7rn7J+k2
wevEStmtjLAiekiMx+b4caWKYmhHjmHfgw9r5obEFN3JSKfLNPBqAbEjyTPjx4U6
BxaxJwaABeY6t8iJKXs+pmbDoh5BrXgOriLzyFy3ws8oP+gK9aSLBZk9/37LMql+
fE/JB4EWTz+9LDeKfRANlWGgcBqqlWb6E9kyV8TqC4ncvIHIcwIDAQABo00wSzAq
BgNVHREEIzAhgh9sb2ctY2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMB0GA1Ud
DgQWBBSBl0c4tqKJnWkK3+ac0edqmXJRSTANBgkqhkiG9w0BAQsFAAOCAQEAVBMv
/+yW4XkZxqzjTm1ZCryXT8+mtD7tYLSuuvHFWKWsAnVjf1Ve2V8OH5caZEoAeT7b
Z4XHCs7sLlg81HLI9tXMOhlrRATgC+ccnuGom2Ts1e4mqworhz5/uhF35ci6+qGv
Zv8a+d9NK1mm5vIPv4y2jE2bE3+tR8ggtrkxPRoRnvlzFx8C7XK9xVHZH88UOJSl
1lVnJAeOd3VTyy3ADqHRmoh3gTq6lL5CfDn4gfdxAxfaLsuwmusu9I4Zt1uKRA2A
kPg01BLHtNI/U4oTDLku7wLJGFhXBAnLVpomuV5m6xwHZ4eOPvsp6XN1laUBwaAu
svTEb6EMuB8T2B9Rtg==
-----END CERTIFICATE-----`

				service.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
					Credential: api.Credential{
						Value: map[string]string{
							"cert_pem": validPEM,
						},
					},
				}, nil)

				// Test filtering for p-bosh
				command.Options.Product = "p-bosh"
				err := command.Execute([]string{})

				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).To(gbytes.Say("Processing 1 certificates"))

				// Verify only p-bosh certificate was processed
				input := service.GetDeployedProductCredentialArgsForCall(0)
				Expect(input.DeployedGUID).To(Equal("p-bosh-guid-456"))
				Expect(input.CredentialReference).To(Equal("bosh-cert"))
			})

			It("processes multiple certificates concurrently", func() {
				expiryTime := time.Now().Add(30 * 24 * time.Hour)

				service.ListDeployedProductsReturns([]api.DeployedProductOutput{
					{GUID: "cf-guid-123", Type: "cf"},
				}, nil)

				service.ListDeployedCertificatesReturns([]api.ExpiringCertificate{
					{
						ProductGUID:       "cf-guid-123",
						PropertyReference: "cert-1",
						ValidUntil:        expiryTime,
					},
					{
						ProductGUID:       "cf-guid-123",
						PropertyReference: "cert-2",
						ValidUntil:        expiryTime,
					},
					{
						ProductGUID:       "cf-guid-123",
						PropertyReference: "cert-3",
						ValidUntil:        expiryTime,
					},
				}, nil)

				validPEM := `-----BEGIN CERTIFICATE-----
MIIDaTCCAlGgAwIBAgIUXjSvpwI5RX6NQdS7iwAAAAAAAMwwDQYJKoZIhvcNAQEL
BQAwRzELMAkGA1UEBhMCVVMxDjAMBgNVBAoMBUR1bW15MSgwJgYDVQQDDB9sb2Mt
Y2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMB4XDTI1MDcyOTE0NTY0NVoXDTI3
MDcyOTE0NTY0NVowRzELMAkGA1UEBhMCVVMxDjAMBgNVBAoMBUR1bW15MSgwJgYD
VQQDDB9sb2ctY2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxkrPEu1uD1QcvVmHgql60r1u0fl4BbmB+5pQ
9J8wnbSOpefbq6YiTb8auHf/ChpwrQnIVv4NiFYOy4s73CutY1vXSalfhrbzMdug
GGOtZB0LVtVHZi1GGxysx9DteDFHsKuPCCa+LrSLR89b2doP6jZ/031L7rn7J+k2
wevEStmtjLAiekiMx+b4caWKYmhHjmHfgw9r5obEFN3JSKfLNPBqAbEjyTPjx4U6
BxaxJwaABeY6t8iJKXs+pmbDoh5BrXgOriLzyFy3ws8oP+gK9aSLBZk9/37LMql+
fE/JB4EWTz+9LDeKfRANlWGgcBqqlWb6E9kyV8TqC4ncvIHIcwIDAQABo00wSzAq
BgNVHREEIzAhgh9sb2ctY2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMB0GA1Ud
DgQWBBSBl0c4tqKJnWkK3+ac0edqmXJRSTANBgkqhkiG9w0BAQsFAAOCAQEAVBMv
/+yW4XkZxqzjTm1ZCryXT8+mtD7tYLSuuvHFWKWsAnVjf1Ve2V8OH5caZEoAeT7b
Z4XHCs7sLlg81HLI9tXMOhlrRATgC+ccnuGom2Ts1e4mqworhz5/uhF35ci6+qGv
Zv8a+d9NK1mm5vIPv4y2jE2bE3+tR8ggtrkxPRoRnvlzFx8C7XK9xVHZH88UOJSl
1lVnJAeOd3VTyy3ADqHRmoh3gTq6lL5CfDn4gfdxAxfaLsuwmusu9I4Zt1uKRA2A
kPg01BLHtNI/U4oTDLku7wLJGFhXBAnLVpomuV5m6xwHZ4eOPvsp6XN1laUBwaAu
svTEb6EMuB8T2B9Rtg==
-----END CERTIFICATE-----`

				service.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
					Credential: api.Credential{
						Value: map[string]string{
							"cert_pem": validPEM,
						},
					},
				}, nil)

				command.Options.Product = "cf"
				err := command.Execute([]string{})

				Expect(err).ToNot(HaveOccurred())
				Expect(stdout).To(gbytes.Say("Processing 3 certificates"))

				// Verify all three certificates were processed
				Expect(service.GetDeployedProductCredentialCallCount()).To(Equal(3))
			})
		})
	})

	Describe("extractSerialFromPEM", func() {
		It("extracts serial number from valid PEM", func() {
			validPEM := `-----BEGIN CERTIFICATE-----
MIIDaTCCAlGgAwIBAgIUXjSvpwI5RX6NQdS7iwAAAAAAAMwwDQYJKoZIhvcNAQEL
BQAwRzELMAkGA1UEBhMCVVMxDjAMBgNVBAoMBUR1bW15MSgwJgYDVQQDDB9sb2Mt
Y2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMB4XDTI1MDcyOTE0NTY0NVoXDTI3
MDcyOTE0NTY0NVowRzELMAkGA1UEBhMCVVMxDjAMBgNVBAoMBUR1bW15MSgwJgYD
VQQDDB9sb2ctY2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxkrPEu1uD1QcvVmHgql60r1u0fl4BbmB+5pQ
9J8wnbSOpefbq6YiTb8auHf/ChpwrQnIVv4NiFYOy4s73CutY1vXSalfhrbzMdug
GGOtZB0LVtVHZi1GGxysx9DteDFHsKuPCCa+LrSLR89b2doP6jZ/031L7rn7J+k2
wevEStmtjLAiekiMx+b4caWKYmhHjmHfgw9r5obEFN3JSKfLNPBqAbEjyTPjx4U6
BxaxJwaABeY6t8iJKXs+pmbDoh5BrXgOriLzyFy3ws8oP+gK9aSLBZk9/37LMql+
fE/JB4EWTz+9LDeKfRANlWGgcBqqlWb6E9kyV8TqC4ncvIHIcwIDAQABo00wSzAq
BgNVHREEIzAhgh9sb2ctY2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMB0GA1Ud
DgQWBBSBl0c4tqKJnWkK3+ac0edqmXJRSTANBgkqhkiG9w0BAQsFAAOCAQEAVBMv
/+yW4XkZxqzjTm1ZCryXT8+mtD7tYLSuuvHFWKWsAnVjf1Ve2V8OH5caZEoAeT7b
Z4XHCs7sLlg81HLI9tXMOhlrRATgC+ccnuGom2Ts1e4mqworhz5/uhF35ci6+qGv
Zv8a+d9NK1mm5vIPv4y2jE2bE3+tR8ggtrkxPRoRnvlzFx8C7XK9xVHZH88UOJSl
1lVnJAeOd3VTyy3ADqHRmoh3gTq6lL5CfDn4gfdxAxfaLsuwmusu9I4Zt1uKRA2A
kPg01BLHtNI/U4oTDLku7wLJGFhXBAnLVpomuV5m6xwHZ4eOPvsp6XN1laUBwaAu
svTEb6EMuB8T2B9Rtg==
-----END CERTIFICATE-----`

			serial, err := commands.ExtractSerialFromPEM(validPEM)
			Expect(err).ToNot(HaveOccurred())
			Expect(serial).ToNot(BeEmpty())
		})

		It("returns error for invalid PEM data", func() {
			invalidPEM := "invalid-pem-data"

			serial, err := commands.ExtractSerialFromPEM(invalidPEM)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to decode PEM"))
			Expect(serial).To(BeEmpty())
		})

		It("returns error for empty PEM data", func() {
			serial, err := commands.ExtractSerialFromPEM("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to decode PEM"))
			Expect(serial).To(BeEmpty())
		})

		It("returns error for malformed certificate data", func() {
			malformedPEM := `-----BEGIN CERTIFICATE-----
invalid-certificate-data
-----END CERTIFICATE-----`

			serial, err := commands.ExtractSerialFromPEM(malformedPEM)
			Expect(err).To(HaveOccurred())
			Expect(serial).To(BeEmpty())
		})
	})
})
