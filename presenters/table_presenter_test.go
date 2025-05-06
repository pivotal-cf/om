package presenters_test

import (
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/models"
	"github.com/pivotal-cf/om/presenters"
)

var _ = Describe("TablePresenter", func() {
	var (
		tablePresenter  presenters.TablePresenter
		fakeTableWriter *fakes.TableWriter
	)

	BeforeEach(func() {
		fakeTableWriter = &fakes.TableWriter{}
		tablePresenter = presenters.NewTablePresenter(fakeTableWriter)
	})

	Describe("PresentAvailableProducts", func() {
		var products []models.Product

		BeforeEach(func() {
			products = []models.Product{
				{Name: "some-name", Version: "some-version"},
				{Name: "some-other-name", Version: "some-other-version"},
			}
		})

		It("creates a table", func() {
			tablePresenter.PresentAvailableProducts(products)

			Expect(fakeTableWriter.SetAlignmentCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetAlignmentArgsForCall(0)).To(Equal(tablewriter.ALIGN_LEFT))

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			headers := fakeTableWriter.SetHeaderArgsForCall(0)
			Expect(headers).To(Equal([]string{"Name", "Version"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(2))

			values := fakeTableWriter.AppendArgsForCall(0)
			Expect(values).To(Equal([]string{"some-name", "some-version"}))
			values = fakeTableWriter.AppendArgsForCall(1)
			Expect(values).To(Equal([]string{"some-other-name", "some-other-version"}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})
	})

	Describe("PresentCredentials", func() {
		var credentials map[string]string

		BeforeEach(func() {
			credentials = map[string]string{"identity": "some-identity", "password": "some-password"}
		})

		It("creates a table", func() {
			tablePresenter.PresentCredentials(credentials)

			Expect(fakeTableWriter.SetAutoFormatHeadersCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetAutoFormatHeadersArgsForCall(0)).To(Equal(false))
			Expect(fakeTableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"identity", "password"}))

			Expect(fakeTableWriter.SetAutoWrapTextCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetAutoWrapTextArgsForCall(0)).To(Equal(false))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(1))
			Expect(fakeTableWriter.AppendArgsForCall(0)).To(Equal([]string{"some-identity", "some-password"}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})
	})

	Describe("PresentCredentialReferences", func() {
		var credentials []string

		BeforeEach(func() {
			credentials = []string{"cred-1", "cred-2", "cred-3"}
		})

		It("creates a table", func() {
			tablePresenter.PresentCredentialReferences(credentials)

			Expect(fakeTableWriter.SetAlignmentCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetAlignmentArgsForCall(0)).To(Equal(tablewriter.ALIGN_LEFT))

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			headers := fakeTableWriter.SetHeaderArgsForCall(0)
			Expect(headers).To(Equal([]string{"Credentials"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(3))

			values := fakeTableWriter.AppendArgsForCall(0)
			Expect(values).To(Equal([]string{"cred-1"}))
			values = fakeTableWriter.AppendArgsForCall(1)
			Expect(values).To(Equal([]string{"cred-2"}))
			values = fakeTableWriter.AppendArgsForCall(2)
			Expect(values).To(Equal([]string{"cred-3"}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})
	})

	Describe("PresentErrands", func() {
		var errands []models.Errand

		BeforeEach(func() {
			errands = []models.Errand{
				{Name: "errand-1", PostDeployEnabled: "post-deploy-1", PreDeleteEnabled: "pre-delete-1"},
				{Name: "errand-2", PostDeployEnabled: "post-deploy-2", PreDeleteEnabled: "pre-delete-2"},
			}
		})

		It("creates a table", func() {
			tablePresenter.PresentErrands(errands)
			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))

			headers := fakeTableWriter.SetHeaderArgsForCall(0)
			Expect(headers).To(Equal([]string{"Name", "Post Deploy Enabled", "Pre Delete Enabled"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(2))

			values := fakeTableWriter.AppendArgsForCall(0)
			Expect(values).To(Equal([]string{errands[0].Name, errands[0].PostDeployEnabled, errands[0].PreDeleteEnabled}))
			values = fakeTableWriter.AppendArgsForCall(1)
			Expect(values).To(Equal([]string{errands[1].Name, errands[1].PostDeployEnabled, errands[1].PreDeleteEnabled}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})
	})

	Describe("PresentCertificateAuthority", func() {
		var certificateAuthority api.CA

		BeforeEach(func() {
			certificateAuthority = api.CA{
				GUID:      "some GUID",
				Issuer:    "some Issuer",
				CreatedOn: "2017-09-12",
				ExpiresOn: "2018-09-12",
				Active:    true,
				CertPEM:   "some CertPem",
			}
		})

		It("creates a table", func() {
			tablePresenter.PresentCertificateAuthority(certificateAuthority)
			Expect(fakeTableWriter.SetAutoWrapTextCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetAutoWrapTextArgsForCall(0)).To(BeFalse())

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"id", "issuer", "active", "created on", "expires on", "certicate pem"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(1))
			Expect(fakeTableWriter.AppendArgsForCall(0)).To(Equal([]string{"some GUID", "some Issuer",
				"true", "2017-09-12", "2018-09-12", "some CertPem"}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})
	})

	Describe("PresentGenerateCAResponse", func() {
		var car api.GenerateCAResponse

		BeforeEach(func() {
			car = api.GenerateCAResponse{
				CA: api.CA{
					GUID:      "some GUID",
					Issuer:    "some Issuer",
					CreatedOn: "2017-09-12",
					ExpiresOn: "2018-09-12",
					Active:    true,
					CertPEM:   "some CertPem",
				},
				Warnings: []string{"something not ideal!", "maybe even two things!"},
			}
		})

		It("creates a table", func() {
			tablePresenter.PresentGenerateCAResponse(car)
			Expect(fakeTableWriter.SetAutoWrapTextCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetAutoWrapTextArgsForCall(0)).To(BeFalse())

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"id", "issuer", "active", "created on", "expires on", "certicate pem", "warnings"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(1))
			Expect(fakeTableWriter.AppendArgsForCall(0)).To(Equal([]string{"some GUID", "some Issuer",
				"true", "2017-09-12", "2018-09-12", "some CertPem", "something not ideal!;maybe even two things!"}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})
	})

	Describe("PresentInstallations", func() {
		var installations []models.Installation

		BeforeEach(func() {
			startedAt := time.Now().Add(1 * time.Hour)
			finishedAt := time.Now().Add(2 * time.Hour)

			installations = []models.Installation{
				{
					Id:         1,
					User:       "some-user",
					Status:     "some-status",
					StartedAt:  &startedAt,
					FinishedAt: &finishedAt,
				},
			}
		})

		It("creates a table", func() {
			tablePresenter.PresentInstallations(installations)
			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))

			headers := fakeTableWriter.SetHeaderArgsForCall(0)
			Expect(headers).To(Equal([]string{"ID", "User", "Status", "Started At", "Finished At"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(1))
			values := fakeTableWriter.AppendArgsForCall(0)
			Expect(values).To(Equal([]string{
				strconv.Itoa(installations[0].Id),
				installations[0].User,
				installations[0].Status,
				installations[0].StartedAt.Format(time.RFC3339Nano),
				installations[0].FinishedAt.Format(time.RFC3339Nano),
			}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})

		When("there are no installations", func() {
			BeforeEach(func() {
				installations = []models.Installation{}
			})

			It("creates an empty table when no installations are present", func() {
				tablePresenter.PresentInstallations(installations)
				Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))

				headers := fakeTableWriter.SetHeaderArgsForCall(0)
				Expect(headers).To(ConsistOf("ID", "User", "Status", "Started At", "Finished At"))

				Expect(fakeTableWriter.AppendCallCount()).To(Equal(0))

				Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
			})
		})
	})

	Describe("PresentPendingChanges", func() {
		var pendingChanges api.PendingChangesOutput
		BeforeEach(func() {
			pendingChanges = api.PendingChangesOutput{
				ChangeList: []api.ProductChange{
					{
						GUID:   "some-product",
						Action: "update",
						Errands: []api.Errand{
							{
								Name:       "some-errand",
								PostDeploy: "on",
								PreDelete:  "false",
							},
							{
								Name:       "some-errand-2",
								PostDeploy: "when-change",
								PreDelete:  "false",
							},
						},
					},
					{
						GUID:    "some-product-without-errand",
						Action:  "install",
						Errands: []api.Errand{},
					},
				},
			}
		})

		It("creates a table", func() {
			tablePresenter.PresentPendingChanges(pendingChanges)

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"PRODUCT", "ACTION", "ERRANDS"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(3))
			Expect(fakeTableWriter.AppendArgsForCall(0)).To(Equal([]string{"some-product", "update", "some-errand"}))
			Expect(fakeTableWriter.AppendArgsForCall(1)).To(Equal([]string{"", "", "some-errand-2"}))
			Expect(fakeTableWriter.AppendArgsForCall(2)).To(Equal([]string{"some-product-without-errand", "install", ""}))
		})
	})

	Describe("PresentProducts", func() {
		It("creates a table with specified columns", func() {
			products := models.ProductsVersionsDisplay{
				ProductVersions: []models.ProductVersions{{
					Name:      "test-product",
					Available: []string{"1"},
					Staged:    "2",
					Deployed:  "3",
				}, {
					Name:      "another-product",
					Available: []string{"4"},
					Staged:    "5",
					Deployed:  "6",
				}},
				Available: true,
				Staged:    false,
				Deployed:  true,
			}
			By("Printing available and deployed columns")
			tablePresenter.PresentProducts(products)

			Expect(fakeTableWriter.SetAlignmentCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetAlignmentArgsForCall(0)).To(Equal(tablewriter.ALIGN_LEFT))

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			headers := fakeTableWriter.SetHeaderArgsForCall(0)
			Expect(headers).To(Equal([]string{"Name", "Available", "Deployed"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(2))

			values := fakeTableWriter.AppendArgsForCall(0)
			Expect(values).To(Equal([]string{"test-product", "1", "3"}))
			values = fakeTableWriter.AppendArgsForCall(1)
			Expect(values).To(Equal([]string{"another-product", "4", "6"}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))

			By("printing only the staged column")
			products.Available = false
			products.Staged = true
			products.Deployed = false

			fakeTableWriter = &fakes.TableWriter{}
			tablePresenter = presenters.NewTablePresenter(fakeTableWriter)

			tablePresenter.PresentProducts(products)

			Expect(fakeTableWriter.SetAlignmentCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetAlignmentArgsForCall(0)).To(Equal(tablewriter.ALIGN_LEFT))

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			headers = fakeTableWriter.SetHeaderArgsForCall(0)
			Expect(headers).To(Equal([]string{"Name", "Staged"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(2))

			values = fakeTableWriter.AppendArgsForCall(0)
			Expect(values).To(Equal([]string{"test-product", "2"}))
			values = fakeTableWriter.AppendArgsForCall(1)
			Expect(values).To(Equal([]string{"another-product", "5"}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})

		It("does not display a product if it has no versions for all of the listed columns", func() {
			products := models.ProductsVersionsDisplay{
				ProductVersions: []models.ProductVersions{{
					Name:      "test-product",
					Available: []string{"1", "2"},
					Staged:    "",
					Deployed:  "3",
				}, {
					Name:      "another-product",
					Available: []string{"4"},
					Staged:    "5",
					Deployed:  "",
				}, {
					Name:      "p-bosh-test",
					Available: []string{},
					Staged:    "6",
					Deployed:  "7",
				}},
				Available: true,
				Staged:    false,
				Deployed:  true,
			}
			By("Printing available and deployed columns")
			tablePresenter.PresentProducts(products)

			Expect(fakeTableWriter.SetAlignmentCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetAlignmentArgsForCall(0)).To(Equal(tablewriter.ALIGN_LEFT))

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			headers := fakeTableWriter.SetHeaderArgsForCall(0)
			Expect(headers).To(Equal([]string{"Name", "Available", "Deployed"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(4))

			values := fakeTableWriter.AppendArgsForCall(0)
			Expect(values).To(Equal([]string{"test-product", "1", "3"}))
			values = fakeTableWriter.AppendArgsForCall(1)
			Expect(values).To(Equal([]string{"", "2", ""}))
			values = fakeTableWriter.AppendArgsForCall(2)
			Expect(values).To(Equal([]string{"another-product", "4", ""}))
			values = fakeTableWriter.AppendArgsForCall(3)
			Expect(values).To(Equal([]string{"p-bosh-test", "", "7"}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))

			By("printing only the staged column")
			products.Available = false
			products.Staged = true
			products.Deployed = false

			fakeTableWriter = &fakes.TableWriter{}
			tablePresenter = presenters.NewTablePresenter(fakeTableWriter)

			tablePresenter.PresentProducts(products)

			Expect(fakeTableWriter.SetAlignmentCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetAlignmentArgsForCall(0)).To(Equal(tablewriter.ALIGN_LEFT))

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			headers = fakeTableWriter.SetHeaderArgsForCall(0)
			Expect(headers).To(Equal([]string{"Name", "Staged"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(2))

			values = fakeTableWriter.AppendArgsForCall(0)
			Expect(values).To(Equal([]string{"another-product", "5"}))
			values = fakeTableWriter.AppendArgsForCall(1)
			Expect(values).To(Equal([]string{"p-bosh-test", "6"}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})
	})

	Describe("PresentStagedProducts", func() {
		var stagedProducts []api.DiagnosticProduct
		BeforeEach(func() {
			stagedProducts = []api.DiagnosticProduct{
				{
					Name:    "some-product",
					Version: "some-version",
				},
				{
					Name:    "acme-product",
					Version: "version-infinity",
				},
			}
		})

		It("creates a table", func() {
			tablePresenter.PresentStagedProducts(stagedProducts)

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"Name", "Version"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(2))
			Expect(fakeTableWriter.AppendArgsForCall(0)).To(Equal([]string{"some-product", "some-version"}))
			Expect(fakeTableWriter.AppendArgsForCall(1)).To(Equal([]string{"acme-product", "version-infinity"}))
			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})
	})

	Describe("PresentDeployedProducts", func() {
		var deployedProducts []api.DiagnosticProduct
		BeforeEach(func() {
			deployedProducts = []api.DiagnosticProduct{
				{
					Name:    "some-product",
					Version: "some-version",
				},
				{
					Name:    "acme-product",
					Version: "version-infinity",
				},
			}
		})

		It("creates a table", func() {
			tablePresenter.PresentDeployedProducts(deployedProducts)

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"Name", "Version"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(2))
			Expect(fakeTableWriter.AppendArgsForCall(0)).To(Equal([]string{"some-product", "some-version"}))
			Expect(fakeTableWriter.AppendArgsForCall(1)).To(Equal([]string{"acme-product", "version-infinity"}))
			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})
	})

	Describe("PresentCertificateAuthorities", func() {
		var certificateAuthorities []api.CA
		BeforeEach(func() {
			certificateAuthorities = []api.CA{
				{
					GUID:      "some-guid",
					Issuer:    "Pivotal",
					CreatedOn: "2017-01-09",
					ExpiresOn: "2021-01-09",
					Active:    true,
					CertPEM:   "-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI....",
				},
				{
					GUID:      "other-guid",
					Issuer:    "Customer",
					CreatedOn: "2017-01-10",
					ExpiresOn: "2021-01-10",
					Active:    false,
					CertPEM:   "-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBBhI....",
				},
			}
		})

		It("creates a table", func() {
			tablePresenter.PresentCertificateAuthorities(certificateAuthorities)

			Expect(fakeTableWriter.SetAutoWrapTextCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetAutoWrapTextArgsForCall(0)).To(BeFalse())

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			Expect(fakeTableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"id", "issuer", "active", "created on", "expires on", "certicate pem"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(2))
			Expect(fakeTableWriter.AppendArgsForCall(0)).To(Equal([]string{"some-guid", "Pivotal", "true", "2017-01-09", "2021-01-09", "-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBAgI...."}))
			Expect(fakeTableWriter.AppendArgsForCall(1)).To(Equal([]string{"other-guid", "Customer", "false", "2017-01-10", "2021-01-10", "-----BEGIN CERTIFICATE-----\nMIIC+zCCAeOgAwIBBhI...."}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})
	})

	Describe("PresentLicensedProducts", func() {
		var products []api.ExpiringLicenseOutput

		BeforeEach(func() {
			expiryDate, _ := time.Parse("2006-01-02", "2026-03-20")
			products = []api.ExpiringLicenseOutput{
				{
					ProductName:    "cf",
					GUID:           "cf-fa24570b6a6e8940ab57",
					ExpiresAt:      expiryDate,
					ProductState:   []string{"staged"},
					LicenseVersion: "10.0",
					ProductVersion: "2.10.1",
				},
				{
					ProductName:    "p-bosh",
					GUID:           "p-bosh-123456789",
					ExpiresAt:      expiryDate.AddDate(0, 1, 0),
					ProductState:   []string{"deployed"},
					LicenseVersion: "10.0",
					ProductVersion: "3.0.0",
				},
			}
		})

		It("creates a table with all product license fields", func() {
			tablePresenter.PresentLicensedProducts(products)

			Expect(fakeTableWriter.SetHeaderCallCount()).To(Equal(1))
			headers := fakeTableWriter.SetHeaderArgsForCall(0)
			Expect(headers).To(Equal([]string{"Name", "GUID", "Product Version", "State", "Licensed Version", "Expiry"}))

			Expect(fakeTableWriter.AppendCallCount()).To(Equal(2))

			firstRow := fakeTableWriter.AppendArgsForCall(0)
			Expect(firstRow).To(Equal([]string{
				"cf",
				"cf-fa24570b6a6e8940ab57",
				"2.10.1",
				"staged",
				"10.0",
				"2026-03-20",
			}))

			secondRow := fakeTableWriter.AppendArgsForCall(1)
			Expect(secondRow).To(Equal([]string{
				"p-bosh",
				"p-bosh-123456789",
				"3.0.0",
				"deployed",
				"10.0",
				"2026-04-20",
			}))

			Expect(fakeTableWriter.RenderCallCount()).To(Equal(1))
		})
	})
})
