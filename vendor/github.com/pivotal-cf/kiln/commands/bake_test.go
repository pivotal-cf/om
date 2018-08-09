package commands_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/kiln/builder"
	"github.com/pivotal-cf/kiln/commands"
	"github.com/pivotal-cf/kiln/commands/fakes"
	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("Bake", func() {
	var (
		fakeBOSHVariablesService     *fakes.BOSHVariablesService
		fakeFormsService             *fakes.FormsService
		fakeIconService              *fakes.IconService
		fakeInstanceGroupsService    *fakes.InstanceGroupsService
		fakeInterpolator             *fakes.Interpolator
		fakeJobsService              *fakes.JobsService
		fakeLogger                   *fakes.Logger
		fakeMetadataService          *fakes.MetadataService
		fakePropertiesService        *fakes.PropertiesService
		fakeReleasesService          *fakes.ReleasesService
		fakeRuntimeConfigsService    *fakes.RuntimeConfigsService
		fakeStemcellService          *fakes.StemcellService
		fakeTemplateVariablesService *fakes.TemplateVariablesService
		fakeTileWriter               *fakes.TileWriter

		otherReleasesDirectory     string
		someBOSHVariablesDirectory string
		someReleasesDirectory      string
		tmpDir                     string

		bake commands.Bake
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "command-test")
		Expect(err).NotTo(HaveOccurred())

		someReleasesDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		otherReleasesDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		someBOSHVariablesDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		nonTarballRelease := filepath.Join(someReleasesDirectory, "some-broken-release")
		err = ioutil.WriteFile(nonTarballRelease, []byte(""), 0644)
		Expect(err).NotTo(HaveOccurred())

		fakeBOSHVariablesService = &fakes.BOSHVariablesService{}
		fakeFormsService = &fakes.FormsService{}
		fakeIconService = &fakes.IconService{}
		fakeInstanceGroupsService = &fakes.InstanceGroupsService{}
		fakeInterpolator = &fakes.Interpolator{}
		fakeJobsService = &fakes.JobsService{}
		fakeLogger = &fakes.Logger{}
		fakeMetadataService = &fakes.MetadataService{}
		fakePropertiesService = &fakes.PropertiesService{}
		fakeReleasesService = &fakes.ReleasesService{}
		fakeRuntimeConfigsService = &fakes.RuntimeConfigsService{}
		fakeStemcellService = &fakes.StemcellService{}
		fakeTemplateVariablesService = &fakes.TemplateVariablesService{}
		fakeTileWriter = &fakes.TileWriter{}

		fakeTemplateVariablesService.FromPathsAndPairsReturns(map[string]interface{}{
			"some-variable-from-file": "some-variable-value-from-file",
			"some-variable":           "some-variable-value",
		}, nil)

		fakeReleasesService.FromDirectoriesReturns(map[string]interface{}{
			"some-release-1": builder.ReleaseManifest{
				Name:    "some-release-1",
				Version: "1.2.3",
				File:    "release1.tgz",
			},
			"some-release-2": builder.ReleaseManifest{
				Name:    "some-release-2",
				Version: "2.3.4",
				File:    "release2.tar.gz",
			},
		}, nil)

		fakeStemcellService.FromTarballReturns(builder.StemcellManifest{
			Version:         "2.3.4",
			OperatingSystem: "an-operating-system",
		}, nil)

		fakeFormsService.FromDirectoriesReturns(map[string]interface{}{
			"some-form": builder.Metadata{
				"name":  "some-form",
				"label": "some-form-label",
			},
		}, nil)

		fakeBOSHVariablesService.FromDirectoriesReturns(map[string]interface{}{
			"some-secret": builder.Metadata{
				"name": "some-secret",
				"type": "password",
			},
		}, nil)

		fakeInstanceGroupsService.FromDirectoriesReturns(map[string]interface{}{
			"some-instance-group": builder.Metadata{
				"name":     "some-instance-group",
				"manifest": "some-manifest",
				"provides": "some-link",
				"release":  "some-release",
			},
		}, nil)

		fakeJobsService.FromDirectoriesReturns(map[string]interface{}{
			"some-job": builder.Metadata{
				"name":     "some-job",
				"release":  "some-release",
				"consumes": "some-link",
			},
		}, nil)

		fakePropertiesService.FromDirectoriesReturns(map[string]interface{}{
			"some-property": builder.Metadata{
				"name":         "some-property",
				"type":         "boolean",
				"configurable": true,
				"default":      false,
			},
		}, nil)

		fakeRuntimeConfigsService.FromDirectoriesReturns(map[string]interface{}{
			"some-runtime-config": builder.Metadata{
				"name":           "some-runtime-config",
				"runtime_config": "some-addon-runtime-config",
			},
		}, nil)

		fakeIconService.EncodeReturns("some-encoded-icon", nil)

		fakeMetadataService.ReadReturns([]byte("some-metadata"), nil)

		fakeInterpolator.InterpolateReturns([]byte("some-interpolated-metadata"), nil)

		bake = commands.NewBake(
			fakeInterpolator,
			fakeTileWriter,
			fakeLogger,
			fakeTemplateVariablesService,
			fakeBOSHVariablesService,
			fakeReleasesService,
			fakeStemcellService,
			fakeFormsService,
			fakeInstanceGroupsService,
			fakeJobsService,
			fakePropertiesService,
			fakeRuntimeConfigsService,
			fakeIconService,
			fakeMetadataService,
		)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	Describe("Execute", func() {
		It("builds the tile", func() {
			err := bake.Execute([]string{
				"--embed", "some-embed-path",
				"--forms-directory", "some-forms-directory",
				"--icon", "some-icon-path",
				"--instance-groups-directory", "some-instance-groups-directory",
				"--jobs-directory", "some-jobs-directory",
				"--metadata", "some-metadata",
				"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
				"--properties-directory", "some-properties-directory",
				"--releases-directory", otherReleasesDirectory,
				"--releases-directory", someReleasesDirectory,
				"--runtime-configs-directory", "some-other-runtime-configs-directory",
				"--runtime-configs-directory", "some-runtime-configs-directory",
				"--stemcell-tarball", "some-stemcell-tarball",
				"--bosh-variables-directory", "some-other-variables-directory",
				"--bosh-variables-directory", "some-variables-directory",
				"--version", "1.2.3", "--migrations-directory", "some-migrations-directory",
				"--migrations-directory", "some-other-migrations-directory",
				"--variable", "some-variable=some-variable-value",
				"--variables-file", "some-variables-file",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeTemplateVariablesService.FromPathsAndPairsCallCount()).To(Equal(1))
			varFiles, variables := fakeTemplateVariablesService.FromPathsAndPairsArgsForCall(0)
			Expect(varFiles).To(Equal([]string{"some-variables-file"}))
			Expect(variables).To(Equal([]string{"some-variable=some-variable-value"}))

			Expect(fakeBOSHVariablesService.FromDirectoriesCallCount()).To(Equal(1))
			Expect(fakeBOSHVariablesService.FromDirectoriesArgsForCall(0)).To(Equal([]string{
				"some-other-variables-directory",
				"some-variables-directory",
			}))

			Expect(fakeReleasesService.FromDirectoriesCallCount()).To(Equal(1))
			Expect(fakeReleasesService.FromDirectoriesArgsForCall(0)).To(Equal([]string{otherReleasesDirectory, someReleasesDirectory}))

			Expect(fakeStemcellService.FromTarballCallCount()).To(Equal(1))
			Expect(fakeStemcellService.FromTarballArgsForCall(0)).To(Equal("some-stemcell-tarball"))

			Expect(fakeFormsService.FromDirectoriesCallCount()).To(Equal(1))
			Expect(fakeFormsService.FromDirectoriesArgsForCall(0)).To(Equal([]string{"some-forms-directory"}))

			Expect(fakeInstanceGroupsService.FromDirectoriesCallCount()).To(Equal(1))
			Expect(fakeInstanceGroupsService.FromDirectoriesArgsForCall(0)).To(Equal([]string{"some-instance-groups-directory"}))

			Expect(fakeJobsService.FromDirectoriesCallCount()).To(Equal(1))
			Expect(fakeJobsService.FromDirectoriesArgsForCall(0)).To(Equal([]string{"some-jobs-directory"}))

			Expect(fakePropertiesService.FromDirectoriesCallCount()).To(Equal(1))
			Expect(fakePropertiesService.FromDirectoriesArgsForCall(0)).To(Equal([]string{"some-properties-directory"}))

			Expect(fakeRuntimeConfigsService.FromDirectoriesCallCount()).To(Equal(1))
			Expect(fakeRuntimeConfigsService.FromDirectoriesArgsForCall(0)).To(Equal([]string{
				"some-other-runtime-configs-directory",
				"some-runtime-configs-directory",
			}))

			Expect(fakeIconService.EncodeCallCount()).To(Equal(1))
			Expect(fakeIconService.EncodeArgsForCall(0)).To(Equal("some-icon-path"))

			Expect(fakeMetadataService.ReadCallCount()).To(Equal(1))
			Expect(fakeMetadataService.ReadArgsForCall(0)).To(Equal("some-metadata"))

			Expect(fakeInterpolator.InterpolateCallCount()).To(Equal(1))

			input, metadata := fakeInterpolator.InterpolateArgsForCall(0)
			Expect(input).To(Equal(builder.InterpolateInput{
				Version: "1.2.3",
				BOSHVariables: map[string]interface{}{
					"some-secret": builder.Metadata{
						"name": "some-secret",
						"type": "password",
					},
				},
				Variables: map[string]interface{}{
					"some-variable-from-file": "some-variable-value-from-file",
					"some-variable":           "some-variable-value",
				},
				ReleaseManifests: map[string]interface{}{
					"some-release-1": builder.ReleaseManifest{
						Name:    "some-release-1",
						Version: "1.2.3",
						File:    "release1.tgz",
					},
					"some-release-2": builder.ReleaseManifest{
						Name:    "some-release-2",
						Version: "2.3.4",
						File:    "release2.tar.gz",
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
				IconImage: "some-encoded-icon",
				InstanceGroups: map[string]interface{}{
					"some-instance-group": builder.Metadata{
						"name":     "some-instance-group",
						"manifest": "some-manifest",
						"provides": "some-link",
						"release":  "some-release",
					},
				},
				Jobs: map[string]interface{}{
					"some-job": builder.Metadata{
						"name":     "some-job",
						"release":  "some-release",
						"consumes": "some-link",
					},
				},
				PropertyBlueprints: map[string]interface{}{
					"some-property": builder.Metadata{
						"name":         "some-property",
						"type":         "boolean",
						"configurable": true,
						"default":      false,
					},
				},
				RuntimeConfigs: map[string]interface{}{
					"some-runtime-config": builder.Metadata{
						"name":           "some-runtime-config",
						"runtime_config": "some-addon-runtime-config",
					},
				},
			}))

			Expect(string(metadata)).To(Equal("some-metadata"))

			Expect(fakeTileWriter.WriteCallCount()).To(Equal(1))
			metadata, writeInput := fakeTileWriter.WriteArgsForCall(0)
			Expect(string(metadata)).To(Equal("some-interpolated-metadata"))
			Expect(writeInput).To(Equal(builder.WriteInput{
				OutputFile:           filepath.Join("some-output-dir", "some-product-file-1.2.3-build.4"),
				StubReleases:         false,
				MigrationDirectories: []string{"some-migrations-directory", "some-other-migrations-directory"},
				ReleaseDirectories:   []string{otherReleasesDirectory, someReleasesDirectory},
				EmbedPaths:           []string{"some-embed-path"},
			}))
		})

		Context("when the optional flags are not specified", func() {
			It("builds the metadata", func() {
				err := bake.Execute([]string{
					"--metadata", "some-metadata",
					"--releases-directory", someReleasesDirectory,
					"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
					"--version", "1.2.3",
				})

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when multiple variable files are provided", func() {
			var otherVariableFile *os.File

			BeforeEach(func() {
				var err error
				otherVariableFile, err = ioutil.TempFile(tmpDir, "variables-file")
				Expect(err).NotTo(HaveOccurred())
				defer otherVariableFile.Close()

				variables := map[string]string{
					"some-variable-from-file":       "override-variable-from-other-file",
					"some-other-variable-from-file": "some-other-variable-value-from-file",
				}
				data, err := yaml.Marshal(&variables)
				Expect(err).NotTo(HaveOccurred())

				n, err := otherVariableFile.Write(data)
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(HaveLen(n))
			})

			It("interpolates variables from both files", func() {
				err := bake.Execute([]string{
					"--embed", "some-embed-path",
					"--forms-directory", "some-forms-directory",
					"--icon", "some-icon-path",
					"--instance-groups-directory", "some-instance-groups-directory",
					"--jobs-directory", "some-jobs-directory",
					"--metadata", "some-metadata",
					"--migrations-directory", "some-migrations-directory",
					"--migrations-directory", "some-other-migrations-directory",
					"--output-file", "some-output-dir/some-product-file-1.2.3-build.4.pivotal",
					"--releases-directory", otherReleasesDirectory,
					"--releases-directory", someReleasesDirectory,
					"--runtime-configs-directory", "some-runtime-configs-directory",
					"--stemcell-tarball", "some-stemcell-tarball",
					"--bosh-variables-directory", "some-variables-directory",
					"--variable", "some-variable=some-variable-value",
					"--variables-file", "some-variable-file-1",
					"--variables-file", "some-variable-file-2",
					"--version", "1.2.3",
				})

				Expect(err).NotTo(HaveOccurred())

				generatedMetadataContents, _ := fakeTileWriter.WriteArgsForCall(0)
				Expect(generatedMetadataContents).To(HelpfullyMatchYAML("some-interpolated-metadata"))
			})
		})

		Context("failure cases", func() {
			Context("when the template variables service errors", func() {
				It("returns an error", func() {
					fakeTemplateVariablesService.FromPathsAndPairsReturns(nil, errors.New("parsing template variables failed"))

					err := bake.Execute([]string{
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
						"--icon", "some-icon-path",
						"--releases-directory", someReleasesDirectory,
					})
					Expect(err).To(MatchError("failed to parse template variables: parsing template variables failed"))
				})
			})

			Context("when the icon service fails", func() {
				It("returns an error", func() {
					fakeIconService.EncodeReturns("", errors.New("encoding icon failed"))

					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
						"--releases-directory", someReleasesDirectory,
					})

					Expect(err).To(MatchError("failed to encode icon: encoding icon failed"))
				})
			})

			Context("when the metadata service fails", func() {
				It("returns an error", func() {
					fakeMetadataService.ReadReturns(nil, errors.New("reading metadata failed"))

					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
						"--releases-directory", someReleasesDirectory,
					})

					Expect(err).To(MatchError("failed to read metadata: reading metadata failed"))
				})
			})

			Context("when the releases service fails", func() {
				It("returns an error", func() {
					fakeReleasesService.FromDirectoriesReturns(nil, errors.New("parsing releases failed"))

					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
						"--releases-directory", someReleasesDirectory,
					})

					Expect(err).To(MatchError("failed to parse releases: parsing releases failed"))
				})
			})

			Context("when the stemcell service fails", func() {
				It("returns an error", func() {
					fakeStemcellService.FromTarballReturns(nil, errors.New("parsing stemcell failed"))

					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
						"--properties-directory", "some-properties-directory",
						"--releases-directory", someReleasesDirectory,
						"--stemcell-tarball", "some-stemcell-tarball",
						"--version", "1.2.3",
					})

					Expect(err).To(MatchError("failed to parse stemcell: parsing stemcell failed"))
				})
			})

			Context("when the forms service fails", func() {
				It("returns an error", func() {
					fakeFormsService.FromDirectoriesReturns(nil, errors.New("parsing forms failed"))

					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
						"--properties-directory", "some-properties-directory",
						"--releases-directory", someReleasesDirectory,
						"--stemcell-tarball", "some-stemcell-tarball",
						"--forms-directory", "some-form-directory",
						"--version", "1.2.3",
					})

					Expect(err).To(MatchError("failed to parse forms: parsing forms failed"))
				})
			})

			Context("when the instance groups service fails", func() {
				It("returns an error", func() {
					fakeInstanceGroupsService.FromDirectoriesReturns(nil, errors.New("parsing instance groups failed"))

					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
						"--properties-directory", "some-properties-directory",
						"--releases-directory", someReleasesDirectory,
						"--stemcell-tarball", "some-stemcell-tarball",
						"--forms-directory", "some-form-directory",
						"--instance-groups-directory", "some-instance-group-directory",
						"--version", "1.2.3",
					})

					Expect(err).To(MatchError("failed to parse instance groups: parsing instance groups failed"))
				})
			})

			Context("when the jobs service fails", func() {
				It("returns an error", func() {
					fakeJobsService.FromDirectoriesReturns(nil, errors.New("parsing jobs failed"))

					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
						"--properties-directory", "some-properties-directory",
						"--releases-directory", someReleasesDirectory,
						"--stemcell-tarball", "some-stemcell-tarball",
						"--forms-directory", "some-form-directory",
						"--instance-groups-directory", "some-instance-group-directory",
						"--version", "1.2.3",
					})

					Expect(err).To(MatchError("failed to parse jobs: parsing jobs failed"))
				})
			})

			Context("when the properties service fails", func() {
				It("returns an error", func() {
					fakePropertiesService.FromDirectoriesReturns(nil, errors.New("parsing properties failed"))

					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
						"--properties-directory", "some-properties-directory",
						"--releases-directory", someReleasesDirectory,
						"--stemcell-tarball", "some-stemcell-tarball",
						"--forms-directory", "some-form-directory",
						"--instance-groups-directory", "some-instance-group-directory",
						"--version", "1.2.3",
					})

					Expect(err).To(MatchError("failed to parse properties: parsing properties failed"))
				})
			})

			Context("when the runtime configs service fails", func() {
				It("returns an error", func() {
					fakeRuntimeConfigsService.FromDirectoriesReturns(nil, errors.New("parsing runtime configs failed"))

					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
						"--runtime-configs-directory", "some-runtime-configs-directory",
						"--releases-directory", someReleasesDirectory,
						"--stemcell-tarball", "some-stemcell-tarball",
						"--forms-directory", "some-form-directory",
						"--instance-groups-directory", "some-instance-group-directory",
						"--version", "1.2.3",
					})

					Expect(err).To(MatchError("failed to parse runtime configs: parsing runtime configs failed"))
				})
			})

			Context("when the template interpolator returns an error", func() {
				It("returns the error", func() {
					fakeInterpolator.InterpolateReturns(nil, errors.New("some-error"))

					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
						"--properties-directory", "some-properties-directory",
						"--releases-directory", someReleasesDirectory,
						"--stemcell-tarball", "some-stemcell-tarball",
						"--forms-directory", "some-form-directory",
						"--version", "1.2.3",
					})

					Expect(err).To(MatchError(ContainSubstring("some-error")))
				})
			})

			Context("when the metadata flag is missing", func() {
				It("returns an error", func() {
					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4.pivotal",
						"--releases-directory", someReleasesDirectory,
						"--stemcell-tarball", "some-stemcell-tarball",
						"--version", "1.2.3",
					})

					Expect(err).To(MatchError("missing required flag \"--metadata\""))
				})
			})

			Context("when the release-tarball flag is missing", func() {
				It("returns an error", func() {
					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4.pivotal",
						"--stemcell-tarball", "some-stemcell-tarball",
						"--version", "1.2.3",
					})

					Expect(err).To(MatchError("missing required flag \"--releases-directory\""))
				})
			})

			Context("when the output-file flag is missing", func() {
				It("returns an error", func() {
					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--metadata", "some-metadata",
						"--releases-directory", someReleasesDirectory,
						"--stemcell-tarball", "some-stemcell-tarball",
						"--version", "1.2.3",
					})

					Expect(err).To(MatchError("missing required flag \"--output-file\""))
				})
			})

			Context("when the jobs-directory flag is passed without the instance-groups-directory flag", func() {
				It("returns an error", func() {
					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--jobs-directory", "some-jobs-directory",
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
						"--releases-directory", someReleasesDirectory,
						"--stemcell-tarball", "some-stemcell-tarball",
						"--version", "1.2.3",
					})

					Expect(err).To(MatchError("--jobs-directory flag requires --instance-groups-directory to also be specified"))
				})
			})

			Context("when an invalid flag is passed", func() {
				It("returns an error", func() {
					err := bake.Execute([]string{
						"--icon", "some-icon-path",
						"--jobs-directory", "some-jobs-directory",
						"--metadata", "some-metadata",
						"--output-file", "some-output-dir/some-product-file-1.2.3-build.4",
						"--releases-directory", someReleasesDirectory,
						"--stemcell-tarball", "some-stemcell-tarball",
						"--version", "1.2.3",
						"--non-existant-flag",
					})

					Expect(err).To(MatchError(ContainSubstring("non-existant-flag")))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			Expect(bake.Usage()).To(Equal(jhanda.Usage{
				Description:      "Bakes tile metadata, stemcell, releases, and migrations into a format that can be consumed by OpsManager.",
				ShortDescription: "bakes a tile",
				Flags:            bake.Options,
			}))
		})
	})
})
