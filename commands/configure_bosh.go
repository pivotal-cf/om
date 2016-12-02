package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-querystring/query"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

const iaasConfigurationPath = "/infrastructure/iaas_configuration/edit"

type ConfigureBosh struct {
	service boshFormService
	logger  logger
	Options struct {
		IAASConfiguration string `short:"i"  long:"iaas-configuration"  description:"iaas specific configuration for the bosh director"`
	}
}

type IAASConfiguration struct {
	// GCP-only configurations
	Project              string `url:"iaas_configuration[project],omitempty" json:"project"`
	DefaultDeploymentTag string `url:"iaas_configuration[default_deployment_tag],omitempty" json:"default_deployment_tag"`
	AuthJSON             string `url:"iaas_configuration[auth_json],omitempty" json:"auth_json"`

	// Azure-only configurations
	SubscriptionID                string `url:"iaas_configuration[subscription_id],omitempty" json:"subscription_id"`
	TenantID                      string `url:"iaas_configuration[tenant_id],omitempty" json:"tenant_id"`
	ClientID                      string `url:"iaas_configuration[client_id],omitempty" json:"client_id"`
	ClientSecret                  string `url:"iaas_configuration[client_secret],omitempty" json:"client_secret"`
	ResourceGroupName             string `url:"iaas_configuration[resource_group_name],omitempty" json:"resource_group_name"`
	BoshStorageAccountName        string `url:"iaas_configuration[bosh_storage_account_name],omitempty" json:"bosh_storage_account_name"`
	DefaultSecurityGroup          string `url:"iaas_configuration[default_security_group],omitempty" json:"default_security_group"`
	SSHPublicKey                  string `url:"iaas_configuration[ssh_public_key],omitempty" json:"ssh_public_key"`
	DeploymentsStorageAccountName string `url:"iaas_configuration[deployments_storage_account_name],omitempty" json:"deployments_storage_account_name"`

	// AWS-only configurations
	AccessKeyID     string `url:"iaas_configuration[access_key_id],omitempty" json:"access_key_id"`
	SecretAccessKey string `url:"iaas_configuration[secret_access_key],omitempty" json:"secret_access_key"`
	VpcID           string `url:"iaas_configuration[vpc_id],omitempty" json:"vpc_id"`
	SecurityGroup   string `url:"iaas_configuration[security_group],omitempty" json:"security_group"`
	KeyPairName     string `url:"iaas_configuration[key_pair_name],omitempty" json:"key_pair_name"`
	Region          string `url:"iaas_configuration[region],omitempty" json:"region"`
	Encrypted       string `url:"iaas_configuration[encrypted],omitempty" json:"encrypted"`

	SSHPrivateKey     string `url:"iaas_configuration[ssh_private_key],omitempty" json:"ssh_private_key"`
	AuthenticityToken string `url:"authenticity_token"`
	Method            string `url:"_method"`
}

//go:generate counterfeiter -o ./fakes/bosh_form_service.go --fake-name BoshFormService . boshFormService
type boshFormService interface {
	GetForm(path string) (api.Form, error)
	ConfigureIAAS(api.ConfigureIAASInput) error
}

func NewConfigureBosh(s boshFormService, l logger) ConfigureBosh {
	return ConfigureBosh{service: s, logger: l}
}

func (c ConfigureBosh) Execute(args []string) error {
	_, err := flags.Parse(&c.Options, args)
	if err != nil {
		return err
	}

	c.logger.Printf("configuring iaas specific options for bosh tile")

	form, err := c.service.GetForm(iaasConfigurationPath)
	if err != nil {
		return fmt.Errorf("could not fetch form: %s", err)
	}

	var initialConfig *IAASConfiguration
	err = json.NewDecoder(strings.NewReader(c.Options.IAASConfiguration)).Decode(&initialConfig)
	if err != nil {
		return fmt.Errorf("could not decode json: %s", err)
	}

	initialConfig.AuthenticityToken = form.AuthenticityToken
	initialConfig.Method = form.RailsMethod

	values, err := query.Values(initialConfig)
	if err != nil {
		return err // cannot be tested
	}

	err = c.service.ConfigureIAAS(api.ConfigureIAASInput{
		Form:           form,
		EncodedPayload: values.Encode(),
	})
	if err != nil {
		return fmt.Errorf("tile failed to configure: %s", err)
	}

	c.logger.Printf("finished configuring bosh tile")
	return nil
}

func (c ConfigureBosh) Usage() Usage {
	return Usage{
		Description:      "configures the bosh director that is deployed by the Ops Manager",
		ShortDescription: "configures Ops Manager deployed bosh director",
		Flags:            c.Options,
	}
}
