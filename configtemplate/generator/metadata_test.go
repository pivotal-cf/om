package generator_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/configtemplate/generator"
)

func getMetadata(filename string) *generator.Metadata {
	fileData, err := os.ReadFile(filename)
	Expect(err).ToNot(HaveOccurred())
	metadata, err := generator.NewMetadata(fileData)
	Expect(err).ToNot(HaveOccurred())
	return metadata
}

var _ = Describe("Metadata", func() {
	Context("UsesServiceNetwork", func() {
		It("Should use service network", func() {
			metadata := getMetadata("fixtures/metadata/p_healthwatch.yml")
			Expect(metadata.UsesServiceNetwork()).Should(BeTrue())
		})

		It("Should not service network", func() {
			metadata := getMetadata("fixtures/metadata/pas.yml")
			Expect(metadata.UsesServiceNetwork()).Should(BeFalse())
		})

		DescribeTable("nested property blueprint scenarios", func(rawMetadata string, expected bool) {
			metadata, err := generator.NewMetadata([]byte(rawMetadata))
			Expect(err).ToNot(HaveOccurred())
			Expect(metadata.UsesServiceNetwork()).Should(BeEquivalentTo(expected))
		},
			Entry("job property nested two levels deep under a selector's option_templates", `
job_types:
- name: some-job
  property_blueprints:
  - name: some_selector
    type: selector
    option_templates:
    - name: option_one
      property_blueprints:
      - name: az
        type: service_network_az_multi_select
`, true),

			Entry("job property nested one level deep under a plain sub-property", `
job_types:
- name: some-job
  property_blueprints:
  - name: parent_prop
    type: string
    property_blueprints:
    - name: az
      type: service_network_az_multi_select
`, true),

			Entry("top-level property nested three levels deep", `
property_blueprints:
- name: level1
  type: string
  property_blueprints:
  - name: level2
    type: string
    property_blueprints:
    - name: level3
      type: service_network_az_single_select
`, true),

			Entry("service_network_az_single_select nested under a job selector's option_templates", `
job_types:
- name: some-job
  property_blueprints:
  - name: some_selector
    type: selector
    option_templates:
    - name: option_one
      property_blueprints:
      - name: az
        type: service_network_az_single_select
`, true),

			Entry("AZ property nested under a non-selector property's option_templates", `
property_blueprints:
- name: not_a_selector
  type: boolean
  option_templates:
  - name: option_one
    property_blueprints:
    - name: az
      type: service_network_az_multi_select
`, true),

			Entry("nested properties present but none match the AZ types", `
job_types:
- name: some-job
  property_blueprints:
  - name: parent
    type: string
    property_blueprints:
    - name: child
      type: boolean
property_blueprints:
- name: top_parent
  type: string
  property_blueprints:
  - name: top_child
    type: integer
`, false),

			Entry("option_templates present with empty or non-matching nested property_blueprints", `
property_blueprints:
- name: some_selector
  type: selector
  option_templates:
  - name: option_one
    property_blueprints: []
  - name: option_two
    property_blueprints:
    - name: other
      type: string
`, false),

			Entry("empty metadata", `{}`, false),

			Entry("empty property_blueprints and option_templates slices", `
job_types:
- name: some-job
  property_blueprints: []
property_blueprints:
- name: top
  type: string
  property_blueprints: []
  option_templates: []
`, false),

			Entry("only the second job (not the first) has the nested match", `
job_types:
- name: job-one
  property_blueprints:
  - name: unrelated
    type: string
- name: job-two
  property_blueprints:
  - name: some_selector
    type: selector
    option_templates:
    - name: option_one
      property_blueprints:
      - name: az
        type: service_network_az_multi_select
`, true),

			Entry("only the second top-level property (not the first) matches", `
property_blueprints:
- name: prop-one
  type: string
- name: prop-two
  type: service_network_az_single_select
`, true),
		)
	})

	Context("GetPropertyBlueprint", func() {
		It("returns a non-job configurable property", func() {
			metadata := getMetadata("fixtures/metadata/p_healthwatch.yml")
			property, err := metadata.GetPropertyBlueprint(".properties.opsman")
			Expect(err).ToNot(HaveOccurred())
			Expect(property.Name).Should(Equal("opsman"))
		})

		It("returns a job configurable property", func() {
			metadata := getMetadata("fixtures/metadata/p_healthwatch.yml")
			property, err := metadata.GetPropertyBlueprint(".healthwatch-forwarder.foundation_name")
			Expect(err).ToNot(HaveOccurred())
			Expect(property).ToNot(BeNil())
			Expect(property.Name).Should(Equal("foundation_name"))
		})
	})

	DescribeTable("ProductName tile metadata fixture tests", func(fixtureFilepath string, expectedName string) {
		metadata := getMetadata(fixtureFilepath)
		Expect(metadata.ProductName()).Should(BeEquivalentTo(expectedName))
	},
		Entry("PAS", "fixtures/metadata/pas.yml", "cf"),
		Entry("healthwatch", "fixtures/metadata/p_healthwatch.yml", "p-healthwatch"),
		Entry("iso-segment", "fixtures/metadata/iso-segment.yml", "p-isolation-segment"),
		Entry("replicated iso-segment", "fixtures/metadata/iso-segment-replicator.yml", "p-isolation-segment-new-seg"),
	)

	DescribeTable("ProductVersion tile metadata fixture tests", func(fixtureFilepath string, expectedVersion string) {
		metadata := getMetadata(fixtureFilepath)
		Expect(metadata.ProductVersion()).Should(BeEquivalentTo(expectedVersion))
	},
		Entry("PAS", "fixtures/metadata/pas.yml", "2.1.3"),
		Entry("healthwatch", "fixtures/metadata/p_healthwatch.yml", "1.2.1-build.1"),
		Entry("iso-segment", "fixtures/metadata/iso-segment.yml", "2.2.4"),
		Entry("replicated iso-segment", "fixtures/metadata/iso-segment-replicator.yml", "2.2.4"),
	)
})
