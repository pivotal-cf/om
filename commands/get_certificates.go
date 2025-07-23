package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

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
	var results []certWithSerial

	for _, cert := range certs {
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
						tmpfile, err := ioutil.TempFile("", "om-cert-*.crt")
						if err == nil {
							defer os.Remove(tmpfile.Name())
							tmpfile.WriteString(pem)
							tmpfile.Close()
							out, err := exec.Command("openssl", "x509", "-noout", "-serial", "-in", tmpfile.Name()).Output()
							if err == nil {
								serialLine := string(out)
								if len(serialLine) > 7 && serialLine[:7] == "serial=" {
									serial = serialLine[7:]
								}
							}
						}
					}
				}
			}
		}
		results = append(results, certWithSerial{
			ExpiringCertificate: cert,
			Serial:              serial,
		})
	}

	// For now, just print the results
	fmt.Printf("%+v\n", results)
	return nil
}
