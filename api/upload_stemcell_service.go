package api

import (
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type StemcellUploadInput struct {
	ContentLength int64
	Stemcell      io.Reader
	ContentType   string
}

type StemcellUploadOutput struct{}

func (a Api) UploadStemcell(input StemcellUploadInput) (StemcellUploadOutput, error) {
	req, err := http.NewRequest("POST", "/api/v0/stemcells", input.Stemcell)
	if err != nil {
		return StemcellUploadOutput{}, err
	}

	req.Header.Set("Content-Type", input.ContentType)
	req.ContentLength = input.ContentLength

	resp, err := a.progressClient.Do(req)
	if err != nil {
		return StemcellUploadOutput{}, errors.Wrap(err, "could not make api request to stemcells endpoint")
	}

	defer resp.Body.Close()

	return StemcellUploadOutput{}, validateStatusOK(resp)
}
