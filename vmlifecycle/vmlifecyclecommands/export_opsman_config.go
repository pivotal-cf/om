package vmlifecyclecommands

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/om/vmlifecycle/configfetchers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
)

type initExportOpsmanConfigFunc func(state *vmmanagers.StateInfo, creds *configfetchers.Credentials) (configfetchers.OpsmanConfigFetcherService, error)

type ExportOpsmanConfig struct {
	stdout                io.Writer
	stderr                io.Writer
	initService           initExportOpsmanConfigFunc
	StateFile             string `long:"state-file"               description:"File that contains the VM identifier info" required:"true"`
	ConfigFile            string `long:"config-file"              description:"File to write the opsman config to" default:"opsman.yml"`
	AWSRegion             string `long:"aws-region"               description:"Region that contains the VM defined in the state file"`
	AWSSecretAccessKey    string `long:"aws-secret-access-key"    description:"NOTE: only required if not using instance profile, AWS_SECRET_ACCESS_KEY"`
	AWSAccessKey          string `long:"aws-access-key-id"        description:"NOTE: only required if not using instance profile, AWS_ACCESS_KEY_ID"`
	AzureSubscriptionID   string `long:"azure-subscription-id"    description:"The subscription ID for target Azure account"`
	AzureTenantID         string `long:"azure-tenant-id"          description:"Tenant ID for the target Azure environment"`
	AzureClientID         string `long:"azure-client-id"          description:"The application (client) ID defined in Azure"`
	AzureClientSecret     string `long:"azure-client-secret"      description:"The application (client) secret defined in Azure"`
	AzureResourceGroup    string `long:"azure-resource-group"     description:"The resource group that contains the VM defined in the state file"`
	AzureCloudName        string `long:"azure-cloud-name"         description:"The azure environment. Valid: (AzureChinaCloud, AzureGermanCloud, AzurePublicCloud, or AzureUSGovernmentCloud)" default:"AzurePublicCloud"`
	GCPServiceAccountJSON string `long:"gcp-service-account-json" description:"File that contains the service account json"`
	GCPProjectID          string `long:"gcp-project-id"           description:"The GCP project id that contains the VM defined in the state file"`
	GCPZone               string `long:"gcp-zone"                 description:"The zone that contains the VM defined in the state file"`
	VSphereURL            string `long:"vsphere-url"              description:"The vCenter server URL"`
	VSphereUsername       string `long:"vsphere-username"         description:"The vCenter server username for login"`
	VSpherePassword       string `long:"vsphere-password"         description:"The vCenter server password for login"`
	VSphereInsecure       bool   `long:"vsphere-insecure"         description:"If set, skip verification of the vSphere endpoint. Not recommended!"`
}

func NewExportOpsmanConfigCommand(stdout, stderr io.Writer, initService initExportOpsmanConfigFunc) ExportOpsmanConfig {
	return ExportOpsmanConfig{
		stdout:      stdout,
		stderr:      stderr,
		initService: initService,
	}
}

func (e *ExportOpsmanConfig) Execute(args []string) error {
	state := vmmanagers.StateInfo{}
	content, err := os.ReadFile(e.StateFile)
	if err != nil {
		return fmt.Errorf("could not read state file %s", err)
	}

	err = yaml.Unmarshal(content, &state)
	if err != nil {
		return fmt.Errorf("could not load state file (%s): %s", e.StateFile, err)
	}

	creds := &configfetchers.Credentials{
		AWS: &vmmanagers.AWSCredential{
			AccessKeyId:     e.AWSAccessKey,
			SecretAccessKey: e.AWSSecretAccessKey,
			Region:          e.AWSRegion,
		},
		GCP: &vmmanagers.GCPCredential{
			ServiceAccount: e.GCPServiceAccountJSON,
			Project:        e.GCPProjectID,
			Zone:           e.GCPZone,
		},
		VSphere: &configfetchers.VCenterCredentialsWrapper{
			VcenterCredential: vmmanagers.VcenterCredential{
				URL:      e.VSphereURL,
				Username: e.VSphereUsername,
				Password: e.VSpherePassword,
			},
			Insecure: e.VSphereInsecure,
		},
		Azure: &configfetchers.AzureCredentialsWrapper{
			AzureCredential: vmmanagers.AzureCredential{
				TenantID:       e.AzureTenantID,
				SubscriptionID: e.AzureSubscriptionID,
				ClientID:       e.AzureClientID,
				ClientSecret:   e.AzureClientSecret,
				ResourceGroup:  e.AzureResourceGroup,
			},
			CloudName: e.AzureCloudName,
		},
	}

	configFetcherService, err := e.initService(&state, creds)
	if err != nil {
		return err
	}

	configOutput, err := configFetcherService.FetchConfig()
	if err != nil {
		return err
	}

	err = e.writeConfigFile(configOutput)
	if err != nil {
		return err
	}

	return nil
}

func (e *ExportOpsmanConfig) writeConfigFile(opsmanConfigFilePayload *vmmanagers.OpsmanConfigFilePayload) error {
	bytes, err := yaml.Marshal(opsmanConfigFilePayload)
	if err != nil {
		return fmt.Errorf("could not marshal the fetch opsman config: %s", err)
	}

	err = os.WriteFile(e.ConfigFile, bytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not write the opsman config file: %s", err)
	}

	_, err = fmt.Fprintf(e.stdout, "successfully wrote the Ops Manager config file to: %s", e.ConfigFile)
	if err != nil {
		return err
	}

	return nil
}
