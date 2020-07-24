package extractor

import (
	"archive/zip"
	"fmt"
	"howett.net/ranger"
	"net/http"
	"net/url"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o ./fakes/httpclient.go --fake-name HttpClient . httpClient
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

func WithHTTPClient(client httpClient) Option{
	return func(me *MetadataExtractor) {
		me.httpClient = client
	}
}

func (me *MetadataExtractor) Do(request *http.Request) (*http.Response, error) {
	resp, err := me.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to perform http request: %v", err)
	}
	resp.Header.Add("Content-Type", "application/octet-stream")
	return resp, err
}

func (me *MetadataExtractor) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return me.Do(req)
}

func (me *MetadataExtractor) Head(url string) (*http.Response, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return me.Do(req)
}

var _ ranger.HTTPClient = &MetadataExtractor{}

func (me *MetadataExtractor) ExtractFromURL(productURL string) (*Metadata, error) {
	url, err := url.Parse(productURL)
	if err != nil {
		return nil, err
	}

	r := &ranger.HTTPRanger{URL: url, Client: me}
	reader, err := ranger.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("could not create ranger: %w", err)
	}

	length, err := reader.Length()
	if err != nil {
		return nil, fmt.Errorf("could not determine reader length: %w", err)
	}

	zipReader, err := zip.NewReader(reader, length)
	if err != nil {
		return nil, fmt.Errorf( "could not create zip reader: %w", err)
	}

	return fromZipFiles(zipReader.File)
}
