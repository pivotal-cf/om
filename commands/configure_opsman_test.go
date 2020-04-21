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
		It("updates the custom banners when given the proper keys", func() {
			bannerConfig := `
banner-settings:
  ui_banner_contents: example UI banner
  ssh_banner_contents: example SSH banner
`
			configFileName := writeTestConfigFile(bannerConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.UpdateBannerCallCount()).To(Equal(1))
			Expect(fakeService.UpdateBannerArgsForCall(0)).To(Equal(api.BannerSettings{
				UIBanner:  "example UI banner",
				SSHBanner: "example SSH banner",
			}))
			Expect(fakeService.UpdatePivnetTokenCallCount()).To(Equal(0))
			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
			Expect(fakeService.EnableRBACCallCount()).To(Equal(0))
		})

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
			Expect(fakeService.UpdateBannerCallCount()).To(Equal(0))
			Expect(fakeService.EnableRBACCallCount()).To(Equal(0))
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
			Expect(fakeService.UpdatePivnetTokenArgsForCall(0)).To(Equal(api.PivnetSettings{
				APIToken: "some-token",
			}))
			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
			Expect(fakeService.UpdateBannerCallCount()).To(Equal(0))
			Expect(fakeService.EnableRBACCallCount()).To(Equal(0))
		})

		It("enables rbac settings for ldap", func() {
			rbacConfig := `
rbac-settings:
  ldap_rbac_admin_group_name: some-ldap-group
`
			configFileName := writeTestConfigFile(rbacConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.EnableRBACCallCount()).To(Equal(1))
			Expect(fakeService.EnableRBACArgsForCall(0)).To(Equal(api.RBACSettings{
				LDAPAdminGroupName: "some-ldap-group",
			}))
			Expect(fakeService.UpdatePivnetTokenCallCount()).To(Equal(0))
			Expect(fakeService.UpdateBannerCallCount()).To(Equal(0))
			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
		})

		It("enables rbac settings for saml", func() {
			rbacConfig := `
rbac-settings:
  rbac_saml_admin_group: some-saml-group
  rbac_saml_groups_attribute: some-saml-attribute
`
			configFileName := writeTestConfigFile(rbacConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.EnableRBACCallCount()).To(Equal(1))
			Expect(fakeService.EnableRBACArgsForCall(0)).To(Equal(api.RBACSettings{
				SAMLAdminGroup:      "some-saml-group",
				SAMLGroupsAttribute: "some-saml-attribute",
			}))
			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
			Expect(fakeService.UpdatePivnetTokenCallCount()).To(Equal(0))
			Expect(fakeService.UpdateBannerCallCount()).To(Equal(0))
		})

		It("updates syslog settings when given the proper keys", func() {
			syslogConfig := `
syslog-settings:
  enabled: true
  address: 1.2.3.4
  port: 999
  transport_protocol: tcp
  tls_enabled: true
  permitted_peer: "*.example.com"
  ssl_ca_certificate: some-cert
  queue_size: 100000
  forward_debug_logs: true
  custom_rsyslog_configuration: some-message
`
			configFileName := writeTestConfigFile(syslogConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.UpdateSyslogSettingsCallCount()).To(Equal(1))
			Expect(fakeService.UpdateSyslogSettingsArgsForCall(0)).To(Equal(api.SyslogSettings{
				Enabled:             "true",
				Address:             "1.2.3.4",
				Port:                "999",
				TransportProtocol:   "tcp",
				TLSEnabled:          "true",
				PermittedPeer:       "*.example.com",
				SSLCACertificate:    "some-cert",
				QueueSize:           "100000",
				ForwardDebugLogs:    "true",
				CustomRsyslogConfig: "some-message",
			}))

			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
			Expect(fakeService.UpdateBannerCallCount()).To(Equal(0))
			Expect(fakeService.EnableRBACCallCount()).To(Equal(0))
			Expect(fakeService.UpdatePivnetTokenCallCount()).To(Equal(0))
		})

		It("returns an error if both ldap and saml keys provided", func() {
			rbacConfig := `
rbac-settings:
  rbac_saml_admin_group: some-saml-group
  rbac_saml_groups_attribute: some-saml-attribute
  ldap_rbac_admin_group_name: some-ldap-group
`
			configFileName := writeTestConfigFile(rbacConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only set SAML or LDAP. Check the config file and use only the appropriate values."))
			Expect(err.Error()).To(ContainSubstring("For example config values, see the docs directory for documentation."))

			Expect(fakeService.EnableRBACCallCount()).To(Equal(0))
			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
			Expect(fakeService.UpdatePivnetTokenCallCount()).To(Equal(0))
			Expect(fakeService.UpdateBannerCallCount()).To(Equal(0))
		})

		It("errors when no config file is provided", func() {
			err := command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing required flag \"--config\""))

			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
		})

		It("returns an error when there is an unrecognized top-level key", func() {
			invalidConfig := `
invalid-key: 1
`
			configFileName := writeTestConfigFile(invalidConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unrecognized top level key(s) in config file:\ninvalid-key"))

			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
			Expect(fakeService.UpdateBannerCallCount()).To(Equal(0))
			Expect(fakeService.UpdatePivnetTokenCallCount()).To(Equal(0))
			Expect(fakeService.EnableRBACCallCount()).To(Equal(0))
		})

		It("returns a nicer error if multiple unrecognized keys", func() {
			invalidConfig := `
invalid-key-one: 1
invalid-key-two: 2
`
			configFileName := writeTestConfigFile(invalidConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unrecognized top level key(s) in config file:\ninvalid-key-one\ninvalid-key-two"))

			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
			Expect(fakeService.UpdateBannerCallCount()).To(Equal(0))
			Expect(fakeService.UpdatePivnetTokenCallCount()).To(Equal(0))
			Expect(fakeService.EnableRBACCallCount()).To(Equal(0))
		})

		It("does not error when top-level key is opsman-configuration", func() {
			opsmanConfiguration := `
opsman-configuration: 1
`
			configFileName := writeTestConfigFile(opsmanConfiguration)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
			Expect(fakeService.UpdateBannerCallCount()).To(Equal(0))
			Expect(fakeService.UpdatePivnetTokenCallCount()).To(Equal(0))
			Expect(fakeService.EnableRBACCallCount()).To(Equal(0))
		})

		It("returns an error if interpolation fails", func() {
			uninterpolatedConfig := `opsman-configuration: ((missing-var))`
			configFileName := writeTestConfigFile(uninterpolatedConfig)

			err := command.Execute([]string{
				"--config", configFileName,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected to find variables: missing-var"))

			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(0))
			Expect(fakeService.UpdateBannerCallCount()).To(Equal(0))
			Expect(fakeService.UpdatePivnetTokenCallCount()).To(Equal(0))
			Expect(fakeService.EnableRBACCallCount()).To(Equal(0))
		})

		Describe("api failures", func() {
			It("returns an error when ssl cert api fails", func() {
				fakeService.UpdateSSLCertificateReturns(errors.New("some error"))

				config := `
ssl-certificate:
  certificate: some-cert-pem
  private_key: some-private-key
`
				configFileName := writeTestConfigFile(config)

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

				config := `
pivotal-network-settings:
  api_token: some-token
`
				configFileName := writeTestConfigFile(config)

				err := command.Execute([]string{
					"--config", configFileName,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("some error"))

				Expect(fakeService.UpdatePivnetTokenCallCount()).To(Equal(1))
			})

			It("returns an error when banner api fails", func() {
				fakeService.UpdateBannerReturns(errors.New("some error"))

				config := `
banner-settings:
  ui_banner_contents: example UI banner
  ssh_banner_contents: example SSH banner
`
				configFileName := writeTestConfigFile(config)

				err := command.Execute([]string{
					"--config", configFileName,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("some error"))

				Expect(fakeService.UpdateBannerCallCount()).To(Equal(1))
			})

			It("returns an error when rbac api fails", func() {
				fakeService.EnableRBACReturns(errors.New("some error"))

				config := `
rbac-settings:
  rbac_saml_admin_group: some-saml-group
  rbac_saml_groups_attribute: some-saml-attribute
`
				configFileName := writeTestConfigFile(config)

				err := command.Execute([]string{
					"--config", configFileName,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("some error"))

				Expect(fakeService.EnableRBACCallCount()).To(Equal(1))
			})

			It("returns an error when syslog api fails", func() {
				fakeService.UpdateSyslogSettingsReturns(errors.New("some error"))

				config := `
syslog-settings:
  enabled: true
  address: 1.2.3.4
  port: 999
  transport_protocol: tcp
  tls_enabled: true
  permitted_peer: "*.example.com"
  ssl_ca_certificate: some-cert
  queue_size: 100000
  forward_debug_logs: true
  custom_rsyslog_configuration: some-message
`
				configFileName := writeTestConfigFile(config)

				err := command.Execute([]string{
					"--config", configFileName,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("some error"))

				Expect(fakeService.UpdateSyslogSettingsCallCount()).To(Equal(1))
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
