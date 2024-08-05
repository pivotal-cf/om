package vmlifecyclecommands_test

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-cf/om/vmlifecycle/vmlifecyclecommands"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf/om/vmlifecycle/runner"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers/fakes"
)

var _ = Describe("upgradeOpsman", func() {
	When("we retrieve the version info", func() {
		Context("error does not come from bad credentials", func() {
			It("runs the create and import commands", func() {
				server := ghttp.NewServer()
				server.SetAllowUnhandledRequests(false)
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/unlock"),
						ghttp.RespondWith(200, "{}"),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/login/ensure_availability"),
						ghttp.RespondWith(302, "", map[string][]string{
							"Location": []string{
								"https://example.com/auth/cloudfoundry",
							},
						}),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/uaa/oauth/token"),
						ghttp.RespondWith(404, "not found"),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/uaa/oauth/token"),
						ghttp.RespondWith(404, "not found"),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/uaa/oauth/token"),
						ghttp.RespondWith(404, "not found"),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/login/ensure_availability"),
						ghttp.RespondWith(302, "", map[string][]string{
							"Location": {"/auth/cloudfoundry"},
						}),
					),
				)
				defer server.Close()

				command, createService, deleteService, _, _ := createUpgradeOpsmanCommand()
				command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
					"target":                server.URL(),
					"request-timeout":       1,
					"connect-timeout":       1,
					"skip-ssl-validation":   true,
					"decryption-passphrase": "decryption-passphrase",
				}))
				err := command.Execute([]string{})
				Expect(err).ToNot(HaveOccurred())
				Expect(createService.CreateVMCallCount()).To(Equal(1))
				Expect(deleteService.DeleteVMCallCount()).To(Equal(0))
			})

			Context("error on the invocation of `om`", func() {
				var server *ghttp.Server

				BeforeEach(func() {
					server = ghttp.NewServer()
					server.SetAllowUnhandledRequests(false)
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/api/v0/unlock"),
							ghttp.RespondWith(200, "{}"),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/login/ensure_availability"),
							ghttp.RespondWith(302, "", map[string][]string{
								"Location": []string{
									"https://example.com/auth/cloudfoundry",
								},
							}),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/uaa/oauth/token"),
							ghttp.RespondWith(500, ""),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/uaa/oauth/token"),
							ghttp.RespondWith(500, ""),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/uaa/oauth/token"),
							ghttp.RespondWith(500, ""),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/login/ensure_availability"),
							ghttp.RespondWith(302, "", map[string][]string{
								"Location": {"/auth/cloudfoundry"},
							}),
						),
					)
				})

				AfterEach(func() {
					server.Close()
				})

				It("runs the create and import commands", func() {
					command, createService, deleteService, _, _ := createUpgradeOpsmanCommand()
					command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
						"target":                server.URL(),
						"request-timeout":       1,
						"connect-timeout":       1,
						"skip-ssl-validation":   true,
						"decryption-passphrase": "decryption-passphrase",
					}))
					err := command.Execute([]string{})

					Expect(err).ToNot(HaveOccurred())
					Expect(createService.CreateVMCallCount()).To(Equal(1))
					Expect(deleteService.DeleteVMCallCount()).To(Equal(0))
				})
			})
		})

		Context("the credentials are wrong", func() {
			var server *ghttp.Server

			BeforeEach(func() {
				server = ghttp.NewServer()
				server.SetAllowUnhandledRequests(false)
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/unlock"),
						ghttp.RespondWith(200, "{}"),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/login/ensure_availability"),
						ghttp.RespondWith(302, "", map[string][]string{
							"Location": []string{
								"https://example.com/auth/cloudfoundry",
							},
						}),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/uaa/oauth/token"),
						ghttp.RespondWith(401, "Bad credentials"),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/uaa/oauth/token"),
						ghttp.RespondWith(401, "Bad credentials"),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/uaa/oauth/token"),
						ghttp.RespondWith(401, "Bad credentials"),
					),
				)
			})

			AfterEach(func() {
				server.Close()
			})

			It("returns an error", func() {
				command, createService, deleteService, _, _ := createUpgradeOpsmanCommand()

				command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
					"target":                server.URL(),
					"username":              "username",
					"password":              "password",
					"request-timeout":       1,
					"connect-timeout":       1,
					"skip-ssl-validation":   true,
					"decryption-passphrase": "decryption-passphrase",
				}))
				err := command.Execute([]string{})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not authenticate with Ops Manager"))
				Expect(createService.CreateVMCallCount()).To(Equal(0))
				Expect(deleteService.DeleteVMCallCount()).To(Equal(0))
			})
		})

		When("opsman login is successful", func() {
			var server *ghttp.Server
			ensureAvailabilityHandler := ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/login/ensure_availability"),
				ghttp.RespondWith(302, "", map[string][]string{
					"Location": {"/auth/cloudfoundry"},
				}),
			)

			BeforeEach(func() {
				server = ghttp.NewServer()
				token := map[string]interface{}{
					"access_token":  "some-random-acceasdfasdfss-token",
					"refresh_token": "some-random-refresh-token",
					"token_type":    "Bearer",
					"expires_in":    3600,
					"scope":         "email address",
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/unlock"),
						ghttp.RespondWith(200, "{}"),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/login/ensure_availability"),
						ghttp.RespondWith(302, "", map[string][]string{
							"Location": []string{
								"https://example.com/auth/cloudfoundry",
							},
						}),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/uaa/oauth/token"),
						ghttp.RespondWithJSONEncoded(200, token),
					),
				)
			})

			AfterEach(func() {
				server.Close()
			})

			When("the versions are the same", func() {
				When("the --Recreate flag is not passed", func() {
					for _, fileNameFixture := range []string{"OpsManager%sonGCP.yml", "[ops-manager,2.2.3]ops-manager-vsphere-%s.ova"} {
						DescribeTable("does not continue the upgrade process and the filename is "+fileNameFixture, func(currentVersion string, newerVersion string) {
							info := map[string]interface{}{
								"info": map[string]interface{}{
									"version": currentVersion,
								},
							}
							server.AppendHandlers(ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/api/v0/info"),
								ghttp.RespondWithJSONEncoded(200, info),
							), ensureAvailabilityHandler)

							command, createService, deleteService, _, _ := createUpgradeOpsmanCommand()

							command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
								"target":                server.URL(),
								"request-timeout":       1,
								"connect-timeout":       1,
								"skip-ssl-validation":   true,
								"decryption-passphrase": "decryption-passphrase",
							}))

							fh, err := os.CreateTemp("", fmt.Sprintf(fileNameFixture, newerVersion))
							Expect(err).ToNot(HaveOccurred())
							Expect(fh.Close()).ToNot(HaveOccurred())

							command.CreateVM.ImageFile = fh.Name()
							err = command.Execute([]string{})

							Expect(err).ToNot(HaveOccurred())
							Expect(createService.CreateVMCallCount()).To(Equal(0))
							Expect(deleteService.DeleteVMCallCount()).To(Equal(0))
						},
							// Same formats
							Entry("semver", "2.10.4", "2.10.4"),
							Entry("semver with build", "2.6.1-build.92", "2.6.1-build.92"),
							Entry("build", "2.3-build.422", "2.3-build.422"),

							// Different formats
							Entry("semver vs semver with build", "2.10.4", "2.10.4-build-75"),
							Entry("semver vs build", "2.10.4", "2.10-build.4"),
							Entry("build vs semver with build", "2.6-build.3", "2.6.3-build.642"),

							// Different formats reversed
							Entry("semver with build vs semver", "2.10.4-build-75", "2.10.4"),
							Entry("build vs semver", "2.10-build.4", "2.10.4"),
							Entry("semver with build vs build", "2.6.3-build.642", "2.6-build.3"),
						)
					}
				})

				When("the --recreate flag is passed", func() {
					for _, fileNameFixture := range []string{"OpsManager%sonGCP.yml", "[ops-manager,2.2.3]ops-manager-vsphere-%s.ova"} {
						DescribeTable("runs the delete, create, and import commands with the filename, and the filename is "+fileNameFixture, func(currentVersion string, newerVersion string) {
							info := map[string]interface{}{
								"info": map[string]interface{}{
									"version": currentVersion,
								},
							}
							server.AppendHandlers(ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/api/v0/info"),
								ghttp.RespondWithJSONEncoded(200, info),
							), ensureAvailabilityHandler)

							command, createService, deleteService, _, _ := createUpgradeOpsmanCommand()
							command.Recreate = true
							command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
								"target":                server.URL(),
								"request-timeout":       1,
								"connect-timeout":       1,
								"skip-ssl-validation":   true,
								"decryption-passphrase": "decryption-passphrase",
							}))

							fh, err := os.CreateTemp("", fmt.Sprintf(fileNameFixture, newerVersion))
							Expect(err).ToNot(HaveOccurred())
							Expect(fh.Close()).ToNot(HaveOccurred())

							command.CreateVM.ImageFile = fh.Name()
							err = command.Execute([]string{})

							Expect(err).ToNot(HaveOccurred())
							Expect(deleteService.DeleteVMCallCount()).To(Equal(1))
							Expect(createService.CreateVMCallCount()).To(Equal(1))
						},
							// Same formats
							Entry("semver", "2.10.4", "2.10.4"),
							Entry("semver with build", "2.6.1-build.92", "2.6.1-build.92"),
							Entry("build", "2.3-build.422", "2.3-build.422"),

							// Different formats
							Entry("semver vs semver with build", "2.10.4", "2.10.4-build-75"),
							Entry("semver vs build", "2.10.4", "2.10-build.4"),
							Entry("build vs semver with build", "2.6-build.3", "2.6.3-build.642"),

							// Different formats reversed
							Entry("semver with build vs semver", "2.10.4-build-75", "2.10.4"),
							Entry("build vs semver", "2.10-build.4", "2.10.4"),
							Entry("semver with build vs build", "2.6.3-build.642", "2.6-build.3"),
						)
					}
				})
			})

			When("the versions are different", func() {
				When("the version to install is smaller than installed version", func() {
					for _, fileNameFixture := range []string{"OpsManager%sonGCP.yml", "[ops-manager,2.2.3]ops-manager-vsphere-%s.ova"} {
						DescribeTable("returns an error on the filename of "+fileNameFixture, func(currentVersion string, newerVersion string) {
							info := map[string]interface{}{
								"info": map[string]interface{}{
									"version": currentVersion,
								},
							}
							server.AppendHandlers(ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/api/v0/info"),
								ghttp.RespondWithJSONEncoded(200, info),
							))

							server.AppendHandlers(ensureAvailabilityHandler)
							command, _, _, _, _ := createUpgradeOpsmanCommand()

							command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
								"target":                server.URL(),
								"request-timeout":       1,
								"connect-timeout":       1,
								"skip-ssl-validation":   true,
								"decryption-passphrase": "decryption-passphrase",
							}))

							fh, err := os.CreateTemp("", fmt.Sprintf(fileNameFixture, newerVersion))
							Expect(err).ToNot(HaveOccurred())
							Expect(fh.Close()).ToNot(HaveOccurred())

							command.CreateVM.ImageFile = fh.Name()
							err = command.Execute([]string{})

							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("downgrading is not supported by Ops Manager"))
						},
							// Same formats
							Entry("semver, via patch", "2.5.3", "2.5.1"),
							Entry("semver, via minor", "2.5.0", "2.4.3"),
							Entry("semver, via major", "3.5.3", "2.5.3"),
							Entry("semver with build, via patch", "2.5.3-build.0", "2.5.1-build.0"),
							Entry("semver with build, via minor", "2.5.0-build.0", "2.4.3-build.0"),
							Entry("semver with build, via major", "3.5.3-build.0", "2.5.3-build.0"),
							Entry("build, via patch", "2.4-build.193", "2.4-build.100"),
							Entry("build, via minor", "2.4-build.193", "2.3-build.193"),
							Entry("build, via major", "2.4-build.193", "1.4-build.193"),

							// Semver vs. semver with build
							Entry("semver vs semver with build, via patch", "2.5.3", "2.5.2-build.0"),
							Entry("semver vs semver with build, via minor", "2.5.3", "2.4.3-build.0"),
							Entry("semver vs semver with build, via major", "2.5.3", "1.5.3-build.0"),
							Entry("semver with build vs semver, via patch", "2.5.3-build.0", "2.5.2"),
							Entry("semver with build vs semver, via minor", "2.5.3-build.0", "2.4.3"),
							Entry("semver with build vs semver, via major", "2.5.3-build.0", "1.5.3"),

							// Semver vs. build
							Entry("semver vs build, via patch", "2.5.1", "2.5-build.0"),
							Entry("semver vs build, via minor", "2.5.1", "2.4-build.1"),
							Entry("semver vs build, via major", "2.5.1", "1.5-build.1"),
							Entry("build vs semver, via patch", "2.5-build.1", "2.5.0"),
							Entry("build vs semver, via minor", "2.5-build.1", "2.4.1"),
							Entry("build vs semver, via major", "2.5-build.1", "1.5.1"),

							// Semver with build vs build
							Entry("semver with build vs build, via patch", "2.5.1-build.103", "2.5-build.0"),
							Entry("semver with build vs build, via minor", "2.5.1-build.103", "2.4-build.1"),
							Entry("semver with build vs build, via major", "2.5.1-build.103", "1.5-build.1"),
							Entry("build vs semver with build, via patch", "2.5-build.1", "2.5.0-build.103"),
							Entry("build vs semver with build, via minor", "2.5-build.1", "2.4.1-build.103"),
							Entry("build vs semver with build, via major", "2.5-build.1", "1.5.1-build.103"),
						)
					}
				})

				When("upgrading case", func() {
					for _, fileNameFixture := range []string{"OpsManager%sonGCP.yml", "[ops-manager,2.2.3]ops-manager-vsphere-%s.ova"} {
						DescribeTable("runs the delete, create, and import commands with the filename "+fileNameFixture, func(currentVersion string, newerVersion string) {
							info := map[string]interface{}{
								"info": map[string]interface{}{
									"version": currentVersion,
								},
							}
							server.AppendHandlers(ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/api/v0/info"),
								ghttp.RespondWithJSONEncoded(200, info),
							))

							server.AppendHandlers(ensureAvailabilityHandler)
							command, createService, deleteService, _, stderr := createUpgradeOpsmanCommand()

							command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
								"target":                server.URL(),
								"request-timeout":       1,
								"connect-timeout":       1,
								"skip-ssl-validation":   true,
								"decryption-passphrase": "decryption-passphrase",
							}))

							image, err := os.CreateTemp("", fmt.Sprintf(fileNameFixture, newerVersion))
							Expect(err).ToNot(HaveOccurred())
							command.CreateVM.ImageFile = image.Name()
							err = command.Execute([]string{})

							Expect(err).ToNot(HaveOccurred())
							Expect(deleteService.DeleteVMCallCount()).To(Equal(1))
							Expect(createService.CreateVMCallCount()).To(Equal(1))

							Eventually(stderr).Should(gbytes.Say(`--skip-ssl-validation curl`))
							Eventually(stderr).Should(gbytes.Say(`--skip-ssl-validation import-installation`))
						},
							// Same formats
							Entry("semver, via patch", "2.5.1", "2.5.3"),
							Entry("semver, via minor", "2.4.3", "2.5.0"),
							Entry("semver, via major", "2.5.3", "3.5.3"),
							Entry("semver with build, via patch", "2.5.1-build.0", "2.5.3-build.0"),
							Entry("semver with build, via minor", "2.4.3-build.0", "2.5.0-build.0"),
							Entry("semver with build, via major", "2.5.3-build.0", "3.5.3-build.0"),
							Entry("build, via patch", "2.4-build.100", "2.4-build.193"),
							Entry("build, via minor", "2.3-build.193", "2.4-build.193"),
							Entry("build, via major", "1.4-build.193", "2.4-build.193"),

							// Semver vs. semver with build
							Entry("semver vs semver with build, via patch", "2.5.2-build.0", "2.5.3"),
							Entry("semver vs semver with build, via minor", "2.4.3-build.0", "2.5.3"),
							Entry("semver vs semver with build, via major", "1.5.3-build.0", "2.5.3"),
							Entry("semver with build vs semver, via patch", "2.5.2", "2.5.3-build.0"),
							Entry("semver with build vs semver, via minor", "2.4.3", "2.5.3-build.0"),
							Entry("semver with build vs semver, via major", "1.5.3", "2.5.3-build.0"),

							// Semver vs. build
							Entry("semver vs build, via patch", "2.5-build.0", "2.5.1"),
							Entry("semver vs build, via minor", "2.4-build.1", "2.5.1"),
							Entry("semver vs build, via major", "1.5-build.1", "2.5.1"),
							Entry("build vs semver, via patch", "2.5.0", "2.5-build.1"),
							Entry("build vs semver, via minor", "2.4.1", "2.5-build.1"),
							Entry("build vs semver, via major", "1.5.1", "2.5-build.1"),

							// Semver with build vs build
							Entry("semver with build vs build, via patch", "2.5-build.0", "2.5.1-build.103"),
							Entry("semver with build vs build, via minor", "2.4-build.1", "2.5.1-build.103"),
							Entry("semver with build vs build, via major", "1.5-build.1", "2.5.1-build.103"),
							Entry("build vs semver with build, via patch", "2.5.0-build.103", "2.5-build.1"),
							Entry("build vs semver with build, via minor", "2.4.1-build.103", "2.5-build.1"),
							Entry("build vs semver with build, via major", "1.5.1-build.103", "2.5-build.1"),
						)
					}

					It("polls till the provided timeout then errors", func(ctx context.Context) {
						info := map[string]interface{}{
							"info": map[string]interface{}{
								"version": "2.10-build.0",
							},
						}
						server.AppendHandlers(ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/info"),
							ghttp.RespondWithJSONEncoded(200, info),
						))

						server.AppendHandlers(ghttp.VerifyRequest("GET", "/login/ensure_availability"))
						server.AppendHandlers(ghttp.VerifyRequest("GET", "/login/ensure_availability"))
						server.AppendHandlers(ghttp.VerifyRequest("GET", "/login/ensure_availability"))
						server.AppendHandlers(ghttp.VerifyRequest("GET", "/login/ensure_availability"))
						command, _, _, _, _ := createUpgradeOpsmanCommand()

						command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
							"target":                server.URL(),
							"request-timeout":       1,
							"connect-timeout":       1,
							"skip-ssl-validation":   true,
							"decryption-passphrase": "decryption-passphrase",
						}))
						image, err := os.CreateTemp("", "OpsManager2.10-build.296onGCP.yml")

						Expect(err).ToNot(HaveOccurred())
						command.CreateVM.ImageFile = image.Name()

						err = command.Execute([]string{})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("exceeded 20ms"))
						//<-ctx.Done()
					}, NodeTimeout(4*time.Minute))

					When("environment variables are provided but not in import installation env file", func() {
						BeforeEach(func() {
							info := map[string]interface{}{
								"info": map[string]interface{}{
									"version": "2.0-build.0",
								},
							}
							server.AppendHandlers(ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/api/v0/info"),
								ghttp.RespondWithJSONEncoded(200, info),
							))
							server.AppendHandlers(ensureAvailabilityHandler)
						})

						It("does not error OM_TARGET is set", func() {
							command, _, _, _, _ := createUpgradeOpsmanCommand()
							err := os.Setenv("OM_TARGET", server.URL())
							Expect(err).ToNot(HaveOccurred())
							defer os.Unsetenv("OM_TARGET")

							command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
								"request-timeout":       1,
								"connect-timeout":       1,
								"skip-ssl-validation":   true,
								"decryption-passphrase": "passphrase",
							}))

							fh, err := os.CreateTemp("", "OpsManager2.2-build.296onGCP.yml")
							Expect(err).ToNot(HaveOccurred())
							Expect(fh.Close()).ToNot(HaveOccurred())

							command.CreateVM.ImageFile = fh.Name()
							err = command.Execute([]string{})

							Expect(err).ToNot(HaveOccurred())
						})

						It("does not error when OM_DECRYPTION_PASSPHRASE is set", func() {
							command, _, _, _, _ := createUpgradeOpsmanCommand()
							err := os.Setenv("OM_DECRYPTION_PASSPHRASE", "passphrase")
							Expect(err).ToNot(HaveOccurred())
							defer os.Unsetenv("OM_DECRYPTION_PASSPHRASE")

							command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
								"target":              server.URL(),
								"request-timeout":     1,
								"connect-timeout":     1,
								"skip-ssl-validation": true,
							}))

							fh, err := os.CreateTemp("", "OpsManager2.2-build.296onGCP.yml")
							Expect(err).ToNot(HaveOccurred())
							Expect(fh.Close()).ToNot(HaveOccurred())

							command.CreateVM.ImageFile = fh.Name()
							err = command.Execute([]string{})

							Expect(err).ToNot(HaveOccurred())
						})
					})

					Describe("interpolation", func() {
						validConfig := `---
opsman-configuration:
  gcp:
    gcp_service_account: something
    project: ((project_name))
    region: us-west1
    zone: us-west1-c
    vpc_subnet: infra
    tags: good
`
						var command *vmlifecyclecommands.UpgradeOpsman
						BeforeEach(func() {
							info := map[string]interface{}{
								"info": map[string]interface{}{
									"version": "2.2.2",
								},
							}
							server.AppendHandlers(ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/api/v0/info"),
								ghttp.RespondWithJSONEncoded(200, info),
							))

							server.AppendHandlers(ensureAvailabilityHandler)
							command, _, _, _, _ = createUpgradeOpsmanCommand()

							command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
								"target":                server.URL(),
								"request-timeout":       1,
								"connect-timeout":       1,
								"skip-ssl-validation":   true,
								"decryption-passphrase": "decryption-passphrase",
							}))

							image, err := os.CreateTemp("", fmt.Sprintf("OpsManager%sonGCP.yml", "2.2.3"))
							Expect(err).ToNot(HaveOccurred())
							command.CreateVM.ImageFile = image.Name()
						})

						It("can interpolate variables into the configuration", func() {
							validVars := `---
project_name: awesome-project
`
							command.CreateVM.VarsFile = []string{writeFile(validVars)}
							command.CreateVM.Config = writeFile(validConfig)

							err := command.Execute([]string{})
							Expect(err).ToNot(HaveOccurred())
						})

						It("returns an error of missing variables", func() {
							command.CreateVM.VarsFile = []string{writeFile(``)}
							command.CreateVM.Config = writeFile(validConfig)

							err := command.Execute([]string{})
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Expected to find variables"))
						})

						It("can interpolate variables from environment variables", func() {
							Expect(os.Setenv("OM_VAR_project_name", "awesome-project")).ToNot(HaveOccurred())
							defer func() {
								err := os.Unsetenv("OM_VAR_project_name")
								Expect(err).ToNot(HaveOccurred())
							}()
							command.CreateVM.VarsEnv = []string{"OM_VAR"}
							command.CreateVM.Config = writeFile(validConfig)
							command.DeleteVM.VarsEnv = []string{"OM_VAR"}

							err := command.Execute([]string{})
							Expect(err).ToNot(HaveOccurred())
						})
					})
				})
			})
		})

		When("it can't unmarshall the Info struct", func() {
			var server *ghttp.Server

			BeforeEach(func() {
				server = ghttp.NewServer()
				server.SetAllowUnhandledRequests(false)

				token := map[string]interface{}{
					"access_token":  "some-random-acceasdfasdfss-token",
					"refresh_token": "some-random-refresh-token",
					"token_type":    "Bearer",
					"expires_in":    3600,
					"scope":         "email address",
				}

				info := map[string]interface{}{
					"not-valid": map[string]interface{}{
						"version": "2.2-build.296",
					},
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/unlock"),
						ghttp.RespondWith(200, "{}"),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/login/ensure_availability"),
						ghttp.RespondWith(302, "", map[string][]string{
							"Location": []string{
								"https://example.com/auth/cloudfoundry",
							},
						}),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/uaa/oauth/token"),
						ghttp.RespondWithJSONEncoded(200, token),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/info"),
						ghttp.RespondWithJSONEncoded(200, info),
					),
				)
			})

			AfterEach(func() {
				server.Close()
			})

			It("returns an error", func() {
				command, _, _, _, _ := createUpgradeOpsmanCommand()

				command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
					"target":                server.URL(),
					"username":              "username",
					"password":              "password",
					"request-timeout":       1,
					"connect-timeout":       1,
					"skip-ssl-validation":   true,
					"decryption-passphrase": "decryption-passphrase",
				}))

				fh, err := os.CreateTemp("", "OpsManager2.2-build.296onGCP.yml")
				Expect(err).ToNot(HaveOccurred())
				Expect(fh.Close()).ToNot(HaveOccurred())

				command.CreateVM.ImageFile = fh.Name()
				err = command.Execute([]string{})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("info struct could not be parsed"))
			})
		})

		It("exits with error when the target is not supplied in the env file", func() {
			command, _, _, _, _ := createUpgradeOpsmanCommand()

			command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
				"request-timeout":     1,
				"connect-timeout":     1,
				"skip-ssl-validation": true,
			}))

			fh, err := os.CreateTemp("", "OpsManager2.2-build.296onGCP.yml")
			Expect(err).ToNot(HaveOccurred())
			Expect(fh.Close()).ToNot(HaveOccurred())

			command.CreateVM.ImageFile = fh.Name()
			err = command.Execute([]string{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("target is a required field in the env configuration"))
		})

		It("errors when the decryption-passphrase is not supplied", func() {
			command, _, _, _, _ := createUpgradeOpsmanCommand()
			command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
				"target":              "https://example.com",
				"request-timeout":     1,
				"connect-timeout":     1,
				"skip-ssl-validation": true,
			}))

			fh, err := os.CreateTemp("", "OpsManager2.2-build.296onGCP.yml")
			Expect(err).ToNot(HaveOccurred())
			Expect(fh.Close()).ToNot(HaveOccurred())

			command.CreateVM.ImageFile = fh.Name()
			err = command.Execute([]string{})

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("decryption-passphrase is a required field in the env configuration"))
		})

		It("gracefully errors when the filename does not have a matching version number", func() {
			command, _, _, _, _ := createUpgradeOpsmanCommand()

			command.ImportInstallation.EnvFile = writeFile(createJsonString(map[string]interface{}{
				"target":                "https://example.com",
				"request-timeout":       1,
				"connect-timeout":       1,
				"skip-ssl-validation":   true,
				"decryption-passphrase": "passphrase",
			}))

			fh, err := os.CreateTemp("", "OpsManager2.2.yml")
			Expect(err).ToNot(HaveOccurred())
			Expect(fh.Close()).ToNot(HaveOccurred())

			command.CreateVM.ImageFile = fh.Name()
			err = command.Execute([]string{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(`the file name '.*' for the Ops Manager image needs to contain the original version number as downloaded from Pivnet \(ie 'OpsManager2.2-build.296onGCP.yml'\)`))
		})
	})

	Describe("validate the inputs to ensure we fail fast", func() {
		var (
			configContent       string
			envContent          string
			imageContent        string
			installationContent string
			stateContent        string
			varsContent         string

			configFile       *os.File
			envFile          *os.File
			imageFile        *os.File
			installationFile string
			stateFile        *os.File
			varsFile         *os.File

			command *vmlifecyclecommands.UpgradeOpsman
		)

		BeforeEach(func() {
			// setting up all the valid configuration before each test case.

			stateContent = `{
              "vm_id": "i-some-id",
              "iaas": "aws"
            }`

			imageContent = `some-bytes` // image is not in the same format across the iaas, it is not easy to verify ahead of time

			configContent = `
			{
              "opsman-configuration": {
                "aws": {
                  "access_key_id": "sample-access-id",
                  "secret_access_key": "sample-secret-access-key",
                  "region": "us-west-2",
                  "vm_name": "ops-manager-vm",
                  "boot_disk_size": 100,
                  "vpc_subnet_id": "subnet-0292bc845215c2cbf",
                  "security_group_id": "sg-0354f804ba7c4bc41",
                  "key_pair_name": "((key_name))",
                  "iam_instance_profile_name": "ops-manager-iam",
                  "public_ip": "1.2.3.4",
                  "private_ip": "10.0.0.2",
                  "instance_type": "m5.large"
                }
              }
            }`

			varsContent = `{ "key_name": "some-key" }`

			envContent = `{
              "target": "https://pcf.example.com",
              "connect-timeout": 30,
              "request-timeout": 1800,
              "skip-ssl-validation": false,
              "username": "username",
              "password": "password",
              "decryption-passphrase": "passphrase"
            }`

			installationContent = `some-bytes`
		})

		JustBeforeEach(func() {
			var err error

			configFile, err = os.CreateTemp("", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(err).ToNot(HaveOccurred())
			_, err = configFile.WriteString(configContent)
			Expect(err).ToNot(HaveOccurred())

			envFile, err = os.CreateTemp("", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(err).ToNot(HaveOccurred())
			_, err = envFile.WriteString(envContent)
			Expect(err).ToNot(HaveOccurred())

			imageFile, err = os.CreateTemp("", "opsman-2.2.2*.yml")
			Expect(err).ToNot(HaveOccurred())
			_, err = imageFile.WriteString(imageContent)
			Expect(err).ToNot(HaveOccurred())

			installationFile = createZipFile([]struct{ Name, Body string }{
				{"installation.yml", ""}})

			stateFile, err = os.CreateTemp("", "")
			Expect(err).ToNot(HaveOccurred())
			_, err = stateFile.WriteString(stateContent)
			Expect(err).ToNot(HaveOccurred())

			varsFile, err = os.CreateTemp("", "")
			Expect(err).ToNot(HaveOccurred())
			_, err = varsFile.WriteString(varsContent)
			Expect(err).ToNot(HaveOccurred())

			deleteService := &fakes.DeleteVMService{}
			createService := &fakes.CreateVMService{}

			deleteVM := vmlifecyclecommands.NewDeleteVMCommand(GinkgoWriter, GinkgoWriter, func(config *vmmanagers.OpsmanConfigFilePayload, image string, state vmmanagers.StateInfo, outWriter, errWriter io.Writer) (vmmanagers.DeleteVMService, error) {
				return deleteService, nil
			})
			createVM := vmlifecyclecommands.NewCreateVMCommand(GinkgoWriter, GinkgoWriter, func(config *vmmanagers.OpsmanConfigFilePayload, image string, state vmmanagers.StateInfo, outWriter, errWriter io.Writer) (vmmanagers.CreateVMService, error) {
				return createService, nil
			})
			om, err := runner.NewRunner("om", GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			command = vmlifecyclecommands.NewUpgradeOpsman(GinkgoWriter, GinkgoWriter, &createVM, &deleteVM, om, 10*time.Millisecond, 20*time.Millisecond)
			command.ImportInstallation.Installation = installationFile
			command.ImportInstallation.EnvFile = envFile.Name()

			command.CreateVM.Config = configFile.Name()
			command.CreateVM.StateFile = stateFile.Name()
			command.CreateVM.VarsFile = []string{varsFile.Name()}
			command.CreateVM.ImageFile = imageFile.Name()
		})

		AfterEach(func() {
			err := configFile.Close()
			Expect(err).ToNot(HaveOccurred())
			err = envFile.Close()
			Expect(err).ToNot(HaveOccurred())
			err = imageFile.Close()
			Expect(err).ToNot(HaveOccurred())
			err = os.Remove(installationFile)
			Expect(err).ToNot(HaveOccurred())
			err = stateFile.Close()
			Expect(err).ToNot(HaveOccurred())
			err = varsFile.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		When("config file is missing", func() {
			JustBeforeEach(func() {
				command.CreateVM.Config = "non-exist.yml"
			})

			It("errors saying file not found", func() {
				err := command.Execute(nil)
				Expect(err).To(MatchError("could not open config file (non-exist.yml): stat non-exist.yml: no such file or directory"))
			})
		})

		When("config file contains unknown iaas", func() {
			BeforeEach(func() {
				configContent = `
                    {
                      "opsman-configuration": {
                        "zzzzz": {
                          "subscription_id": "90f35f10-ea9e-4e80-aac4-d6778b995532"
                    }}}`
			})

			It("errors saying config contains unknown iaas", func() {
				err := command.Execute(nil)
				Expect(err.Error()).To(ContainSubstring("could not validate config file (" + command.CreateVM.Config + "): unknown iaas: zzzzz, please refer to documentation"))
			})
		})

		When("installation file is missing", func() {
			JustBeforeEach(func() {
				command.ImportInstallation.Installation = "non-exist.yml"
			})

			It("errors saying file not found", func() {
				err := command.Execute(nil)
				Expect(err).To(MatchError("could not open installation file (non-exist.yml): stat non-exist.yml: no such file or directory"))
			})
		})

		When("installation file is a not a valid zip file", func() {
			var notZippedInstallationFile string
			JustBeforeEach(func() {
				notZippedInstallationFile = createZipFile([]struct{ Name, Body string }{})
				command.ImportInstallation.Installation = notZippedInstallationFile
			})

			It("errors saying file is not valid", func() {
				err := command.Execute(nil)
				expectedError := fmt.Sprintf("file: \"%s\" is not a valid installation file", notZippedInstallationFile)
				Expect(err).To(MatchError(expectedError))
			})
		})

		When("installation zip file does not have required installation.yml", func() {
			var invalidInstallationZipName string
			JustBeforeEach(func() {
				invalidInstallationZip, err := os.CreateTemp("", "")
				Expect(err).ToNot(HaveOccurred())
				_, err = invalidInstallationZip.WriteString(installationContent)
				Expect(err).ToNot(HaveOccurred())
				command.ImportInstallation.Installation = invalidInstallationZip.Name()
				invalidInstallationZipName = invalidInstallationZip.Name()
			})

			It("errors saying file is not valid", func() {
				err := command.Execute(nil)
				expectedError := fmt.Sprintf("file: \"%s\" is not a valid zip file", invalidInstallationZipName)
				Expect(err).To(MatchError(expectedError))
			})
		})

		When("env file is missing", func() {
			JustBeforeEach(func() {
				command.ImportInstallation.EnvFile = "non-exist.yml"
			})

			It("errors saying file not found", func() {
				err := command.Execute(nil)
				Expect(err).To(MatchError("could not open env file (non-exist.yml): stat non-exist.yml: no such file or directory"))
			})
		})

		When("image file is missing", func() {
			JustBeforeEach(func() {
				command.CreateVM.ImageFile = "non-exist.yml"
			})

			It("errors saying file not found", func() {
				err := command.Execute(nil)
				Expect(err).To(MatchError("could not open image file (non-exist.yml): stat non-exist.yml: no such file or directory"))
			})
		})

		When("state file is missing", func() {
			JustBeforeEach(func() {
				command.CreateVM.StateFile = "non-exist.yml"
			})

			It("errors saying file not found", func() {
				err := command.Execute(nil)
				Expect(err).To(MatchError("could not open state file (non-exist.yml): stat non-exist.yml: no such file or directory"))
			})
		})

		When("vars file is missing", func() {
			JustBeforeEach(func() {
				command.CreateVM.VarsFile = []string{"non-exist.yml"}
			})

			It("errors saying file not found", func() {
				err := command.Execute(nil)
				Expect(err).To(MatchError("could not open vars file (non-exist.yml): stat non-exist.yml: no such file or directory"))
			})
		})

		When("top level keys are not recognized", func() {
			BeforeEach(func() {
				configContent = `
                 {
                   "opsman-configuration": {
                     "aws": {
                       "access_key_id": "sample-access-id",
                       "secret_access_key": "sample-secret-access-key",
                       "region": "us-west-2",
                       "vm_name": "ops-manager-vm",
                       "boot_disk_size": 100,
                       "vpc_subnet_id": "subnet-0292bc845215c2cbf",
                       "security_group_id": "sg-0354f804ba7c4bc41",
                       "key_pair_name": "some-key-name",
                       "iam_instance_profile_name": "ops-manager-iam",
                       "public_ip": "1.2.3.4",
                       "private_ip": "10.0.0.2",
                       "instance_type": "m5.large"
                     }
                   },
                   "unused-top-level-key": {
                     "unused-nested-key": "some-value"
                   }
                 }`
			})

			It("does not return a configuration validation error", func() {
				err := command.Execute([]string{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("20ms waiting for opsman to respond"))
			})
		})

		When("opsman-configuration is missing from the config", func() {
			BeforeEach(func() {
				configContent = `{
                  "unused-top-level-key-1": {
                    "unused-nested-key": "some-value"
                  },
                  "unused-top-level-key-2": {
                    "unused-nested-key": "some-value"
                  }
                }`
			})

			It("returns an error highlighting the missing key", func() {
				err := command.Execute([]string{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("top-level-key 'opsman-configuration' is a required key."))
				Expect(err.Error()).To(ContainSubstring("Ensure the correct file is passed, the 'opsman-configuration' key is present, and the key is spelled correctly with a dash(-)."))
				Expect(err.Error()).To(ContainSubstring("Found keys:\n  'unused-top-level-key-1'\n  'unused-top-level-key-2'"))
			})
		})
	})
})

func createZipFile(files []struct{ Name, Body string }) string {
	tmpFile, err := os.CreateTemp("", "")
	w := zip.NewWriter(tmpFile)

	Expect(err).ToNot(HaveOccurred())
	for _, file := range files {
		f, err := w.Create(file.Name)
		if err != nil {
			Expect(err).ToNot(HaveOccurred())
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			Expect(err).ToNot(HaveOccurred())
		}
	}
	err = w.Close()
	Expect(err).ToNot(HaveOccurred())

	return tmpFile.Name()
}

func createUpgradeOpsmanCommand() (*vmlifecyclecommands.UpgradeOpsman, *fakes.CreateVMService, *fakes.DeleteVMService, *gbytes.Buffer, *gbytes.Buffer) {
	deleteService := &fakes.DeleteVMService{}
	createService := &fakes.CreateVMService{}

	deleteCmd := vmlifecyclecommands.NewDeleteVMCommand(GinkgoWriter, GinkgoWriter, func(config *vmmanagers.OpsmanConfigFilePayload, image string, state vmmanagers.StateInfo, outWriter, errWriter io.Writer) (vmmanagers.DeleteVMService, error) {
		return deleteService, nil
	})
	createCmd := vmlifecyclecommands.NewCreateVMCommand(GinkgoWriter, GinkgoWriter, func(config *vmmanagers.OpsmanConfigFilePayload, image string, state vmmanagers.StateInfo, outWriter, errWriter io.Writer) (vmmanagers.CreateVMService, error) {
		return createService, nil
	})
	createCmd.ImageFile = writeFile("")
	createCmd.Config = writeFile("")
	createCmd.StateFile = writeFile(`{"iaas": "gcp", "vm_id": "some-id" }`)
	createCmd.Config = writeFile(`{"opsman-configuration": {"gcp": {}}}`)

	stdout := gbytes.NewBuffer()
	stderr := gbytes.NewBuffer()

	omRunner, err := runner.NewRunner("om", io.MultiWriter(stdout, GinkgoWriter), io.MultiWriter(stderr, GinkgoWriter))
	Expect(err).ToNot(HaveOccurred())

	upgradeOpsman := vmlifecyclecommands.NewUpgradeOpsman(io.MultiWriter(stdout, GinkgoWriter), io.MultiWriter(stderr, GinkgoWriter), &createCmd, &deleteCmd, omRunner, 10*time.Millisecond, 20*time.Millisecond)
	upgradeOpsman.ImportInstallation.Installation = createZipFile([]struct{ Name, Body string }{
		{"installation.yml", ""},
	})

	return upgradeOpsman, createService, deleteService, stdout, stderr
}

func createJsonString(t interface{}) string {
	bytes, err := json.Marshal(t)
	Expect(err).ToNot(HaveOccurred())
	return string(bytes)
}
