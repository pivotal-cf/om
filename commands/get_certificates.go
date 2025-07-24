package commands

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"sync"

	"github.com/pivotal-cf/om/api"
)

type GetCertificates struct {
	api api.Api
}

func NewGetCertificates(apiClient api.Api) *GetCertificates {
	return &GetCertificates{
		api: apiClient,
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

	type certWithSerial struct {
		api.ExpiringCertificate
		Serial string `json:"serial_number"`
	}

	results := make([]certWithSerial, len(certs))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // limit concurrency to 10
	progressEvery := 10
	progressCount := 0
	progressLock := sync.Mutex{}

	for i, cert := range certs {
		wg.Add(1)
		go func(i int, cert api.ExpiringCertificate) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			serial := ""
			if cert.ProductGUID != "" && cert.PropertyReference != "" {
				_, ok := guidToType[cert.ProductGUID]
				if ok {
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
			}
			results[i] = certWithSerial{
				ExpiringCertificate: cert,
				Serial:              serial,
			}

			progressLock.Lock()
			progressCount++
			if progressCount%progressEvery == 0 {
				fmt.Fprint(os.Stderr, ".")
			}
			progressLock.Unlock()
		}(i, cert)
	}

	wg.Wait()
	fmt.Fprintln(os.Stderr) // finish progress line

	// Pretty-print the results as JSON, with serial number included in each cert
	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}
	fmt.Println(string(jsonBytes))
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
