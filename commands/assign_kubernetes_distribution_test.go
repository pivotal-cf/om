package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"

	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("AssignKubernetesDistribution", func() {
	var (
		fakeService *fakes.AssignKubernetesDistributionService
		logger      *fakes.Logger
		command     *commands.AssignKubernetesDistribution
	)

	BeforeEach(func() {
		fakeService = &fakes.AssignKubernetesDistributionService{}
		logger = &fakes.Logger{}
		command = commands.NewAssignKubernetesDistribution(fakeService, logger)

		fakeService.InfoReturns(api.Info{Version: "3.3.0"}, nil)
	})

	When("--distribution exists for the specified product", func() {
		BeforeEach(func() {
			fakeService.ListKubernetesDistributionsReturns(api.KubernetesDistributionAssociationsResponse{
				Products: []api.KubernetesProductDistributionEntry{
					{
						GUID:        "redis-xyz789",
						ProductName: "redis",
						AvailableKubernetesDistributions: []api.KubernetesDistribution{
							{Identifier: "unmanaged-k8s", Version: "50.0"},
						},
					},
					{
						GUID:        "postgres-abc123",
						ProductName: "postgres",
						AvailableKubernetesDistributions: []api.KubernetesDistribution{
							{Identifier: "managed-k8s", Version: "4.567"},
							{Identifier: "unmanaged-k8s", Version: "99.0"},
							{Identifier: "managed-k8s", Version: "8.910"},
							{Identifier: "managed-k8s", Version: "6.0"},
						},
					},
				},
			}, nil)
		})

		It("assigns the distribution", func() {
			err := executeCommand(command, []string{"--product", "postgres", "--distribution", "managed-k8s:4.567"})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.AssignKubernetesDistributionArgsForCall(0)).To(Equal(api.AssignKubernetesDistributionInput{
				Products: []api.AssignKubernetesDistributionProduct{
					{
						GUID: "postgres-abc123",
						KubernetesDistribution: api.KubernetesDistribution{
							Identifier: "managed-k8s",
							Version:    "4.567",
						},
					},
				},
			}))
		})

		When("--distribution latest is used", func() {
			It("assigns the latest distribution matching the kubernetes distribution name", func() {
				err := executeCommand(command, []string{"--product", "postgres", "--distribution", "managed-k8s:latest"})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeService.AssignKubernetesDistributionArgsForCall(0)).To(Equal(api.AssignKubernetesDistributionInput{
					Products: []api.AssignKubernetesDistributionProduct{
						{
							GUID: "postgres-abc123",
							KubernetesDistribution: api.KubernetesDistribution{
								Identifier: "managed-k8s",
								Version:    "8.910",
							},
						},
					},
				}))
			})
		})
	})

	When("--distribution is malformed", func() {
		BeforeEach(func() {
			fakeService.ListKubernetesDistributionsReturns(api.KubernetesDistributionAssociationsResponse{
				Products: []api.KubernetesProductDistributionEntry{
					{
						GUID:        "postgres-abc123",
						ProductName: "postgres",
						AvailableKubernetesDistributions: []api.KubernetesDistribution{
							{Identifier: "managed-k8s", Version: "1.0.0"},
						},
					},
				},
			}, nil)
		})

		It("returns a format error for bare 'latest' without identifier", func() {
			err := executeCommand(command, []string{"--product", "postgres", "--distribution", "latest"})
			Expect(err).To(MatchError(ContainSubstring(`expected "--distribution" format as "identifier:version"`)))
			Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
		})

		It("returns a format error for a value with no colon separator", func() {
			err := executeCommand(command, []string{"--product", "postgres", "--distribution", "nocolon"})
			Expect(err).To(MatchError(ContainSubstring(`expected "--distribution" format as "identifier:version"`)))
			Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
		})

		It("returns a format error when identifier is empty", func() {
			err := executeCommand(command, []string{"--product", "postgres", "--distribution", ":1.0.0"})
			Expect(err).To(MatchError(ContainSubstring(`expected "--distribution" format as "identifier:version"`)))
			Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
		})

		It("returns a format error when version is empty", func() {
			err := executeCommand(command, []string{"--product", "postgres", "--distribution", "managed-k8s:"})
			Expect(err).To(MatchError(ContainSubstring(`expected "--distribution" format as "identifier:version"`)))
			Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
		})
	})

	When("the given distribution is not available", func() {
		BeforeEach(func() {
			fakeService.ListKubernetesDistributionsReturns(api.KubernetesDistributionAssociationsResponse{
				Products: []api.KubernetesProductDistributionEntry{
					{
						GUID:        "redis-abc123",
						ProductName: "redis",
						AvailableKubernetesDistributions: []api.KubernetesDistribution{
							{Identifier: "managed-k8s", Version: "1.23"},
							{Identifier: "unmanaged-k8s", Version: "4.56"},
						},
					},
				},
			}, nil)
		})
		It("returns an error with available distributions", func() {
			err := executeCommand(command, []string{"--product", "redis", "--distribution", "nonexistent-k8s:0.12"})
			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("kubernetes distribution nonexistent-k8s version 0.12 not found"),
				ContainSubstring("managed-k8s 1.23"),
				ContainSubstring("unmanaged-k8s 4.56"),
			)))

			Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
		})
	})

	When("the product is not found", func() {
		Context("because an error was encountered enumerating kubernetes product distribution associations", func() {
			It("returns an error with a clear error message", func() {
				fakeService.ListKubernetesDistributionsReturns(api.KubernetesDistributionAssociationsResponse{}, errors.New("api call failed"))

				err := executeCommand(command, []string{"--product", "rabbitmq", "--distribution", "managed-k8s:1.0"})
				Expect(err).To(MatchError(ContainSubstring("api call failed")))

				Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
			})
		})

		Context("because the product was not found", func() {
			It("returns an error", func() {
				fakeService.ListKubernetesDistributionsReturns(api.KubernetesDistributionAssociationsResponse{
					Products: []api.KubernetesProductDistributionEntry{
						{GUID: "other-guid", ProductName: "other-product"},
					},
				}, nil)

				err := executeCommand(command, []string{"--product", "rabbitmq", "--distribution", "managed-k8s:1.0"})
				Expect(err).To(MatchError(ContainSubstring(`product "rabbitmq" not found`)))

				Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
			})
		})
	})

	When("the product is staged for deletion", func() {
		It("returns an error", func() {
			fakeService.ListKubernetesDistributionsReturns(api.KubernetesDistributionAssociationsResponse{
				Products: []api.KubernetesProductDistributionEntry{
					{
						GUID:              "rabbitmq-abc123",
						ProductName:       "rabbitmq",
						StagedForDeletion: true,
					},
				},
			}, nil)

			err := executeCommand(command, []string{"--product", "rabbitmq", "--distribution", "managed-k8s:1.0"})
			Expect(err).To(MatchError(ContainSubstring(`product "rabbitmq" is staged for deletion`)))

			Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
		})
	})

	When("no available distributions returned from the api", func() {
		It("returns an error", func() {
			fakeService.ListKubernetesDistributionsReturns(api.KubernetesDistributionAssociationsResponse{
				Products: []api.KubernetesProductDistributionEntry{
					{
						GUID:                             "rabbitmq-abc123",
						ProductName:                      "rabbitmq",
						AvailableKubernetesDistributions: []api.KubernetesDistribution{},
					},
				},
			}, nil)

			err := executeCommand(command, []string{"--product", "rabbitmq", "--distribution", "managed-k8s:1.0"})
			Expect(err).To(MatchError(ContainSubstring(`no kubernetes distributions are available for "rabbitmq"`)))

			Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
		})
	})

	When("the assign API call fails", func() {
		It("returns the error", func() {
			fakeService.ListKubernetesDistributionsReturns(api.KubernetesDistributionAssociationsResponse{
				Products: []api.KubernetesProductDistributionEntry{
					{
						GUID:        "postgres-abc123",
						ProductName: "postgres",
						AvailableKubernetesDistributions: []api.KubernetesDistribution{
							{Identifier: "managed-k8s", Version: "1.0.0"},
						},
					},
				},
			}, nil)
			fakeService.AssignKubernetesDistributionReturns(errors.New("server returned 422: incompatible distribution"))

			err := executeCommand(command, []string{"--product", "postgres", "--distribution", "managed-k8s:1.0.0"})
			Expect(err).To(MatchError(ContainSubstring("incompatible distribution")))
		})
	})

	When("--distribution latest is used but no distributions match the identifier", func() {
		It("returns an error listing available distributions", func() {
			fakeService.ListKubernetesDistributionsReturns(api.KubernetesDistributionAssociationsResponse{
				Products: []api.KubernetesProductDistributionEntry{
					{
						GUID:        "redis-def456",
						ProductName: "redis",
						AvailableKubernetesDistributions: []api.KubernetesDistribution{
							{Identifier: "managed-k8s", Version: "1.23.0"},
							{Identifier: "unmanaged-k8s", Version: "2.0.0"},
						},
					},
				},
			}, nil)

			err := executeCommand(command, []string{"--product", "redis", "--distribution", "nonexistent-k8s:latest"})
			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring(`no available kubernetes distribution with identifier "nonexistent-k8s"`),
				ContainSubstring("managed-k8s 1.23.0"),
				ContainSubstring("unmanaged-k8s 2.0.0"),
			)))

			Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
		})
	})

	When("--distribution latest is used but a version string cannot be parsed", func() {
		It("returns a version parse error", func() {
			fakeService.ListKubernetesDistributionsReturns(api.KubernetesDistributionAssociationsResponse{
				Products: []api.KubernetesProductDistributionEntry{
					{
						GUID:        "redis-def456",
						ProductName: "redis",
						AvailableKubernetesDistributions: []api.KubernetesDistribution{
							{Identifier: "managed-k8s", Version: "not-a-version"},
						},
					},
				},
			}, nil)

			err := executeCommand(command, []string{"--product", "redis", "--distribution", "managed-k8s:latest"})
			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("could not parse version"),
				ContainSubstring("not-a-version"),
			)))

			Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
		})
	})

	When("fetching the Ops Manager version fails", func() {
		It("returns an error and does not call AssignKubernetesDistribution", func() {
			fakeService.InfoReturns(api.Info{}, errors.New("could not make request to info endpoint"))

			err := executeCommand(command, []string{
				"--product", "postgres",
				"--distribution", "managed-k8s:3.21",
			})
			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("failed to get Ops Manager version"),
				ContainSubstring("could not make request to info endpoint"),
			)))

			Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
		})
	})

	When("Ops Manager version is older than 3.3", func() {
		It("returns an error with a clear message", func() {
			fakeService.InfoReturns(api.Info{Version: "3.0.0"}, nil)

			err := executeCommand(command, []string{
				"--product", "postgres",
				"--distribution", "managed-k8s:1.234",
			})
			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("assign-kubernetes-distribution requires Ops Manager 3.3 or newer"),
				ContainSubstring("(current version: 3.0.0)"),
			)))

			Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
		})
	})

	When("determining the OpsManager version fails", func() {
		It("returns an error with a clear message", func() {
			fakeService.InfoReturns(api.Info{Version: "not-a-valid-semver"}, nil)

			err := executeCommand(command, []string{
				"--product", "postgres",
				"--distribution", "managed-k8s:1.234",
			})
			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("assign-kubernetes-distribution requires Ops Manager 3.3 or newer"),
				ContainSubstring("not-a-valid-semver"),
			)))

			Expect(fakeService.AssignKubernetesDistributionCallCount()).To(Equal(0))
		})
	})
})
