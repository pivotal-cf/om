package commands

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/pivotal-cf/om/api"
)

type GetCertificates struct {
	api     api.Api
	logger  logger
	Options struct {
		Product string `long:"product" short:"p" required:"true" description:"product type to filter certificates (e.g., p-bosh, cf)"`
	}
}

func NewGetCertificates(apiClient api.Api, logger logger) *GetCertificates {
	return &GetCertificates{
		api:    apiClient,
		logger: logger,
	}
}

func (cmd *GetCertificates) Execute(args []string) error {
	certs, err := cmd.api.ListDeployedCertificates()
	if err != nil {
		return fmt.Errorf("failed to fetch deployed certificates: %w", err)
	}

	products, err := cmd.api.ListDeployedProducts()
	if err != nil {
		return fmt.Errorf("failed to fetch deployed products: %w", err)
	}
	guidToType := map[string]string{}
	for _, p := range products {
		guidToType[p.GUID] = p.Type
	}

	// Filter certificates by product type
	var filteredCerts []api.ExpiringCertificate
	for _, cert := range certs {
		if cert.ProductGUID != "" {
			if productType, ok := guidToType[cert.ProductGUID]; ok && productType == cmd.Options.Product {
				filteredCerts = append(filteredCerts, cert)
			}
		}
	}

	if len(filteredCerts) == 0 {
		cmd.logger.Printf("No certificates found for product '%s'", cmd.Options.Product)
		return nil
	}

	cmd.logger.Printf("Processing %d certificates (this may take a moment)...", len(filteredCerts))

	type certWithSerial struct {
		api.ExpiringCertificate
		Serial string `json:"serial_number"`
	}

	results := make([]certWithSerial, len(filteredCerts))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // limit concurrency to 10

	for i, cert := range filteredCerts {
		wg.Add(1)
		go func(i int, cert api.ExpiringCertificate) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			serial := ""
			if cert.ProductGUID != "" && cert.PropertyReference != "" {
				cred, err := cmd.api.GetDeployedProductCredential(api.GetDeployedProductCredentialInput{
					DeployedGUID:        cert.ProductGUID,
					CredentialReference: cert.PropertyReference,
				})
				if err == nil {
					pem, ok := cred.Credential.Value["cert_pem"]
					if ok && pem != "" {
						serial, _ = extractSerialFromPEM(pem)
					}
				}
			}
			results[i] = certWithSerial{
				ExpiringCertificate: cert,
				Serial:              serial,
			}
		}(i, cert)
	}

	wg.Wait()

	// Pretty-print the results as JSON using logger
	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}
	cmd.logger.Print(string(jsonBytes))
	return nil
}

func extractSerialFromPEM(pemData string) (string, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	return cert.SerialNumber.Text(16), nil // hex string, like openssl
}
