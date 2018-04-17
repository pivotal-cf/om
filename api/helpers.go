package api

import "fmt"

// In general we'd like to keep the API layer as simple as possible, e.g. each method hits a single API endpoint.
// This file contains helper methods that we wish we API endpoints, but the Ops Manager API does not yet support these actions.

func (a Api) CheckProductAvailability(productName string, productVersion string) (bool, error) {
	availableProducts, err := a.ListAvailableProducts()
	if err != nil {
		return false, err
	}

	for _, product := range availableProducts.ProductsList {
		if product.Name == productName && product.Version == productVersion {
			return true, nil
		}
	}

	return false, nil
}

func (a Api) RunningInstallation() (InstallationsServiceOutput, error) {
	installationOutput, err := a.ListInstallations()
	if err != nil {
		return InstallationsServiceOutput{}, err
	}
	if len(installationOutput) > 0 && installationOutput[0].Status == StatusRunning {
		return installationOutput[0], nil
	}
	return InstallationsServiceOutput{}, nil
}

type StagedProductsFindOutput struct {
	Product StagedProduct
}

func (a Api) GetStagedProductByName(productName string) (StagedProductsFindOutput, error) {
	productsOutput, err := a.ListStagedProducts()
	if err != nil {
		return StagedProductsFindOutput{}, err
	}

	var foundProduct StagedProduct
	for _, product := range productsOutput.Products {
		if product.Type == productName {
			foundProduct = product
			break
		}
	}

	if (foundProduct == StagedProduct{}) {
		return StagedProductsFindOutput{}, fmt.Errorf("could not find product %q", productName)
	}

	return StagedProductsFindOutput{Product: foundProduct}, nil
}
