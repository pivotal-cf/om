package acceptance

import (
	"archive/zip"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("bake command", func() {
	var (
		cfSHA1                           string
		diegoSHA1                        string
		metadata                         string
		otherReleasesDirectory           string
		outputFile                       string
		someBOSHVariablesDirectory       string
		someFormsDirectory               string
		someIconPath                     string
		someInstanceGroupsDirectory      string
		someJobsDirectory                string
		someOtherFormsDirectory          string
		someOtherInstanceGroupsDirectory string
		someOtherJobsDirectory           string
		somePropertiesDirectory          string
		someReleasesDirectory            string
		someRuntimeConfigsDirectory      string
		someVarFile                      string
		stemcellTarball                  string
		tmpDir                           string
		variableFile                     *os.File

		commandWithArgs []string
	)

	BeforeEach(func() {
		var err error

		tmpDir, err = ioutil.TempDir("", "kiln-main-test")
		Expect(err).NotTo(HaveOccurred())

		tileDir, err := ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		outputFile = filepath.Join(tileDir, "cool-product-1.2.3-build.4.pivotal")

		someIconFile, err := ioutil.TempFile("", "icon")
		Expect(err).NotTo(HaveOccurred())
		defer someIconFile.Close()
		someIconPath = someIconFile.Name()

		someImageData := "i-am-some-image"
		_, err = someIconFile.Write([]byte(someImageData))
		Expect(err).NotTo(HaveOccurred())

		somePropertiesDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		someReleasesDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		otherReleasesDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		someRuntimeConfigsDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		someBOSHVariablesDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		someFormsDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		someOtherFormsDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		someInstanceGroupsDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		someOtherInstanceGroupsDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		someJobsDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		someOtherJobsDirectory, err = ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		someVarDir, err := ioutil.TempDir(tmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		variableFile, err = ioutil.TempFile(tmpDir, "variables-file")
		Expect(err).NotTo(HaveOccurred())
		defer variableFile.Close()

		variables := map[string]string{"some-variable": "some-variable-value"}
		data, err := yaml.Marshal(&variables)
		Expect(err).NotTo(HaveOccurred())

		n, err := variableFile.Write(data)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(HaveLen(n))

		someVarFile = filepath.Join(someVarDir, "var-file.yml")

		cfReleaseManifest := `---
name: cf
version: 235
`
		err = ioutil.WriteFile(filepath.Join(somePropertiesDirectory, "some-templated-property.yml"), []byte(`---
name: some_templated_property_blueprint
type: boolean
configurable: false
default: true
`), 0644)

		Expect(err).NotTo(HaveOccurred())

		_, err = ioutil.TempFile(someReleasesDirectory, "")
		Expect(err).NotTo(HaveOccurred())

		_, err = createTarball(someReleasesDirectory, "cf-release-235.0.0-3215.4.0.tgz", "release.MF", cfReleaseManifest)
		Expect(err).NotTo(HaveOccurred())

		f, err := os.Open(filepath.Join(someReleasesDirectory, "cf-release-235.0.0-3215.4.0.tgz"))
		Expect(err).NotTo(HaveOccurred())

		hash := sha1.New()
		_, err = io.Copy(hash, f)
		Expect(err).NotTo(HaveOccurred())

		cfSHA1 = fmt.Sprintf("%x", hash.Sum(nil))

		diegoReleaseManifest := `---
name: diego
version: 0.1467.1
key: value
`

		_, err = createTarball(otherReleasesDirectory, "diego-release-0.1467.1-3215.4.0.tgz", "release.MF", diegoReleaseManifest)
		Expect(err).NotTo(HaveOccurred())

		f, err = os.Open(filepath.Join(otherReleasesDirectory, "diego-release-0.1467.1-3215.4.0.tgz"))
		Expect(err).NotTo(HaveOccurred())

		hash = sha1.New()
		_, err = io.Copy(hash, f)
		Expect(err).NotTo(HaveOccurred())

		diegoSHA1 = fmt.Sprintf("%x", hash.Sum(nil))

		notATarball := filepath.Join(someReleasesDirectory, "not-a-tarball.txt")
		_ = ioutil.WriteFile(notATarball, []byte(`this is not a tarball`), 0644)
		stemcellManifest := `---
version: "3215.4"
operating_system: ubuntu-trusty
`

		stemcellTarball, err = createTarball(tmpDir, "stemcell.tgz", "stemcell.MF", stemcellManifest)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(someFormsDirectory, "some-config.yml"), []byte(`---
name: some-config
label: some-form-label
description: some-form-description
`), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(someFormsDirectory, "some-other-config.yml"), []byte(`---
name: some-other-config
label: some-other-form-label
description: some-other-form-description
`), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(someOtherFormsDirectory, "some-more-config.yml"), []byte(`---
name: some-more-config
label: some-form-label
description: some-form-description
`), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(someInstanceGroupsDirectory, "some-instance-group.yml"), []byte(`---
name: some-instance-group
label: Some Instance Group
templates:
- $( job "some-job" )
`), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(someOtherInstanceGroupsDirectory, "some-other-instance-group.yml"), []byte(`---
name: some-other-instance-group
label: Some Other Instance Group
templates:
- $( job "some-job-alias" )
`), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(someJobsDirectory, "some-job.yml"), []byte(`---
name: some-job
release: some-release
`), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(someOtherJobsDirectory, "some-other-job.yml"), []byte(`---
name: some-other-job
alias: some-job-alias
release: some-other-release
`), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(someRuntimeConfigsDirectory, "some-runtime-config.yml"), []byte(`---
name: some-runtime-config
runtime_config: |
  releases:
  - name: some-addon
    version: some-addon-version
`), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(someBOSHVariablesDirectory, "variable-1.yml"), []byte(`---
- name: variable-1
  type: certificate
  options:
    some_option: Option value
- name: variable-2
  type: password
`), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(someVarFile, []byte(`---
some-boolean-variable: true
some-literal-variable: |
  { "some": "value" }
`), 0644)
		Expect(err).NotTo(HaveOccurred())

		metadata = filepath.Join(tmpDir, "metadata.yml")
		err = ioutil.WriteFile(metadata, untemplatedMetadata, 0644)
		Expect(err).NotTo(HaveOccurred())

		commandWithArgs = []string{
			"bake",
			"--bosh-variables-directory", someBOSHVariablesDirectory,
			"--forms-directory", someFormsDirectory,
			"--forms-directory", someOtherFormsDirectory,
			"--icon", someIconPath,
			"--instance-groups-directory", someInstanceGroupsDirectory,
			"--instance-groups-directory", someOtherInstanceGroupsDirectory,
			"--jobs-directory", someJobsDirectory,
			"--jobs-directory", someOtherJobsDirectory,
			"--metadata", metadata,
			"--output-file", outputFile,
			"--properties-directory", somePropertiesDirectory,
			"--releases-directory", otherReleasesDirectory,
			"--releases-directory", someReleasesDirectory,
			"--runtime-configs-directory", someRuntimeConfigsDirectory,
			"--stemcell-tarball", stemcellTarball,
			"--variable", "some-variable=some-variable-value",
			"--variables-file", filepath.Join(someVarFile),
			"--version", "1.2.3",
		}
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	It("generates a tile with the correct metadata", func() {
		commandWithArgs = append(commandWithArgs, "--migrations-directory",
			"fixtures/extra-migrations",
			"--migrations-directory",
			"fixtures/migrations",
			"--variables-file",
			variableFile.Name())

		f, err := os.OpenFile(metadata, os.O_APPEND|os.O_WRONLY, 0644)
		Expect(err).NotTo(HaveOccurred())

		defer f.Close()
		_, err = f.WriteString("icon_img: $( icon )")

		Expect(err).NotTo(HaveOccurred())

		command := exec.Command(pathToMain, commandWithArgs...)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		archive, err := os.Open(outputFile)
		Expect(err).NotTo(HaveOccurred())

		archiveInfo, err := archive.Stat()
		Expect(err).NotTo(HaveOccurred())

		zr, err := zip.NewReader(archive, archiveInfo.Size())
		Expect(err).NotTo(HaveOccurred())

		var file io.ReadCloser
		for _, f := range zr.File {
			if f.Name == "metadata/metadata.yml" {
				file, err = f.Open()
				Expect(err).NotTo(HaveOccurred())
				break
			}
		}

		Expect(file).NotTo(BeNil(), "metadata was not found in built tile")
		metadataContents, err := ioutil.ReadAll(file)
		Expect(err).NotTo(HaveOccurred())

		renderedYAML := fmt.Sprintf(expectedMetadata, diegoSHA1, cfSHA1)
		Expect(metadataContents).To(HelpfullyMatchYAML(renderedYAML))

		// Bosh Variables
		Expect(string(metadataContents)).To(ContainSubstring("name: variable-1"))
		Expect(string(metadataContents)).To(ContainSubstring("name: variable-2"))
		Expect(string(metadataContents)).To(ContainSubstring("type: certificate"))
		Expect(string(metadataContents)).To(ContainSubstring("some_option: Option value"))

		// Template Variables
		Expect(string(metadataContents)).To(ContainSubstring("custom_variable: some-variable-value"))

		var (
			archivedMigration1 io.ReadCloser
			archivedMigration2 io.ReadCloser
			archivedMigration3 io.ReadCloser
		)

		for _, f := range zr.File {
			if f.Name == "migrations/v1/201603041539_custom_buildpacks.js" {
				archivedMigration1, err = f.Open()
				Expect(err).NotTo(HaveOccurred())
			}

			if f.Name == "migrations/v1/201603071158_auth_enterprise_sso.js" {
				archivedMigration2, err = f.Open()
				Expect(err).NotTo(HaveOccurred())
			}

			if f.Name == "migrations/v1/some_migration.js" {
				archivedMigration3, err = f.Open()
				Expect(err).NotTo(HaveOccurred())
			}
		}

		contents, err := ioutil.ReadAll(archivedMigration1)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("custom-buildpack-migration\n"))

		contents, err = ioutil.ReadAll(archivedMigration2)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("auth-enterprise-sso-migration\n"))

		contents, err = ioutil.ReadAll(archivedMigration3)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("some_migration\n"))

		Eventually(session.Out).Should(gbytes.Say("Reading release manifests"))
		Eventually(session.Out).Should(gbytes.Say("Reading stemcell manifest"))
		Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf("Building %s", outputFile)))
		Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf("Adding metadata/metadata.yml to %s...", outputFile)))
		Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf("Adding migrations/v1/201603041539_custom_buildpacks.js to %s...", outputFile)))
		Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf("Adding migrations/v1/201603071158_auth_enterprise_sso.js to %s...", outputFile)))
		Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf("Adding releases/diego-release-0.1467.1-3215.4.0.tgz to %s...", outputFile)))
		Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf("Adding releases/cf-release-235.0.0-3215.4.0.tgz to %s...", outputFile)))
		Eventually(session.Out).ShouldNot(gbytes.Say(fmt.Sprintf("Adding releases/not-a-tarball.txt to %s...", outputFile)))
		Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf("Calculating md5 sum of %s...", outputFile)))
		Eventually(session.Out).Should(gbytes.Say("Calculated md5 sum: [0-9a-f]{32}"))
	})

	Context("when the --stub-releases flag is specified", func() {
		It("creates a tile with empty release tarballs", func() {
			commandWithArgs = append(commandWithArgs, "--stub-releases")

			command := exec.Command(pathToMain, commandWithArgs...)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			archive, err := os.Open(outputFile)
			Expect(err).NotTo(HaveOccurred())

			archiveInfo, err := archive.Stat()
			Expect(err).NotTo(HaveOccurred())

			zr, err := zip.NewReader(archive, archiveInfo.Size())
			Expect(err).NotTo(HaveOccurred())

			for _, f := range zr.File {
				if f.Name == "releases/cf-release-235.0.0-3215.4.0.tgz" {
					Expect(f.UncompressedSize64).To(Equal(uint64(0)))
				}

				if f.Name == "releases/diego-release-0.1467.1-3215.4.0.tgz" {
					Expect(f.UncompressedSize64).To(Equal(uint64(0)))
				}
			}
		})
	})

	Context("when no migrations are provided", func() {
		It("creates empty migrations folder", func() {
			command := exec.Command(pathToMain, commandWithArgs...)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			archive, err := os.Open(outputFile)
			Expect(err).NotTo(HaveOccurred())

			archiveInfo, err := archive.Stat()
			Expect(err).NotTo(HaveOccurred())

			zr, err := zip.NewReader(archive, archiveInfo.Size())
			Expect(err).NotTo(HaveOccurred())

			var emptyMigrationsFolderMode os.FileMode
			for _, f := range zr.File {
				if f.Name == "migrations/v1/" {
					emptyMigrationsFolderMode = f.Mode()
					break
				}
			}

			Expect(emptyMigrationsFolderMode.IsDir()).To(BeTrue())

			Eventually(session.Out).Should(gbytes.Say(fmt.Sprintf("Creating empty migrations folder in %s...", outputFile)))
		})
	})

	Context("when the --embed flag is specified", func() {
		Context("when only file paths are specified", func() {
			It("creates a tile with the specified file copied into the embed directory", func() {
				someFileToEmbed := filepath.Join(tmpDir, "some-file-to-embed")
				otherFileToEmbed := filepath.Join(tmpDir, "other-file-to-embed")

				err := ioutil.WriteFile(someFileToEmbed, []byte("content-of-some-file"), 0600)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(otherFileToEmbed, []byte("content-of-other-file"), 0755)
				Expect(err).NotTo(HaveOccurred())

				commandWithArgs = append(commandWithArgs,
					"--embed", otherFileToEmbed,
					"--embed", someFileToEmbed,
					"--stub-releases")

				command := exec.Command(pathToMain, commandWithArgs...)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))

				archive, err := os.Open(outputFile)
				Expect(err).NotTo(HaveOccurred())

				archiveInfo, err := archive.Stat()
				Expect(err).NotTo(HaveOccurred())

				zr, err := zip.NewReader(archive, archiveInfo.Size())
				Expect(err).NotTo(HaveOccurred())

				seenSomeFile := false
				seenOtherFile := false
				for _, f := range zr.File {
					if f.Name == "embed/some-file-to-embed" {
						seenSomeFile = true
						r, err := f.Open()
						Expect(err).NotTo(HaveOccurred())

						content, err := ioutil.ReadAll(r)
						Expect(err).NotTo(HaveOccurred())

						Expect(content).To(Equal([]byte("content-of-some-file")))
					}

					if f.Name == "embed/other-file-to-embed" {
						seenOtherFile = true
						r, err := f.Open()
						Expect(err).NotTo(HaveOccurred())

						content, err := ioutil.ReadAll(r)
						Expect(err).NotTo(HaveOccurred())

						mode := f.FileHeader.Mode()
						Expect(mode).To(Equal(os.FileMode(0755)))

						Expect(content).To(Equal([]byte("content-of-other-file")))
					}
				}

				Expect(seenSomeFile).To(BeTrue())
				Expect(seenOtherFile).To(BeTrue())
			})
		})

		Context("when an embed directory is specified", func() {
			It("embeds the root directory and retains its structure", func() {
				dirToAdd := filepath.Join(tmpDir, "some-dir")
				nestedDir := filepath.Join(dirToAdd, "some-nested-dir")
				someFileToEmbed := filepath.Join(nestedDir, "some-file-to-embed")

				err := os.MkdirAll(nestedDir, 0700)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(someFileToEmbed, []byte("content-of-some-file"), 0600)
				Expect(err).NotTo(HaveOccurred())

				commandWithArgs = append(commandWithArgs,
					"--embed", dirToAdd,
					"--stub-releases")
				command := exec.Command(pathToMain, commandWithArgs...)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))

				archive, err := os.Open(outputFile)
				Expect(err).NotTo(HaveOccurred())

				archiveInfo, err := archive.Stat()
				Expect(err).NotTo(HaveOccurred())

				zr, err := zip.NewReader(archive, archiveInfo.Size())
				Expect(err).NotTo(HaveOccurred())

				seenFile := false
				for _, f := range zr.File {
					if f.Name == "embed/some-dir/some-nested-dir/some-file-to-embed" {
						seenFile = true
						r, err := f.Open()
						Expect(err).NotTo(HaveOccurred())

						content, err := ioutil.ReadAll(r)
						Expect(err).NotTo(HaveOccurred())

						Expect(content).To(Equal([]byte("content-of-some-file")))
					}
				}

				Expect(seenFile).To(BeTrue())
			})
		})
	})

	Context("failure cases", func() {
		Context("when a release tarball does not exist", func() {
			It("prints an error and exits 1", func() {
				commandWithArgs = append(commandWithArgs, "--releases-directory", "missing-directory")
				command := exec.Command(pathToMain, commandWithArgs...)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(string(session.Err.Contents())).To(ContainSubstring("lstat missing-directory: no such file or directory"))
			})
		})

		Context("when the output directory is not writable", func() {
			It("prints an error and exit 1", func() {
				command := exec.Command(pathToMain,
					"bake",
					"--forms-directory", someFormsDirectory,
					"--forms-directory", someOtherFormsDirectory,
					"--icon", someIconPath,
					"--metadata", metadata,
					"--output-file", "/path/to/missing/dir/product.zip",
					"--releases-directory", someReleasesDirectory,
					"--releases-directory", otherReleasesDirectory,
					"--instance-groups-directory", someInstanceGroupsDirectory,
					"--instance-groups-directory", someOtherInstanceGroupsDirectory,
					"--jobs-directory", someJobsDirectory,
					"--jobs-directory", someOtherJobsDirectory,
					"--properties-directory", somePropertiesDirectory,
					"--runtime-configs-directory", someRuntimeConfigsDirectory,
					"--stemcell-tarball", stemcellTarball,
					"--bosh-variables-directory", someBOSHVariablesDirectory,
					"--variable", "some-variable=some-variable-value",
					"--variables-file", filepath.Join(someVarFile),
					"--version", "1.2.3",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(string(string(session.Err.Contents()))).To(ContainSubstring("no such file or directory"))
			})
		})
	})
})
