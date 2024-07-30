package configfetchers_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"

	"github.com/pivotal-cf/om/vmlifecycle/configfetchers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
)

var _ = Describe("gcp", func() {
	When("the api returns valid responses", func() {
		var (
			state   *vmmanagers.StateInfo
			service *compute.Service
			server  *ghttp.Server
		)

		BeforeEach(func() {
			var err error
			server = ghttp.NewServer()

			ctx := context.Background()
			service, err = compute.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(server.URL()))
			Expect(err).ToNot(HaveOccurred())

			state = &vmmanagers.StateInfo{
				IAAS: "gcp",
				ID:   "some-vm-name",
			}

			server.AppendHandlers(
				ghttp.RespondWith(http.StatusOK, `{
					"machineType": "https://www.googleapis.com/compute/v1/projects/some-project-id/zones/some-region-a/machineTypes/n1-highmem-2",
					"name": "opsman-vm",
					"disks": [{
						"source": "https://www.googleapis.com/compute/v1/projects/some-project-id/zones/some-region-a/disks/some-disk-id"
					}],
					"networkInterfaces": [{
	 					"accessConfigs": [{
		   					"natIP": "1.2.3.4"
						}],
	 					"networkIP": "5.6.7.8",
	 					"subnetwork": "https://www.googleapis.com/compute/v1/projects/some-project-id/regions/some-region-a/subnetworks/some-subnet-id"
  					}],
					"serviceAccounts":[{
	 					"scopes": [
							"some-scope"
	 					]
					}],
					"tags": {
  						"items": [
	 						"ops-manager",
	 						"some-tag"
  						]
					},
					"zone": "some-region-a"
				}`),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects/some-project-id/zones/some-region-a/machineTypes/n1-highmem-2"),
					ghttp.RespondWith(http.StatusOK, `{
						"guestCpus": 1,
						"memoryMb": 3840
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects/some-project-id/zones/some-region-a/disks/some-disk-id"),
					ghttp.RespondWith(http.StatusOK, `{
						"sizeGb": "160"
					}`),
				),
			)
		})

		AfterEach(func() {
			server.Close()
		})

		When("the credentials passed in contain a service account json", func() {
			It("creates an opsman.yml that includes placeholders for credentials", func() {
				creds := &configfetchers.Credentials{
					GCP: &vmmanagers.GCPCredential{
						ServiceAccount: `{}`,
						Project:        "some-project-id",
						Zone:           "some-region-a",
					},
				}

				fetcher := configfetchers.NewGCPConfigFetcher(state, creds, service)

				output, err := fetcher.FetchConfig()
				Expect(err).ToNot(HaveOccurred())

				Expect(output).To(Equal(&vmmanagers.OpsmanConfigFilePayload{
					OpsmanConfig: vmmanagers.OpsmanConfig{
						GCP: &vmmanagers.GCPConfig{
							GCPCredential: vmmanagers.GCPCredential{
								ServiceAccount: "((gcp-service-account-json))",
								Project:        "some-project-id",
								Region:         "some-region",
								Zone:           "some-region-a",
							},
							VpcSubnet:    "projects/some-project-id/regions/some-region-a/subnetworks/some-subnet-id",
							PublicIP:     "1.2.3.4",
							PrivateIP:    "5.6.7.8",
							VMName:       "opsman-vm",
							Tags:         "ops-manager,some-tag",
							CPU:          "1",
							Memory:       "3840MB",
							BootDiskSize: "160GB",
							Scopes:       []string{"some-scope"},
							SSHPublicKey: "((ssh-public-key))",
						},
					},
				}))
			})
		})

		When("the credentials passed in do not contain a service account key", func() {
			It("creates an opsman.yml that includes placeholders for credentials", func() {
				creds := &configfetchers.Credentials{
					GCP: &vmmanagers.GCPCredential{
						Project: "some-project-id",
						Zone:    "some-region-a",
					},
				}

				fetcher := configfetchers.NewGCPConfigFetcher(state, creds, service)

				output, err := fetcher.FetchConfig()
				Expect(err).ToNot(HaveOccurred())

				Expect(output).To(Equal(&vmmanagers.OpsmanConfigFilePayload{
					OpsmanConfig: vmmanagers.OpsmanConfig{
						GCP: &vmmanagers.GCPConfig{
							GCPCredential: vmmanagers.GCPCredential{
								ServiceAccountName: "((gcp-service-account-name))",
								Project:            "some-project-id",
								Region:             "some-region",
								Zone:               "some-region-a",
							},
							VpcSubnet:    "projects/some-project-id/regions/some-region-a/subnetworks/some-subnet-id",
							PublicIP:     "1.2.3.4",
							PrivateIP:    "5.6.7.8",
							VMName:       "opsman-vm",
							Tags:         "ops-manager,some-tag",
							CPU:          "1",
							Memory:       "3840MB",
							BootDiskSize: "160GB",
							Scopes:       []string{"some-scope"},
							SSHPublicKey: "((ssh-public-key))",
						},
					},
				}))
			})
		})
	})

	When("the api response is missing required values", func() {
		DescribeTable("returns an error", func(payload, expectedError string) {
			server := ghttp.NewServer()
			defer server.Close()

			ctx := context.Background()
			service, err := compute.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(server.URL()))
			Expect(err).ToNot(HaveOccurred())

			state := &vmmanagers.StateInfo{
				IAAS: "gcp",
				ID:   "some-vm-name",
			}

			server.AppendHandlers(
				ghttp.RespondWith(http.StatusOK, payload),
			)

			creds := &configfetchers.Credentials{
				GCP: &vmmanagers.GCPCredential{
					ServiceAccount: `{}`,
					Project:        "some-project-id",
					Zone:           "some-region-a",
				},
			}

			fetcher := configfetchers.NewGCPConfigFetcher(state, creds, service)

			_, err = fetcher.FetchConfig()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring(expectedError)))
		},
			Entry("there are no disks", `{"disks":[]}`, "expected a boot disk to be attached to the VM"),
			Entry("there are no network interfaces", `{
				"disks":[{}],
				"networkInterfaces":[]
			}`, "expected a network interface to be attached to the VM"),
		)
	})

	When("the api is missing optional values", func() {
		It("returns a valid config with those properties empty", func() {
			server := ghttp.NewServer()
			defer server.Close()

			ctx := context.Background()
			service, err := compute.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(server.URL()))
			Expect(err).ToNot(HaveOccurred())

			state := &vmmanagers.StateInfo{
				IAAS: "gcp",
				ID:   "some-vm-name",
			}

			server.AppendHandlers(
				ghttp.RespondWith(http.StatusOK, `{
					"machineType": "https://www.googleapis.com/compute/v1/projects/some-project-id/zones/some-region-a/machineTypes/n1-highmem-2",
					"name": "opsman-vm",
					"disks": [{
						"source": "https://www.googleapis.com/compute/v1/projects/some-project-id/zones/some-region-a/disks/some-disk-id"
					}],
					"networkInterfaces": [{
	 					"networkIP": "5.6.7.8",
	 					"subnetwork": "some-subnet"
  					}],
					"zone": "some-region-a"
				}`),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects/some-project-id/zones/some-region-a/machineTypes/n1-highmem-2"),
					ghttp.RespondWith(http.StatusOK, `{
						"guestCpus": 1,
						"memoryMb": 3840
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects/some-project-id/zones/some-region-a/disks/some-disk-id"),
					ghttp.RespondWith(http.StatusOK, `{
						"sizeGb": "160"
					}`),
				),
			)

			creds := &configfetchers.Credentials{
				GCP: &vmmanagers.GCPCredential{
					ServiceAccount: `{}`,
					Project:        "some-project-id",
					Zone:           "some-region-a",
				},
			}

			fetcher := configfetchers.NewGCPConfigFetcher(state, creds, service)

			output, err := fetcher.FetchConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(&vmmanagers.OpsmanConfigFilePayload{
				OpsmanConfig: vmmanagers.OpsmanConfig{
					GCP: &vmmanagers.GCPConfig{
						GCPCredential: vmmanagers.GCPCredential{
							ServiceAccount: "((gcp-service-account-json))",
							Project:        "some-project-id",
							Region:         "some-region",
							Zone:           "some-region-a",
						},
						VpcSubnet:    "some-subnet",
						PrivateIP:    "5.6.7.8",
						VMName:       "opsman-vm",
						CPU:          "1",
						Memory:       "3840MB",
						BootDiskSize: "160GB",
						SSHPublicKey: "((ssh-public-key))",
					},
				},
			}))
		})
	})

	When("the API calls return http errors", func() {
		It("returns an error", func() {
			server := ghttp.NewServer()
			defer server.Close()

			ctx := context.Background()
			service, err := compute.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(server.URL()))
			Expect(err).ToNot(HaveOccurred())

			state := &vmmanagers.StateInfo{
				IAAS: "gcp",
				ID:   "some-vm-name",
			}

			server.AppendHandlers(
				ghttp.RespondWith(http.StatusTeapot, `{}`),
				ghttp.RespondWith(http.StatusOK, `{"disks": [{}], "networkInterfaces": [{"networkIP": "5.6.7.8"}]}`),
				ghttp.RespondWith(http.StatusTeapot, `{}`),
				ghttp.RespondWith(http.StatusOK, `{"disks": [{}], "networkInterfaces": [{"networkIP": "5.6.7.8"}]}`),
				ghttp.RespondWith(http.StatusOK, `{"guestCpus": 1, "memoryMb": 3840}`),
				ghttp.RespondWith(http.StatusTeapot, `{}`),
			)

			creds := &configfetchers.Credentials{
				GCP: &vmmanagers.GCPCredential{
					ServiceAccount: `{}`,
					Project:        "some-project-id",
					Zone:           "some-region-a",
				},
			}

			fetcher := configfetchers.NewGCPConfigFetcher(state, creds, service)

			_, err = fetcher.FetchConfig()
			Expect(err).To(MatchError(ContainSubstring("could not fetch instance data")))

			_, err = fetcher.FetchConfig()
			Expect(err).To(MatchError(ContainSubstring("could not fetch machine type data")))

			_, err = fetcher.FetchConfig()
			Expect(err).To(MatchError(ContainSubstring("could not fetch disk data")))
		})
	})
})
