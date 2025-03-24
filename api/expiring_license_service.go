package api

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type ExpiringLicenseOutput struct {
	ProductName  string    `json:"product_name"`
	GUID         string    `json:"guid"`
	ExpiresAt    time.Time `json:"expires_at"`
	ProductState string    `json:"product_state"`
}

type expiringProduct struct {
	Type            string
	GUID            string
	LicenseMetadata []ExpiryInfo `json:"license_metadata"`
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
			t, err := time.Parse(layout, license.ExpiresAt)
			if err != nil {
				return nil, fmt.Errorf("could not convert expiry date string to time: %w", err)
			}

			//expiresWithin is never null. Defaults to 3 months when nothing is passed
			if t.Before(calcEndDate(expiresWithin)) {
				expiredLicense = append(expiredLicense, ExpiringLicenseOutput{
					ProductName:  expiredProduct.Type,
					GUID:         expiredProduct.GUID,
					ExpiresAt:    t,
					ProductState: expiredProduct.ProductState,
				})
			}
		}
	}
	return expiredLicense, err
}

func (a Api) getProductsLicenseInfo(staged bool, deployed bool) ([]expiringProduct, error) {
	expiringProducts := []expiringProduct{}

	noModifiersSelected := !staged && !deployed
	if staged || noModifiersSelected {
		err := a.addStagedProducts(&expiringProducts)

		if err != nil {
			return nil, fmt.Errorf("could not get staged products: %w", err)
		}

	}
	if deployed || noModifiersSelected {
		err := a.addDeployedProducts(&expiringProducts)

		if err != nil {
			return nil, fmt.Errorf("could not get deployed products: %w", err)
		}

	}
	removeDuplicates(&expiringProducts)
	return expiringProducts, nil
}

func (a Api) addStagedProducts(expiringProducts *[]expiringProduct) error {
	stagedProducts, err := a.ListStagedProducts()

	if err != nil {
		return fmt.Errorf("could not make a call to ListStagedProducts api: %w", err)
	}

	for _, stagedProduct := range stagedProducts.Products {
		*expiringProducts = append(*expiringProducts, expiringProduct{GUID: stagedProduct.GUID,
			Type:            stagedProduct.Type,
			LicenseMetadata: stagedProduct.LicenseMetadata,
			ProductState:    ProductStateStaged,
		})
	}

	return nil
}

func (a Api) addDeployedProducts(expiringProducts *[]expiringProduct) error {
	deployedProducts, err := a.ListDeployedProducts()

	if err != nil {
		return fmt.Errorf("could not make a call to ListDeployedProducts api: %w", err)
	}

	for _, deployedProduct := range deployedProducts {
		*expiringProducts = append(*expiringProducts, expiringProduct{GUID: deployedProduct.GUID,
			Type:            deployedProduct.Type,
			LicenseMetadata: deployedProduct.LicenseMetadata,
			ProductState:    ProductStateDeployed,
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
