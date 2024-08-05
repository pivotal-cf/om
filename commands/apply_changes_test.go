package commands_test

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/onsi/gomega/gbytes"
	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ApplyChanges", func() {
	var (
		service        *fakes.ApplyChangesService
		pendingService *fakes.PendingChangesService
		logger         *log.Logger
		stderr         *gbytes.Buffer
		writer         *fakes.LogWriter
	)

	BeforeEach(func() {
		service = &fakes.ApplyChangesService{}
		pendingService = &fakes.PendingChangesService{}
		stderr = gbytes.NewBuffer()
		logger = log.New(stderr, "", 0)
		writer = &fakes.LogWriter{}
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			// 2.9 due to apply changes recreate only working on 2.9+
			service.InfoReturns(api.Info{Version: "2.9"}, nil)
			service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)
			service.RunningInstallationReturns(api.InstallationsServiceOutput{}, nil)

			service.GetInstallationReturnsOnCall(0, api.InstallationsServiceOutput{Status: "running"}, nil)
			service.GetInstallationReturnsOnCall(1, api.InstallationsServiceOutput{Status: "running"}, nil)
			service.GetInstallationReturnsOnCall(2, api.InstallationsServiceOutput{Status: "succeeded"}, nil)

			service.GetInstallationLogsReturnsOnCall(0, api.InstallationsServiceOutput{Logs: "start of logs"}, nil)
			service.GetInstallationLogsReturnsOnCall(1, api.InstallationsServiceOutput{Logs: "these logs"}, nil)
			service.GetInstallationLogsReturnsOnCall(2, api.InstallationsServiceOutput{Logs: "some other logs"}, nil)
		})

		It("applies changes to the Ops Manager", func() {
			command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(service.CreateInstallationCallCount()).To(Equal(1))

			ignoreWarnings, deployProducts, forceLatestVariables, _, _ := service.CreateInstallationArgsForCall(0)
			Expect(ignoreWarnings).To(Equal(false))
			Expect(deployProducts).To(Equal(true))
			Expect(forceLatestVariables).To(Equal(false))

			Expect(stderr).To(gbytes.Say("attempting to apply changes to the targeted Ops Manager"))

			Expect(service.GetInstallationArgsForCall(0)).To(Equal(311))
			Expect(service.GetInstallationCallCount()).To(Equal(3))

			Expect(service.GetInstallationLogsArgsForCall(0)).To(Equal(311))
			Expect(service.GetInstallationLogsCallCount()).To(Equal(3))

			Expect(writer.FlushCallCount()).To(Equal(3))
			Expect(writer.FlushArgsForCall(0)).To(Equal("start of logs"))
			Expect(writer.FlushArgsForCall(1)).To(Equal("these logs"))
			Expect(writer.FlushArgsForCall(2)).To(Equal("some other logs"))
		})

		It("retries apply changes to the Ops Manager", func() {
			service.GetInstallationReturnsOnCall(0, api.InstallationsServiceOutput{}, errors.New("some error"))
			service.GetInstallationReturnsOnCall(1, api.InstallationsServiceOutput{Status: "running"}, nil)
			service.GetInstallationReturnsOnCall(2, api.InstallationsServiceOutput{Status: "succeeded"}, nil)

			command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(service.CreateInstallationCallCount()).To(Equal(1))

			ignoreWarnings, deployProducts, _, _, _ := service.CreateInstallationArgsForCall(0)
			Expect(ignoreWarnings).To(Equal(false))
			Expect(deployProducts).To(Equal(true))

			Expect(stderr).To(gbytes.Say("attempting to apply changes to the targeted Ops Manager"))

			Expect(service.GetInstallationArgsForCall(0)).To(Equal(311))
			Expect(service.GetInstallationCallCount()).To(Equal(3))

			Expect(service.GetInstallationLogsArgsForCall(0)).To(Equal(311))
			Expect(service.GetInstallationLogsCallCount()).To(Equal(2))

			Expect(writer.FlushCallCount()).To(Equal(2))
			Expect(writer.FlushArgsForCall(0)).To(Equal("start of logs"))
			Expect(writer.FlushArgsForCall(1)).To(Equal("these logs"))
		})

		When("passed the ignore-warnings flag", func() {
			It("applies changes while ignoring warnings", func() {
				service.InfoReturns(api.Info{Version: "2.3-build43"}, nil)

				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

				err := executeCommand(command, []string{"--ignore-warnings"})
				Expect(err).ToNot(HaveOccurred())

				ignoreWarnings, _, _, _, _ := service.CreateInstallationArgsForCall(0)
				Expect(ignoreWarnings).To(Equal(true))
			})
		})

		When("passed the force-latest-variables flag", func() {
			It("applies changes while forcing the latest variable versions to be used", func() {
				service.InfoReturns(api.Info{Version: "2.3-build43"}, nil)

				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

				err := executeCommand(command, []string{"--force-latest-variables"})
				Expect(err).ToNot(HaveOccurred())

				_, _, forceLatestVariables, _, _ := service.CreateInstallationArgsForCall(0)
				Expect(forceLatestVariables).To(Equal(true))
			})
		})

		When("passed the skip-deploy-products flag", func() {
			It("applies changes while not deploying products", func() {
				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

				err := executeCommand(command, []string{"--skip-deploy-products"})
				Expect(err).ToNot(HaveOccurred())

				_, _, deployProducts, _, _ := service.CreateInstallationArgsForCall(0)
				Expect(deployProducts).To(Equal(false))
			})

			It("fails if product names were specified", func() {
				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)
				err := executeCommand(command, []string{"--skip-deploy-products", "--product-name", "product1"})
				Expect(err).To(HaveOccurred())
			})
		})

		When("passed the product-name flag", func() {
			It("passes product names to the installation service", func() {
				service.InfoReturns(api.Info{Version: "2.2-build243"}, nil)
				service.CreateInstallationReturns(api.InstallationsServiceOutput{}, errors.New("error"))
				service.RunningInstallationReturns(api.InstallationsServiceOutput{}, nil)

				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)
				err := executeCommand(command, []string{"--product-name", "product1", "--product-name", "product2"})
				Expect(err).To(HaveOccurred())

				_, _, _, productNames, _ := service.CreateInstallationArgsForCall(0)
				Expect(productNames).To(ConsistOf("product1", "product2"))
			})
		})

		When("passed the reattach flag", func() {
			It("re-attaches to an ongoing installation", func() {
				installationStartedAt := time.Date(2017, time.February, 25, 02, 31, 1, 0, time.UTC)

				service.RunningInstallationReturns(api.InstallationsServiceOutput{
					ID:        200,
					Status:    "running",
					StartedAt: &installationStartedAt,
				}, nil)

				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

				err := executeCommand(command, []string{"--reattach"})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.CreateInstallationCallCount()).To(Equal(0))

				Expect(stderr).To(gbytes.Say(regexp.QuoteMeta("found already running installation... re-attaching (Installation ID: 200, Started: Sat Feb 25 02:31:01 UTC 2017)")))
				Expect(stderr).To(gbytes.Say(regexp.QuoteMeta("found already running installation... re-attaching (Installation ID: 200, Started: Sat Feb 25 02:31:01 UTC 2017)")))

				Expect(service.GetInstallationArgsForCall(0)).To(Equal(200))
				Expect(service.GetInstallationLogsArgsForCall(0)).To(Equal(200))
			})

			When("the recreate-vms flag is also passed", func() {
				It("errors because this is a conflict", func() {
					command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

					err := executeCommand(command, []string{"--reattach", "--recreate-vms"})
					Expect(err).To(MatchError(ContainSubstring("--recreate-vms cannot be used with --reattach because it requires the ability to update a director property")))
				})
			})
		})

		When("not passed reattach", func() {
			It("errors of an already running installation", func() {
				installationStartedAt := time.Date(2017, time.February, 25, 02, 31, 1, 0, time.UTC)

				service.RunningInstallationReturns(api.InstallationsServiceOutput{
					ID:        200,
					Status:    "running",
					StartedAt: &installationStartedAt,
				}, nil)

				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

				err := executeCommand(command, []string{})
				Expect(err).To(HaveOccurred())

				Expect(service.CreateInstallationCallCount()).To(Equal(0))

				Expect(stderr).To(gbytes.Say(regexp.QuoteMeta("found already running installation... not re-attaching (Installation ID: 200, Started: Sat Feb 25 02:31:01 UTC 2017)")))
			})
		})

		When("passed the recreate-vms", func() {
			It("ensures all vms are recreated", func() {
				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

				err := executeCommand(command, []string{"--recreate-vms"})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.UpdateStagedDirectorPropertiesCallCount()).To(Equal(1))
				Expect(string(service.UpdateStagedDirectorPropertiesArgsForCall(0))).To(MatchJSON(`
					{
						"director_configuration": {
							"bosh_recreate_on_next_deploy": true,
							"bosh_director_recreate_on_next_deploy": true
						}
					}
				`))

				Expect(stderr.Contents()).To(ContainSubstring("setting director to recreate all vms (available in Ops Manager 2.9+)"))
			})

			It("ensures only the director is recreated", func() {
				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

				err := executeCommand(command, []string{
					"--recreate-vms",
					"--skip-deploy-products",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.UpdateStagedDirectorPropertiesCallCount()).To(Equal(1))
				Expect(string(service.UpdateStagedDirectorPropertiesArgsForCall(0))).To(MatchJSON(`
					{
						"director_configuration": {
							"bosh_director_recreate_on_next_deploy": true
						}
					}
				`))

				Expect(stderr).To(gbytes.Say(`setting director to recreate director vm \(available in Ops Manager 2\.9\+\)`))
			})

			It("ensures only products are updated", func() {
				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

				err := executeCommand(command, []string{
					"--recreate-vms",
					"--product-name", "cf",
					"-n", "example-product",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.UpdateStagedDirectorPropertiesCallCount()).To(Equal(1))
				Expect(string(service.UpdateStagedDirectorPropertiesArgsForCall(0))).To(MatchJSON(`
					{
						"director_configuration": {
							"bosh_recreate_on_next_deploy": true
						}
					}
				`))

				Expect(stderr).To(gbytes.Say("setting director to recreate all VMs for the following products:"))
				Expect(stderr).To(gbytes.Say("- cf"))
				Expect(stderr).To(gbytes.Say("- example-product"))
				Expect(stderr).To(gbytes.Say("this will also recreate the director vm if there are changes"))
			})

			When("on a version less than 2.9", func() {
				It("ensures only products are updated", func() {
					service.InfoReturns(api.Info{Version: "2.6.0"}, nil)

					command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

					err := executeCommand(command, []string{
						"--recreate-vms",
						"--product-name", "cf",
						"-n", "example-product",
					})
					Expect(err).ToNot(HaveOccurred())

					Expect(service.UpdateStagedDirectorPropertiesCallCount()).To(Equal(1))
					Expect(string(service.UpdateStagedDirectorPropertiesArgsForCall(0))).To(MatchJSON(`
						{
							"director_configuration": {
								"bosh_recreate_on_next_deploy": true
							}
						}
					`))

					Expect(stderr).To(gbytes.Say("setting director to recreate all VMs for the following products:"))
					Expect(stderr).To(gbytes.Say("- cf"))
					Expect(stderr).To(gbytes.Say("- example-product"))
					Expect(stderr).To(gbytes.Say("this will also recreate the director vm if there are changes"))
				})
			})

			When("the service returns an error", func() {
				It("displays that error message", func() {
					service.UpdateStagedDirectorPropertiesReturns(errors.New("testing"))
					command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

					err := executeCommand(command, []string{"--recreate-vms"})
					Expect(err).To(MatchError(ContainSubstring("testing")))
				})
			})
		})

		Context("Load config file", func() {
			var fileName string

			Context("given a valid config file", func() {
				BeforeEach(func() {
					fh, err := os.CreateTemp("", "")
					defer func() { _ = fh.Close() }()
					Expect(err).ToNot(HaveOccurred())
					_, err = fh.WriteString(`
---
errands:
  product1_name:
    run_post_deploy:
      errand_c: "default"
    run_pre_delete:
      errand_a: true
      errand_b: false
  product2_name:
    run_post_deploy:
      errand_a: false
    run_pre_delete:
      errand_b: "default"
`)

					Expect(err).ToNot(HaveOccurred())
					fileName = fh.Name()
				})

				It("parses the config file correctly", func() {
					fh, err := os.Open(fileName)
					Expect(err).ToNot(HaveOccurred())
					defer fh.Close()

					applyErrandChanges := api.ApplyErrandChanges{}
					err = yaml.NewDecoder(fh).Decode(&applyErrandChanges)
					Expect(err).ToNot(HaveOccurred())

					Expect(applyErrandChanges).To(Equal(api.ApplyErrandChanges{
						Errands: map[string]api.ProductErrand{
							"product1_name": {
								RunPostDeploy: map[string]interface{}{
									"errand_c": "default",
								},
								RunPreDelete: map[string]interface{}{
									"errand_a": true,
									"errand_b": false,
								},
							},
							"product2_name": {
								RunPostDeploy: map[string]interface{}{
									"errand_a": false,
								},
								RunPreDelete: map[string]interface{}{
									"errand_b": "default",
								},
							},
						}}))
				})

				It("calls the api with correct arguments", func() {
					command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

					err := executeCommand(command, []string{"--config", fileName})
					Expect(err).ToNot(HaveOccurred())

					Expect(service.CreateInstallationCallCount()).To(Equal(1))

					ignoreWarnings, deployProducts, forceLatestVariables, _, errands := service.CreateInstallationArgsForCall(0)
					Expect(ignoreWarnings).To(Equal(false))
					Expect(deployProducts).To(Equal(true))
					Expect(forceLatestVariables).To(Equal(false))
					Expect(errands).To(Equal(api.ApplyErrandChanges{
						Errands: map[string]api.ProductErrand{
							"product1_name": {
								RunPostDeploy: map[string]interface{}{
									"errand_c": "default",
								},
								RunPreDelete: map[string]interface{}{
									"errand_a": true,
									"errand_b": false,
								},
							},
							"product2_name": {
								RunPostDeploy: map[string]interface{}{
									"errand_a": false,
								},
								RunPreDelete: map[string]interface{}{
									"errand_b": "default",
								},
							},
						}}))

					Expect(stderr).To(gbytes.Say("attempting to apply changes to the targeted Ops Manager"))

					Expect(service.GetInstallationArgsForCall(0)).To(Equal(311))
					Expect(service.GetInstallationCallCount()).To(Equal(3))

					Expect(service.GetInstallationLogsArgsForCall(0)).To(Equal(311))
					Expect(service.GetInstallationLogsCallCount()).To(Equal(3))

					Expect(writer.FlushCallCount()).To(Equal(3))
					Expect(writer.FlushArgsForCall(0)).To(Equal("start of logs"))
					Expect(writer.FlushArgsForCall(1)).To(Equal("these logs"))
					Expect(writer.FlushArgsForCall(2)).To(Equal("some other logs"))
				})
			})

			Context("given a file that does not exist", func() {
				It("returns an error", func() {
					command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

					err := executeCommand(command, []string{"--config", "filedoesnotexist"})
					Expect(err).To(MatchError("could not load config: open filedoesnotexist: no such file or directory"))
				})
			})

			Context("given a invalid yaml file", func() {
				BeforeEach(func() {
					fh, err := os.CreateTemp("", "")
					defer func() { _ = fh.Close() }()
					Expect(err).ToNot(HaveOccurred())
					_, err = fh.WriteString(`
---
errands: lolololol
`)

					Expect(err).ToNot(HaveOccurred())
					fileName = fh.Name()
				})

				It("returns an error", func() {
					command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

					err := executeCommand(command, []string{"--config", fileName})
					Expect(err).To(MatchError(ContainSubstring("line 3: cannot unmarshal !!str `lolololol`")))
				})
			})
		})

		It("handles a failed installation", func() {
			service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)
			service.GetInstallationReturnsOnCall(0, api.InstallationsServiceOutput{Status: "failed"}, nil)
			service.GetInstallationLogsReturnsOnCall(0, api.InstallationsServiceOutput{Logs: "start of logs"}, nil)

			command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

			err := executeCommand(command, []string{})
			Expect(err).To(MatchError("installation was unsuccessful"))
		})

		When("checking for an already running installation returns an error", func() {
			It("returns an error", func() {
				service.RunningInstallationReturns(api.InstallationsServiceOutput{}, errors.New("some error"))

				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

				err := executeCommand(command, []string{})
				Expect(err).To(MatchError("could not check for any already running installation: some error"))
			})
		})

		When("--product-name is used with an old version of ops manager", func() {
			It("returns an error", func() {
				versions := []string{"2.1-build.326", "1.12-build99"}
				for _, version := range versions {
					service.InfoReturns(api.Info{Version: version}, nil)

					command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)
					err := executeCommand(command, []string{"--product-name", "p-mysql"})
					Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("--product-name is only available with Ops Manager 2.2 or later: you are running %s", version)))
				}
			})
		})

		When("an installation cannot be triggered", func() {
			It("returns an error", func() {
				service.CreateInstallationReturns(api.InstallationsServiceOutput{}, errors.New("some error"))

				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

				err := executeCommand(command, []string{})
				Expect(err).To(MatchError("installation failed to trigger: some error"))
			})
		})

		When("getting the installation status has an error", func() {
			It("returns an error", func() {
				service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)
				service.GetInstallationReturnsOnCall(0, api.InstallationsServiceOutput{}, errors.New("first error"))
				service.GetInstallationReturnsOnCall(1, api.InstallationsServiceOutput{}, errors.New("second error"))
				service.GetInstallationReturnsOnCall(2, api.InstallationsServiceOutput{}, errors.New("third error"))

				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

				err := executeCommand(command, []string{})
				Expect(err).To(MatchError("installation failed to get status after 3 attempts: third error"))
			})
		})

		When("there is an error fetching the logs", func() {
			It("returns an error", func() {
				service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)
				service.GetInstallationReturnsOnCall(0, api.InstallationsServiceOutput{Status: "running"}, nil)
				service.GetInstallationLogsReturnsOnCall(0, api.InstallationsServiceOutput{}, errors.New("no"))

				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

				err := executeCommand(command, []string{})
				Expect(err).To(MatchError("installation failed to get logs: no"))
			})
		})

		When("there is an error flushing the logs", func() {
			It("returns an error", func() {
				service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)
				service.GetInstallationReturnsOnCall(0, api.InstallationsServiceOutput{Status: "running"}, nil)
				service.GetInstallationLogsReturnsOnCall(0, api.InstallationsServiceOutput{Logs: "some logs"}, nil)

				writer.FlushReturns(errors.New("yes"))

				command := commands.NewApplyChanges(service, pendingService, writer, logger, 1)

				err := executeCommand(command, []string{})
				Expect(err).To(MatchError("installation failed to flush logs: yes"))
			})
		})
	})
})
