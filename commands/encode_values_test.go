package commands_test

import (
	"github.com/google/go-querystring/query"
	"github.com/pivotal-cf/om/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("", func() {
	Describe("NetworksConfiguration", func() {
		Describe("EncodeValues", func() {
			It("turns the network configuration into urlencoded form values", func() {
				var isServiceNetwork = true

				n := commands.NetworksConfiguration{
					ICMP: true,
					Networks: []commands.NetworkConfiguration{
						{
							Name:           "foo",
							ServiceNetwork: &isServiceNetwork,
							Subnets: []commands.Subnet{
								{
									IAASIdentifier:        "something",
									CIDR:                  "some-cidr",
									ReservedIPRanges:      "reserved-ips",
									DNS:                   "some-dns",
									Gateway:               "some-gateway",
									AvailabilityZoneGUIDs: []string{"one", "two"},
								},
							},
						},
					},
				}

				values, err := query.Values(n)
				Expect(err).NotTo(HaveOccurred())

				Expect(values).To(HaveKeyWithValue("infrastructure[icmp_checks_enabled]", []string{"1"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][name]", []string{"foo"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][service_network]", []string{"1"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][subnets][0][iaas_identifier]", []string{"something"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][subnets][0][cidr]", []string{"some-cidr"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][subnets][0][reserved_ip_ranges]", []string{"reserved-ips"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][subnets][0][dns]", []string{"some-dns"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][subnets][0][gateway]", []string{"some-gateway"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][subnets][0][availability_zone_references][]", []string{"one", "two"}))
			})

			It("does submit service network when false", func() {
				var isServiceNetwork = false

				n := commands.NetworksConfiguration{
					ICMP: true,
					Networks: []commands.NetworkConfiguration{
						{
							Name:           "foo",
							ServiceNetwork: &isServiceNetwork,
							Subnets: []commands.Subnet{
								{
									IAASIdentifier:        "something",
									CIDR:                  "some-cidr",
									ReservedIPRanges:      "reserved-ips",
									DNS:                   "some-dns",
									Gateway:               "some-gateway",
									AvailabilityZoneGUIDs: []string{"one", "two"},
								},
							},
						},
					},
				}

				values, err := query.Values(n)
				Expect(err).NotTo(HaveOccurred())

				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][service_network]", []string{"0"}))
			})

			It("does not submit service network when nil", func() {
				n := commands.NetworksConfiguration{
					ICMP: true,
					Networks: []commands.NetworkConfiguration{
						{
							Name: "foo",
							Subnets: []commands.Subnet{
								{
									IAASIdentifier:        "something",
									CIDR:                  "some-cidr",
									ReservedIPRanges:      "reserved-ips",
									DNS:                   "some-dns",
									Gateway:               "some-gateway",
									AvailabilityZoneGUIDs: []string{"one", "two"},
								},
							},
						},
					},
				}

				values, err := query.Values(n)
				Expect(err).NotTo(HaveOccurred())

				Expect(values).NotTo(HaveKey("network_collection[networks_attributes][0][service_network]"))
			})
		})
	})
})
