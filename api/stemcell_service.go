package api

import (
	"fmt"
	"io"
	"net/http"
)

type StemcellUploadInput struct {
	ContentLength int64
	Stemcell      io.Reader
	ContentType   string
}

type StemcellUploadOutput struct{}

type UploadStemcellService struct {
	client httpClient
}

func NewUploadStemcellService(client httpClient) UploadStemcellService {
	return UploadStemcellService{
		client: client,
	}
}

func (us UploadStemcellService) UploadStemcell(input StemcellUploadInput) (StemcellUploadOutput, error) {
	req, err := http.NewRequest("POST", "/api/v0/stemcells", input.Stemcell)
	if err != nil {
		return StemcellUploadOutput{}, err
	}

	req.Header.Set("Content-Type", input.ContentType)
	req.ContentLength = input.ContentLength

	resp, err := us.client.Do(req)
	if err != nil {
		return StemcellUploadOutput{}, fmt.Errorf("could not make api request to stemcells endpoint: %s", err)
	}

	defer resp.Body.Close()

	return StemcellUploadOutput{}, ValidateStatusOK(resp)
}
