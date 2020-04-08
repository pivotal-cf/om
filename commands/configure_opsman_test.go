package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"io/ioutil"
	"log"
	//"github.com/pivotal-cf/om/commands/fakes"
)

var _ = FDescribe("ConfigureOpsman", func() {
	var (
		stdout      *gbytes.Buffer
		command     commands.ConfigureOpsman
		fakeService *fakes.ConfigureOpsmanService
	)

	BeforeEach(func() {
		fakeService = &fakes.ConfigureOpsmanService{}
		logger := log.New(stdout, "", 0)
		command = commands.NewConfigureOpsman(func() []string { return []string{} }, fakeService, logger)
	})

	Describe("Execute", func() {
		It("updates the ssl certificate when given the proper keys", func() {
			sslCertConfig := `
opsman-configuration:
	ssl-certificates:
		certificate-pem: some-cert-pem
		certificate-private-key: some-private-key
`
			configFile, err := ioutil.TempFile("", "sslCertConfig.yml")
			Expect(err).ToNot(HaveOccurred())
			defer configFile.Close()

			_, err = configFile.WriteString(sslCertConfig)
			Expect(err).ToNot(HaveOccurred())
			err = command.Execute([]string{
				"--config", configFile.Name(),
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.UpdateSSLCertificateCallCount()).To(Equal(1))
			Expect(fakeService.UpdateSSLCertificateArgsForCall(0)).To(Equal(api.SSLCertificateInput{
				CertPem:       "some CertPem",
				PrivateKeyPem: "some PrivateKey",
			}))
		})

		It("errors when no config file is provided", func() {

		})

		It("does not fail when given an unrecognized top-level key", func() {

		})

		//When("interpolating", func() {
		//	var (
		//		configFile *os.File
		//		err        error
		//	)
		//
		//	BeforeEach(func() {
		//		service.ListStagedProductsReturns(api.StagedProductsOutput{
		//			Products: []api.StagedProduct{
		//				{GUID: "some-product-guid", Type: "cf"},
		//				{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
		//			},
		//		}, nil)
		//		service.ListStagedProductJobsReturns(map[string]string{
		//			"some-job":       "a-guid",
		//			"some-other-job": "a-different-guid",
		//			"bad":            "do-not-use",
		//		}, nil)
		//	})
		//
		//	AfterEach(func() {
		//		os.RemoveAll(configFile.Name())
		//	})
		//
		//	When("the config file contains variables", func() {
		//		Context("passed in a vars-file", func() {
		//			It("can interpolate variables into the configuration", func() {
		//				client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
		//
		//				configFile, err = ioutil.TempFile("", "")
		//				Expect(err).ToNot(HaveOccurred())
		//
		//				_, err = configFile.WriteString(productPropertiesWithVariableTemplate)
		//				Expect(err).ToNot(HaveOccurred())
		//
		//				varsFile, err := ioutil.TempFile("", "")
		//				Expect(err).ToNot(HaveOccurred())
		//
		//				_, err = varsFile.WriteString(`password: something-secure`)
		//				Expect(err).ToNot(HaveOccurred())
		//
		//				err = client.Execute([]string{
		//					"--config", configFile.Name(),
		//					"--vars-file", varsFile.Name(),
		//				})
		//				Expect(err).ToNot(HaveOccurred())
		//			})
		//		})
		//
		//		Context("given vars", func() {
		//			It("can interpolate variables into the configuration", func() {
		//				client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
		//
		//				configFile, err = ioutil.TempFile("", "")
		//				Expect(err).ToNot(HaveOccurred())
		//
		//				_, err = configFile.WriteString(productPropertiesWithVariableTemplate)
		//				Expect(err).ToNot(HaveOccurred())
		//
		//				err = client.Execute([]string{
		//					"--config", configFile.Name(),
		//					"--var", "password=something-secure",
		//				})
		//				Expect(err).ToNot(HaveOccurred())
		//			})
		//		})
		//
		//		Context("passed as environment variables", func() {
		//			It("can interpolate variables into the configuration", func() {
		//				client := commands.NewConfigureProduct(func() []string { return []string{"OM_VAR_password=something-secure"} }, service, "", logger)
		//
		//				configFile, err = ioutil.TempFile("", "")
		//				Expect(err).ToNot(HaveOccurred())
		//
		//				_, err = configFile.WriteString(productPropertiesWithVariableTemplate)
		//				Expect(err).ToNot(HaveOccurred())
		//
		//				err = client.Execute([]string{
		//					"--config", configFile.Name(),
		//					"--vars-env", "OM_VAR",
		//				})
		//				Expect(err).ToNot(HaveOccurred())
		//			})
		//
		//			It("supports the experimental feature of OM_VARS_ENV", func() {
		//				os.Setenv("OM_VARS_ENV", "OM_VAR")
		//				defer os.Unsetenv("OM_VARS_ENV")
		//
		//				client := commands.NewConfigureProduct(func() []string { return []string{"OM_VAR_password=something-secure"} }, service, "", logger)
		//
		//				configFile, err = ioutil.TempFile("", "")
		//				Expect(err).ToNot(HaveOccurred())
		//
		//				_, err = configFile.WriteString(productPropertiesWithVariableTemplate)
		//				Expect(err).ToNot(HaveOccurred())
		//
		//				err = client.Execute([]string{
		//					"--config", configFile.Name(),
		//				})
		//				Expect(err).ToNot(HaveOccurred())
		//			})
		//		})
		//
		//		It("returns an error if missing variables", func() {
		//			client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
		//
		//			configFile, err = ioutil.TempFile("", "")
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			_, err = configFile.WriteString(productPropertiesWithVariableTemplate)
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			err = client.Execute([]string{
		//				"--config", configFile.Name(),
		//			})
		//			Expect(err).To(MatchError(ContainSubstring("Expected to find variables")))
		//		})
		//	})
		//
		//	When("an ops-file is provided", func() {
		//		It("can interpolate ops-files into the configuration", func() {
		//			client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
		//
		//			configFile, err = ioutil.TempFile("", "")
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			_, err = configFile.WriteString(ymlProductProperties)
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			opsFile, err := ioutil.TempFile("", "")
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			_, err = opsFile.WriteString(productOpsFile)
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			err = client.Execute([]string{
		//				"--config", configFile.Name(),
		//				"--ops-file", opsFile.Name(),
		//			})
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			Expect(service.ListStagedProductsCallCount()).To(Equal(1))
		//			Expect(service.UpdateStagedProductPropertiesCallCount()).To(Equal(1))
		//			Expect(service.UpdateStagedProductPropertiesArgsForCall(0).GUID).To(Equal("some-product-guid"))
		//			Expect(service.UpdateStagedProductPropertiesArgsForCall(0).Properties).To(MatchJSON(productPropertiesWithOpsFileInterpolated))
		//		})
		//
		//		It("returns an error if the ops file is invalid", func() {
		//			client := commands.NewConfigureProduct(func() []string { return nil }, service, "", logger)
		//
		//			configFile, err = ioutil.TempFile("", "")
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			_, err = configFile.WriteString(ymlProductProperties)
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			opsFile, err := ioutil.TempFile("", "")
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			_, err = opsFile.WriteString(`%%%`)
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			err = client.Execute([]string{
		//				"-c", configFile.Name(),
		//				"-o", opsFile.Name(),
		//			})
		//			Expect(err).To(MatchError(ContainSubstring("could not find expected directive name")))
		//		})
		//	})
		//})


		Describe("Usage", func() {
			It("returns usage info", func() {
				usage := command.Usage()
				Expect(usage).To(Equal(jhanda.Usage{
					Description:      "This authenticated command configures settings available on the \"Settings\" page in the Ops Manager UI",
					ShortDescription: "configures values present on the Ops Manager settings page",
					Flags:            command.Options,
				}))
			})
		})
	})
})
