package commands_test

import (
	"bytes"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("KubernetesDistributions", func() {
	var (
		fakeService *fakes.KubernetesDistributionsService
		stdout      *bytes.Buffer
		command     *commands.KubernetesDistributions
	)

	BeforeEach(func() {
		fakeService = &fakes.KubernetesDistributionsService{}
		stdout = &bytes.Buffer{}
		command = commands.NewKubernetesDistributions(fakeService, stdout)

		fakeService.InfoReturns(api.Info{Version: "3.3.0"}, nil)
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			fakeService.ListKubernetesDistributionsReturns(api.KubernetesDistributionAssociationsResponse{
				Products: []api.KubernetesProductDistributionEntry{
					{
						GUID:        "rabbitmq-abc123",
						ProductName: "rabbitmq",
						StagedKubernetesDistribution: &api.KubernetesDistribution{
							Identifier: "managed-k8s",
							Version:    "0.2.0",
						},
						DeployedKubernetesDistribution: &api.KubernetesDistribution{
							Identifier: "managed-k8s",
							Version:    "0.1.0",
						},
						AvailableKubernetesDistributions: []api.KubernetesDistribution{
							{Identifier: "managed-k8s", Version: "0.1.0"},
							{Identifier: "managed-k8s", Version: "0.2.0"},
						},
					},
					{
						GUID:        "mysql-def456",
						ProductName: "mysql",
						StagedKubernetesDistribution: &api.KubernetesDistribution{
							Identifier: "managed-k8s",
							Version:    "0.2.0",
						},
						AvailableKubernetesDistributions: []api.KubernetesDistribution{
							{Identifier: "managed-k8s", Version: "0.2.0"},
						},
					},
				},
				Library: []api.KubernetesDistributionLibraryEntry{
					{Identifier: "managed-k8s", Version: "0.1.0", Rank: 1, Label: "Managed Kubernetes"},
					{Identifier: "managed-k8s", Version: "0.2.0", Rank: 1, Label: "Managed Kubernetes"},
					{Identifier: "unmanaged-k8s", Version: "0.1.0", Rank: 50, Label: "Unmanaged Kubernetes"},
					{Identifier: "unmanaged-k8s", Version: "0.3.0", Rank: 50, Label: "Unmanaged Kubernetes"},
				},
			}, nil)
		})

		It("renders a distribution-centric table without product columns", func() {
			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.ListKubernetesDistributionsCallCount()).To(Equal(1))

			output := stdout.String()
			Expect(output).To(ContainSubstring("DISTRIBUTION"))
			Expect(output).To(ContainSubstring("VERSION"))
			Expect(output).ToNot(ContainSubstring("STAGED"))
			Expect(output).ToNot(ContainSubstring("DEPLOYED"))

			Expect(output).To(ContainSubstring("unmanaged-k8s"))
			Expect(output).To(ContainSubstring("0.1.0"))
			Expect(output).To(ContainSubstring("0.3.0"))

			Expect(output).To(ContainSubstring("managed-k8s"))
		})

		It("shows unassociated distributions (unmanaged-k8s has no products)", func() {
			err := executeCommand(command, []string{})
			Expect(err).ToNot(HaveOccurred())

			output := stdout.String()
			Expect(output).To(ContainSubstring("unmanaged-k8s"))
		})

		When("--format json is provided", func() {
			It("renders JSON output", func() {
				err := executeCommand(command, []string{"--format", "json"})
				Expect(err).ToNot(HaveOccurred())

				var rows []commands.K8sDistributionRow
				Expect(json.Unmarshal(stdout.Bytes(), &rows)).To(Succeed())

				Expect(rows).To(HaveLen(4))

				Expect(rows[0].Distribution).To(Equal("managed-k8s"))
				Expect(rows[0].Version).To(Equal("0.1.0"))
				Expect(rows[0].Products).To(HaveLen(1))
				Expect(rows[0].Products[0].Name).To(Equal("rabbitmq"))
				Expect(rows[0].Products[0].Staged).To(BeFalse())
				Expect(rows[0].Products[0].Deployed).To(BeTrue())

				Expect(rows[1].Distribution).To(Equal("managed-k8s"))
				Expect(rows[1].Version).To(Equal("0.2.0"))
				Expect(rows[1].Products).To(HaveLen(2))
				Expect(rows[1].Products[0].Name).To(Equal("mysql"))
				Expect(rows[1].Products[0].Staged).To(BeTrue())
				Expect(rows[1].Products[1].Name).To(Equal("rabbitmq"))
				Expect(rows[1].Products[1].Staged).To(BeTrue())

				Expect(rows[2].Distribution).To(Equal("unmanaged-k8s"))
				Expect(rows[2].Version).To(Equal("0.1.0"))
				Expect(rows[2].Products).To(BeEmpty())

				Expect(rows[3].Distribution).To(Equal("unmanaged-k8s"))
				Expect(rows[3].Version).To(Equal("0.3.0"))
				Expect(rows[3].Products).To(BeEmpty())
			})
		})

		When("the API returns unsorted data", func() {
			BeforeEach(func() {
				fakeService.ListKubernetesDistributionsReturns(api.KubernetesDistributionAssociationsResponse{
					Products: []api.KubernetesProductDistributionEntry{
						{
							GUID:        "redis-abc123",
							ProductName: "redis",
							StagedKubernetesDistribution: &api.KubernetesDistribution{
								Identifier: "managed-k8s",
								Version:    "0.10.0",
							},
							AvailableKubernetesDistributions: []api.KubernetesDistribution{
								{Identifier: "managed-k8s", Version: "0.10.0"},
							},
						},
						{
							GUID:        "kafka-def456",
							ProductName: "kafka",
							StagedKubernetesDistribution: &api.KubernetesDistribution{
								Identifier: "managed-k8s",
								Version:    "0.10.0",
							},
							AvailableKubernetesDistributions: []api.KubernetesDistribution{
								{Identifier: "managed-k8s", Version: "0.10.0"},
							},
						},
					},
					Library: []api.KubernetesDistributionLibraryEntry{
						{Identifier: "managed-k8s", Version: "0.10.0", Rank: 1, Label: "Managed Kubernetes"},
						{Identifier: "unmanaged-k8s", Version: "0.3.0", Rank: 50, Label: "Unmanaged Kubernetes"},
						{Identifier: "managed-k8s", Version: "0.2.0", Rank: 1, Label: "Managed Kubernetes"},
						{Identifier: "unmanaged-k8s", Version: "0.1.0", Rank: 50, Label: "Unmanaged Kubernetes"},
					},
				}, nil)
			})

			It("sorts distributions alphabetically, versions by semver, and products alphabetically", func() {
				err := executeCommand(command, []string{"--format", "json"})
				Expect(err).ToNot(HaveOccurred())

				var rows []commands.K8sDistributionRow
				Expect(json.Unmarshal(stdout.Bytes(), &rows)).To(Succeed())

				Expect(rows).To(HaveLen(4))

				Expect(rows[0].Distribution).To(Equal("managed-k8s"))
				Expect(rows[0].Version).To(Equal("0.2.0"))
				Expect(rows[0].Products).To(BeEmpty())

				Expect(rows[1].Distribution).To(Equal("managed-k8s"))
				Expect(rows[1].Version).To(Equal("0.10.0"))
				Expect(rows[1].Products).To(HaveLen(2))
				Expect(rows[1].Products[0].Name).To(Equal("kafka"))
				Expect(rows[1].Products[1].Name).To(Equal("redis"))

				Expect(rows[2].Distribution).To(Equal("unmanaged-k8s"))
				Expect(rows[2].Version).To(Equal("0.1.0"))
				Expect(rows[2].Products).To(BeEmpty())

				Expect(rows[3].Distribution).To(Equal("unmanaged-k8s"))
				Expect(rows[3].Version).To(Equal("0.3.0"))
				Expect(rows[3].Products).To(BeEmpty())
			})
		})

		When("--product filter is provided", func() {
			It("shows distributions with staged/deployed columns", func() {
				err := executeCommand(command, []string{"--product", "rabbitmq"})
				Expect(err).ToNot(HaveOccurred())

				output := stdout.String()
				Expect(output).To(ContainSubstring("DISTRIBUTION"))
				Expect(output).To(ContainSubstring("VERSION"))
				Expect(output).To(ContainSubstring("STAGED"))
				Expect(output).To(ContainSubstring("DEPLOYED"))

				Expect(output).To(ContainSubstring("managed-k8s"))
				Expect(output).To(ContainSubstring("0.1.0"))
				Expect(output).To(ContainSubstring("0.2.0"))
				Expect(output).To(ContainSubstring("yes"))
				Expect(output).ToNot(ContainSubstring("unmanaged-k8s"))
			})

			It("returns an error when the product is not found", func() {
				err := executeCommand(command, []string{"--product", "nonexistent"})
				Expect(err).To(MatchError(ContainSubstring(`product "nonexistent" not found`)))
			})
		})

		When("no distributions exist in the library", func() {
			BeforeEach(func() {
				fakeService.ListKubernetesDistributionsReturns(api.KubernetesDistributionAssociationsResponse{
					Products: []api.KubernetesProductDistributionEntry{},
					Library:  []api.KubernetesDistributionLibraryEntry{},
				}, nil)
			})

			It("renders an empty table without error", func() {
				err := executeCommand(command, []string{})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("the list API call fails", func() {
			BeforeEach(func() {
				fakeService.ListKubernetesDistributionsReturns(
					api.KubernetesDistributionAssociationsResponse{}, errors.New("api call failed"))
			})

			It("returns the error", func() {
				err := executeCommand(command, []string{})
				Expect(err).To(MatchError(ContainSubstring("api call failed")))
			})
		})
	})

	When("fetching the Ops Manager version fails", func() {
		It("returns an error", func() {
			fakeService.InfoReturns(api.Info{}, errors.New("could not make request to info endpoint"))

			err := executeCommand(command, []string{})
			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("failed to get Ops Manager version"),
				ContainSubstring("could not make request to info endpoint"),
			)))
		})
	})

	When("Ops Manager version is older than 3.3", func() {
		It("returns an error", func() {
			fakeService.InfoReturns(api.Info{Version: "3.2.0"}, nil)

			err := executeCommand(command, []string{})
			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("kubernetes-distributions requires Ops Manager 3.3 or newer"),
				ContainSubstring("(current version: 3.2.0)"),
			)))
		})
	})

	When("determining the Ops Manager version fails", func() {
		It("returns an error", func() {
			fakeService.InfoReturns(api.Info{Version: "not-a-valid-semver"}, nil)

			err := executeCommand(command, []string{})
			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("kubernetes-distributions requires Ops Manager 3.3 or newer"),
				ContainSubstring("not-a-valid-semver"),
			)))
		})
	})
})
