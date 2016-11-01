package api

import (
	"fmt"
	"io"
	"net/http"
)

type RequestServiceInvokeInput struct {
	Path   string
	Method string
	Data   io.Reader
}

type RequestServiceInvokeOutput struct {
	StatusCode int
	Headers    http.Header
	Body       io.Reader
}

type RequestService struct {
	client httpClient
}

func NewRequestService(client httpClient) RequestService {
	return RequestService{client: client}
}

func (rs RequestService) Invoke(input RequestServiceInvokeInput) (RequestServiceInvokeOutput, error) {
	request, err := http.NewRequest(input.Method, input.Path, input.Data)
	if err != nil {
		return RequestServiceInvokeOutput{}, fmt.Errorf("failed constructing request: %s", err)
	}

	request.Header.Set("Content-Type", "application/json")
	response, err := rs.client.Do(request)
	if err != nil {
		return RequestServiceInvokeOutput{}, fmt.Errorf("failed submitting request: %s", err)
	}

	output := RequestServiceInvokeOutput{
		StatusCode: response.StatusCode,
		Headers:    response.Header,
		Body:       response.Body,
	}

	return output, nil
}
