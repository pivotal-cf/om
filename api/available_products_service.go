package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const availableProductsEndpoint = "/api/v0/available_products"

type UploadProductInput struct {
	ContentLength   int64
	Product         io.Reader
	ContentType     string
	PollingInterval int
}

type ProductInfo struct {
	Name    string `json:"name"`
	Version string `json:"product_version"`
}

type UploadProductOutput struct{}

type AvailableProductsOutput struct {
	ProductsList []ProductInfo
}

type AvailableProductsInput struct {
	ProductName    string
	ProductVersion string
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

	req, err := http.NewRequest("POST", availableProductsEndpoint, body)
	if err != nil {
		return UploadProductOutput{}, err
	}

	req.Header.Set("Content-Type", input.ContentType)
	req.ContentLength = input.ContentLength

	requestComplete := make(chan bool)
	progressComplete := make(chan bool)

	go func() {
		ap.progress.Kickoff()
		ap.liveWriter.Start()

		for {
			select {
			case <-requestComplete:
				ap.progress.End()
				ap.liveWriter.Stop()
				progressComplete <- true
				return
			default:
				if ap.progress.GetCurrent() != ap.progress.GetTotal() {
					time.Sleep(time.Second)
					continue
				}

				ap.progress.End()

				liveLog := log.New(ap.liveWriter, "", 0)
				startTime := time.Now().Round(time.Second)
				ticker := time.NewTicker(time.Duration(input.PollingInterval) * time.Second)

				for {
					select {
					case <-requestComplete:
						ticker.Stop()
						ap.liveWriter.Stop()
						progressComplete <- true
					case now := <-ticker.C:
						liveLog.Printf("%s elapsed, waiting for response from Ops Manager...\r", now.Round(time.Second).Sub(startTime).String())
					}
				}
			}
		}
	}()

	resp, err := ap.client.Do(req)
	requestComplete <- true
	<-progressComplete

	if err != nil {
		return UploadProductOutput{}, fmt.Errorf("could not make api request to available_products endpoint: %s", err)
	}

	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return UploadProductOutput{}, err
	}

	return UploadProductOutput{}, nil
}

func (ap AvailableProductsService) List() (AvailableProductsOutput, error) {
	avReq, err := http.NewRequest("GET", availableProductsEndpoint, nil)
	if err != nil {
		return AvailableProductsOutput{}, err
	}

	resp, err := ap.client.Do(avReq)
	if err != nil {
		return AvailableProductsOutput{}, fmt.Errorf("could not make api request to available_products endpoint: %s", err)
	}
	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return AvailableProductsOutput{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return AvailableProductsOutput{}, err
	}

	var availableProducts []ProductInfo
	err = json.Unmarshal(respBody, &availableProducts)
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

func (ap AvailableProductsService) Delete(input AvailableProductsInput, all bool) error {
	req, err := http.NewRequest("DELETE", availableProductsEndpoint, nil)

	if !all {
		query := url.Values{}
		query.Add("product_name", input.ProductName)
		query.Add("version", input.ProductVersion)

		req.URL.RawQuery = query.Encode()
	}

	resp, err := ap.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make api request to available_products endpoint: %s", err)
	}

	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
