package generator_test

import (
	"github.com/pivotal-cf/om/configtemplate/generator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CollectionPropertyMetadata", func() {
	Context("IsDefaultAnArray", func() {
		It("is true", func() {
			Expect(generator.IsDefaultAnArray(make([]interface{}, 0))).To(BeTrue())
		})
		It("is false", func() {
			Expect(generator.IsDefaultAnArray("foo")).To(BeFalse())
		})

		It("is false due to nil", func() {
			Expect(generator.IsDefaultAnArray(nil)).To(BeFalse())
		})
	})
	Context("DefaultsArrayToCollectionArray", func() {
		It("contains simple strings", func() {
			defaults := make([]interface{}, 2)
			defaults[0] = map[interface{}]interface{}{
				"simple": "simple-value",
			}
			defaults[1] = map[interface{}]interface{}{
				"other": "other-value",
			}
			collectionArray, err := generator.DefaultsArrayToCollectionArray("foo", defaults, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(collectionArray).ToNot(BeNil())
			Expect(collectionArray[0]["simple"]).Should(Equal(generator.SimpleString("simple-value")))
			Expect(collectionArray[1]["other"]).Should(Equal(generator.SimpleString("other-value")))
		})

		It("contains simple ints", func() {
			defaults := make([]interface{}, 2)
			defaults[0] = map[interface{}]interface{}{
				"simple": 0,
			}
			defaults[1] = map[interface{}]interface{}{
				"other": 1,
			}

			collectionArray, err := generator.DefaultsArrayToCollectionArray("foo", defaults, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(collectionArray).ToNot(BeNil())
			Expect(collectionArray[0]["simple"]).Should(Equal(generator.SimpleInteger(0)))
			Expect(collectionArray[1]["other"]).Should(Equal(generator.SimpleInteger(1)))
		})

		It("contains simple booleans", func() {
			defaults := make([]interface{}, 2)
			defaults[0] = map[interface{}]interface{}{
				"simple": true,
			}
			defaults[1] = map[interface{}]interface{}{
				"other": false,
			}
			collectionArray, err := generator.DefaultsArrayToCollectionArray("foo", defaults, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(collectionArray).ToNot(BeNil())
			Expect(len(collectionArray)).Should(Equal(2))

			Expect(collectionArray[0]["simple"]).Should(Equal(generator.SimpleBoolean(true)))
			Expect(collectionArray[1]["other"]).Should(Equal(generator.SimpleBoolean(false)))
		})

		It("default does not contain sub property key", func() {
			defaults := make([]interface{}, 1)
			defaults[0] = map[interface{}]interface{}{
				"simple": "simple-value",
			}

			collectionArray, err := generator.DefaultsArrayToCollectionArray("foo", defaults, []generator.PropertyBlueprint{
				{
					Name: "other-name",
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(collectionArray)).Should(Equal(1))
			Expect(collectionArray[0]["simple"]).Should(Equal(generator.SimpleString("simple-value")))
			Expect(collectionArray[0]["other-name"]).Should(Equal(generator.SimpleString("((foo_other-name))")))
		})

		It("contains unknown type", func() {
			defaults := make([]interface{}, 2)
			defaults[0] = map[interface{}]interface{}{
				"simple": int64(0),
			}
			collectionArray, err := generator.DefaultsArrayToCollectionArray("foo", defaults, nil)
			Expect(err).Should(MatchError("value int64 is not known"))
			Expect(len(collectionArray)).Should(Equal(0))
		})

		It("contains float32 types", func() {
			defaults := make([]interface{}, 1)
			defaults[0] = map[interface{}]interface{}{
				"simple": float32(0),
			}
			collectionArray, err := generator.DefaultsArrayToCollectionArray("foo", defaults, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(collectionArray)).Should(Equal(1))
			Expect(collectionArray[0]["simple"]).Should(Equal(generator.SimpleFloat(0.0)))
		})

		It("contains float64 types", func() {
			defaults := make([]interface{}, 1)
			defaults[0] = map[interface{}]interface{}{
				"simple": float64(0),
			}
			collectionArray, err := generator.DefaultsArrayToCollectionArray("foo", defaults, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(collectionArray)).Should(Equal(1))
			Expect(collectionArray[0]["simple"]).Should(Equal(generator.SimpleFloat(0.0)))
		})
	})

	Context("DefaultsToArray", func() {
		It("contains simple string", func() {

			propertyArray := generator.DefaultsToArray("foo", []generator.PropertyBlueprint{
				{
					Configurable: "true",
					Name:         "simple-value",
				},
			},
			)
			Expect(len(propertyArray)).Should(Equal(1))
			Expect(propertyArray["simple-value"]).To(Equal(generator.SimpleString("((foo_simple-value))")))
		})
		It("contains secret type", func() {
			propertyArray := generator.DefaultsToArray("foo", []generator.PropertyBlueprint{
				{
					Configurable: "true",
					Name:         "secret-value",
					Type:         "secret",
				},
			})
			Expect(len(propertyArray)).Should(Equal(1))
			Expect(propertyArray["secret-value"]).To(Equal(&generator.SecretValue{
				Value: "((foo_secret-value))",
			}))
		})

		It("contains credential type", func() {
			propertyArray := generator.DefaultsToArray("foo", []generator.PropertyBlueprint{
				{
					Configurable: "true",
					Name:         "certificate-value",
					Type:         "rsa_cert_credentials",
				},
			})
			Expect(len(propertyArray)).Should(Equal(1))
			Expect(propertyArray["certificate-value"]).To(Equal(generator.NewCertificateValue("foo")))
		})
	})
})
