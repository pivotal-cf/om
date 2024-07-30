package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/configtemplate/generator"
	"github.com/pivotal-cf/om/configtemplate/generator/fakes"

	"gopkg.in/yaml.v2"
)

var withServiceNetwork = `network:
  name: ((network_name))
other_availability_zones:
- name: ((singleton_availability_zone))
service_network:
  name: ((service_network_name))
singleton_availability_zone:
  name: ((singleton_availability_zone))`

var withoutServiceNetwork = `network:
  name: ((network_name))
other_availability_zones:
- name: ((singleton_availability_zone))
singleton_availability_zone:
  name: ((singleton_availability_zone))`

var _ = Describe("NetworkProperties", func() {
	Context("CreateNetworkProperties", func() {
		var (
			metadata *fakes.FakeMetadata
		)
		BeforeEach(func() {
			metadata = new(fakes.FakeMetadata)
		})
		It("Should return network properties with network", func() {
			networkProps := generator.CreateNetworkProperties(metadata)
			Expect(networkProps).ToNot(BeNil())
			Expect(networkProps.Network).ToNot(BeNil())
			Expect(networkProps.Network.Name).Should(BeEquivalentTo("((network_name))"))
		})
		It("Should return network properties with service network", func() {
			metadata.UsesServiceNetworkReturns(true)
			networkProps := generator.CreateNetworkProperties(metadata)
			Expect(networkProps).ToNot(BeNil())
			Expect(networkProps.ServiceNetwork).ToNot(BeNil())
			Expect(networkProps.ServiceNetwork.Name).Should(BeEquivalentTo("((service_network_name))"))
		})

		It("Should return network properties without service network", func() {
			metadata.UsesServiceNetworkReturns(false)
			networkProps := generator.CreateNetworkProperties(metadata)
			Expect(networkProps).ToNot(BeNil())
			Expect(networkProps.ServiceNetwork).Should(BeNil())
		})

		It("Should return singleton availability zone", func() {
			networkProps := generator.CreateNetworkProperties(metadata)
			Expect(networkProps).ToNot(BeNil())
			Expect(networkProps.SingletonAvailabilityZone).ToNot(BeNil())
			Expect(networkProps.SingletonAvailabilityZone.Name).Should(BeEquivalentTo("((singleton_availability_zone))"))
		})

		It("Should return single az in other azs", func() {
			networkProps := generator.CreateNetworkProperties(metadata)
			Expect(networkProps).ToNot(BeNil())
			Expect(len(networkProps.OtherAvailabilityZones)).Should(Equal(1))
			Expect(networkProps.OtherAvailabilityZones[0].Name).Should(BeEquivalentTo("((singleton_availability_zone))"))
		})

		It("Should marshall to yaml with service network", func() {
			metadata.UsesServiceNetworkReturns(true)
			networkProps := generator.CreateNetworkProperties(metadata)
			Expect(networkProps).ToNot(BeNil())
			yml, err := yaml.Marshal(networkProps)
			Expect(err).ToNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(withServiceNetwork))
		})

		It("Should marshall to yaml without service network", func() {
			metadata.UsesServiceNetworkReturns(false)
			networkProps := generator.CreateNetworkProperties(metadata)
			Expect(networkProps).ToNot(BeNil())
			yml, err := yaml.Marshal(networkProps)
			Expect(err).ToNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(withoutServiceNetwork))
		})
	})
})
