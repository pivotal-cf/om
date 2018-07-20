package configparser_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/pivotal-cf/om/commands"
)

var _ = Describe("Abc", func() {

	It("writes a config file to stdout", func() {
		command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
		err := command.Execute([]string{
			"--product-name", "some-product",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeService.GetStagedProductByNameCallCount()).To(Equal(1))
		Expect(fakeService.GetStagedProductByNameArgsForCall(0)).To(Equal("some-product"))

		Expect(fakeService.GetStagedProductPropertiesCallCount()).To(Equal(1))
		Expect(fakeService.GetStagedProductPropertiesArgsForCall(0)).To(Equal("some-product-guid"))

		Expect(fakeService.GetStagedProductNetworksAndAZsCallCount()).To(Equal(1))
		Expect(fakeService.GetStagedProductNetworksAndAZsArgsForCall(0)).To(Equal("some-product-guid"))

		Expect(fakeService.ListStagedProductJobsCallCount()).To(Equal(1))
		Expect(fakeService.ListStagedProductJobsArgsForCall(0)).To(Equal("some-product-guid"))

		Expect(fakeService.GetStagedProductJobResourceConfigCallCount()).To(Equal(1))
		productGuid, jobsGuid := fakeService.GetStagedProductJobResourceConfigArgsForCall(0)
		Expect(productGuid).To(Equal("some-product-guid"))
		Expect(jobsGuid).To(Equal("some-job-guid"))

		Expect(logger.PrintlnCallCount()).To(Equal(1))
		output := logger.PrintlnArgsForCall(0)
		Expect(output).To(ContainElement(MatchYAML(`---
product-properties:
 .properties.collection:
   value:
   - name: Certificate
 .properties.some-string-property:
   value: some-value
 .properties.some-selector:
   value: internal
network-properties:
 singleton_availability_zone:
   name: az-one
resource-config:
 some-job:
   instances: 1
   instance_type:
     id: automatic
`)))
	})
})

It("replace *** with interpolatable placeholder and removes non-configurable properties", func() {
	command := commands.NewStagedConfig(fakeService, logger)
	err := command.Execute([]string{
		"--product-name", "some-product",
		"--include-placeholder",
	})
	Expect(err).NotTo(HaveOccurred())

	Expect(logger.PrintlnCallCount()).To(Equal(1))
	output := logger.PrintlnArgsForCall(0)
	Expect(output).To(ContainElement(MatchYAML(`---
product-properties:
 ".properties.some-string-property":
   value: some-value
 ".properties.some-secret-property":
   value:
     secret: "((properties_some-secret-property.secret))"
 ".properties.some-selector":
   value: internal
 ".properties.simple-credentials":
   value:
     identity: "((properties_simple-credentials.identity))"
     password: "((properties_simple-credentials.password))"
 ".properties.rsa-cert-credentials":
   value:
     cert_pem: "((properties_rsa-cert-credentials.cert_pem))"
     private_key_pem: "((properties_rsa-cert-credentials.private_key_pem))"
 ".properties.rsa-pkey-credentials":
   value:
     private_key_pem: "((properties_rsa-pkey-credentials.private_key_pem))"
 ".properties.salted-credentials":
   value:
     identity: "((properties_salted-credentials.identity))"
     password: "((properties_salted-credentials.password))"
     salt: "((properties_salted-credentials.salt))"
 ".properties.collection":
   value:
   - certificate:
       private_key_pem: "((properties_collection_0_certificate.private_key_pem))"
       cert_pem: "((properties_collection_0_certificate.cert_pem))"
     name: Certificate
   - certificate2:
       private_key_pem: "((properties_collection_1_certificate2.private_key_pem))"
       cert_pem: "((properties_collection_1_certificate2.cert_pem))"
network-properties:
 singleton_availability_zone:
   name: az-one
resource-config:
 some-job:
   instances: 1
   instance_type:
     id: automatic

`)))
})

It("includes secret values in the output", func() {
	command := commands.NewStagedConfig(fakeService, logger)
	err := command.Execute([]string{
		"--product-name", "some-product",
		"--include-credentials",
	})
	Expect(err).NotTo(HaveOccurred())

	Expect(fakeService.GetDeployedProductCredentialCallCount()).To(Equal(7))

	apiInputs := []api.GetDeployedProductCredentialInput{}
	for i := 0; i < 7; i++ {
		apiInputs = append(apiInputs, fakeService.GetDeployedProductCredentialArgsForCall(i))
	}
	Expect(apiInputs).To(ConsistOf([]api.GetDeployedProductCredentialInput{
		{
			DeployedGUID:        "some-product-guid",
			CredentialReference: ".properties.some-secret-property",
		},
		{
			DeployedGUID:        "some-product-guid",
			CredentialReference: ".properties.salted-credentials",
		},
		{
			DeployedGUID:        "some-product-guid",
			CredentialReference: ".properties.collection[0].certificate",
		},
		{
			DeployedGUID:        "some-product-guid",
			CredentialReference: ".properties.collection[1].certificate2",
		},
		{
			DeployedGUID:        "some-product-guid",
			CredentialReference: ".properties.simple-credentials",
		},
		{
			DeployedGUID:        "some-product-guid",
			CredentialReference: ".properties.rsa-cert-credentials",
		},
		{
			DeployedGUID:        "some-product-guid",
			CredentialReference: ".properties.rsa-pkey-credentials",
		},
	}))

	Expect(logger.PrintlnCallCount()).To(Equal(1))
	output := logger.PrintlnArgsForCall(0)
	Expect(output).To(ContainElement(MatchYAML(`product-properties:
 .properties.collection:
   value:
   - certificate:
       some-secret-key: some-secret-value
     name: Certificate
   - certificate2:
       some-secret-key: some-secret-value
 .properties.some-string-property:
   value: some-value
 .properties.some-secret-property:
   value:
     some-secret-key: some-secret-value
 .properties.simple-credentials:
   value:
     some-secret-key: some-secret-value
 .properties.rsa-cert-credentials:
   value:
     some-secret-key: some-secret-value
 .properties.rsa-pkey-credentials:
   value:
     some-secret-key: some-secret-value
 .properties.salted-credentials:
   value:
     some-secret-key: some-secret-value
 .properties.some-selector:
   value: internal
network-properties:
 singleton_availability_zone:
   name: az-one
resource-config:
 some-job:
   instances: 1
   instance_type:
     id: automatic
`)))
})