package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/configtemplate/generator"
)

var _ = Describe("Product Properties", func() {
	Context("GetRequiredPropertyVars", func() {
		When("GetPropertyBlueprint returns an error", func() {
			It("returns an error", func() {
				metadata := &generator.Metadata{
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_multi_select_property",
								},
							},
						},
					},
				}

				_, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).To(MatchError(ContainSubstring("could not create required-vars file")))
			})
		})

		When("there are simple properties", func() {
			It("adds required properties without defaults", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Name:               "some_property",
							Optional:           false,
							Configurable:       "true",
							Type:               "string",
							Default:            nil,
							Options:            nil,
							OptionTemplates:    nil,
							PropertyBlueprints: nil,
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_property",
									PropertyInputs: []generator.PropertyInput{
										{
											Reference: "collection_object",
										},
									},
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(1))
				Expect(requiredVars).To(HaveKey("some_property"))
				Expect(requiredVars).To(HaveKeyWithValue("some_property", ""))
			})

			It("does not add required properties with defaults", func() {
				{
					metadata := &generator.Metadata{
						PropertyBlueprints: []generator.PropertyBlueprint{
							{
								Name:            "some_property",
								Optional:        false,
								Configurable:    "true",
								Type:            "string",
								Default:         "some-default",
								Options:         nil,
								OptionTemplates: nil,
							},
						},
						FormTypes: []generator.FormType{
							{
								PropertyInputs: []generator.PropertyInput{
									{
										Reference: ".properties.some_property",
									},
								},
							},
						},
					}
					requiredVars, err := generator.GetRequiredPropertyVars(metadata)
					Expect(err).ToNot(HaveOccurred())
					Expect(requiredVars).To(HaveLen(0))
					Expect(requiredVars).ToNot(HaveKey("some_property"))
				}
			})
		})

		When("there are multi-select properties", func() {
			It("does not add required properties with defaults", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Name:            "some_property",
							Optional:        false,
							Configurable:    "true",
							Type:            "multi_select",
							Default:         []interface{}{"some-default"},
							Options:         []generator.Option{{Name: "some-default"}},
							OptionTemplates: nil,
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_property",
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(0))
				Expect(requiredVars).ToNot(HaveKey("some_property"))
			})

			It("does not add multi-select properties without defaults", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Name:         "some_property",
							Optional:     false,
							Configurable: "true",
							Type:         "multi_select_options",
							Options: []generator.Option{
								{
									Name: "some-option",
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_property",
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(0))
				Expect(requiredVars).ToNot(HaveKey("some_property"))
			})
		})

		When("there is a collection property", func() {
			It("adds required properties when they don't have defaults", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Configurable: "true",
							Optional:     false,
							Name:         "some_property",
							Type:         "collection",
							PropertyBlueprints: []generator.PropertyBlueprint{
								{
									Name:         "collection_object",
									Optional:     false,
									Configurable: "true",
									Type:         "string",
								},
								{
									Name:         "another_object",
									Optional:     false,
									Configurable: "true",
									Type:         "string",
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_property",
									PropertyInputs: []generator.PropertyInput{
										{
											Reference: "collection_object",
										},
										{
											Reference: "another_object",
										},
									},
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(2))
				Expect(requiredVars).To(HaveKey("some_property_0_collection_object"))
				Expect(requiredVars).To(HaveKeyWithValue("some_property_0_collection_object", ""))
				Expect(requiredVars).To(HaveKey("some_property_0_another_object"))
				Expect(requiredVars).To(HaveKeyWithValue("some_property_0_another_object", ""))
			})

			DescribeTable("does not add properties", func(
				optional bool,
				defaultValues interface{},
			) {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Configurable: "true",
							Optional:     optional,
							Name:         "some_property",
							Type:         "collection",
							PropertyBlueprints: []generator.PropertyBlueprint{
								{
									Name:         "collection_object",
									Optional:     false,
									Configurable: "true",
									Type:         "string",
									Default:      defaultValues,
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_property",
									PropertyInputs: []generator.PropertyInput{
										{
											Reference: "collection_object",
										},
									},
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(0))
				Expect(requiredVars).ToNot(HaveKey("some_property_0_collection_object"))
			},
				Entry("when optional with no defaults", true, nil),
				Entry("when optional with defaults", true, "some-default"),
				Entry("when required with defaults", false, "some-default"),
			)

			It("adds secret properties", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Configurable: "true",
							Optional:     false,
							Name:         "some_property",
							Type:         "collection",
							PropertyBlueprints: []generator.PropertyBlueprint{
								{
									Name:         "secret_object_1",
									Optional:     false,
									Configurable: "true",
									Type:         "secret",
								},
								{
									Name:         "secret_object_2",
									Optional:     false,
									Configurable: "true",
									Type:         "simple_credentials",
								},
								{
									Name:         "secret_object_3",
									Optional:     false,
									Configurable: "true",
									Type:         "rsa_cert_credentials",
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_property",
									PropertyInputs: []generator.PropertyInput{
										{
											Reference: "secret_object_1",
										},
										{
											Reference: "secret_object_2",
										},
										{
											Reference: "secret_object_3",
										},
									},
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(4))
				Expect(requiredVars).To(HaveKey("some_property_0_secret_object_1"))
				Expect(requiredVars).To(HaveKeyWithValue("some_property_0_secret_object_1", ""))
				Expect(requiredVars).To(HaveKey("some_property_0_secret_object_2"))
				Expect(requiredVars).To(HaveKeyWithValue("some_property_0_secret_object_2", ""))
				Expect(requiredVars).To(HaveKey("some_property_0_certificate"))
				Expect(requiredVars).To(HaveKeyWithValue("some_property_0_certificate", ""))
				Expect(requiredVars).To(HaveKey("some_property_0_privatekey"))
				Expect(requiredVars).To(HaveKeyWithValue("some_property_0_privatekey", ""))
			})
		})

		When("there is a dropdown property", func() {
			DescribeTable("does not output the 'required' property", func(
				propertyType string,
				optional bool,
			) {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Type:         propertyType,
							Name:         "some_dropdown",
							Optional:     optional,
							Configurable: "true",
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_dropdown",
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(0))
				Expect(requiredVars).ToNot(HaveKey("some_dropdown"))
			},
				Entry("optional vm_type_dropdown", "vm_type_dropdown", true),
				Entry("required vm_type_dropdown", "vm_type_dropdown", false),
				Entry("optional disk_type_dropdown", "disk_type_dropdown", true),
				Entry("required disk_type_dropdown", "disk_type_dropdown", false),
			)
		})

		When("there is a selector", func() {
			It("outputs the required selector property when there is no default", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Type:         "selector",
							Name:         "some_selector",
							Optional:     false,
							Configurable: "true",
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_selector",
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(1))
				Expect(requiredVars).To(HaveKey("some_selector"))
				Expect(requiredVars).To(HaveKeyWithValue("some_selector", ""))
			})

			It("outputs the required properties when there is a default", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Type:         "selector",
							Name:         "some_selector",
							Default:      "DEFAULT",
							Configurable: "true",
							Optional:     false,
							OptionTemplates: []generator.OptionTemplate{
								{
									Name:        "default_option",
									SelectValue: "DEFAULT",
									PropertyBlueprints: []generator.PropertyBlueprint{
										{
											Configurable: "true",
											Optional:     false,
											Name:         "first_property",
											Type:         "string",
										},
										{
											Configurable: "true",
											Optional:     false,
											Default:      "some-default-value",
											Name:         "second_property",
											Type:         "string",
										},
										{
											Configurable: "true",
											Optional:     true,
											Name:         "third_property",
											Type:         "string",
										},
									},
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_selector",
									SelectorPropertyInputs: []generator.SelectorPropertyInput{
										{
											Reference: ".properties.some_selector.default_option",
											PropertyInputs: []generator.PropertyInput{
												{
													Reference: ".properties.some_selector.default_option.first_property",
												},
												{
													Reference: ".properties.some_selector.default_option.second_property",
												},
												{
													Reference: ".properties.some_selector.default_option.third_property",
												},
											},
										},
									},
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(1))
				Expect(requiredVars).To(HaveKey("some_selector_default_option_first_property"))
				Expect(requiredVars).ToNot(HaveKey("some_selector_default_option_second_property"))
				Expect(requiredVars).ToNot(HaveKey("some_selector_default_option_third_property"))
			})

			It("does not add nested 'required' multi-select's properties", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Type:         "selector",
							Name:         "some_selector",
							Default:      "DEFAULT",
							Configurable: "true",
							Optional:     false,
							OptionTemplates: []generator.OptionTemplate{
								{
									Name:        "default_option",
									SelectValue: "DEFAULT",
									PropertyBlueprints: []generator.PropertyBlueprint{
										{
											Configurable: "true",
											Optional:     false,
											Name:         "multi_selector",
											Type:         "multi_select_options",
											Options: []generator.Option{
												{
													Name: "some_multi_select_property",
												},
											},
										},
									},
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_selector",
									SelectorPropertyInputs: []generator.SelectorPropertyInput{
										{
											Reference: ".properties.some_selector.default_option",
											PropertyInputs: []generator.PropertyInput{
												{
													Reference: ".properties.some_selector.default_option.some_multi_select_property",
												},
											},
										},
									},
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(0))
				Expect(requiredVars).ToNot(HaveKey("some_selector_default_option_multi_selector"))
				Expect(requiredVars).ToNot(HaveKey("some_selector_default_option_multi_selector_some_multi_select_property"))
			})

			It("does not output the nested 'required' vm_type_dropdown property", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Type:         "selector",
							Name:         "some_selector",
							Default:      "DEFAULT",
							Configurable: "true",
							Optional:     false,
							OptionTemplates: []generator.OptionTemplate{
								{
									Name:        "default_option",
									SelectValue: "DEFAULT",
									PropertyBlueprints: []generator.PropertyBlueprint{
										{
											Configurable: "true",
											Optional:     false,
											Name:         "some_dropdown",
											Type:         "vm_type_dropdown",
										},
									},
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_selector",
									SelectorPropertyInputs: []generator.SelectorPropertyInput{
										{
											Reference: ".properties.some_selector.some_dropdown",
										},
									},
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(0))
				Expect(requiredVars).ToNot(HaveKey("some_selector_some_dropdown"))
			})

			It("does not output the nested 'optional' vm_type_dropdown property", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Type:         "selector",
							Name:         "some_selector",
							Default:      "DEFAULT",
							Configurable: "true",
							Optional:     false,
							OptionTemplates: []generator.OptionTemplate{
								{
									Name:        "default_option",
									SelectValue: "DEFAULT",
									PropertyBlueprints: []generator.PropertyBlueprint{
										{
											Configurable: "true",
											Optional:     true,
											Name:         "some_dropdown",
											Type:         "vm_type_dropdown",
										},
									},
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_selector",
									SelectorPropertyInputs: []generator.SelectorPropertyInput{
										{
											Reference: ".properties.some_selector.some_dropdown",
										},
									},
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(0))
				Expect(requiredVars).ToNot(HaveKey("some_selector_some_dropdown"))
			})
		})

		When("there is a secret property", func() {
			It("adds secret properties", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Configurable: "true",
							Optional:     false,
							Name:         "secret_1",
							Type:         "secret",
						},
						{
							Configurable: "true",
							Optional:     false,
							Name:         "secret_2",
							Type:         "simple_credentials",
						},
						{
							Configurable: "true",
							Optional:     false,
							Name:         "secret_3",
							Type:         "rsa_cert_credentials",
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.secret_1",
								},
								{
									Reference: ".properties.secret_2",
								},
								{
									Reference: ".properties.secret_3",
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(4))
				Expect(requiredVars).To(HaveKey("secret_1"))
				Expect(requiredVars).To(HaveKeyWithValue("secret_1", ""))
				Expect(requiredVars).To(HaveKey("secret_2"))
				Expect(requiredVars).To(HaveKeyWithValue("secret_2", ""))
				Expect(requiredVars).To(HaveKey("secret_3_certificate"))
				Expect(requiredVars).To(HaveKeyWithValue("secret_3_certificate", ""))
				Expect(requiredVars).To(HaveKey("secret_3_privatekey"))
				Expect(requiredVars).To(HaveKeyWithValue("secret_3_privatekey", ""))
			})
		})

		When("network configuration exists", func() {
			It("adds network properties", func() {
				metadata := &generator.Metadata{
					JobTypes: []generator.JobType{{
						Name: "",
					}},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)

				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(2))
				Expect(requiredVars).To(HaveKey("network_name"))
				Expect(requiredVars).To(HaveKeyWithValue("network_name", ""))
				Expect(requiredVars).To(HaveKey("singleton_availability_zone"))
				Expect(requiredVars).To(HaveKeyWithValue("singleton_availability_zone", ""))
			})

			When("the service network is in use", func() {
				It("adds the service_network_name property", func() {
					metadata := &generator.Metadata{
						JobTypes: []generator.JobType{{
							PropertyBlueprint: []generator.PropertyBlueprint{{
								Type: "service_network_az_single_select",
							}},
						}},
					}
					requiredVars, err := generator.GetRequiredPropertyVars(metadata)
					Expect(err).ToNot(HaveOccurred())
					Expect(requiredVars).To(HaveLen(3))
					Expect(requiredVars).To(HaveKey("network_name"))
					Expect(requiredVars).To(HaveKeyWithValue("network_name", ""))
					Expect(requiredVars).To(HaveKey("singleton_availability_zone"))
					Expect(requiredVars).To(HaveKeyWithValue("singleton_availability_zone", ""))
					Expect(requiredVars).To(HaveKey("service_network_name"))
					Expect(requiredVars).To(HaveKeyWithValue("service_network_name", ""))
				})
			})
		})

		When("there is a bool property without a default", func() {
			It("does not add it", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Configurable: "true",
							Optional:     false,
							Name:         "bool_property",
							Type:         "boolean",
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.bool_property",
								},
							},
						},
					},
				}
				requiredVars, err := generator.GetRequiredPropertyVars(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(requiredVars).To(HaveLen(0))
				Expect(requiredVars).ToNot(HaveKey("bool_property"))
			})
		})
	})

	Context("CreateProductPropertiesFeaturesOpsFiles", func() {
		When("GetPropertyBlueprint returns an error", func() {
			It("returns an error", func() {
				metadata := &generator.Metadata{
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_multi_select_property",
								},
							},
						},
					},
				}

				_, err := generator.CreateProductPropertiesFeaturesOpsFiles(metadata)
				Expect(err).To(MatchError(ContainSubstring("could not create feature ops files")))
			})
		})

		When("there is a property that is a multi-select", func() {
			It("adds a replace statement in order to add each multi-select option for that property", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Configurable: "true",
							Optional:     true,
							Default:      []interface{}{"first_option", "third_option"},
							Name:         "some_multi_select_property",
							Type:         "multi_select_options",
							Options: []generator.Option{
								{
									Name: "first_option",
								},
								{
									Name: "second_option",
								},
								{
									Name: "third_option",
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_multi_select_property",
								},
							},
						},
					},
				}
				opsfiles, err := generator.CreateProductPropertiesFeaturesOpsFiles(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(opsfiles["some_multi_select_property_first_option"]).To(ContainElement(generator.Ops{
					Type:  "replace",
					Path:  "/product-properties/.properties.some_multi_select_property?/value/-",
					Value: generator.StringOpsValue("first_option"),
				}))
				Expect(opsfiles["some_multi_select_property_second_option"]).To(ContainElement(generator.Ops{
					Type:  "replace",
					Path:  "/product-properties/.properties.some_multi_select_property?/value/-",
					Value: generator.StringOpsValue("second_option"),
				}))
				Expect(opsfiles["some_multi_select_property_third_option"]).To(ContainElement(generator.Ops{
					Type:  "replace",
					Path:  "/product-properties/.properties.some_multi_select_property?/value/-",
					Value: generator.StringOpsValue("third_option"),
				}))
			})
		})

		When("there is a property that is a selector", func() {
			It("returns the value and selected value", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Type: "selector",
							Name: "some_selector",
							OptionTemplates: []generator.OptionTemplate{
								{
									Name:        "gcp",
									SelectValue: "GCP",
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_selector",
									SelectorPropertyInputs: []generator.SelectorPropertyInput{
										{
											Reference: ".properties.some_selector.gcp",
										},
									},
								},
							},
						},
					},
				}
				opsfiles, err := generator.CreateProductPropertiesFeaturesOpsFiles(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(opsfiles["some_selector-gcp"]).To(ContainElement(generator.Ops{
					Type: "replace",
					Path: "/product-properties/.properties.some_selector?",
					Value: &generator.OpsValue{
						Value:          "GCP",
						SelectedOption: "gcp",
					},
				}))
			})

			It("adds a remove statement to remove each property associated with the default selector", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Type:    "selector",
							Name:    "some_selector",
							Default: "DEFAULT",
							OptionTemplates: []generator.OptionTemplate{
								{
									Name:        "default_option",
									SelectValue: "DEFAULT",
									PropertyBlueprints: []generator.PropertyBlueprint{
										{
											Configurable: "true",
											Optional:     true,
											Name:         "some_property",
											Type:         "string",
										},
									},
								},
								{
									Name:        "replace_option",
									SelectValue: "REPLACE",
									PropertyBlueprints: []generator.PropertyBlueprint{
										{
											Configurable: "true",
											Optional:     true,
											Name:         "some_property",
											Type:         "string",
										},
									},
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_selector",
									SelectorPropertyInputs: []generator.SelectorPropertyInput{
										{
											Reference: ".properties.some_selector.default_option",
											PropertyInputs: []generator.PropertyInput{
												{
													Reference: ".properties.some_selector.default_option.some_property",
												},
											},
										},
										{
											Reference: ".properties.some_selector.replace_option",
											PropertyInputs: []generator.PropertyInput{
												{
													Reference: ".properties.some_selector.replace_option.some_property",
												},
											},
										},
									},
								},
							},
						},
					},
				}
				opsfiles, err := generator.CreateProductPropertiesFeaturesOpsFiles(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(opsfiles["some_selector-replace_option"]).To(ContainElement(generator.Ops{
					Type: "remove",
					Path: "/product-properties/.properties.some_selector.default_option.some_property?",
				}))
			})

			It("adds a replace statement for each multi-select option for the non-default selector property", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Type:    "selector",
							Name:    "some_selector",
							Default: "DEFAULT",
							OptionTemplates: []generator.OptionTemplate{
								{
									Name:        "default_option",
									SelectValue: "DEFAULT",
									PropertyBlueprints: []generator.PropertyBlueprint{
										{
											Configurable: "true",
											Optional:     true,
											Name:         "some_property",
											Type:         "string",
										},
									},
								},
								{
									Name:        "replace_option",
									SelectValue: "REPLACE",
									PropertyBlueprints: []generator.PropertyBlueprint{
										{
											Configurable: "true",
											Optional:     true,
											Default:      []interface{}{"first_option", "third_option"},
											Name:         "some_multi_select_property",
											Type:         "multi_select_options",
											Options: []generator.Option{
												{
													Name: "first_option",
												},
												{
													Name: "second_option",
												},
												{
													Name: "third_option",
												},
											},
										},
									},
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_selector",
									SelectorPropertyInputs: []generator.SelectorPropertyInput{
										{
											Reference: ".properties.some_selector.default_option",
											PropertyInputs: []generator.PropertyInput{
												{
													Reference: ".properties.some_selector.default_option.some_multi_select_property",
												},
											},
										},
										{
											Reference: ".properties.some_selector.replace_option",
											PropertyInputs: []generator.PropertyInput{
												{
													Reference: ".properties.some_selector.replace_option.some_property",
												},
											},
										},
									},
								},
							},
						},
					},
				}
				opsfiles, err := generator.CreateProductPropertiesFeaturesOpsFiles(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(opsfiles["some_selector-replace_option-some_multi_select_property_first_option"]).To(ContainElement(generator.Ops{
					Type:  "replace",
					Path:  "/product-properties/.properties.some_selector.replace_option.some_multi_select_property?/value/-",
					Value: generator.StringOpsValue("first_option"),
				}))
				Expect(opsfiles["some_selector-replace_option-some_multi_select_property_second_option"]).To(ContainElement(generator.Ops{
					Type:  "replace",
					Path:  "/product-properties/.properties.some_selector.replace_option.some_multi_select_property?/value/-",
					Value: generator.StringOpsValue("second_option"),
				}))
				Expect(opsfiles["some_selector-replace_option-some_multi_select_property_third_option"]).To(ContainElement(generator.Ops{
					Type:  "replace",
					Path:  "/product-properties/.properties.some_selector.replace_option.some_multi_select_property?/value/-",
					Value: generator.StringOpsValue("third_option"),
				}))
			})

			It("adds a replace statement for each required and configurable non-default selector property", func() {
				metadata := &generator.Metadata{
					PropertyBlueprints: []generator.PropertyBlueprint{
						{
							Type:    "selector",
							Name:    "some_selector",
							Default: "DEFAULT",
							OptionTemplates: []generator.OptionTemplate{
								{
									Name:        "default_option",
									SelectValue: "DEFAULT",
									PropertyBlueprints: []generator.PropertyBlueprint{
										{
											Configurable: "true",
											Optional:     true,
											Name:         "some_property",
											Type:         "string",
										},
									},
								},
								{
									Name:        "replace_option",
									SelectValue: "REPLACE",
									PropertyBlueprints: []generator.PropertyBlueprint{
										{
											Configurable: "true",
											Optional:     false,
											Name:         "some_property",
											Type:         "string",
										},
									},
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							PropertyInputs: []generator.PropertyInput{
								{
									Reference: ".properties.some_selector",
									SelectorPropertyInputs: []generator.SelectorPropertyInput{
										{
											Reference: ".properties.some_selector.default_option",
											PropertyInputs: []generator.PropertyInput{
												{
													Reference: ".properties.some_selector.default_option.some_property",
												},
											},
										},
										{
											Reference: ".properties.some_selector.replace_option",
											PropertyInputs: []generator.PropertyInput{
												{
													Reference: ".properties.some_selector.replace_option.some_property",
												},
											},
										},
									},
								},
							},
						},
					},
				}
				opsfiles, err := generator.CreateProductPropertiesFeaturesOpsFiles(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(opsfiles["some_selector-replace_option"]).To(ContainElement(generator.Ops{
					Type: "replace",
					Path: "/product-properties/.properties.some_selector.replace_option.some_property?",
					Value: &generator.SimpleValue{
						Value: "((some_selector_replace_option_some_property))",
					},
				}))
			})
		})
	})
})
