package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/renderers"
)

type BoshEnvironment struct {
	service         boshEnvironmentService
	logger          logger
	rendererFactory rendererFactory
	opsmanHost      string
	Options         struct {
		ShellType     string `long:"shell-type" description:"Prints for the given shell (posix|powershell)"`
		SSHPrivateKey string `long:"ssh-private-key" short:"i" description:"location of ssh private key"`
	}
}

//go:generate counterfeiter -o ./fakes/bosh_environment_service.go --fake-name BoshEnvironmentService . boshEnvironmentService
type boshEnvironmentService interface {
	GetBoshEnvironment() (api.GetBoshEnvironmentOutput, error)
	ListCertificateAuthorities() (api.CertificateAuthoritiesOutput, error)
}

//go:generate counterfeiter -o ./fakes/renderer_factory.go --fake-name RendererFactory . rendererFactory
type rendererFactory interface {
	Create(shellType string) (renderers.Renderer, error)
}

func NewBoshEnvironment(service boshEnvironmentService, logger logger, opsmanHost string, rendererFactory rendererFactory) BoshEnvironment {
	return BoshEnvironment{
		service:         service,
		logger:          logger,
		rendererFactory: rendererFactory,
		opsmanHost:      opsmanHost,
	}
}

func (be BoshEnvironment) Execute(args []string) error {
	if _, err := jhanda.Parse(&be.Options, args); err != nil {
		return fmt.Errorf("could not parse bosh-env flags: %s", err)
	}

	renderer, err := be.rendererFactory.Create(be.Options.ShellType)
	if err != nil {
		return err
	}
	boshEnvironment, err := be.service.GetBoshEnvironment()
	if err != nil {
		return err
	}

	certificateAuthorities, err := be.service.ListCertificateAuthorities()
	if err != nil {
		return err
	}

	var boshCACerts string

	for _, ca := range certificateAuthorities.CAs {
		if ca.Active {
			if boshCACerts != "" {
				boshCACerts = boshCACerts + "\n"
			}
			boshCACerts = boshCACerts + ca.CertPEM
		}
	}

	variables := make(map[string]string)
	variables["BOSH_CLIENT"] = boshEnvironment.Client
	variables["BOSH_CLIENT_SECRET"] = boshEnvironment.ClientSecret
	variables["BOSH_ENVIRONMENT"] = boshEnvironment.Environment
	variables["BOSH_CA_CERT"] = boshCACerts

	variables["CREDHUB_CLIENT"] = boshEnvironment.Client
	variables["CREDHUB_SECRET"] = boshEnvironment.ClientSecret
	variables["CREDHUB_SERVER"] = fmt.Sprintf("https://%s:8844", boshEnvironment.Environment)
	variables["CREDHUB_CA_CERT"] = boshCACerts

	if be.Options.SSHPrivateKey != "" {
		variables["BOSH_ALL_PROXY"] = fmt.Sprintf("ssh+socks5://ubuntu@%s:22?private-key=%s", be.opsmanHost, be.Options.SSHPrivateKey)
		variables["CREDHUB_PROXY"] = variables["BOSH_ALL_PROXY"]
	}
	be.renderVariables(renderer, variables)

	return nil
}

func (be BoshEnvironment) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This prints bosh environment variables to target bosh director",
		ShortDescription: "prints bosh environment variables",
		Flags:            be.Options,
	}
}

func (be BoshEnvironment) renderVariables(renderer renderers.Renderer, variables map[string]string) {
	for k, v := range variables {
		be.logger.Println(renderer.RenderEnvironmentVariable(k, v))
	}
}
