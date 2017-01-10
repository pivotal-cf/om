package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type StageProductInput struct {
	ProductName    string `json:"name"`
	ProductVersion string `json:"product_version"`
}

type StagedProductsOutput struct {
	Products []StagedProduct
}

type StagedProduct struct {
	GUID string
	Type string
}

type ProductsConfigurationInput struct {
	GUID          string
	Configuration string
	Network       string
}

type StagedProductsService struct {
	client httpClient
}

type DeployedProductInfo struct {
	Type             string
	GUID             string
	InstallationName string `json:"installation_name"`
}

type UpgradeRequest struct {
	ToVersion string `json:"to_version"`
}

type ConfigurationRequest struct {
	Method        string
	URL           string
	Configuration string
}

func NewStagedProductsService(client httpClient) StagedProductsService {
	return StagedProductsService{
		client: client,
	}
}

func (p StagedProductsService) Stage(input StageProductInput) error {
	deployedGUID, err := p.checkDeployedProducts(input.ProductName)
	if err != nil {
		return err
	}

	stagedGUID, err := p.checkStagedProducts(input.ProductName)
	if err != nil {
		return err
	}

	var stReq *http.Request
	if deployedGUID == "" && stagedGUID == "" {
		stagedProductBody, err := json.Marshal(input)
		if err != nil {
			return err
		}

		stReq, err = http.NewRequest("POST", "/api/v0/staged/products", bytes.NewBuffer(stagedProductBody))
		if err != nil {
			return err
		}
	} else if deployedGUID != "" {
		upgradeReq := UpgradeRequest{
			ToVersion: input.ProductVersion,
		}

		upgradeReqBody, err := json.Marshal(upgradeReq)
		if err != nil {
			return err
		}

		stReq, err = http.NewRequest("PUT", fmt.Sprintf("/api/v0/staged/products/%s", deployedGUID), bytes.NewBuffer(upgradeReqBody))
		if err != nil {
			return err
		}
	} else if stagedGUID != "" {
		upgradeReq := UpgradeRequest{
			ToVersion: input.ProductVersion,
		}

		upgradeReqBody, err := json.Marshal(upgradeReq)
		if err != nil {
			return err
		}

		stReq, err = http.NewRequest("PUT", fmt.Sprintf("/api/v0/staged/products/%s", stagedGUID), bytes.NewBuffer(upgradeReqBody))
		if err != nil {
			return err
		}
	}

	stReq.Header.Set("Content-Type", "application/json")
	stResp, err := p.client.Do(stReq)
	if err != nil {
		return fmt.Errorf("could not make %s api request to staged products endpoint: %s", stReq.Method, err)
	}
	defer stResp.Body.Close()

	if err = ValidateStatusOK(stResp); err != nil {
		return err
	}

	return nil
}

func (p StagedProductsService) StagedProducts() (StagedProductsOutput, error) {
	req, err := http.NewRequest("GET", "/api/v0/staged/products", nil)
	if err != nil {
		return StagedProductsOutput{}, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return StagedProductsOutput{}, fmt.Errorf("could not make api request to staged products endpoint: %s", err)
	}
	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return StagedProductsOutput{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return StagedProductsOutput{}, err
	}

	var stagedProducts []StagedProduct
	err = json.Unmarshal(respBody, &stagedProducts)
	if err != nil {
		return StagedProductsOutput{}, fmt.Errorf("could not unmarshal staged products response: %s", err)
	}

	return StagedProductsOutput{
		Products: stagedProducts,
	}, nil
}

func (p StagedProductsService) Configure(input ProductsConfigurationInput) error {
	reqList, err := createConfigureRequests(input)
	if err != nil {
		return err
	}

	for _, req := range reqList {
		resp, err := p.client.Do(req)
		if err != nil {
			return fmt.Errorf("could not make api request to staged product properties endpoint: %s", err)
		}
		defer resp.Body.Close()

		if err = ValidateStatusOK(resp); err != nil {
			return err
		}
	}

	return nil
}

func createConfigureRequests(input ProductsConfigurationInput) ([]*http.Request, error) {
	var reqList []*http.Request

	var configurations []ConfigurationRequest

	if input.Configuration != "" {
		configurations = append(configurations,
			ConfigurationRequest{
				Method:        "PUT",
				URL:           fmt.Sprintf("/api/v0/staged/products/%s/properties", input.GUID),
				Configuration: fmt.Sprintf(`{"properties": %s}`, input.Configuration),
			},
		)
	}

	if input.Network != "" {
		configurations = append(configurations,
			ConfigurationRequest{
				Method:        "PUT",
				URL:           fmt.Sprintf("/api/v0/staged/products/%s/networks_and_azs", input.GUID),
				Configuration: fmt.Sprintf(`{"networks_and_azs": %s}`, input.Network),
			},
		)
	}

	for _, config := range configurations {
		body := bytes.NewBufferString(config.Configuration)
		req, err := http.NewRequest(config.Method, config.URL, body)
		if err != nil {
			return reqList, err
		}

		req.Header.Set("Content-Type", "application/json")

		reqList = append(reqList, req)
	}

	return reqList, nil
}

func (p StagedProductsService) checkDeployedProducts(productName string) (string, error) {
	depReq, err := http.NewRequest("GET", "/api/v0/deployed/products", nil)
	if err != nil {
		return "", err
	}

	depResp, err := p.client.Do(depReq)
	if err != nil {
		return "", fmt.Errorf("could not make api request to deployed products endpoint: %s", err)
	}
	defer depResp.Body.Close()

	if err = ValidateStatusOK(depResp); err != nil {
		return "", err
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

func (p StagedProductsService) checkStagedProducts(productName string) (string, error) {
	stagedProductsOutput, err := p.StagedProducts()
	if err != nil {
		return "", err
	}

	for _, product := range stagedProductsOutput.Products {
		if productName == product.Type {
			return product.GUID, nil
		}
	}

	return "", nil
}
