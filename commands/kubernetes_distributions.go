package commands

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/hashicorp/go-version"
	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf/om/api"
)

//counterfeiter:generate -o ./fakes/kubernetes_distributions_service.go --fake-name KubernetesDistributionsService . kubernetesDistributionsService
type kubernetesDistributionsService interface {
	ListKubernetesDistributions() (api.KubernetesDistributionAssociationsResponse, error)
	Info() (api.Info, error)
}

type KubernetesDistributions struct {
	service kubernetesDistributionsService
	stdout  io.Writer
	Options struct {
		ProductName string `long:"product" short:"p" description:"filter to distributions available for a specific product"`
		Format      string `long:"format"  short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

func NewKubernetesDistributions(service kubernetesDistributionsService, stdout io.Writer) *KubernetesDistributions {
	return &KubernetesDistributions{
		service: service,
		stdout:  stdout,
	}
}

type K8sDistributionRow struct {
	Distribution string            `json:"distribution"`
	Version      string            `json:"version"`
	Products     []K8sProductAssoc `json:"products,omitempty"`
}

type K8sProductAssoc struct {
	Name     string `json:"name"`
	Staged   bool   `json:"staged,omitempty"`
	Deployed bool   `json:"deployed,omitempty"`
}

func (kd KubernetesDistributions) Execute(_ []string) error {
	info, err := kd.service.Info()
	if err != nil {
		return fmt.Errorf("failed to get Ops Manager version: %w", err)
	}
	if ok, verErr := info.VersionAtLeast(3, 3); !ok {
		if verErr != nil {
			return fmt.Errorf("kubernetes-distributions requires Ops Manager 3.3 or newer: %w", verErr)
		}
		return fmt.Errorf("kubernetes-distributions requires Ops Manager 3.3 or newer (current version: %s)", info.Version)
	}

	response, err := kd.service.ListKubernetesDistributions()
	if err != nil {
		return fmt.Errorf("failed to list kubernetes distributions: %w", err)
	}

	library := response.Library
	products := response.Products

	if kd.Options.ProductName != "" {
		product, found := findK8sProduct(products, kd.Options.ProductName)
		if !found {
			return fmt.Errorf("kubernetes product %q not found", kd.Options.ProductName)
		}
		library = filterK8sDistributionLibraryForProduct(library, product)
		products = []api.KubernetesProductDistributionEntry{product}
	}

	rows := buildK8sDistributionRows(library, products)

	if kd.Options.Format == "json" {
		return kd.renderJSON(rows)
	}
	kd.renderTable(rows, kd.Options.ProductName != "")
	return nil
}

func (kd KubernetesDistributions) renderJSON(rows []K8sDistributionRow) error {
	encoder := json.NewEncoder(kd.stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(rows); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func (kd KubernetesDistributions) renderTable(rows []K8sDistributionRow, showProductColumns bool) {
	table := tablewriter.NewWriter(kd.stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoMergeCellsByColumnIndex([]int{0, 1})

	if showProductColumns {
		table.SetHeader([]string{"Distribution", "Version", "Staged", "Deployed"})
		for _, row := range rows {
			staged, deployed := "", ""
			for _, p := range row.Products {
				if p.Staged {
					staged = "yes"
				}
				if p.Deployed {
					deployed = "yes"
				}
			}
			table.Append([]string{row.Distribution, row.Version, staged, deployed})
		}
	} else {
		table.SetHeader([]string{"Distribution", "Version"})
		for _, row := range rows {
			table.Append([]string{row.Distribution, row.Version})
		}
	}

	table.Render()
}

func findK8sProduct(products []api.KubernetesProductDistributionEntry, name string) (api.KubernetesProductDistributionEntry, bool) {
	for _, p := range products {
		if p.ProductName == name {
			return p, true
		}
	}
	return api.KubernetesProductDistributionEntry{}, false
}

func filterK8sDistributionLibraryForProduct(library []api.KubernetesDistributionLibraryEntry, product api.KubernetesProductDistributionEntry) []api.KubernetesDistributionLibraryEntry {
	// Using a simple loop for readability. Ops Manager rarely returns more
	// than a handful of available distributions per product, making
	// optimizations unnecessary overkill here.
	var filtered []api.KubernetesDistributionLibraryEntry
	for _, entry := range library {
		isAvailable := slices.ContainsFunc(product.AvailableKubernetesDistributions, func(d api.KubernetesDistribution) bool {
			return d.Identifier == entry.Identifier && d.Version == entry.Version
		})

		if isAvailable {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func buildK8sDistributionRows(library []api.KubernetesDistributionLibraryEntry, products []api.KubernetesProductDistributionEntry) []K8sDistributionRow {
	compareVersions := func(vs1, vs2 string) int {
		v1, err1 := version.NewVersion(vs1)
		v2, err2 := version.NewVersion(vs2)

		if err1 == nil && err2 == nil {
			return v1.Compare(v2)
		}

		// fallback to simple lexicographical comparison
		return cmp.Compare(vs1, vs2)
	}

	sortedLibrary := slices.SortedFunc(slices.Values(library), func(a, b api.KubernetesDistributionLibraryEntry) int {
		return cmp.Or(
			cmp.Compare(a.Identifier, b.Identifier),
			compareVersions(a.Version, b.Version),
		)
	})

	sortedProducts := slices.SortedFunc(slices.Values(products), func(a, b api.KubernetesProductDistributionEntry) int {
		return cmp.Compare(a.ProductName, b.ProductName)
	})

	var rows []K8sDistributionRow
	for _, entry := range sortedLibrary {
		row := K8sDistributionRow{
			Distribution: entry.Identifier,
			Version:      entry.Version,
		}

		for _, p := range sortedProducts {
			isStaged := p.StagedKubernetesDistribution != nil &&
				p.StagedKubernetesDistribution.Identifier == entry.Identifier &&
				p.StagedKubernetesDistribution.Version == entry.Version
			isDeployed := p.DeployedKubernetesDistribution != nil &&
				p.DeployedKubernetesDistribution.Identifier == entry.Identifier &&
				p.DeployedKubernetesDistribution.Version == entry.Version

			if isStaged || isDeployed {
				row.Products = append(row.Products, K8sProductAssoc{
					Name:     p.ProductName,
					Staged:   isStaged,
					Deployed: isDeployed,
				})
			}
		}

		rows = append(rows, row)
	}

	return rows
}
