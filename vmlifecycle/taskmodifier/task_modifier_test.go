package taskmodifier_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/vmlifecycle/taskmodifier"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("When modifying task files", func() {
	When("task directory exists", func() {
		var taskDir string
		BeforeEach(func() {
			var err error
			taskDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
		})

		When("config directories are specified", func() {
			It("errors when one of the config directories does not exist", func() {
				configDir, err := ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				taskModifier := taskmodifier.NewTaskModifier()
				Expect(err).NotTo(HaveOccurred())
				Expect(taskModifier).NotTo(BeNil())

				configPaths := []string{configDir, "does not exist"}
				err = taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, configPaths, nil)
				Expect(err).To(HaveOccurred())
			})
		})

		When("the task directory contains YAML files", func() {
			var configDir string
			BeforeEach(func() {
				var err error
				configDir, err = ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())
			})

			It("adds a params block if one is not present", func() {
				writeFile(filepath.Join(taskDir, "without-params.yml"), `not-params: {}`)
				writeFile(filepath.Join(configDir, "secrets.yml"), `some: ((secret))`)

				taskModifier := taskmodifier.NewTaskModifier()
				Expect(taskModifier).NotTo(BeNil())

				err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(readFile(filepath.Join(taskDir, "without-params.yml"))).To(MatchYAML(`{not-params: {}, params: { OM_VAR_secret: ((secret)), OM_VARS_ENV: OM_VAR }}`))
			})

			It("modifies tasks recursively", func() {
				writeFile(filepath.Join(taskDir, "some-dir/with-params.yml"), `params: {}`)
				writeFile(filepath.Join(configDir, "secrets.yml"), `some: ((secret))`)

				taskModifier := taskmodifier.NewTaskModifier()
				Expect(taskModifier).NotTo(BeNil())

				err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(readFile(filepath.Join(taskDir, "some-dir/with-params.yml"))).To(MatchYAML(`{
					params: { OM_VAR_secret: ((secret)), OM_VARS_ENV: OM_VAR }
				}`))
			})

			It("reads recursively from config directories", func() {
				writeFile(filepath.Join(taskDir, "some-dir/with-params.yml"), `params: {}`)
				writeFile(filepath.Join(configDir, "secrets.yml"), `some: ((secret))`)
				writeFile(filepath.Join(configDir, "some-dir/nested/secrets.yml"), `many: ((much))`)

				taskModifier := taskmodifier.NewTaskModifier()
				Expect(taskModifier).NotTo(BeNil())

				err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(readFile(filepath.Join(taskDir, "some-dir/with-params.yml"))).To(MatchYAML(`{
					params: { OM_VAR_secret: ((secret)), OM_VAR_much: ((much)), OM_VARS_ENV: OM_VAR }
				}`))
			})

			It("does not duplicate secrets found across multiple files", func() {
				writeFile(filepath.Join(taskDir, "some-dir/with-params.yml"), `params: {}`)
				writeFile(filepath.Join(configDir, "secrets.yml"), `some: ((secret))`)
				writeFile(filepath.Join(configDir, "secrets1.yml"), `some: ((secret))`)

				taskModifier := taskmodifier.NewTaskModifier()
				Expect(taskModifier).NotTo(BeNil())

				err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(readFile(filepath.Join(taskDir, "some-dir/with-params.yml"))).To(MatchYAML(`{
					params: { OM_VAR_secret: ((secret)), OM_VARS_ENV: OM_VAR}
				}`))
			})

			It("handles the `.yaml` prefix", func() {
				writeFile(filepath.Join(taskDir, "some-dir/with-params.yaml"), `params: {}`)
				writeFile(filepath.Join(configDir, "secrets.yaml"), `some: ((secret))`)

				taskModifier := taskmodifier.NewTaskModifier()
				Expect(taskModifier).NotTo(BeNil())

				err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(readFile(filepath.Join(taskDir, "some-dir/with-params.yaml"))).To(MatchYAML(`{
					params: { OM_VAR_secret: ((secret)), OM_VARS_ENV: OM_VAR}
				}`))
			})

			It("maintains the original fields in task unmodified", func() {
				writeFile(filepath.Join(taskDir, "some-dir/with-fields.yaml"), `{task: true, nested: {true: it is}, params: {}}`)
				writeFile(filepath.Join(configDir, "secrets.yaml"), `some: ((secret))`)

				taskModifier := taskmodifier.NewTaskModifier()
				Expect(taskModifier).NotTo(BeNil())

				err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(readFile(filepath.Join(taskDir, "some-dir/with-fields.yaml"))).To(MatchYAML(`{
					task: true,
					nested: {true: it is},
					params: { OM_VAR_secret: ((secret)), OM_VARS_ENV: OM_VAR}
				}`))
			})

			It("errors when a task file cannot be unmarshalled", func() {
				writeFile(filepath.Join(taskDir, "some-dir/not-valid.yaml"), `{[[[[[`)

				taskModifier := taskmodifier.NewTaskModifier()
				Expect(taskModifier).NotTo(BeNil())

				err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, nil)
				Expect(err).To(HaveOccurred())
			})

			It("does not modify any non-YAML files", func() {
				writeFile(filepath.Join(taskDir, "some-dir/with-params.yaml"), `params: {}`)
				writeFile(filepath.Join(taskDir, "some-dir/not-yaml.txt"), `params: {}`)

				taskModifier := taskmodifier.NewTaskModifier()
				Expect(taskModifier).NotTo(BeNil())

				err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(readFile(filepath.Join(taskDir, "some-dir/not-yaml.txt"))).To(Equal(`params: {}`))
			})

			It("writes a message about modifying the task file", func() {
				writeFile(filepath.Join(taskDir, "with-params.yml"), `params: {}`)
				writeFile(filepath.Join(configDir, "secrets.yml"), `some: ((secret))`)

				stderr := gbytes.NewBuffer()
				taskModifier := taskmodifier.NewTaskModifier()
				Expect(taskModifier).NotTo(BeNil())

				err := taskModifier.ModifyTasksWithSecrets(io.MultiWriter(stderr, GinkgoWriter), taskDir, []string{configDir}, nil)
				Expect(err).NotTo(HaveOccurred())

				Eventually(stderr).Should(gbytes.Say(`modifying task file.*with-params.yml`))
				Consistently(stderr).ShouldNot(gbytes.Say(`modifying task file.*without-params.yml`))
			})

			It("writes a message about reading secrets from a config", func() {
				writeFile(filepath.Join(taskDir, "with-params.yml"), `params: {}`)
				writeFile(filepath.Join(configDir, "secrets.yml"), `some: ((secret))`)
				writeFile(filepath.Join(configDir, "other.yml"), `no: secrets`)

				stderr := gbytes.NewBuffer()
				taskModifier := taskmodifier.NewTaskModifier()
				Expect(taskModifier).NotTo(BeNil())

				err := taskModifier.ModifyTasksWithSecrets(io.MultiWriter(stderr, GinkgoWriter), taskDir, []string{configDir}, nil)
				Expect(err).NotTo(HaveOccurred())

				Consistently(stderr).ShouldNot(gbytes.Say(`found secrets in.*other.yml`))
				Eventually(stderr).Should(gbytes.Say(`found secrets in.*secrets.yml`))
				Consistently(stderr).ShouldNot(gbytes.Say(`found secrets in.*other.yml`))
			})

			It("can read config files (not just directories)", func() {
				writeFile(filepath.Join(taskDir, "task.yml"), `params: {}`)

				configFile := filepath.Join(configDir, "config.yml")
				writeFile(configFile, `some: ((secret))`)

				taskModifier := taskmodifier.NewTaskModifier()
				Expect(taskModifier).NotTo(BeNil())

				err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(readFile(filepath.Join(taskDir, "task.yml"))).To(MatchYAML(`{
					params: { OM_VAR_secret: ((secret)), OM_VARS_ENV: OM_VAR }
				}`))
			})

			When("the config files have no parameters", func() {
				It("does not change the task files", func() {
					writeFile(filepath.Join(taskDir, "some-dir/with-params.yml"), `params: {}`)
					writeFile(filepath.Join(configDir, "secrets.yml"), `some: non-secret`)

					taskModifier := taskmodifier.NewTaskModifier()
					Expect(taskModifier).NotTo(BeNil())

					err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, nil)
					Expect(err).NotTo(HaveOccurred())

					Expect(readFile(filepath.Join(taskDir, "some-dir/with-params.yml"))).To(MatchYAML(`{
						params: {}
					}`))
				})
			})
		})

		Context("vars files", func() {
			var configDir string

			BeforeEach(func() {
				var err error
				configDir, err = ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				writeFile(filepath.Join(taskDir, "with-params.yml"), `params: {}`)
				writeFile(filepath.Join(configDir, "non-vars-file-secret.yml"), "non-vars-file-secret: ((non-vars-file-secret))")
				writeFile(filepath.Join(configDir, "nested-secret.yml"), "nested-secret: ((nested.secret))")
				writeFile(filepath.Join(configDir, "secrets.yml"), "non-nested-secret: ((non-nested-secret))\nanother-non-nested-secret: ((another-secret))\na-third-non-nested-secret: ((third-secret))")
				writeFile(filepath.Join(configDir, "duplicate-secrets.yml"), "non-nested-secret: ((non-nested-secret))")
			})

			When("vars files satisfy some of the params in config files", func() {
				var varsDir string

				BeforeEach(func() {
					var err error
					varsDir, err = ioutil.TempDir("", "")
					Expect(err).NotTo(HaveOccurred())
					writeFile(filepath.Join(varsDir, "whitelist-secrets.yml"), "nested:\n  secret: secret-nested-value\nnon-nested-secret: non-nested-value")
				})

				It("does not add params for secrets found in the vars files", func() {
					taskModifier := taskmodifier.NewTaskModifier()
					Expect(taskModifier).NotTo(BeNil())

					err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, []string{varsDir})
					Expect(err).NotTo(HaveOccurred())

					Expect(readFile(filepath.Join(taskDir, "with-params.yml"))).To(MatchYAML(`{
					params: {
						OM_VAR_non-vars-file-secret: ((non-vars-file-secret)),
						OM_VAR_another-secret: ((another-secret)),
						OM_VAR_third-secret: ((third-secret)),
						OM_VARS_ENV: OM_VAR
					}}`))
				})
			})

			When("vars files satisfy all of the params in config files", func() {
				var varsDir string

				BeforeEach(func() {
					var err error
					varsDir, err = ioutil.TempDir("", "")
					Expect(err).NotTo(HaveOccurred())
					writeFile(filepath.Join(varsDir, "all-secrets.yml"), "nested:\n  secret: secret-nested-value\nnon-nested-secret: non-nested-value\nnon-vars-file-secret: actually-a-secret\nanother-secret: secret2\nthird-secret: secret3")
				})

				It("does not change the task files", func() {
					taskModifier := taskmodifier.NewTaskModifier()
					Expect(taskModifier).NotTo(BeNil())

					err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, []string{varsDir})
					Expect(err).NotTo(HaveOccurred())

					Expect(readFile(filepath.Join(taskDir, "with-params.yml"))).To(MatchYAML(`{
						params: {}
					}`))
				})
			})

			It("returns an error when the vars path does not exist", func() {
				taskModifier := taskmodifier.NewTaskModifier()
				Expect(taskModifier).NotTo(BeNil())

				err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, []string{"doesnt-exist"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("vars directory 'doesnt-exist' does not exist:"))
			})

			When("the vars directory has an invalid name", func() {
				var invalidVarsDir string

				BeforeEach(func() {
					var err error

					invalidVarsDir, err = ioutil.TempDir("", "{invalid")
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					taskModifier := taskmodifier.NewTaskModifier()
					Expect(taskModifier).NotTo(BeNil())

					err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, []string{invalidVarsDir})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("could not find vars files in '%s'", invalidVarsDir)))
				})
			})

			When("the vars file is not valid yaml", func() {
				var varsDir string

				BeforeEach(func() {
					var err error
					varsDir, err = ioutil.TempDir("", "")
					invalidFile := filepath.Join(varsDir, "invalid.yml")
					Expect(err).NotTo(HaveOccurred())
					writeFile(filepath.Join(varsDir, "all-secrets.yml"), "nested:\n  secret: secret-nested-value\nnon-nested-secret: non-nested-value\nnon-vars-file-secret: actually-a-secret")
					writeFile(invalidFile, "invalid")
				})

				It("returns an error", func() {
					taskModifier := taskmodifier.NewTaskModifier()
					Expect(taskModifier).NotTo(BeNil())

					err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, taskDir, []string{configDir}, []string{varsDir})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("could not interpolate vars from '%s' into config file", []string{varsDir})))
				})
			})
		})
	})

	When("task directory does not exist", func() {
		It("returns an error", func() {
			taskModifier := taskmodifier.NewTaskModifier()

			err := taskModifier.ModifyTasksWithSecrets(GinkgoWriter, "unknown-task-dir", nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("task directory 'unknown-task-dir' does not exist"))
		})
	})
})
