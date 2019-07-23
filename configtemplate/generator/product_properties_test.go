package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not create required-vars file"))
			})
		})

		DescribeTable("required properties without defaults", func(
			propertyType string,
			defaultValue interface{},
			options []generator.Option,
			optionTemplates []generator.OptionTemplate,
			blueprints []generator.PropertyBlueprint,
		) {
			metadata := &generator.Metadata{
				PropertyBlueprints: []generator.PropertyBlueprint{
					{
						Name:               "some_property",
						Optional:           false,
						Configurable:       "true",
						Type:               propertyType,
						Default:            defaultValue,
						Options:            options,
						OptionTemplates:    optionTemplates,
						PropertyBlueprints: blueprints,
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
		},
			Entry("adds a simple property", "string", nil, nil, nil, nil),
		)

		DescribeTable("required properties with defaults", func(
			propertyType string,
			defaultValue interface{},
			options []generator.Option,
			optionTemplates []generator.OptionTemplate,
		) {
			metadata := &generator.Metadata{
				PropertyBlueprints: []generator.PropertyBlueprint{
					{
						Name:            "some_property",
						Optional:        false,
						Configurable:    "true",
						Type:            propertyType,
						Default:         defaultValue,
						Options:         options,
						OptionTemplates: optionTemplates,
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
		},
			Entry("does not add simple properties", "string", "some-default", nil, nil),
			Entry("does not add multi-select properties",
				"multi_select_options",
				[]interface{}{"some-default"},
				[]generator.Option{{Name: "some-default"}},
				nil,
			),
		)

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

		It("adds required collection's properties when they don't have defaults", func() {
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
			Expect(requiredVars).To(HaveLen(1))
			Expect(requiredVars).To(HaveKey("some_property_0/collection_object"))
			Expect(requiredVars).To(HaveKeyWithValue("some_property_0/collection_object", ""))
		})

		DescribeTable("does not add collection properties", func(
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
			Expect(requiredVars).ToNot(HaveKey("some_property_0/collection_object"))
		},
			Entry("optional with no defaults", true, nil),
			Entry("optional with defaults", true, "some-default"),
			Entry("required with defaults", false, "some-default"),
		)

		When("there is a selector", func() {
			When("there is no default", func() {
				It("outputs the required selector property", func() {
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
			})

			When("there is a default", func() {
				It("outputs the required properties for that default", func() {
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
					Expect(requiredVars).To(HaveKey("some_selector/default_option/first_property"))
					Expect(requiredVars).ToNot(HaveKey("some_selector/default_option/second_property"))
					Expect(requiredVars).ToNot(HaveKey("some_selector/default_option/third_property"))
				})
			})

			When("an option has a multi-select ", func() {
				It("does not add the 'required' multi-select's properties", func() {
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
					Expect(requiredVars).ToNot(HaveKey("some_selector/default_option/multi_selector"))
					Expect(requiredVars).ToNot(HaveKey("some_selector/default_option/multi_selector/some_multi_select_property"))
				})
			})

			When("an option has a vm_type_dropdown property", func() {
				It("does not output the 'required' property", func() {
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
					Expect(requiredVars).ToNot(HaveKey("some_selector/some_dropdown"))
				})

				It("does not output the 'optional' property", func() {
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
					Expect(requiredVars).ToNot(HaveKey("some_selector/some_dropdown"))
				})
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
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not create feature ops files"))
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

			When("the selector has a default", func() {
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
			})

			When("there is an option with a multi-select property", func() {
				It("adds a replace statement in order to add each multi-select option for that property", func() {
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
			})

			When("there is an option with a property that is required and configurable", func() {
				It("adds a replace statement in order to add that property", func() {
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
							Value: "((some_selector/replace_option/some_property))",
						},
					}))
				})
			})
		})
	})
})
