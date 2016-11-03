package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

type UploadProductInput struct {
	ContentLength int64
	Product       io.Reader
	ContentType   string
}

type UploadProductOutput struct{}

type StageProductInput struct {
	ProductName    string
	ProductVersion string
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

type ProductConfiguration struct {
	Properties map[string]interface{}
}

type ProductsService struct {
	client     httpClient
	progress   progress
	liveWriter liveWriter
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

func NewProductsService(client httpClient, progress progress, liveWriter liveWriter) ProductsService {
	return ProductsService{
		client:     client,
		progress:   progress,
		liveWriter: liveWriter,
	}
}

func (p ProductsService) Upload(input UploadProductInput) (UploadProductOutput, error) {
	p.progress.SetTotal(input.ContentLength)
	body := p.progress.NewBarReader(input.Product)

	req, err := http.NewRequest("POST", "/api/v0/available_products", body)
	if err != nil {
		return UploadProductOutput{}, err
	}

	req.Header.Set("Content-Type", input.ContentType)
	req.ContentLength = input.ContentLength

	p.progress.Kickoff()
	respChan := make(chan error)
	go func() {
		var elapsedTime int
		var liveLog logger
		for {
			select {
			case _ = <-respChan:
				p.liveWriter.Stop()
				return
			default:
				time.Sleep(1 * time.Second)
				if p.progress.GetCurrent() == p.progress.GetTotal() { // so that we only start logging time elapsed after the progress bar is done
					p.progress.End()
					if elapsedTime == 0 {
						p.liveWriter.Start()
						liveLog = log.New(p.liveWriter, "", 0)
					}
					elapsedTime++
					liveLog.Printf("%ds elapsed, waiting for response from Ops Manager...\r", elapsedTime)
				}
			}
		}
	}()

	resp, err := p.client.Do(req)
	respChan <- err
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

func (p ProductsService) Stage(input StageProductInput) error {
	productToStage, err := p.checkAvailableProducts(input.ProductName, input.ProductVersion)
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
		out, err := httputil.DumpResponse(stResp, true)
		if err != nil {
			return fmt.Errorf("request failed: unexpected response: %s", err)
		}
		return fmt.Errorf("could not make api request to staged products endpoint: unexpected response. Please make sure the product you are adding is compatible with everything that is currently staged/deployed.\n%s", out)
	}

	return nil
}

func (p ProductsService) StagedProducts() (StagedProductsOutput, error) {
	req, err := http.NewRequest("GET", "/api/v0/staged/products", nil)
	if err != nil {
		return StagedProductsOutput{}, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return StagedProductsOutput{}, fmt.Errorf("could not make api request to staged products endpoint: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return StagedProductsOutput{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}
		return StagedProductsOutput{}, fmt.Errorf("could not make api request to staged products endpoint: unexpected response.\n%s", out)
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

func (p ProductsService) Configure(input ProductsConfigurationInput) error {
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

		if resp.StatusCode != http.StatusOK {
			out, err := httputil.DumpResponse(resp, true)
			if err != nil {
				return fmt.Errorf("request failed: unexpected response: %s", err)
			}
			return fmt.Errorf("could not make api request to staged product properties endpoint: unexpected response.\n%s", out)
		}
	}

	return nil
}

func createConfigureRequests(input ProductsConfigurationInput) ([]*http.Request, error) {
	var reqList []*http.Request

	configurations := []struct {
		Method        string
		URL           string
		Configuration string
	}{
		{
			Method:        "PUT",
			URL:           fmt.Sprintf("/api/v0/staged/products/%s/properties", input.GUID),
			Configuration: fmt.Sprintf(`{"properties": %s}`, input.Configuration),
		},
		{
			Method:        "PUT",
			URL:           fmt.Sprintf("/api/v0/staged/products/%s/networks_and_azs", input.GUID),
			Configuration: fmt.Sprintf(`{"networks_and_azs": %s}`, input.Network),
		},
	}

	for _, config := range configurations {
		if config.Configuration == "" {
			continue
		}

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

func (p ProductsService) checkAvailableProducts(productName string, productVersion string) (ProductInfo, error) {
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
		out, err := httputil.DumpResponse(avResp, true)
		if err != nil {
			return ProductInfo{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}
		return ProductInfo{}, fmt.Errorf("could not make api request to available_products endpoint: unexpected response.\n%s", out)
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
		if product.Name == productName && product.Version == productVersion {
			foundProduct = product
			prodFound = true
			break
		}
	}

	if !prodFound {
		return ProductInfo{}, fmt.Errorf("cannot find product %s %s", productName, productVersion)
	}

	return foundProduct, nil
}

func (p ProductsService) checkDeployedProducts(productName string) (string, error) {
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
		out, err := httputil.DumpResponse(depResp, true)
		if err != nil {
			return "", fmt.Errorf("request failed: unexpected response: %s", err)
		}
		return "", fmt.Errorf("could not make api request to deployed products endpoint: unexpected response.\n%s", out)
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
