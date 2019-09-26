package api

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
)

const listProductVerifiersEndpointTemplate = "/api/v0/staged/products/%s/verifiers/install_time"
const disableProductVerifiersEndpointTemplate = "/api/v0/staged/products/%s/verifiers/install_time/%s"

func (a Api) ListProductVerifiers(productName string) ([]Verifier, string, error) {
	stagedProduct, err := a.GetStagedProductByName(productName)
	if err != nil {
		return nil, "", err
	}

	resp, err := a.sendAPIRequest("GET", fmt.Sprintf(listProductVerifiersEndpointTemplate, stagedProduct.Product.GUID), nil)
	if err != nil {
		return nil, "", errors.Wrap(err, "could not make api request to list_product_verifiers endpoint")
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return nil, "", err
	}

	verifiersBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var verifierResponse verifiersResponse
	if err := json.Unmarshal(verifiersBytes, &verifierResponse); err != nil {
		return nil, "", errors.Wrap(err, "could not unmarshal list_product_verifiers response")
	}

	return verifierResponse.Verifiers, stagedProduct.Product.GUID, nil
}

func (a Api) DisableProductVerifiers(verifiers []string, productGUID string) error {
	for _, verifier := range verifiers {
		resp, err := a.sendAPIRequest("PUT", fmt.Sprintf(disableProductVerifiersEndpointTemplate, productGUID, verifier), []byte(`{ "enabled": false }`))
		if err != nil {
			return errors.Wrap(err, "could not make api request to disable_product_verifiers endpoint")
		}
		resp.Body.Close()

		if err = validateStatusOK(resp); err != nil {
			return err
		}
	}

	return nil
}
