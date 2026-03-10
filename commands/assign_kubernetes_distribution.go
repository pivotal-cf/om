package commands

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/pivotal-cf/om/api"
)

type AssignKubernetesDistribution struct {
	logger  logger
	service assignKubernetesDistributionService
	Options struct {
		InterpolateOptions interpolateConfigFileOptions `group:"config file interpolation"`
		ProductName        string                       `long:"product"      short:"p" description:"name of Ops Manager tile to associate a kubernetes distribution to" required:"true"`
		Distribution       string                       `long:"distribution" short:"d" description:"kubernetes distribution to assign (identifier:version, e.g. 'embedded-kubernetes:1.0.0' or 'embedded-kubernetes:latest')" required:"true"`
	}
}

//counterfeiter:generate -o ./fakes/assign_kubernetes_distribution_service.go --fake-name AssignKubernetesDistributionService . assignKubernetesDistributionService
type assignKubernetesDistributionService interface {
	ListKubernetesDistributions() (api.KubernetesDistributionAssociationsResponse, error)
	AssignKubernetesDistribution(input api.AssignKubernetesDistributionInput) error
	Info() (api.Info, error)
}

func NewAssignKubernetesDistribution(service assignKubernetesDistributionService, logger logger) *AssignKubernetesDistribution {
	return &AssignKubernetesDistribution{
		service: service,
		logger:  logger,
	}
}

func (ak AssignKubernetesDistribution) Execute(_ []string) error {
	info, err := ak.service.Info()
	if err != nil {
		return fmt.Errorf("failed to get Ops Manager version: %w", err)
	}
	if ok, verErr := info.VersionAtLeast(3, 3); !ok {
		if verErr != nil {
			return fmt.Errorf("assign-kubernetes-distribution requires Ops Manager 3.3 or newer: %w", verErr)
		}
		return fmt.Errorf("assign-kubernetes-distribution requires Ops Manager 3.3 or newer (current version: %s)", info.Version)
	}

	ak.logger.Printf("finding available kubernetes distributions for product: %q...", ak.Options.ProductName)

	product, err := ak.getProduct()
	if err != nil {
		return err
	}

	if product.StagedForDeletion {
		return fmt.Errorf("could not assign kubernetes distribution: product %q is staged for deletion", ak.Options.ProductName)
	}

	if len(product.AvailableKubernetesDistributions) == 0 {
		return fmt.Errorf(
			"no kubernetes distributions are available for %q",
			ak.Options.ProductName,
		)
	}

	ak.logger.Println("validating that kubernetes distribution exists in Ops Manager...")
	distribution, err := ak.resolveDistribution(product)
	if err != nil {
		return err
	}

	ak.logger.Printf("assigning kubernetes distribution: \"%s:%s\" to product %q...\n",
		distribution.Identifier, distribution.Version, ak.Options.ProductName)

	err = ak.service.AssignKubernetesDistribution(api.AssignKubernetesDistributionInput{
		Products: []api.AssignKubernetesDistributionProduct{
			{
				GUID: product.GUID,
				KubernetesDistribution: api.KubernetesDistribution{
					Identifier: distribution.Identifier,
					Version:    distribution.Version,
				},
			},
		},
	})
	if err != nil {
		return err
	}

	ak.logger.Println("assigned kubernetes distribution successfully")
	return nil
}

func (ak AssignKubernetesDistribution) getProduct() (api.KubernetesProductDistributionEntry, error) {
	var result api.KubernetesProductDistributionEntry

	products, err := ak.service.ListKubernetesDistributions()
	if err != nil {
		return result, err
	}

	for _, p := range products.Products {
		if p.ProductName == ak.Options.ProductName {
			return p, nil
		}
	}

	return result, fmt.Errorf("kubernetes product %q not found", ak.Options.ProductName)
}

func (ak AssignKubernetesDistribution) resolveDistribution(product api.KubernetesProductDistributionEntry) (api.KubernetesDistribution, error) {
	available := product.AvailableKubernetesDistributions

	identifier, distributionVersion, err := ak.parseDistributionOption()
	if err != nil {
		return api.KubernetesDistribution{}, err
	}

	if distributionVersion == "latest" {
		return ak.resolveLatest(available, identifier)
	}

	for _, d := range available {
		if d.Identifier == identifier && d.Version == distributionVersion {
			return d, nil
		}
	}

	return api.KubernetesDistribution{}, fmt.Errorf(
		"kubernetes distribution %s:%s not a valid option for product %q.\nAvailable distributions for %q: %s",
		identifier, distributionVersion, ak.Options.ProductName, ak.Options.ProductName, formatAvailableKubernetesDistributions(available),
	)
}

func (ak AssignKubernetesDistribution) resolveLatest(available []api.KubernetesDistribution, identifier string) (api.KubernetesDistribution, error) {
	var best api.KubernetesDistribution
	var bestVersion *version.Version

	for _, d := range available {
		if d.Identifier != identifier {
			continue
		}

		v, err := version.NewVersion(d.Version)
		if err != nil {
			return api.KubernetesDistribution{}, fmt.Errorf(
				"could not parse version %q for distribution %q: %w",
				d.Version, d.Identifier, err,
			)
		}

		if bestVersion == nil || v.GreaterThan(bestVersion) {
			best = d
			bestVersion = v
		}
	}

	if bestVersion == nil {
		return api.KubernetesDistribution{}, fmt.Errorf(
			"no supported kubernetes distribution with identifier %q for product %q.\nAvailable distributions: %s",
			identifier, ak.Options.ProductName, formatAvailableKubernetesDistributions(available),
		)
	}

	return best, nil
}

func (ak AssignKubernetesDistribution) parseDistributionOption() (identifier, version string, err error) {
	parts := strings.SplitN(ak.Options.Distribution, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf(`expected "--distribution" format as "identifier:version" (e.g. "embedded-kubernetes:1.0.0" or "embedded-kubernetes:latest")`)
	}

	return parts[0], parts[1], nil
}

func formatAvailableKubernetesDistributions(distributions []api.KubernetesDistribution) string {
	var builder strings.Builder

	for _, d := range distributions {
		_, _ = fmt.Fprintf(&builder, "\n  - %s:%s", d.Identifier, d.Version)
	}

	return builder.String()
}
