package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

type StemcellUploadInput struct {
	ContentLength int64
	Stemcell      io.Reader
	ContentType   string
}

type StemcellUploadOutput struct{}

type UploadStemcellService struct {
	client   httpClient
	progress progress
}

func NewUploadStemcellService(client httpClient, progress progress) UploadStemcellService {
	return UploadStemcellService{
		client:   client,
		progress: progress,
	}
}

func (us UploadStemcellService) Upload(input StemcellUploadInput) (StemcellUploadOutput, error) {
	us.progress.SetTotal(input.ContentLength)
	body := us.progress.NewBarReader(input.Stemcell)

	req, err := http.NewRequest("POST", "/api/v0/stemcells", body)
	if err != nil {
		return StemcellUploadOutput{}, err
	}

	req.Header.Set("Content-Type", input.ContentType)
	req.ContentLength = input.ContentLength

	us.progress.Kickoff()

	resp, err := us.client.Do(req)
	if err != nil {
		return StemcellUploadOutput{}, fmt.Errorf("could not make api request to stemcells endpoint: %s", err)
	}

	defer resp.Body.Close()

	us.progress.End()

	if resp.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return StemcellUploadOutput{}, fmt.Errorf("request failed: unexpected response: %s", err)
		}

		return StemcellUploadOutput{}, fmt.Errorf("request failed: unexpected response:\n%s", out)
	}

	return StemcellUploadOutput{}, nil
}
