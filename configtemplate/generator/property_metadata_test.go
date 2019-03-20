package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/configtemplate/generator"
)

var _ = Describe("PropertyMetadata", func() {
	Context("IsConfigurable", func() {
		It("is true", func() {
			propertyMetaData := &generator.PropertyMetadata{}
			Expect(propertyMetaData.IsConfigurable()).To(BeTrue())
		})
		It("is false", func() {
			propertyMetaData := &generator.PropertyMetadata{
				Configurable: "false",
			}
			Expect(propertyMetaData.IsConfigurable()).To(BeFalse())
		})
	})

	Context("IsExplicityConfigurable", func() {
		It("is true", func() {
			propertyMetaData := &generator.PropertyMetadata{
				Configurable: "true",
			}
			Expect(propertyMetaData.IsExplicityConfigurable()).To(BeTrue())
		})
		It("is false", func() {
			propertyMetaData := &generator.PropertyMetadata{
				Configurable: "false",
			}
			Expect(propertyMetaData.IsExplicityConfigurable()).To(BeFalse())
		})
	})

	Context("DefaultSelector with no matching option template", func() {
		It("selector path equals", func() {
			propertyMetaData := &generator.PropertyMetadata{
				Default: "bar",
			}
			Expect(propertyMetaData.DefaultSelectorPath("foo")).To(Equal("foo.bar"))
		})
		It("default selector", func() {
			propertyMetaData := &generator.PropertyMetadata{
				Default: "bar",
			}
			Expect(propertyMetaData.DefaultSelector()).To(Equal("bar"))
		})
	})
	Context("DefaultSelector with no matching option template", func() {
		var (
			propertyMetaData *generator.PropertyMetadata
		)
		BeforeEach(func() {
			propertyMetaData = &generator.PropertyMetadata{
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
			propertyMetaData := &generator.PropertyMetadata{
				Optional: false,
			}
			Expect(propertyMetaData.IsRequired()).To(BeTrue())
		})
		It("is false", func() {
			propertyMetaData := &generator.PropertyMetadata{
				Optional: true,
			}
			Expect(propertyMetaData.IsRequired()).To(BeFalse())
		})
	})

	Context("OptionTemplate", func() {
		var (
			propertyMetaData *generator.PropertyMetadata
		)
		BeforeEach(func() {
			propertyMetaData = &generator.PropertyMetadata{
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
			optionTemplate, err := propertyMetaData.OptionTemplate("bar-name")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(optionTemplate).ShouldNot(BeNil())
		})
		It("doesn't find matching option template", func() {
			optionTemplate, err := propertyMetaData.OptionTemplate("foo-name")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(optionTemplate).Should(BeNil())
		})
	})
})
