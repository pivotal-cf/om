package api

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type ExpiringLicenseOutput struct {
	ProductName    string    `json:"product_name"`
	GUID           string    `json:"guid"`
	ExpiresAt      time.Time `json:"expires_at"`
	ProductState   []string  `json:"product_state"`
	ProductVersion string    `json:"product_version"`
	LicenseVersion string    `json:"licensed_version"`
}

type licenseInfoProduct struct {
	Type            string
	GUID            string
	LicenseMetadata []LicenseMetadata `json:"license_metadata"`
	ProductVersion  string
	ProductState    string
}

const (
	ProductStateStaged   = "staged"
	ProductStateDeployed = "deployed"
	layout               = "2006-01-02" // golang's time format for the library not a true date
)

func (a Api) ListExpiringLicenses(expiresWithin string, staged bool, deployed bool) ([]ExpiringLicenseOutput, error) {
	expiredLicense := []ExpiringLicenseOutput{}

	expiringProducts, err := a.getProductsLicenseInfo(staged, deployed)

	if err != nil {
		return nil, fmt.Errorf("could not list licensed products: %w", err)
	}
	for _, expiredProduct := range expiringProducts {

		for _, license := range expiredProduct.LicenseMetadata {

			if license.ExpiresAt == "" {
				currentConfiguredProperties, err := a.GetStagedProductProperties(expiredProduct.GUID, false)
				if err != nil {
					return nil, err
				}
				configuredProperty := currentConfiguredProperties[license.PropertyReference]

				if configuredProperty.Value == nil {

					expiredLicense = append(expiredLicense, ExpiringLicenseOutput{
						ProductName:    expiredProduct.Type,
						GUID:           expiredProduct.GUID,
						ProductState:   []string{expiredProduct.ProductState},
						ProductVersion: expiredProduct.ProductVersion,
					})
				}

			} else {
				t, err := time.Parse(layout, license.ExpiresAt)
				if err != nil {
					return nil, fmt.Errorf("could not convert expiry date string to time: %w", err)
				}

				//expiresWithin is never null. Defaults to 3 months when nothing is passed
				if t.Before(calcEndDate(expiresWithin)) {
					// Get the product version from the license metadata
					productVersion := license.ProductVersion

					expiredLicense = append(expiredLicense, ExpiringLicenseOutput{
						ProductName:    expiredProduct.Type,
						GUID:           expiredProduct.GUID,
						ExpiresAt:      t,
						ProductState:   []string{expiredProduct.ProductState},
						ProductVersion: expiredProduct.ProductVersion,
						LicenseVersion: productVersion,
					})
				}
			}
		}
	}

	mergeDuplicates(&expiredLicense)
	return expiredLicense, err
}

func (a Api) getProductsLicenseInfo(staged bool, deployed bool) ([]licenseInfoProduct, error) {
	var allProducts []licenseInfoProduct

	noModifiersSelected := !staged && !deployed

	if staged || noModifiersSelected {
		stagedProducts, err := a.getStagedProducts()
		if err != nil {
			return nil, fmt.Errorf("could not get staged products: %w", err)
		}
		allProducts = append(allProducts, stagedProducts...)
	}

	if deployed || noModifiersSelected {
		deployedProducts, err := a.getDeployedProducts()
		if err != nil {
			return nil, fmt.Errorf("could not get deployed products: %w", err)
		}
		allProducts = append(allProducts, deployedProducts...)
	}

	return allProducts, nil
}

func (a Api) getStagedProducts() ([]licenseInfoProduct, error) {
	stagedProducts, err := a.ListStagedProducts()
	if err != nil {
		return nil, fmt.Errorf("could not make a call to ListStagedProducts api: %w", err)
	}

	var expiringProducts []licenseInfoProduct
	for _, stagedProduct := range stagedProducts.Products {
		expiringProducts = append(expiringProducts, licenseInfoProduct{
			GUID:            stagedProduct.GUID,
			Type:            stagedProduct.Type,
			LicenseMetadata: stagedProduct.LicenseMetadata,
			ProductVersion:  stagedProduct.ProductVersion,
			ProductState:    ProductStateStaged,
		})
	}

	return expiringProducts, nil
}

func (a Api) getDeployedProducts() ([]licenseInfoProduct, error) {
	deployedProducts, err := a.ListDeployedProducts()
	if err != nil {
		return nil, fmt.Errorf("could not make a call to ListDeployedProducts api: %w", err)
	}

	var expiringProducts []licenseInfoProduct
	for _, deployedProduct := range deployedProducts {
		expiringProducts = append(expiringProducts, licenseInfoProduct{
			GUID:            deployedProduct.GUID,
			Type:            deployedProduct.Type,
			LicenseMetadata: deployedProduct.LicenseMetadata,
			ProductVersion:  deployedProduct.ProductVersion,
			ProductState:    ProductStateDeployed,
		})
	}
	return expiringProducts, nil
}

func calcEndDate(expiresWithin string) time.Time {
	exp := regexp.MustCompile(`(?P<duration>^[1-9]\d*)+(?P<type>[dwmy]$)`)
	match := exp.FindStringSubmatch(expiresWithin)

	switch durationType := match[2]; durationType {
	case "d":
		days, _ := strconv.Atoi(match[1])
		return time.Now().AddDate(0, 0, days)
	case "w":
		weeks, _ := strconv.Atoi(match[1])
		return time.Now().AddDate(0, 0, weeks*7)
	case "m":
		months, _ := strconv.Atoi(match[1])
		return time.Now().AddDate(0, months, 0)
	case "y":
		years, _ := strconv.Atoi(match[1])
		return time.Now().AddDate(years, 0, 0)
	default:
		return time.Now().AddDate(0, 3, 0)
	}
}

func mergeDuplicates(expiringProducts *[]ExpiringLicenseOutput) {
	type loc struct {
		seen  bool
		index int
	}
	seen := make(map[string]loc)
	result := []ExpiringLicenseOutput{}

	for _, expiringProduct := range *expiringProducts {
		// cannot use license key  in the unique key combination because the key in staged product properties API automatically gets updated with
		// the new license key value when a user updates with a new license key in the UI
		// There is no endpoint to check for the deployed product properties. So relying on the expiry date being different.
		key := expiringProduct.GUID + expiringProduct.ProductVersion + expiringProduct.ExpiresAt.String()
		if !seen[key].seen {
			seen[key] = loc{true, len(result)}
			result = append(result, expiringProduct)
		} else {
			result[seen[key].index].ProductState = append(result[seen[key].index].ProductState, expiringProduct.ProductState...)

		}
	}

	*expiringProducts = result
}
