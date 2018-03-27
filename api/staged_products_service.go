package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type StageProductInput struct {
	ProductName    string `json:"name"`
	ProductVersion string `json:"product_version"`
}

type StagedProductsOutput struct {
	Products []StagedProduct
}

type StagedProductsFindOutput struct {
	Product StagedProduct
}

type StagedProduct struct {
	GUID string
	Type string
}

type UnstageProductInput struct {
	ProductName string `json:"name"`
}

type ProductsConfigurationInput struct {
	GUID          string
	Configuration string
	Network       string
}

type StagedProductsService struct {
	client httpClient
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

func (p StagedProductsService) Stage(input StageProductInput, deployedGUID string) error {
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

func (p StagedProductsService) Unstage(input UnstageProductInput) error {
	stagedGUID, err := p.checkStagedProducts(input.ProductName)
	if err != nil {
		return err
	}

	if len(stagedGUID) == 0 {
		return fmt.Errorf("product is not staged: %s", input.ProductName)
	}

	var req *http.Request
	req, err = http.NewRequest("DELETE", fmt.Sprintf("/api/v0/staged/products/%s", stagedGUID), strings.NewReader("{}"))

	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make %s api request to staged products endpoint: %s", req.Method, err)
	}
	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (p StagedProductsService) List() (StagedProductsOutput, error) {
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

func (p StagedProductsService) Find(productName string) (StagedProductsFindOutput, error) {
	productsOutput, err := p.List()
	if err != nil {
		return StagedProductsFindOutput{}, err
	}

	var foundProduct StagedProduct
	for _, product := range productsOutput.Products {
		if product.Type == productName {
			foundProduct = product
			break
		}
	}

	if (foundProduct == StagedProduct{}) {
		return StagedProductsFindOutput{}, fmt.Errorf("could not find product %q", productName)
	}

	return StagedProductsFindOutput{Product: foundProduct}, nil
}

func (p StagedProductsService) Manifest(guid string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/manifest", guid), nil)
	if err != nil {
		return "", err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not make api request to staged products manifest endpoint: %s", err)
	}

	if err = ValidateStatusOK(resp); err != nil {
		return "", err
	}

	defer resp.Body.Close()
	var contents struct {
		Manifest interface{} `json:"manifest"`
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err = yaml.Unmarshal(body, &contents); err != nil {
		return "", fmt.Errorf("could not parse json: %s", err)
	}

	manifest, err := yaml.Marshal(contents.Manifest)
	if err != nil {
		return "", err // this should never happen, all valid json can be marshalled
	}

	return string(manifest), nil
}

func (p StagedProductsService) checkStagedProducts(productName string) (string, error) {
	stagedProductsOutput, err := p.List()
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
