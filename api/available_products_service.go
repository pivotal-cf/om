package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

const availableProductsEndpoint = "/api/v0/available_products"

type UploadAvailableProductInput struct {
	ContentLength   int64
	Product         io.Reader
	ContentType     string
	PollingInterval int
}

type ProductInfo struct {
	Name    string `json:"name"`
	Version string `json:"product_version"`
}

type UploadAvailableProductOutput struct{}

type AvailableProductsOutput struct {
	ProductsList []ProductInfo
}

type DeleteAvailableProductsInput struct {
	ProductName             string
	ProductVersion          string
	ShouldDeleteAllProducts bool
}

func (a Api) UploadAvailableProduct(input UploadAvailableProductInput) (UploadAvailableProductOutput, error) {
	req, err := http.NewRequest("POST", availableProductsEndpoint, input.Product)
	if err != nil {
		return UploadAvailableProductOutput{}, err
	}

	req.Header.Set("Content-Type", input.ContentType)
	req.ContentLength = input.ContentLength

	req = req.WithContext(context.WithValue(req.Context(), "polling-interval", time.Duration(input.PollingInterval)*time.Second))

	resp, err := a.progressClient.Do(req)
	if err != nil {
		return UploadAvailableProductOutput{}, errors.Wrap(err, "could not make api request to available_products endpoint")
	}

	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return UploadAvailableProductOutput{}, err
	}

	return UploadAvailableProductOutput{}, nil
}

func (a Api) ListAvailableProducts() (AvailableProductsOutput, error) {
	resp, err := a.sendAPIRequest("GET", availableProductsEndpoint, nil)
	if err != nil {
		return AvailableProductsOutput{}, errors.Wrap(err, "could not make api request to available_products endpoint")
	}

	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return AvailableProductsOutput{}, err
	}

	var availableProducts []ProductInfo
	if err := json.NewDecoder(resp.Body).Decode(&availableProducts); err != nil {
		return AvailableProductsOutput{}, errors.Wrap(err, "could not unmarshal available_products response")
	}

	return AvailableProductsOutput{ProductsList: availableProducts}, nil
}

func (a Api) DeleteAvailableProducts(input DeleteAvailableProductsInput) error {
	req, err := http.NewRequest("DELETE", availableProductsEndpoint, nil)

	if !input.ShouldDeleteAllProducts {
		query := url.Values{}
		query.Add("product_name", input.ProductName)
		query.Add("version", input.ProductVersion)

		req.URL.RawQuery = query.Encode()
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not make api request to available_products endpoint")
	}

	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
