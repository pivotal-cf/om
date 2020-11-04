package configfetchers

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-09-01/network"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/soap"
	google "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

//go:generate counterfeiter -o ./fakes/export_opsman_config.go --fake-name OpsmanConfigFetcherService . OpsmanConfigFetcherService
type OpsmanConfigFetcherService interface {
	FetchConfig() (*vmmanagers.OpsmanConfigFilePayload, error)
}

type Credentials struct {
	AWS     *vmmanagers.AWSCredential
	GCP     *vmmanagers.GCPCredential
	VSphere *VCenterCredentialsWrapper
	Azure   *AzureCredentialsWrapper
}

type VCenterCredentialsWrapper struct {
	vmmanagers.VcenterCredential
	Insecure bool
}

type AzureCredentialsWrapper struct {
	vmmanagers.AzureCredential
	CloudName string
}

func NewOpsmanConfigFetcher(state *vmmanagers.StateInfo, creds *Credentials) (OpsmanConfigFetcherService, error) {
	err := validateState(state)
	if err != nil {
		return nil, err
	}

	switch state.IAAS {
	case "aws":
		err := validateAWSCreds(creds)
		if err != nil {
			return nil, err
		}

		config := &aws.Config{
			Region:                        &creds.AWS.Region,
			CredentialsChainVerboseErrors: aws.Bool(true),
		}

		if creds.AWS.AccessKeyId != "" && creds.AWS.SecretAccessKey != "" {
			config.Credentials = credentials.NewStaticCredentials(creds.AWS.AccessKeyId, creds.AWS.SecretAccessKey, "")
		}

		awsSession, err := session.NewSession(config)
		if err != nil {
			return nil, errors.New("could not create aws session")
		}

		return NewAWSConfigFetcher(state, creds, ec2.New(awsSession)), nil
	case "gcp":
		err := validateGCPCreds(creds)
		if err != nil {
			return nil, err
		}

		ctx := context.Background()

		var service *google.Service
		if creds.GCP.ServiceAccount != "" {
			service, err = google.NewService(ctx, option.WithCredentialsJSON([]byte(creds.GCP.ServiceAccount)))
			if err != nil {
				return nil, fmt.Errorf("could not create gcp client with service account json: %s", err)
			}
		} else {
			service, err = google.NewService(ctx)
			if err != nil {
				return nil, fmt.Errorf("could not create gcp client with service account: %s", err)
			}
		}

		return NewGCPConfigFetcher(state, creds, service), nil
	case "vsphere":
		err := validateVCenterCreds(creds)
		if err != nil {
			return nil, err
		}

		parsedURL, err := buildURL(creds)
		if err != nil {
			return nil, err
		}

		u, err := soap.ParseURL(parsedURL)
		if err != nil {
			return nil, fmt.Errorf("could not parse url: %s", err)
		}

		ctx := context.Background()
		client, err := govmomi.NewClient(ctx, u, creds.VSphere.Insecure)
		if err != nil {
			return nil, fmt.Errorf("could not create vcenter client: %s", err)
		}

		return NewVSphereConfigFetcher(state, creds, client), nil

	case "azure":
		err := validateAzureCreds(creds)
		if err != nil {
			return nil, err
		}

		vmClient, networkClient, ipClient, imageClient, err := createAzureClients(creds)
		if err != nil {
			return nil, err
		}

		return NewAzureConfigFetcher(state, creds, vmClient, networkClient, ipClient, imageClient), nil
	}

	return nil, errors.New("unexpected error creating a config fetcher. Please contact Pivotal Support")
}

func createAzureClients(creds *Credentials) (AzureVMClient, AzureNetworkClient, AzureIPClient, AzureImageClient, error) {
	subID := creds.Azure.SubscriptionID
	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("could not create azure auth settings: %s", err)
	}

	settings.Values[auth.SubscriptionID] = subID
	settings.Values[auth.TenantID] = creds.Azure.TenantID
	settings.Values[auth.ClientID] = creds.Azure.ClientID
	settings.Values[auth.ClientSecret] = creds.Azure.ClientSecret

	settings.Environment, err = azure.EnvironmentFromName(creds.Azure.CloudName)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("could not create azure environment from name: %s", err)
	}

	// Switch cloud names to work with Azure cli (from azure/environments.go in sdk)
	if strings.ToUpper(creds.Azure.CloudName) == "AZUREPUBLICCLOUD" {
		creds.Azure.CloudName = "AzureCloud"
	}

	if strings.ToUpper(creds.Azure.CloudName) == "AZUREUSGOVERNMENTCLOUD" {
		creds.Azure.CloudName = "AzureUSGovernment"
	}

	authorizer, err := settings.GetAuthorizer()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("could not create azure authorizer: %s", err)
	}

	vmClient := compute.NewVirtualMachinesClient(subID)
	vmClient.Authorizer = authorizer

	networkClient := network.NewInterfacesClient(subID)
	networkClient.Authorizer = authorizer

	ipClient := network.NewPublicIPAddressesClient(subID)
	ipClient.Authorizer = authorizer

	imageClient := compute.NewImagesClient(subID)
	imageClient.Authorizer = authorizer

	return vmClient, networkClient, ipClient, imageClient, nil
}

func validateState(state *vmmanagers.StateInfo) (err error) {
	supportedIAASes := []string{"aws", "azure", "gcp", "openstack", "vsphere"}

	var errString string
	if state.IAAS == "" {
		errString = "'iaas' is required in the provided state file\n"
	}

	if state.ID == "" {
		errString = errString + "'vm_id' is required in the provided state file\n"
	}

	var foundIAAS bool
	for _, iaas := range supportedIAASes {
		if iaas == state.IAAS {
			foundIAAS = true
		}
	}

	if !foundIAAS {
		errString = errString + fmt.Sprintf("IAAS: %s is not supported. Use %s\n", state.IAAS, strings.Join(supportedIAASes, "|"))
	}

	if errString != "" {
		err = errors.New(errString)
	}

	return err
}

func validateAWSCreds(creds *Credentials) error {
	if creds.AWS.Region == "" {
		return errors.New("the required flag '--aws-region' was not specified")
	}

	if creds.AWS.AccessKeyId != "" && creds.AWS.SecretAccessKey == "" || creds.AWS.AccessKeyId == "" && creds.AWS.SecretAccessKey != "" {
		return errors.New("both '--aws-access-key-id' and '--aws-secret-access-key' need to be specified if not using iam instance profiles")
	}

	return nil
}

func validateGCPCreds(creds *Credentials) (err error) {
	var errs []string
	if creds.GCP.Zone == "" {
		errs = append(errs, "the required flag '--gcp-zone' was not specified\n")
	}

	if creds.GCP.Project == "" {
		errs = append(errs, "the required flag '--gcp-project-id' was not specified\n")
	}

	if creds.GCP.ServiceAccount != "" {
		if _, err := os.Stat(creds.GCP.ServiceAccount); os.IsNotExist(err) {
			errs = append(errs, fmt.Sprintf("gcp-service-account-json file (%s) cannot be found", creds.GCP.ServiceAccount))

		}

		contents, err := ioutil.ReadFile(creds.GCP.ServiceAccount)
		if err != nil {
			errs = append(errs, fmt.Sprintf("could not read gcp-service-account-json file (%s): %s", creds.GCP.ServiceAccount, err))
		}
		creds.GCP.ServiceAccount = string(contents)
	}

	if len(errs) > 0 {
		err = fmt.Errorf(strings.Join(errs, "\n"))
	}

	return err
}

func validateVCenterCreds(creds *Credentials) (err error) {
	var errs []string
	if creds.VSphere.URL == "" {
		errs = append(errs, "the required flag '--vsphere-url' was not specified\n")
	}

	if creds.VSphere.Username == "" {
		errs = append(errs, "the required flag '--vsphere-username' was not specified\n")
	}

	if creds.VSphere.Password == "" {
		errs = append(errs, "the required flag '--vsphere-password' was not specified\n")
	}

	if len(errs) > 0 {
		err = fmt.Errorf(strings.Join(errs, "\n"))
	}

	return err
}

func validateAzureCreds(creds *Credentials) (err error) {
	var errs []string
	if creds.Azure.SubscriptionID == "" {
		errs = append(errs, "the required flag '--azure-subscription-id' was not specified\n")
	}
	if creds.Azure.TenantID == "" {
		errs = append(errs, "the required flag '--azure-tenant-id' was not specified\n")
	}
	if creds.Azure.ClientID == "" {
		errs = append(errs, "the required flag '--azure-client-id' was not specified\n")
	}
	if creds.Azure.ClientSecret == "" {
		errs = append(errs, "the required flag '--azure-client-secret' was not specified\n")
	}
	if creds.Azure.ResourceGroup == "" {
		errs = append(errs, "the required flag '--azure-resource-group' was not specified\n")
	}

	if len(errs) > 0 {
		err = fmt.Errorf(strings.Join(errs, "\n"))
	}

	return err
}

func buildURL(creds *Credentials) (string, error) {
	parsedURL, err := url.Parse(creds.VSphere.URL)
	if err != nil {
		return "", fmt.Errorf("the '--vsphere-url=%s' was not provided with the correct format, like https://vcenter.example.com", creds.VSphere.URL)
	}

	if parsedURL.Scheme == "" {
		return "", fmt.Errorf("the '--vsphere-url=%s' was not supplied a protocol (http or https), like https://vcenter.example.com", creds.VSphere.URL)
	}

	return fmt.Sprintf("%s://%s:%s@%s/sdk", parsedURL.Scheme, creds.VSphere.Username, creds.VSphere.Password, parsedURL.Host), nil
}
