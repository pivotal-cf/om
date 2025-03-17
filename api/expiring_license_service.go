package api

import (
	"fmt"
	"time"
)

// not the acutal response, will need to be updated to match the actual response
type ExpiringLicenseOutPut struct {
	ProductName string    `json:"product_name"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (a Api) ListExpiringLicenses(expiresWithin string) ([]ExpiringLicenseOutPut, error) {
	// TODO: Implement, should call to deployed products and staged services

	expiredLicense := []ExpiringLicenseOutPut{}
	expiredProducts, err := a.ListDeployedProducts()
	layout := "2006-01-02"

	if err != nil {
		return nil, fmt.Errorf("Cannot list deployed products: %w", err)
	}

	for _, expiredProduct := range expiredProducts {

		if len(expiredProduct.LicenseMetadata) == 0 {
			continue
		}

		t, err := time.Parse(layout, expiredProduct.LicenseMetadata[0].ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("could not make convert expiry date string to time: %w", err)
		}
		expiredLicense = append(expiredLicense, ExpiringLicenseOutPut{ProductName: expiredProduct.GUID, ExpiresAt: t})

	}
	return expiredLicense, err
}
