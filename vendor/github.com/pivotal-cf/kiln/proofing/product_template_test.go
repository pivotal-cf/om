package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProductTemplate", func() {
	var productTemplate proofing.ProductTemplate

	BeforeEach(func() {
		f, err := os.Open("fixtures/metadata.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err = proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())
	})

	It("parses a metadata file", func() {
		Expect(productTemplate.IconImage).To(Equal("some-icon-image"))
		Expect(productTemplate.Label).To(Equal("some-label"))
		Expect(productTemplate.MetadataVersion).To(Equal("some-metadata-version"))
		Expect(productTemplate.MinimumVersionForUpgrade).To(Equal("some-minimum-version-for-upgrade"))
		Expect(productTemplate.Name).To(Equal("some-name"))
		Expect(productTemplate.ProductVersion).To(Equal("some-product-version"))
		Expect(productTemplate.Rank).To(Equal(1))
		Expect(productTemplate.Serial).To(BeTrue())
		Expect(productTemplate.OriginalMetadataVersion).To(Equal("some-original-metadata-version"))
		Expect(productTemplate.ServiceBroker).To(BeTrue())
		Expect(productTemplate.DeprecatedTileImage).To(Equal("some-deprecated-tile-image"))
		Expect(productTemplate.BaseReleasesURL).To(Equal("some-base-releases-url"))
		Expect(productTemplate.Cloud).To(Equal("some-cloud"))
		Expect(productTemplate.Network).To(Equal("some-network"))

		Expect(productTemplate.FormTypes).To(HaveLen(1))
		Expect(productTemplate.InstallTimeVerifiers).To(HaveLen(1))
		Expect(productTemplate.JobTypes).To(HaveLen(1))
		Expect(productTemplate.PostDeployErrands).To(HaveLen(1))
		Expect(productTemplate.PreDeleteErrands).To(HaveLen(1))
		Expect(productTemplate.PropertyBlueprints).To(HaveLen(1))
		Expect(productTemplate.RequiresProductVersions).To(HaveLen(1))
		Expect(productTemplate.Releases).To(HaveLen(1))
		Expect(productTemplate.RuntimeConfigs).To(HaveLen(1))
		Expect(productTemplate.StemcellCriteria).To(BeAssignableToTypeOf(proofing.StemcellCriteria{}))
		Expect(productTemplate.Variables).To(HaveLen(1))
	})

	Describe("AllPropertyBlueprints", func() {
		BeforeEach(func() {
			f, err := os.Open("fixtures/property_blueprints.yml")
			defer f.Close()
			Expect(err).NotTo(HaveOccurred())

			productTemplate, err = proofing.Parse(f)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns all property blueprints as a list", func() {
			propertyBlueprints := productTemplate.AllPropertyBlueprints()

			Expect(propertyBlueprints).To(HaveLen(8))

			simpleBlueprint := propertyBlueprints[0]
			Expect(simpleBlueprint.Property).To(Equal(".properties.some-simple-name"))
			Expect(simpleBlueprint.Default).To(Equal("some-default"))
			Expect(simpleBlueprint.Configurable).To(BeTrue())

			selectorBlueprint := propertyBlueprints[1]
			Expect(selectorBlueprint.Property).To(Equal(".properties.some-selector-name"))
			Expect(selectorBlueprint.Configurable).To(BeTrue())

			selectorOptionBlueprint := propertyBlueprints[2]
			Expect(selectorOptionBlueprint.Property).To(Equal(
				".properties.some-selector-name.some-option-template-name.some-nested-simple-name"))
			Expect(selectorOptionBlueprint.Default).To(Equal(1))
			Expect(selectorOptionBlueprint.Configurable).To(BeTrue())

			collection := propertyBlueprints[3]
			Expect(collection.Property).To(Equal(".properties.some-collection-name"))
			Expect(collection.Configurable).To(BeTrue())

			instanceGroupBlueprint := propertyBlueprints[4]
			Expect(instanceGroupBlueprint.Property).To(Equal(".some-job-type-name.some-name"))
			Expect(instanceGroupBlueprint.Default).To(Equal("some-default"))
			Expect(instanceGroupBlueprint.Configurable).To(BeTrue())

			instanceGroupCollection := propertyBlueprints[5]
			Expect(instanceGroupCollection.Property).To(Equal(
				".some-job-type-name.some-nested-collection-name"))
			Expect(instanceGroupCollection.Configurable).To(BeTrue())

			instanceGroupSelector := propertyBlueprints[6]
			Expect(instanceGroupSelector.Property).To(Equal(
				".some-job-type-name.some-nested-selector-name"))
			Expect(instanceGroupSelector.Configurable).To(BeTrue())

			instanceGroupSelectorOptionBlueprint := propertyBlueprints[7]
			Expect(instanceGroupSelectorOptionBlueprint.Property).To(Equal(
				".some-job-type-name.some-nested-selector-name.some-option-template-name.some-nested-simple-name"))
			Expect(instanceGroupSelectorOptionBlueprint.Default).To(Equal(1))
			Expect(instanceGroupSelectorOptionBlueprint.Configurable).To(BeTrue())
		})
	})
})
