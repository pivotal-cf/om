package commands_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"io/ioutil"
	"log"
)

var _ = Describe("ConfigureOpsman", func() {
	var (
		command     commands.ConfigureOpsman
		fakeService *fakes.ConfigureOpsmanService
	)

	BeforeEach(func() {
		fakeService = &fakes.ConfigureOpsmanService{}
		logger := log.New(ioutil.Discard, "", 0)
		command = commands.NewConfigureOpsman(func() []string { return []string{} }, fakeService, logger)
	})

	Describe("Execute", func() {
		It("updates the ssl certificate when given the proper keys", func() {
			sslCertConfig := `
ssl-certificate:
  certificate: some-cert-pem
  private_key: some-private-key
`
			configFileName := writeTestConfigFile(sslCertConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(1))
			Expect(fakeService.UpdateSSLCertificateArgsForCall(0)).To(Equal(api.SSLCertificateInput{
				CertPem:       "some-cert-pem",
				PrivateKeyPem: "some-private-key",
			}))
			Expect(fakeService.UpdatePivnetTokenCallCount()).To(Equal(0))
		})

		It("updates the pivnet token when given the proper keys", func() {
			pivnetConfig := `
pivotal-network-settings:
  api_token: some-token
`
			configFileName := writeTestConfigFile(pivnetConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.UpdatePivnetTokenCallCount()).To(Equal(1))
			Expect(fakeService.UpdatePivnetTokenArgsForCall(0)).To(Equal("some-token"))
			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
		})

		It("errors when no config file is provided", func() {
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing required flag \"--config\""))

			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
		})

		It("returns an error when there is an unrecognized top-level key", func() {
			sslCertConfig := `
invalid-key: 1
`
			configFileName := writeTestConfigFile(sslCertConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unrecognized top level key(s) in config file:\ninvalid-key"))

			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
		})

		It("returns a nicer error if multiple unrecognized keys", func() {
			sslCertConfig := `
invalid-key-one: 1
invalid-key-two: 2
`
			configFileName := writeTestConfigFile(sslCertConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unrecognized top level key(s) in config file:\ninvalid-key-one\ninvalid-key-two"))

			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
		})

		It("does not error when top-level key is opsman-configuration", func() {
			sslCertConfig := `
opsman-configuration: 1
`
			configFileName := writeTestConfigFile(sslCertConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
		})

		It("returns an error if interpolation fails", func() {
			sslCertConfig := `opsman-configuration: ((missing-var))`
			configFileName := writeTestConfigFile(sslCertConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected to find variables: missing-var"))
		})

		Describe("api failures", func() {
			It("returns an error when ssl cert api fails", func() {
				fakeService.UpdateSSLCertificateReturns(errors.New("some error"))

				sslCertConfig := `
ssl-certificate:
  certificate: some-cert-pem
  private_key: some-private-key
`
				configFileName := writeTestConfigFile(sslCertConfig)

				err := command.Execute([]string{
					"--config", configFileName,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("some error"))

				Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(1))
				Expect(fakeService.UpdateSSLCertificateArgsForCall(0)).To(Equal(api.SSLCertificateInput{
					CertPem:       "some-cert-pem",
					PrivateKeyPem: "some-private-key",
				}))
			})

			It("returns an error when pivnet token api fails", func() {
				fakeService.UpdatePivnetTokenReturns(errors.New("some error"))

				sslCertConfig := `
pivotal-network-settings:
  api_token: some-token
`
				configFileName := writeTestConfigFile(sslCertConfig)

				err := command.Execute([]string{
					"--config", configFileName,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("some error"))

				Expect(fakeService.UpdatePivnetTokenCallCount()).To(Equal(1))
			})
		})

		Describe("Usage", func() {
			It("returns usage info", func() {
				usage := command.Usage()
				Expect(usage).To(Equal(jhanda.Usage{
					Description:      "This authenticated command configures settings available on the \"Settings\" page in the Ops Manager UI. For an example config, reference the docs directory for this command.",
					ShortDescription: "configures values present on the Ops Manager settings page",
					Flags:            command.Options,
				}))
			})
		})
	})
})
