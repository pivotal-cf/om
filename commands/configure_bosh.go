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

type ConfigureBOSH struct {
	service boshFormService
	logger  logger
	Options struct {
		IAASConfiguration string `short:"i"  long:"iaas-configuration"  description:"iaas specific configuration for the bosh director"`
	}
}

type IAASConfiguration struct {
	Project              string `url:"iaas_configuration[project]" json:"project"`
	DefaultDeploymentTag string `url:"iaas_configuration[default_deployment_tag]" json:"default_deployment_tag"`
	AuthJSON             string `url:"iaas_configuration[auth_json]" json:"auth_json"`
	AuthenticityToken    string `url:"authenticity_token"`
	Method               string `url:"_method"`
}

//go:generate counterfeiter -o ./fakes/bosh_form_service.go --fake-name BoshFormService . boshFormService
type boshFormService interface {
	GetForm(path string) (api.Form, error)
	ConfigureIAAS(api.ConfigureIAASInput) error
}

func NewConfigureBOSH(s boshFormService, l logger) ConfigureBOSH {
	return ConfigureBOSH{service: s, logger: l}
}

func (c ConfigureBOSH) Execute(args []string) error {
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

func (c ConfigureBOSH) Usage() Usage {
	return Usage{
		Description:      "configures the bosh director that is deployed by the Ops Manager",
		ShortDescription: "configures Ops Manager deployed bosh director",
		Flags:            c.Options,
	}
}
