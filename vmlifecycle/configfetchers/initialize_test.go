package configfetchers_test

import (
	"os"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware/govmomi/simulator"

	"github.com/pivotal-cf/om/vmlifecycle/configfetchers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
)

var _ = Describe("selects the correct config fetcher based on the state file", func() {
	Context("aws", func() {
		When("the config is valid", func() {
			When("no credentials are passed in", func() {
				It("returns the aws config fetcher", func() {
					state := &vmmanagers.StateInfo{
						IAAS: "aws",
						ID:   "some-vm-id",
					}

					creds := &configfetchers.Credentials{
						AWS: &vmmanagers.AWSCredential{
							Region: "some-region",
						},
					}

					awsConfigFetcher, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
					Expect(err).ToNot(HaveOccurred())

					Expect(reflect.TypeOf(awsConfigFetcher)).To(Equal(reflect.TypeOf(&configfetchers.AWSConfigFetcher{})))
				})
			})

			It("returns the aws config fetcher", func() {
				state := &vmmanagers.StateInfo{
					IAAS: "aws",
					ID:   "some-vm-id",
				}

				creds := &configfetchers.Credentials{
					AWS: &vmmanagers.AWSCredential{
						Region:          "some-region",
						SecretAccessKey: "secret-access-key",
						AccessKeyId:     "access-key",
					},
				}

				awsConfigFetcher, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).ToNot(HaveOccurred())

				Expect(reflect.TypeOf(awsConfigFetcher)).To(Equal(reflect.TypeOf(&configfetchers.AWSConfigFetcher{})))
			})
		})

		When("aws access key and secret access key are not both set or both unset", func() {
			It("returns an error", func() {
				state := &vmmanagers.StateInfo{
					ID:   "some-vm-id",
					IAAS: "aws",
				}

				creds := &configfetchers.Credentials{
					AWS: &vmmanagers.AWSCredential{
						AccessKeyId:     "key",
						SecretAccessKey: "",
						Region:          "some-region",
					},
				}

				expectedError := "both '--aws-access-key-id' and '--aws-secret-access-key' need to be specified if not using iam instance profiles"

				_, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedError))

				creds = &configfetchers.Credentials{
					AWS: &vmmanagers.AWSCredential{
						AccessKeyId:     "",
						SecretAccessKey: "secret",
						Region:          "some-region",
					},
				}

				_, err = configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedError))
			})
		})

		When("region is not set in the credentials", func() {
			It("returns an error", func() {
				state := &vmmanagers.StateInfo{
					ID:   "some-vm-id",
					IAAS: "aws",
				}

				creds := &configfetchers.Credentials{
					AWS: &vmmanagers.AWSCredential{
						AccessKeyId:     "key",
						SecretAccessKey: "secret",
						Region:          "",
					},
				}

				_, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("the required flag '--aws-region' was not specified"))
			})
		})
	})

	Context("gcp", func() {
		When("the config is valid", func() {
			When("the gcp service account is set", func() {
				It("returns the gcp config fetcher", func() {
					state := &vmmanagers.StateInfo{
						IAAS: "gcp",
						ID:   "some-vm-name",
					}

					creds := &configfetchers.Credentials{
						GCP: &vmmanagers.GCPCredential{
							ServiceAccount: writeFile(`{
      						  	"type": "service_account",
      						  	"project_id": "some-project-id",
      						  	"private_key_id": "private-key-id",
      						  	"private_key": "-----BEGIN PRIVATE KEY-----fake-key-----END PRIVATE KEY-----\n",
      						  	"client_email": "user@some-project-id.iam.gserviceaccount.com",
      						  	"client_id": "123456789098765432123",
      						  	"auth_uri": "https://accounts.google.com/o/oauth2/auth",
      						  	"token_uri": "https://accounts.google.com/o/oauth2/token",
      						  	"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
      						  	"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/user%40some-project-id.iam.gserviceaccount.com"
      						}`),
							Project: "some-project-id",
							Zone:    "some-zone",
						},
					}

					gcpConfigFetcher, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
					Expect(err).ToNot(HaveOccurred())

					Expect(reflect.TypeOf(gcpConfigFetcher)).To(Equal(reflect.TypeOf(&configfetchers.GCPConfigFetcher{})))
				})
			})

			When("the gcp account name is set", func() {
				It("returns the gcp config fetcher", func() {
					err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", writeFile(`{
						"type": "service_account",
						"project_id": "some-project-id",
						"private_key_id": "private-key-id",
						"private_key": "-----BEGIN PRIVATE KEY-----fake-key-----END PRIVATE KEY-----\n",
						"client_email": "user@some-project-id.iam.gserviceaccount.com",
						"client_id": "123456789098765432123",
						"auth_uri": "https://accounts.google.com/o/oauth2/auth",
						"token_uri": "https://accounts.google.com/o/oauth2/token",
						"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
						"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/user%40some-project-id.iam.gserviceaccount.com"
					}`))
					Expect(err).ToNot(HaveOccurred())
					defer os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")

					state := &vmmanagers.StateInfo{
						IAAS: "gcp",
						ID:   "some-vm-name",
					}

					creds := &configfetchers.Credentials{
						GCP: &vmmanagers.GCPCredential{
							Project: "some-project-id",
							Zone:    "some-zone",
						},
					}

					gcpConfigFetcher, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
					Expect(err).ToNot(HaveOccurred())

					Expect(reflect.TypeOf(gcpConfigFetcher)).To(Equal(reflect.TypeOf(&configfetchers.GCPConfigFetcher{})))
				})
			})
		})

		When("required properties are not set in the credentials", func() {
			It("returns an error", func() {
				state := &vmmanagers.StateInfo{
					ID:   "some-vm-name",
					IAAS: "gcp",
				}

				creds := &configfetchers.Credentials{
					GCP: &vmmanagers.GCPCredential{},
				}

				_, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("the required flag '--gcp-zone' was not specified"))
				Expect(err.Error()).To(ContainSubstring("the required flag '--gcp-project-id' was not specified"))
			})
		})

		When("the gcp service account file is not valid json", func() {
			It("returns an error", func() {
				state := &vmmanagers.StateInfo{
					IAAS: "gcp",
					ID:   "some-vm-name",
				}

				creds := &configfetchers.Credentials{
					GCP: &vmmanagers.GCPCredential{
						ServiceAccount: writeFile(`{
      					  	invalid-json
      					}`),
						Project: "some-project-id",
						Zone:    "some-zone",
					},
				}

				_, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).To(HaveOccurred())
			})
		})

		When("the gcp service account file does not exist", func() {
			It("returns an error", func() {
				state := &vmmanagers.StateInfo{
					IAAS: "gcp",
					ID:   "some-vm-name",
				}

				creds := &configfetchers.Credentials{
					GCP: &vmmanagers.GCPCredential{
						ServiceAccount: "never-gonna-give-you-up.txt",
						Project:        "some-project-id",
						Zone:           "some-zone",
					},
				}

				_, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).To(MatchError(ContainSubstring("gcp-service-account-json file (never-gonna-give-you-up.txt) cannot be found")))
			})
		})
	})

	Context("vsphere", func() {
		When("the config is valid", func() {
			It("returns the vsphere config fetcher", func() {
				state := &vmmanagers.StateInfo{
					IAAS: "vsphere",
					ID:   "/DC0/vm/DC0_H0_VM0",
				}

				model := simulator.VPX()
				err := model.Create()
				Expect(err).ToNot(HaveOccurred())

				model.Service.TLS = nil
				s := model.Service.NewServer()
				defer s.Close()

				creds := &configfetchers.Credentials{
					VSphere: &configfetchers.VCenterCredentialsWrapper{
						VcenterCredential: vmmanagers.VcenterCredential{
							URL:      s.URL.String(),
							Username: "some-username",
							Password: "some-password",
						},
						Insecure: true,
					},
				}

				vSphereConfigFetcher, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).ToNot(HaveOccurred())

				Expect(reflect.TypeOf(vSphereConfigFetcher)).To(Equal(reflect.TypeOf(&configfetchers.VSphereConfigFetcher{})))
			})
		})

		When("the username and password include URI-reserved characters", func() {
			It("URI-encodes them for use in requests in the fetcher", func() {
				state := &vmmanagers.StateInfo{
					IAAS: "vsphere",
					ID:   "/DC0/vm/DC0_H0_VM0",
				}

				model := simulator.VPX()
				err := model.Create()
				Expect(err).ToNot(HaveOccurred())

				model.Service.TLS = nil
				s := model.Service.NewServer()
				defer s.Close()

				creds := &configfetchers.Credentials{
					VSphere: &configfetchers.VCenterCredentialsWrapper{
						VcenterCredential: vmmanagers.VcenterCredential{
							URL:      s.URL.String(),
							Username: `some\username`,
							Password: "some-password-with-ampersand-&",
						},
						Insecure: true,
					},
				}

				vSphereConfigFetcher, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).ToNot(HaveOccurred())

				Expect(reflect.TypeOf(vSphereConfigFetcher)).To(Equal(reflect.TypeOf(&configfetchers.VSphereConfigFetcher{})))
			})
		})

		When("the insecure value is set to false and the server url is https", func() {
			It("does not return an error", func() {
				state := &vmmanagers.StateInfo{
					IAAS: "vsphere",
					ID:   "/DC0/vm/DC0_H0_VM0",
				}

				model := simulator.VPX()
				err := model.Create()
				Expect(err).ToNot(HaveOccurred())

				s := model.Service.NewServer()
				defer s.Close()

				creds := &configfetchers.Credentials{
					VSphere: &configfetchers.VCenterCredentialsWrapper{
						VcenterCredential: vmmanagers.VcenterCredential{
							URL:      s.URL.String(),
							Username: "some-username",
							Password: "some-password",
						},
						Insecure: true,
					},
				}

				_, err = configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("required properties are not set in the credentials", func() {
			It("returns an error", func() {
				state := &vmmanagers.StateInfo{
					IAAS: "vsphere",
					ID:   "/DC0/vm/DC0_H0_VM0",
				}

				model := simulator.VPX()
				err := model.Create()
				Expect(err).ToNot(HaveOccurred())

				model.Service.TLS = nil
				s := model.Service.NewServer()
				defer s.Close()

				creds := &configfetchers.Credentials{
					VSphere: &configfetchers.VCenterCredentialsWrapper{
						Insecure: false,
					},
				}

				_, err = configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).To(MatchError(ContainSubstring("the required flag '--vsphere-url' was not specified")))
				Expect(err).To(MatchError(ContainSubstring("the required flag '--vsphere-password' was not specified")))
				Expect(err).To(MatchError(ContainSubstring("the required flag '--vsphere-username' was not specified")))
			})
		})

		When("the vcenter url cannot be parsed", func() {
			It("returns an error", func() {
				state := &vmmanagers.StateInfo{
					IAAS: "vsphere",
					ID:   "/DC0/vm/DC0_H0_VM0",
				}

				model := simulator.VPX()
				err := model.Create()
				Expect(err).ToNot(HaveOccurred())

				model.Service.TLS = nil
				s := model.Service.NewServer()
				defer s.Close()

				creds := &configfetchers.Credentials{
					VSphere: &configfetchers.VCenterCredentialsWrapper{
						VcenterCredential: vmmanagers.VcenterCredential{
							URL:      "%%%",
							Username: "username",
							Password: "password",
						},
						Insecure: false,
					},
				}

				_, err = configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).To(MatchError(ContainSubstring("the '--vsphere-url=%%%' was not provided with the correct format, like https://vcenter.example.com")))
			})
		})

		When("the vcenter url does not specify a protocol", func() {
			It("returns an error", func() {
				state := &vmmanagers.StateInfo{
					IAAS: "vsphere",
					ID:   "/DC0/vm/DC0_H0_VM0",
				}

				model := simulator.VPX()
				err := model.Create()
				Expect(err).ToNot(HaveOccurred())

				model.Service.TLS = nil
				s := model.Service.NewServer()
				defer s.Close()

				creds := &configfetchers.Credentials{
					VSphere: &configfetchers.VCenterCredentialsWrapper{
						VcenterCredential: vmmanagers.VcenterCredential{
							URL:      "example.org",
							Username: "username",
							Password: "password",
						},
						Insecure: false,
					},
				}

				_, err = configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).To(MatchError(ContainSubstring("the '--vsphere-url=example.org' was not supplied a protocol (http or https), like https://vcenter.example.com")))
			})
		})
	})

	Context("azure", func() {
		When("the config is valid", func() {
			It("returns the azure config fetcher", func() {
				state := &vmmanagers.StateInfo{
					IAAS: "azure",
					ID:   "some-vm-id",
				}

				creds := &configfetchers.Credentials{
					Azure: &configfetchers.AzureCredentialsWrapper{
						AzureCredential: vmmanagers.AzureCredential{
							TenantID:       "some-tenant-id",
							SubscriptionID: "some-subscription-id",
							ClientID:       "some-client-id",
							ClientSecret:   "some-client-secret",
							ResourceGroup:  "some-resource-group",
						},
						CloudName: "AzurePublicCloud",
					},
				}

				azureConfigFetcher, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).ToNot(HaveOccurred())

				Expect(reflect.TypeOf(azureConfigFetcher)).To(Equal(reflect.TypeOf(&configfetchers.AzureConfigFetcher{})))
			})
		})
		When("the config does not contain all needed credentials", func() {
			It("gives an error specifying the needed information", func() {
				state := &vmmanagers.StateInfo{
					IAAS: "azure",
					ID:   "some-vm-id",
				}

				creds := &configfetchers.Credentials{
					Azure: &configfetchers.AzureCredentialsWrapper{
						AzureCredential: vmmanagers.AzureCredential{},
					},
				}

				_, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("the required flag '--azure-subscription-id' was not specified"))
				Expect(err.Error()).To(ContainSubstring("the required flag '--azure-tenant-id' was not specified"))
				Expect(err.Error()).To(ContainSubstring("the required flag '--azure-client-id' was not specified"))
				Expect(err.Error()).To(ContainSubstring("the required flag '--azure-client-secret' was not specified"))
				Expect(err.Error()).To(ContainSubstring("the required flag '--azure-resource-group' was not specified"))
			})
		})
	})

	Context("state", func() {
		When("the state file doesn't have an ID", func() {
			It("returns an error", func() {
				state := &vmmanagers.StateInfo{
					IAAS: "aws",
				}

				creds := &configfetchers.Credentials{
					AWS: &vmmanagers.AWSCredential{
						Region: "some-region",
					},
				}

				_, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("'vm_id' is required in the provided state file"))
			})
		})

		When("the state file doesn't have an IAAS defined", func() {
			It("returns an error", func() {
				state := &vmmanagers.StateInfo{
					ID: "some-vm-id",
				}

				creds := &configfetchers.Credentials{
					AWS: &vmmanagers.AWSCredential{
						Region: "some-region",
					},
				}

				_, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("'iaas' is required in the provided state file"))
			})
		})

		When("the state file doesn't have a supported IAAS", func() {
			It("returns an error", func() {
				state := &vmmanagers.StateInfo{
					ID:   "some-vm-id",
					IAAS: "unsupported-iaas",
				}

				creds := &configfetchers.Credentials{
					AWS: &vmmanagers.AWSCredential{
						Region: "some-region",
					},
				}

				_, err := configfetchers.NewOpsmanConfigFetcher(state, creds)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("IAAS: unsupported-iaas is not supported. Use aws|azure|gcp|openstack|vsphere"))
			})
		})
	})
})

func writeFile(contents string) string {
	tempfile, err := os.CreateTemp("", "some*.yaml")
	Expect(err).ToNot(HaveOccurred())
	_, err = tempfile.WriteString(contents)
	Expect(err).ToNot(HaveOccurred())
	err = tempfile.Close()
	Expect(err).ToNot(HaveOccurred())

	return tempfile.Name()
}
