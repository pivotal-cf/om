package builder_test

import (
	"github.com/pivotal-cf/kiln/builder"
	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("interpolator", func() {
	const templateYAML = `
name: $( variable "some-variable" )
icon_img: $( icon )
some_releases:
- $(release "some-release")
stemcell_criteria: $( stemcell )
some_form_types:
- $( form "some-form" )
some_job_types:
- $( instance_group "some-instance-group" )
version: $( version )
some_property_blueprints:
- $( property "some-templated-property" )
- $( property "some-other-templated-property" )
some_runtime_configs:
- $( runtime_config "some-runtime-config" )
some_bosh_variables:
- $( bosh_variable "some-bosh-variable" )

selected_value: $( release "some-release" | select "version" )
`

	var (
		input        builder.InterpolateInput
		interpolator builder.Interpolator
	)

	BeforeEach(func() {
		interpolator = builder.NewInterpolator()

		input = builder.InterpolateInput{
			Version: "3.4.5",
			BOSHVariables: map[string]interface{}{
				"some-bosh-variable": builder.Metadata{
					"name": "some-bosh-variable",
					"type": "some-bosh-type",
				},
			},
			Variables: map[string]interface{}{
				"some-variable": "some-value",
			},
			ReleaseManifests: map[string]interface{}{
				"some-release": builder.ReleaseManifest{
					Name:    "some-release",
					Version: "1.2.3",
					File:    "some-release-1.2.3.tgz",
					SHA1:    "123abc",
				},
			},
			StemcellManifest: builder.StemcellManifest{
				Version:         "2.3.4",
				OperatingSystem: "an-operating-system",
			},
			FormTypes: map[string]interface{}{
				"some-form": builder.Metadata{
					"name":  "some-form",
					"label": "some-form-label",
				},
			},
			IconImage: "some-icon-image",
			InstanceGroups: map[string]interface{}{
				"some-instance-group": builder.Metadata{
					"name": "some-instance-group",
					"templates": []string{
						"$( job \"some-job\" )",
					},
				},
			},
			Jobs: map[string]interface{}{
				"some-job": builder.Metadata{
					"name":    "some-job",
					"release": "some-release",
				},
			},
			PropertyBlueprints: map[string]interface{}{
				"some-templated-property": builder.Metadata{
					"name":         "some-templated-property",
					"type":         "boolean",
					"configurable": true,
					"default":      false,
				},
				"some-other-templated-property": builder.Metadata{
					"name":         "some-other-templated-property",
					"type":         "string",
					"configurable": false,
					"default":      "some-value",
				},
			},
			RuntimeConfigs: map[string]interface{}{
				"some-runtime-config": builder.Metadata{
					"name":           "some-runtime-config",
					"runtime_config": "some-addon-runtime-config\n",
				},
			},
		}
	})

	It("interpolates metadata templates", func() {
		interpolatedYAML, err := interpolator.Interpolate(input, []byte(templateYAML))
		Expect(err).NotTo(HaveOccurred())
		Expect(interpolatedYAML).To(HelpfullyMatchYAML(`
name: some-value
icon_img: some-icon-image
some_releases:
- name: some-release
  file: some-release-1.2.3.tgz
  version: 1.2.3
  sha1: 123abc
stemcell_criteria:
  version: 2.3.4
  os: an-operating-system
some_form_types:
- name: some-form
  label: some-form-label
some_job_types:
- name: some-instance-group
  templates:
  - name: some-job
    release: some-release
version: 3.4.5
some_property_blueprints:
- name: some-templated-property
  type: boolean
  configurable: true
  default: false
- name: some-other-templated-property
  type: string
  configurable: false
  default: some-value
some_runtime_configs:
- name: some-runtime-config
  runtime_config: |
    some-addon-runtime-config
some_bosh_variables:
- name: some-bosh-variable
  type: some-bosh-type

selected_value: 1.2.3	
`))
		Expect(string(interpolatedYAML)).To(ContainSubstring("file: some-release-1.2.3.tgz\n"))
	})

	It("allows interpolation helpers inside forms", func() {
		templateYAML := `
---
some_form_types:
- $( form "some-form" )`

		input := builder.InterpolateInput{
			Variables: map[string]interface{}{
				"some-form-variable": "variable-form-label",
			},
			FormTypes: map[string]interface{}{
				"some-form": builder.Metadata{
					"name":  "some-form",
					"label": `$( variable "some-form-variable" )`,
				},
			},
		}

		interpolatedYAML, err := interpolator.Interpolate(input, []byte(templateYAML))
		Expect(err).NotTo(HaveOccurred())
		Expect(interpolatedYAML).To(HelpfullyMatchYAML(`
some_form_types:
- name: some-form
  label: variable-form-label
`))
	})

	Context("when the runtime config is provided", func() {

		var templateYAML string
		var input builder.InterpolateInput

		BeforeEach(func() {
			templateYAML = `
---
some_runtime_configs:
- $( runtime_config "some-runtime-config" )`

			input = builder.InterpolateInput{
				ReleaseManifests: map[string]interface{}{
					"some-release": builder.ReleaseManifest{
						Name:    "some-release",
						Version: "1.2.3",
						File:    "some-release-1.2.3.tgz",
						SHA1:    "123abc",
					},
				},
				RuntimeConfigs: map[string]interface{}{
					"some-runtime-config": builder.Metadata{
						"name": "some-runtime-config",
						"runtime_config": `releases:
- $( release "some-release" )`,
					},
				},
			}
		})

		It("allows interpolation helpers inside runtime_configs", func() {
			interpolatedYAML, err := interpolator.Interpolate(input, []byte(templateYAML))
			Expect(err).NotTo(HaveOccurred())

			var output map[string]interface{}
			err = yaml.Unmarshal(interpolatedYAML, &output)
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(HaveKey("some_runtime_configs"))
			configs, ok := output["some_runtime_configs"].([]interface{})
			Expect(ok).To(BeTrue())
			config, ok := configs[0].(map[interface{}]interface{})
			Expect(ok).To(BeTrue())

			Expect(config).To(HaveKeyWithValue("name", "some-runtime-config"))
			Expect(config["runtime_config"]).To(HelpfullyMatchYAML(`
releases:
- file: some-release-1.2.3.tgz
  name: some-release
  sha1: 123abc
  version: 1.2.3`))
		})

		Context("when the interpolated runtime config does not have a runtime_config key", func() {
			JustBeforeEach(func() {
				input.RuntimeConfigs = map[string]interface{}{
					"some-runtime-config": builder.Metadata{
						"name": "some-runtime-config",
					},
				}
			})
			It("does not error", func() {
				interpolatedYAML, err := interpolator.Interpolate(input, []byte(templateYAML))
				Expect(err).NotTo(HaveOccurred())
				Expect(interpolatedYAML).To(HelpfullyMatchYAML(`
some_runtime_configs:
- name: some-runtime-config
`))
			})
		})
	})

	Context("failure cases", func() {
		Context("when the requested form name is not found", func() {
			It("returns an error", func() {
				input.FormTypes = map[string]interface{}{}
				interpolator := builder.NewInterpolator()
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find form with key 'some-form'"))
			})
		})

		Context("when the requested property blueprint is not found", func() {
			It("returns an error", func() {
				input.PropertyBlueprints = map[string]interface{}{}
				interpolator := builder.NewInterpolator()
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find property blueprint with name 'some-templated-property'"))
			})
		})

		Context("when the requested runtime config is not found", func() {
			It("returns an error", func() {
				input.RuntimeConfigs = map[string]interface{}{}
				interpolator := builder.NewInterpolator()
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find runtime_config with name 'some-runtime-config'"))
			})
		})

		Context("when the nested form contains invalid templating", func() {
			It("returns an error", func() {
				input.FormTypes = map[string]interface{}{
					"some-form": builder.Metadata{
						"name":  "some-form",
						"label": "$( invalid_helper )",
					},
				}
				interpolator := builder.NewInterpolator()
				_, err := interpolator.Interpolate(input, []byte(templateYAML))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to interpolate value"))
			})
		})

		Context("when template parsing fails", func() {
			It("returns an error", func() {

				interpolator := builder.NewInterpolator()
				_, err := interpolator.Interpolate(input, []byte("$(variable"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("template parsing failed"))
			})
		})

		Context("when template execution fails", func() {
			It("returns an error", func() {

				interpolator := builder.NewInterpolator()
				_, err := interpolator.Interpolate(input, []byte(`name: $( variable "some-missing-variable" )`))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("template execution failed"))
				Expect(err.Error()).To(ContainSubstring("could not find variable with key"))
			})
		})

		Context("when release tgz file does not exist but is provided", func() {
			It("returns an error", func() {

				interpolator := builder.NewInterpolator()
				_, err := interpolator.Interpolate(input, []byte(`releases: [$(release "some-release-does-not-exist")]`))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find release with name 'some-release-does-not-exist'"))
			})
		})

		Context("when the bosh_variable helper is used without providing the flag", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				input.BOSHVariables = nil
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("--bosh-variables-directory must be specified"))
			})
		})

		Context("when the form helper is used without providing the flag", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				input.FormTypes = nil
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("--forms-directory must be specified"))
			})
		})

		Context("when the property helper is used without providing the flag", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				input.PropertyBlueprints = nil
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("--properties-directory must be specified"))
			})
		})

		Context("when the release helper is used without providing the flag", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				input.ReleaseManifests = nil
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("--releases-directory must be specified"))
			})
		})

		Context("when the stemcell helper is used without providing the flag", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				input.StemcellManifest = nil
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("--stemcell-tarball must be specified"))
			})
		})

		Context("when the version helper is used without providing the flag", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				input.Version = ""
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("--version must be specified"))
			})
		})

		Context("when the variable helper is used without providing the flag", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				input.Variables = nil
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("--variable or --variables-file must be specified"))
			})
		})

		Context("when the icon helper is used without providing the flag", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				input.IconImage = ""
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("--icon must be specified"))
			})
		})

		Context("when the instance_group helper is used without providing the flag", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				input.InstanceGroups = nil
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("--instance-groups-directory must be specified"))
			})
		})

		Context("when the job helper is used without providing the flag", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				input.Jobs = nil
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("--jobs-directory must be specified"))
			})
		})

		Context("when the runtime_config helper is used without providing the flag", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				input.RuntimeConfigs = nil
				_, err := interpolator.Interpolate(input, []byte(templateYAML))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("--runtime-configs-directory must be specified"))
			})
		})

		Context("when a specified instance group is not included in the interpolate input", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				_, err := interpolator.Interpolate(input, []byte(`job_types: [$(instance_group "some-instance-group-does-not-exist")]`))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find instance_group with name 'some-instance-group-does-not-exist'"))
			})
		})

		Context("when a specified job is not included in the interpolate input", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				_, err := interpolator.Interpolate(input, []byte(`job: [$(job "some-job-does-not-exist")]`))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not find job with name 'some-job-does-not-exist'"))
			})
		})

		Context("input to the select function cannot be JSON unmarshalled", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				_, err := interpolator.Interpolate(input, []byte(`job: [$( "%%%" | select "value" )]`))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not JSON unmarshal \"%%%\": invalid character"))
			})
		})

		Context("input to the select function cannot be JSON unmarshalled", func() {
			It("returns an error", func() {
				interpolator := builder.NewInterpolator()
				_, err := interpolator.Interpolate(input, []byte(`release: [$( release "some-release" | select "key-not-there" )]`))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not select \"key-not-there\", key does not exist"))
			})
		})
	})
})
