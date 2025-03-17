package api

import "time"

// not the acutal response, will need to be updated to match the actual response
type ExpiringLicense struct {
	ProductName string    `json:"product_name"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (a Api) ListExpiringLicenses(expiresWithin string) ([]ExpiringLicense, error) {
	// TODO: Implement, should call to deployed proejcts and staged services
	return nil, nil
}
