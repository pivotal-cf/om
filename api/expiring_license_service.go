package api

import (
	"fmt"
	"time"
)

// not the acutal response, will need to be updated to match the actual response
type ExpiringLicenseOutPut struct {
	ProductName string    `json:"product_name"`
	GUID        string    `json:"guid"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type expiringProduct struct {
	Type            string
	GUID            string
	LicenseMetadata []ExpiryInfo `json:"license_metadata"`
}

func (a Api) ListExpiringLicenses(expiresWithin string, staged bool, deployed bool) ([]ExpiringLicenseOutPut, error) {

	expiredLicense := []ExpiringLicenseOutPut{}

	expiringProducts := []expiringProduct{}
	err := a.getProductsLicenseInfo(&expiringProducts, staged, deployed)
	layout := "2006-01-02"

	if err != nil {
		return nil, fmt.Errorf("Cannot list deployed products: %w", err)
	}

	for _, expiredProduct := range expiringProducts {

		//handling multiple license associated with a single product
		for _, license := range expiredProduct.LicenseMetadata {

			t, err := time.Parse(layout, license.ExpiresAt)
			if err != nil {
				return nil, fmt.Errorf("could not make convert expiry date string to time: %w", err)
			}
			expiredLicense = append(expiredLicense, ExpiringLicenseOutPut{ProductName: expiredProduct.Type, GUID: expiredProduct.GUID, ExpiresAt: t})
		}
	}
	return expiredLicense, err
}

func (a Api) getProductsLicenseInfo(expiringProducts *[]expiringProduct, staged bool, deployed bool) error {

	if staged {
		err := a.getStagedProducts(expiringProducts)

		if err != nil {
			return fmt.Errorf("could not get staged products: %w", err)
		}

	} else if deployed {
		err := a.getDeployedProducts(expiringProducts)

		if err != nil {
			return fmt.Errorf("could not get staged products: %w", err)
		}

	} else {
		err := a.getStagedProducts(expiringProducts)

		if err != nil {
			return fmt.Errorf("could not get staged products: %w", err)
		}

		err = a.getDeployedProducts(expiringProducts)

		if err != nil {
			return fmt.Errorf("could not get staged products: %w", err)
		}

	}
	removeDuplicates(expiringProducts)
	return nil
}

func (a Api) getStagedProducts(expiringProducts *[]expiringProduct) error {

	stagedProducts, err := a.ListStagedProducts()

	if err != nil {
		return fmt.Errorf("could not make a call to ListStagedProducts api: %w", err)
	}

	for _, stagedProduct := range stagedProducts.Products {
		*expiringProducts = append(*expiringProducts, expiringProduct{GUID: stagedProduct.GUID,
			Type:            stagedProduct.Type,
			LicenseMetadata: stagedProduct.LicenseMetadata,
		})
	}

	return nil
}

func (a Api) getDeployedProducts(expiringProducts *[]expiringProduct) error {
	deployedProducts, err := a.ListDeployedProducts()

	if err != nil {
		return fmt.Errorf("could not make a call to ListDeployedProducts api: %w", err)
	}

	for _, deployedProduct := range deployedProducts {
		*expiringProducts = append(*expiringProducts, expiringProduct{GUID: deployedProduct.GUID,
			Type:            deployedProduct.Type,
			LicenseMetadata: deployedProduct.LicenseMetadata,
		})
	}

	return nil
}

func removeDuplicates(expiringProducts *[]expiringProduct) {
	seen := make(map[string]bool)
	result := []expiringProduct{}

	for _, expiringProduct := range *expiringProducts {
		if !seen[expiringProduct.GUID] {
			seen[expiringProduct.GUID] = true
			result = append(result, expiringProduct)
		}
	}

	*expiringProducts = result
}
