package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/configtemplate/generator"
)

var _ = Describe("PropertyBlueprints", func() {
	Context("IsConfigurable", func() {
		It("is true", func() {
			propertyMetaData := &generator.PropertyBlueprint{
				Configurable: "true",
			}
			Expect(propertyMetaData.IsConfigurable()).To(BeTrue())
		})
		It("is false", func() {
			propertyMetaData := &generator.PropertyBlueprint{
				Configurable: "false",
			}
			Expect(propertyMetaData.IsConfigurable()).To(BeFalse())

			propertyMetaData = &generator.PropertyBlueprint{
				Configurable: "",
			}
			Expect(propertyMetaData.IsConfigurable()).To(BeFalse())
		})
	})

	Context("DefaultSelector with no matching option template", func() {
		It("selector path equals", func() {
			propertyMetaData := &generator.PropertyBlueprint{
				Default: "bar",
			}
			Expect(propertyMetaData.DefaultSelectorPath("foo")).To(Equal("foo.bar"))
		})
		It("default selector", func() {
			propertyMetaData := &generator.PropertyBlueprint{
				Default: "bar",
			}
			Expect(propertyMetaData.DefaultSelector()).To(Equal("bar"))
		})
	})
	Context("DefaultSelector with no matching option template", func() {
		var (
			propertyMetaData *generator.PropertyBlueprint
		)
		BeforeEach(func() {
			propertyMetaData = &generator.PropertyBlueprint{
				OptionTemplates: []generator.OptionTemplate{
					{
						Name:        "bar-name",
						SelectValue: "bar",
					},
					{
						Name:        "other-name",
						SelectValue: "other",
					},
				},
				Default: "bar",
			}
		})
		It("selector path equals", func() {
			Expect(propertyMetaData.DefaultSelectorPath("foo")).To(Equal("foo.bar-name"))
		})
		It("default selector", func() {
			Expect(propertyMetaData.DefaultSelector()).To(Equal("bar-name"))
		})
	})

	Context("IsRequired", func() {
		It("is true", func() {
			propertyMetaData := &generator.PropertyBlueprint{
				Optional: false,
			}
			Expect(propertyMetaData.IsRequired()).To(BeTrue())
		})
		It("is false", func() {
			propertyMetaData := &generator.PropertyBlueprint{
				Optional: true,
			}
			Expect(propertyMetaData.IsRequired()).To(BeFalse())
		})
	})

	Context("OptionTemplate", func() {
		var (
			propertyMetaData *generator.PropertyBlueprint
		)
		BeforeEach(func() {
			propertyMetaData = &generator.PropertyBlueprint{
				OptionTemplates: []generator.OptionTemplate{
					{
						Name:        "bar-name",
						SelectValue: "bar",
					},
					{
						Name:        "other-name",
						SelectValue: "other",
					},
				},
				Default: "bar",
			}
		})
		It("finds matching option template", func() {
			optionTemplate := propertyMetaData.OptionTemplate("bar-name")
			Expect(optionTemplate).ToNot(BeNil())
		})
		It("doesn't find matching option template", func() {
			optionTemplate := propertyMetaData.OptionTemplate("foo-name")
			Expect(optionTemplate).Should(BeNil())
		})
	})
})
