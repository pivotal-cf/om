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
const directorConfigurationPath = "/infrastructure/director_configuration/edit"

type ConfigureBosh struct {
	service boshFormService
	logger  logger
	Options struct {
		IaaSConfiguration     string `short:"i"  long:"iaas-configuration"  description:"iaas specific configuration for the bosh director"`
		DirectorConfiguration string `short:"d"  long:"director-configuration"  description:"director-specific configuration for the bosh director"`
	}
}

//go:generate counterfeiter -o ./fakes/bosh_form_service.go --fake-name BoshFormService . boshFormService
type boshFormService interface {
	GetForm(path string) (api.Form, error)
	PostForm(api.PostFormInput) error
}

func NewConfigureBosh(s boshFormService, l logger) ConfigureBosh {
	return ConfigureBosh{service: s, logger: l}
}

func (c ConfigureBosh) ConfigureForm(path, configuration string) error {
	form, err := c.service.GetForm(path)
	if err != nil {
		return fmt.Errorf("could not fetch form: %s", err)
	}

	var initialConfig *BoshConfiguration
	err = json.NewDecoder(strings.NewReader(configuration)).Decode(&initialConfig)
	if err != nil {
		return fmt.Errorf("could not decode json: %s", err)
	}

	initialConfig.AuthenticityToken = form.AuthenticityToken
	initialConfig.Method = form.RailsMethod

	values, err := query.Values(initialConfig)
	if err != nil {
		return err // cannot be tested
	}

	err = c.service.PostForm(api.PostFormInput{
		Form:           form,
		EncodedPayload: values.Encode(),
	})
	if err != nil {
		return fmt.Errorf("tile failed to configure: %s", err)
	}

	return nil
}

func (c ConfigureBosh) Execute(args []string) error {
	_, err := flags.Parse(&c.Options, args)
	if err != nil {
		return err
	}

	if c.Options.IaaSConfiguration != "" {
		c.logger.Printf("configuring iaas specific options for bosh tile")
		err = c.ConfigureForm(iaasConfigurationPath, c.Options.IaaSConfiguration)
		if err != nil {
			return err
		}
	}

	if c.Options.DirectorConfiguration != "" {
		c.logger.Printf("configuring director options for bosh tile")
		err = c.ConfigureForm(directorConfigurationPath, c.Options.DirectorConfiguration)
		if err != nil {
			return err
		}
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
