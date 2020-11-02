package commands

import (
	"fmt"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
	"sort"

	"github.com/pivotal-cf/om/presenters"
)

//counterfeiter:generate -o ./fakes/product_service.go --fake-name ProductService . productService
type productService interface {
	GetDiagnosticReport() (api.DiagnosticReport, error)
	ListAvailableProducts() (api.AvailableProductsOutput, error)
}

type byProductName []models.ProductVersions

func (a byProductName) Len() int           { return len(a) }
func (a byProductName) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a byProductName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type Products struct {
	presenter      presenters.FormattedPresenter
	productService productService
	Options        struct {
		Available bool   `long:"available" short:"a" description:"Specify to include available products. Can be used with other options."`
		Staged    bool   `long:"staged" short:"s" description:"Specify to include staged products. Can be used with other options."`
		Deployed  bool   `long:"deployed" short:"d" description:"Specify to deployed products. Can be used with other options."`
		Format    string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

func NewProducts(presenter presenters.FormattedPresenter, productService productService) *Products {
	return &Products{
		presenter:      presenter,
		productService: productService,
	}
}

func (sp Products) Execute(args []string) error {
	diagnosticReport, err := sp.productService.GetDiagnosticReport()
	if err != nil {
		return fmt.Errorf("failed to retrieve staged and deployed products %s", err)
	}

	stagedProducts := diagnosticReport.StagedProducts
	deployedProducts := diagnosticReport.DeployedProducts

	availableProducts, err := sp.productService.ListAvailableProducts()
	if err != nil {
		return fmt.Errorf("failed to retrieve available products %s", err)
	}

	productVersionsCombiner := make(map[string]models.ProductVersions)
	sp.combineStagedProducts(stagedProducts, productVersionsCombiner)
	sp.combineDeployedProducts(deployedProducts, productVersionsCombiner)
	sp.combineAvailableProducts(availableProducts, productVersionsCombiner)

	var productVersions []models.ProductVersions
	for _, versions := range productVersionsCombiner {
		productVersions = append(productVersions, versions)
	}

	sort.Sort(byProductName(productVersions))

	productVersionsOutput := models.ProductsVersionsDisplay{
		ProductVersions: productVersions,
	}

	noModifiersSelected := !sp.Options.Available && !sp.Options.Staged && !sp.Options.Deployed

	if sp.Options.Available || noModifiersSelected {
		productVersionsOutput.Available = true
	}

	if sp.Options.Staged || noModifiersSelected {
		productVersionsOutput.Staged = true
	}

	if sp.Options.Deployed || noModifiersSelected {
		productVersionsOutput.Deployed = true
	}

	sp.presenter.SetFormat(sp.Options.Format)
	sp.presenter.PresentProducts(productVersionsOutput)

	return nil
}

func (sp Products) combineStagedProducts(stagedProducts []api.DiagnosticProduct, productVersionsCombiner map[string]models.ProductVersions) {
	for _, product := range stagedProducts {
		if _, ok := productVersionsCombiner[product.Name]; !ok {
			productVersionsCombiner[product.Name] = models.ProductVersions{
				Name: product.Name,
			}
		}

		productVersions := productVersionsCombiner[product.Name]
		productVersions.Staged = product.Version
		productVersionsCombiner[product.Name] = productVersions
	}
}

func (sp Products) combineDeployedProducts(deployedProducts []api.DiagnosticProduct, productVersionsCombiner map[string]models.ProductVersions) {
	for _, product := range deployedProducts {
		if _, ok := productVersionsCombiner[product.Name]; !ok {
			productVersionsCombiner[product.Name] = models.ProductVersions{
				Name: product.Name,
			}
		}

		productVersions := productVersionsCombiner[product.Name]
		productVersions.Deployed = product.Version
		productVersionsCombiner[product.Name] = productVersions
	}
}

func (sp Products) combineAvailableProducts(availableProducts api.AvailableProductsOutput, productVersionsCombiner map[string]models.ProductVersions) {
	for _, product := range availableProducts.ProductsList {
		if _, ok := productVersionsCombiner[product.Name]; !ok {
			productVersionsCombiner[product.Name] = models.ProductVersions{
				Name: product.Name,
			}
		}

		productVersions := productVersionsCombiner[product.Name]
		productVersions.Available = append(productVersions.Available, product.Version)
		productVersionsCombiner[product.Name] = productVersions
	}
}
