package api

import (
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type RequestServiceCurlInput struct {
	Path    string
	Method  string
	Data    io.Reader
	Headers http.Header
}

type RequestServiceCurlOutput struct {
	StatusCode int
	Headers    http.Header
	Body       io.ReadCloser
}

func (a Api) Curl(input RequestServiceCurlInput) (RequestServiceCurlOutput, error) {
	request, err := http.NewRequest(input.Method, input.Path, input.Data)
	if err != nil {
		return RequestServiceCurlOutput{}, errors.Wrap(err, "failed constructing request")
	}

	request.Header = input.Headers
	response, err := a.client.Do(request)
	if err != nil {
		return RequestServiceCurlOutput{}, errors.Wrap(err, "failed submitting request")
	}

	output := RequestServiceCurlOutput{
		StatusCode: response.StatusCode,
		Headers:    response.Header,
		Body:       response.Body,
	}

	return output, nil
}
