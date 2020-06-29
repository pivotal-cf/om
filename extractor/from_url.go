package extractor

import (
	"archive/zip"
	"fmt"
	"howett.net/ranger"
	"net/http"
	"net/url"
)

type wrappedClient struct {}

func (w wrappedClient) Do(request *http.Request) (*http.Response, error) {
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to perform http request: %v", err)
	}
	resp.Header.Add("Content-Type", "application/octet-stream")
	return resp, err
}

func (w wrappedClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return w.Do(req)
}

func (w wrappedClient) Head(url string) (*http.Response, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return w.Do(req)
}

var _ ranger.HTTPClient = wrappedClient{}

func (me MetadataExtractor) ExtractFromURL(productURL string) (Metadata, error) {
	url, err := url.Parse(productURL)
	if err != nil {
		return Metadata{}, err
	}

	r := &ranger.HTTPRanger{URL: url, Client: wrappedClient{}}
	reader, err := ranger.NewReader(r)
	if err != nil {
		return Metadata{}, fmt.Errorf("could not create ranger: %w", err)
	}

	length, err := reader.Length()
	if err != nil {
		return Metadata{}, fmt.Errorf("could not determine reader length: %w", err)
	}

	zipReader, err := zip.NewReader(reader, length)
	if err != nil {
		return Metadata{}, fmt.Errorf( "could not create zip reader: %w", err)
	}

	return fromZipFiles(zipReader.File)
}
