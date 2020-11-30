package commands

import (
	"errors"
	"fmt"
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
		Settings api.SSLCertificateSettings `yaml:",inline"`
	} `yaml:"ssl-certificate"`
	PivotalNetwork *struct {
		Settings api.PivnetSettings `yaml:",inline"`
	} `yaml:"pivotal-network-settings"`
	RBAC *struct {
		Settings api.RBACSettings `yaml:",inline"`
	} `yaml:"rbac-settings"`
	Banner *struct {
		Settings api.BannerSettings `yaml:",inline"`
	} `yaml:"banner-settings"`
	Syslog *struct {
		Settings api.SyslogSettings `yaml:",inline"`
	} `yaml:"syslog-settings"`
	TokenExpirations *struct {
		Settings api.TokensExpiration `yaml:",inline"`
	} `yaml:"tokens-expiration"`
	Field map[string]interface{} `yaml:",inline"`
}

//counterfeiter:generate -o ./fakes/configure_opsman_service.go --fake-name ConfigureOpsmanService . configureOpsmanService
type configureOpsmanService interface {
	UpdateBanner(settings api.BannerSettings) error
	UpdateSSLCertificate(api.SSLCertificateSettings) error
	UpdatePivnetToken(settings api.PivnetSettings) error
	EnableRBAC(rbacSettings api.RBACSettings) error
	UpdateSyslogSettings(syslogSettings api.SyslogSettings) error
	UpdateTokensExpiration(tokenExpirations api.TokensExpiration) error
}

func NewConfigureOpsman(environFunc func() []string, service configureOpsmanService, logger logger) *ConfigureOpsman {
	return &ConfigureOpsman{
		environFunc: environFunc,
		service:     service,
		logger:      logger,
	}
}

func (c ConfigureOpsman) Execute(args []string) error {
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
		err = c.service.UpdatePivnetToken(config.PivotalNetwork.Settings)
		if err != nil {
			return err
		}
		c.logger.Printf("Successfully applied Pivnet Token.\n")
	}

	if config.RBAC != nil {
		c.logger.Printf("Updating RBAC Settings...\n")
		err = c.service.EnableRBAC(config.RBAC.Settings)
		if err != nil {
			return err
		}
		c.logger.Printf("Successfully applied RBAC Settings.\n")
	}

	if config.Banner != nil {
		c.logger.Printf("Updating Banner...\n")
		err = c.service.UpdateBanner(config.Banner.Settings)
		if err != nil {
			return err
		}
		c.logger.Printf("Successfully applied Banner.\n")
	}

	if config.Syslog != nil {
		c.logger.Printf("Updating Syslog...\n")
		payload := config.Syslog.Settings
		err = c.service.UpdateSyslogSettings(payload)
		if err != nil {
			return err
		}
		c.logger.Printf("Successfully applied Syslog.\n")
	}

	if config.TokenExpirations != nil {
		c.logger.Printf("Updating tokens expiration...\n")
		payload := config.TokenExpirations.Settings
		err = c.service.UpdateTokensExpiration(payload)
		if err != nil {
			return err
		}
		c.logger.Printf("Successfully updated tokens expiration.\n")
	}

	return nil
}

func (c ConfigureOpsman) updateSSLCertificate(config *opsmanConfig) error {
	return c.service.UpdateSSLCertificate(config.SSLCertificate.Settings)
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
		for key := range config.Field {
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

	if config.RBAC != nil {
		if config.RBAC.Settings.SAMLGroupsAttribute != "" &&
			config.RBAC.Settings.SAMLAdminGroup != "" &&
			config.RBAC.Settings.LDAPAdminGroupName != "" {
			return errors.New("can only set SAML or LDAP. Check the config file and use only the appropriate values.\nFor example config values, see the docs directory for documentation.")
		}
	}
	return nil
}
