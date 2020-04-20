package commands

import (
	"errors"
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/interpolate"
	"gopkg.in/yaml.v2"
	"sort"
	"strings"
)

type ConfigureOpsman struct {
	service     configureOpsmanService
	logger      logger
	environFunc func() []string
	Options     struct {
		ConfigFile string   `long:"config"    short:"c"         description:"path to yml file containing all config fields (see docs/configure-director/README.md for format)" required:"true"`
		VarsFile   []string `long:"vars-file" short:"l"         description:"load variables from a YAML file"`
		VarsEnv    []string `long:"vars-env"  env:"OM_VARS_ENV" description:"load variables from environment variables (e.g.: 'MY' to load MY_var=value)"`
		Vars       []string `long:"var"       short:"v"         description:"load variable from the command line. Format: VAR=VAL"`
		OpsFile    []string `long:"ops-file"                    description:"YAML operations file"`
	}
}

type opsmanConfig struct {
	SSLCertificate *struct {
		Certificate string `yaml:"certificate"`
		PrivateKey  string `yaml:"private_key"`
	} `yaml:"ssl-certificate"`
	PivotalNetwork *struct {
		Token string `yaml:"api_token"`
	} `yaml:"pivotal-network-settings"`
	RBACSettings *struct {
		LDAPAdminGroupName  string `yaml:"ldap_rbac_admin_group_name"`
		SAMLAdminGroup      string `yaml:"rbac_saml_admin_group"`
		SAMLGroupsAttribute string `yaml:"rbac_saml_groups_attribute"`
	} `yaml:"rbac-settings"`
	Field map[string]interface{} `yaml:",inline"`
}

//counterfeiter:generate -o ./fakes/configure_opsman_service.go --fake-name ConfigureOpsmanService . configureOpsmanService
type configureOpsmanService interface {
	UpdateSSLCertificate(api.SSLCertificateInput) error
	UpdatePivnetToken(token string) error
	EnableRBAC(rbacSettings api.RBACSettings) error
}

func NewConfigureOpsman(environFunc func() []string, service configureOpsmanService, logger logger) ConfigureOpsman {
	return ConfigureOpsman{
		environFunc: environFunc,
		service:     service,
		logger:      logger,
	}
}

func (c ConfigureOpsman) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse configure-opsman flags: %s", err)
	}

	config, err := c.interpolateConfig()
	if err != nil {
		return err
	}

	err = c.validate(config)
	if err != nil {
		return err
	}

	if config.SSLCertificate != nil {
		c.logger.Printf("Updating SSL Certificate...\n")
		err = c.updateSSLCertificate(config)
		if err != nil {
			return err
		}
		c.logger.Printf("Successfully applied custom SSL Certificate.\n")
		c.logger.Printf("Please allow about 1 min for the new certificate to take effect.\n")
	}

	if config.PivotalNetwork != nil {
		c.logger.Printf("Updating Pivnet Token...\n")
		err = c.service.UpdatePivnetToken(config.PivotalNetwork.Token)
		if err != nil {
			return err
		}
		c.logger.Printf("Successfully applied Pivnet Token.\n")
	}

	if config.RBACSettings != nil {
		c.logger.Printf("Updating RBAC Settings...\n")
		payload := api.RBACSettings{
			SAMLAdminGroup:      config.RBACSettings.SAMLAdminGroup,
			SAMLGroupsAttribute: config.RBACSettings.SAMLGroupsAttribute,
			LDAPAdminGroupName:  config.RBACSettings.LDAPAdminGroupName,
		}
		err = c.service.EnableRBAC(payload)
		if err != nil {
			return err
		}
		c.logger.Printf("Successfully applied RBAC Settings.\n")
	}
	return nil
}

func (c ConfigureOpsman) updateSSLCertificate(config *opsmanConfig) error {
	return c.service.UpdateSSLCertificate(api.SSLCertificateInput{
		CertPem:       config.SSLCertificate.Certificate,
		PrivateKeyPem: config.SSLCertificate.PrivateKey,
	})
}

func (c ConfigureOpsman) interpolateConfig() (*opsmanConfig, error) {
	configContents, err := interpolate.Execute(interpolate.Options{
		TemplateFile:  c.Options.ConfigFile,
		VarsFiles:     c.Options.VarsFile,
		EnvironFunc:   c.environFunc,
		Vars:          c.Options.Vars,
		VarsEnvs:      c.Options.VarsEnv,
		OpsFiles:      c.Options.OpsFile,
		ExpectAllKeys: true,
	})
	if err != nil {
		return nil, err
	}

	var config opsmanConfig
	err = yaml.UnmarshalStrict(configContents, &config)
	if err != nil {
		return nil, fmt.Errorf("could not be parsed as valid configuration: %s: %s", c.Options.ConfigFile, err)
	}
	return &config, nil
}

func (c ConfigureOpsman) validate(config *opsmanConfig) error {
	invalidFields := []string{}
	if len(config.Field) > 0 {
		for key, _ := range config.Field {
			if key == "opsman-configuration" {
				continue
			}
			invalidFields = append(invalidFields, key)
		}
	}
	if len(invalidFields) > 0 {
		sort.Strings(invalidFields)
		return fmt.Errorf("unrecognized top level key(s) in config file:\n%s", strings.Join(invalidFields, "\n"))
	}

	if config.RBACSettings != nil {
		if config.RBACSettings.SAMLGroupsAttribute != "" &&
			config.RBACSettings.SAMLAdminGroup != "" &&
			config.RBACSettings.LDAPAdminGroupName != "" {
			return errors.New("can only set SAML or LDAP. Check the config file and use only the appropriate values.\nFor example config values, see the docs directory for documentation.")
		}
	}
	return nil
}

func (c ConfigureOpsman) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description: "This authenticated command " +
			"configures settings available on the \"Settings\" page " +
			"in the Ops Manager UI. For an example config, " +
			"reference the docs directory for this command.",
		ShortDescription: "configures values present on the Ops Manager settings page",
		Flags:            c.Options,
	}
}
