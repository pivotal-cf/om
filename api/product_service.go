package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

type UploadProductInput struct {
	ContentLength int64
	Product       io.Reader
	ContentType   string
}

type UploadProductOutput struct{}

type UploadProductService struct {
	client   httpClient
	progress progress
}

func NewUploadProductService(client httpClient, progress progress) UploadProductService {
	return UploadProductService{
		client:   client,
		progress: progress,
	}
}

func (up UploadProductService) Upload(input UploadProductInput) (UploadProductOutput, error) {
	up.progress.SetTotal(input.ContentLength)
	body := up.progress.NewBarReader(input.Product)

	req, err := http.NewRequest("POST", "/api/v0/available_products", body)
	if err != nil {
		return UploadProductOutput{}, err
	}

	req.Header.Set("Content-Type", input.ContentType)
	req.ContentLength = input.ContentLength

	up.progress.Kickoff()

	resp, err := up.client.Do(req)
	if err != nil {
		return UploadProductOutput{}, fmt.Errorf("could not make api request to available_products endpoint: %s", err)
	}

	defer resp.Body.Close()

	up.progress.End()

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return UploadProductOutput{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return UploadProductOutput{}, fmt.Errorf("request failed: unexpected response:\n%s", out)
	}

	return UploadProductOutput{}, nil
}
