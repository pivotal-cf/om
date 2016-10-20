package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

type UploadProductInput struct {
	ContentLength int64
	Product       io.Reader
	ContentType   string
}

type StageProductInput struct {
	ProductName string
}

type UploadProductOutput struct{}

type ProductService struct {
	client   httpClient
	progress progress
}

type ProductInfo struct {
	Name    string `json:"name"`
	Version string `json:"product_version"`
}

type DeployedProductInfo struct {
	Type             string
	GUID             string
	InstallationName string `json:"installation_name"`
}

type UpgradeRequest struct {
	ToVersion string `json:"to_version"`
}

func NewProductService(client httpClient, progress progress) ProductService {
	return ProductService{
		client:   client,
		progress: progress,
	}
}

func (p ProductService) Upload(input UploadProductInput) (UploadProductOutput, error) {
	p.progress.SetTotal(input.ContentLength)
	body := p.progress.NewBarReader(input.Product)

	req, err := http.NewRequest("POST", "/api/v0/available_products", body)
	if err != nil {
		return UploadProductOutput{}, err
	}

	req.Header.Set("Content-Type", input.ContentType)
	req.ContentLength = input.ContentLength

	p.progress.Kickoff()

	resp, err := p.client.Do(req)
	if err != nil {
		return UploadProductOutput{}, fmt.Errorf("could not make api request to available_products endpoint: %s", err)
	}

	defer resp.Body.Close()

	p.progress.End()

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return UploadProductOutput{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return UploadProductOutput{}, fmt.Errorf("request failed: unexpected response:\n%s", out)
	}

	return UploadProductOutput{}, nil
}

func (p ProductService) Stage(input StageProductInput) error {
	productToStage, err := p.checkAvailableProducts(input.ProductName)
	if err != nil {
		return err
	}

	deployedGuid, err := p.checkDeployedProducts(input.ProductName)
	if err != nil {
		return err
	}

	var stReq *http.Request
	if deployedGuid == "" {
		stagedProductBody, err := json.Marshal(productToStage)
		if err != nil {
			return err
		}

		stReq, err = http.NewRequest("POST", "/api/v0/staged/products", bytes.NewBuffer(stagedProductBody))
		if err != nil {
			return err
		}
	} else {
		upgradeReq := UpgradeRequest{
			ToVersion: productToStage.Version,
		}

		upgradeReqBody, err := json.Marshal(upgradeReq)
		if err != nil {
			return err
		}

		stReq, err = http.NewRequest("PUT", fmt.Sprintf("/api/v0/staged/products/%s", deployedGuid), bytes.NewBuffer(upgradeReqBody))
		if err != nil {
			return err
		}
	}

	stReq.Header.Set("Content-Type", "application/json")
	stResp, err := p.client.Do(stReq)
	if err != nil {
		return fmt.Errorf("could not make api request to staged products endpoint: %s", err)
	}
	defer stResp.Body.Close()

	if stResp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not make api request to staged products endpoint: unexpected response %d. Please make sure the product you are adding is compatible with everything that is currently staged/deployed.", stResp.StatusCode)
	}

	return nil
}

func (p ProductService) checkAvailableProducts(productName string) (ProductInfo, error) {
	avReq, err := http.NewRequest("GET", "/api/v0/available_products", nil)
	if err != nil {
		return ProductInfo{}, err
	}

	avResp, err := p.client.Do(avReq)
	if err != nil {
		return ProductInfo{}, fmt.Errorf("could not make api request to available_products endpoint: %s", err)
	}
	defer avResp.Body.Close()

	if avResp.StatusCode != http.StatusOK {
		return ProductInfo{}, fmt.Errorf("could not make api request to available_products endpoint: unexpected response %d", avResp.StatusCode)
	}

	avRespBody, err := ioutil.ReadAll(avResp.Body)
	if err != nil {
		return ProductInfo{}, err
	}

	var availableProducts []ProductInfo
	err = json.Unmarshal(avRespBody, &availableProducts)
	if err != nil {
		return ProductInfo{}, fmt.Errorf("could not unmarshal available_products response: %s", err)
	}

	var foundProduct ProductInfo
	var prodFound bool
	for _, product := range availableProducts {
		if product.Name == productName {
			foundProduct = product
			prodFound = true
			break
		}
	}

	if !prodFound {
		return ProductInfo{}, fmt.Errorf("cannot find product %s", productName)
	}

	return foundProduct, nil
}

func (p ProductService) checkDeployedProducts(productName string) (string, error) {
	depReq, err := http.NewRequest("GET", "/api/v0/deployed/products", nil)
	if err != nil {
		return "", err
	}

	depResp, err := p.client.Do(depReq)
	if err != nil {
		return "", fmt.Errorf("could not make api request to deployed products endpoint: %s", err)
	}
	defer depResp.Body.Close()

	if depResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("could not make api request to deployed products endpoint: unexpected response %d", depResp.StatusCode)
	}

	depRespBody, err := ioutil.ReadAll(depResp.Body)
	if err != nil {
		return "", err
	}

	var deployedProducts []DeployedProductInfo
	err = json.Unmarshal(depRespBody, &deployedProducts)
	if err != nil {
		return "", err
	}

	for _, product := range deployedProducts {
		if product.Type == productName {
			return product.GUID, nil
		}
	}

	return "", nil
}
