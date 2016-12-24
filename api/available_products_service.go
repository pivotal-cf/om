package api

import (
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

type ProductInfo struct {
	Name    string `json:"name"`
	Version string `json:"product_version"`
}

type UploadProductOutput struct{}

type AvailableProductsOutput struct {
	ProductsList []ProductInfo
}

type AvailableProductsService struct {
	client     httpClient
	progress   progress
	liveWriter liveWriter
}

func NewAvailableProductsService(client httpClient, progress progress, liveWriter liveWriter) AvailableProductsService {
	return AvailableProductsService{
		client:     client,
		progress:   progress,
		liveWriter: liveWriter,
	}
}

func (ap AvailableProductsService) Upload(input UploadProductInput) (UploadProductOutput, error) {
	ap.progress.SetTotal(input.ContentLength)
	body := ap.progress.NewBarReader(input.Product)

	req, err := http.NewRequest("POST", "/api/v0/available_products", body)
	if err != nil {
		return UploadProductOutput{}, err
	}

	req.Header.Set("Content-Type", input.ContentType)
	req.ContentLength = input.ContentLength

	ap.progress.Kickoff()
	respChan := make(chan error)
	go func() {
		for {
			if ap.progress.GetCurrent() == ap.progress.GetTotal() {
				ap.progress.End()
				break
			}

			time.Sleep(1 * time.Second)
		}

		var elapsedTime int
		ap.liveWriter.Start()
		liveLog := log.New(ap.liveWriter, "", 0)

		for {
			select {
			case _ = <-respChan:
				ap.liveWriter.Stop()
				return
			default:
				time.Sleep(1 * time.Second)
				elapsedTime++
				liveLog.Printf("%ds elapsed, waiting for response from Ops Manager...\r", elapsedTime)
			}
		}
	}()

	resp, err := ap.client.Do(req)
	respChan <- err
	if err != nil {
		return UploadProductOutput{}, fmt.Errorf("could not make api request to available_products endpoint: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return UploadProductOutput{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return UploadProductOutput{}, fmt.Errorf("request failed: unexpected response:\n%s", out)
	}

	return UploadProductOutput{}, nil
}

func (ap AvailableProductsService) Trash() error {
	req, err := http.NewRequest("DELETE", "/api/v0/available_products", nil)
	if err != nil {
		return err
	}

	resp, err := ap.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make api request to available_products endpoint: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return fmt.Errorf("request failed: unexpected response: %s", err)
		}
		return fmt.Errorf("could not make api request to available_products endpoint: unexpected response.\n%s", out)
	}

	return nil
}

func (ap AvailableProductsService) List() (AvailableProductsOutput, error) {
	avReq, err := http.NewRequest("GET", "/api/v0/available_products", nil)
	if err != nil {
		return AvailableProductsOutput{}, err
	}

	avResp, err := ap.client.Do(avReq)
	if err != nil {
		return AvailableProductsOutput{}, fmt.Errorf("could not make api request to available_products endpoint: %s", err)
	}
	defer avResp.Body.Close()

	if avResp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(avResp, true)
		if err != nil {
			return AvailableProductsOutput{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}
		return AvailableProductsOutput{}, fmt.Errorf("could not make api request to available_products endpoint: unexpected response.\n%s", out)
	}

	avRespBody, err := ioutil.ReadAll(avResp.Body)
	if err != nil {
		return AvailableProductsOutput{}, err
	}

	var availableProducts []ProductInfo
	err = json.Unmarshal(avRespBody, &availableProducts)
	if err != nil {
		return AvailableProductsOutput{}, fmt.Errorf("could not unmarshal available_products response: %s", err)
	}

	return AvailableProductsOutput{ProductsList: availableProducts}, nil
}

func (ap AvailableProductsService) CheckProductAvailability(productName string, productVersion string) (bool, error) {
	availableProducts, err := ap.List()
	if err != nil {
		return false, err
	}

	for _, product := range availableProducts.ProductsList {
		if product.Name == productName && product.Version == productVersion {
			return true, nil
		}
	}

	return false, nil
}
