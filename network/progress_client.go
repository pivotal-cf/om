package network

import (
	"github.com/cheggaaa/pb/v3"
	"io"
	"net/http"
)

type ProgressClient struct {
	client httpClient
	stderr io.Writer
}

func NewProgressClient(client httpClient, stderr io.Writer) ProgressClient {
	return ProgressClient{
		client: client,
		stderr: stderr,
	}
}

func (pc ProgressClient) Do(req *http.Request) (*http.Response, error) {
	bar := pb.Default.New(0)
	bar.SetWriter(pc.stderr)
	bar.Set(pb.Bytes, true)
	bar.SetMaxWidth(80)

	switch req.Method {
	case http.MethodPost, http.MethodPut:
		req.Body = bar.NewProxyReader(req.Body)
		bar.SetTotal(req.ContentLength)
	}

	bar.Start()

	resp, err := pc.client.Do(req)
	if err != nil {
		return nil, err
	}

	if req.Method == http.MethodGet {
		resp.Body = bar.NewProxyReader(resp.Body)
		bar.SetTotal(resp.ContentLength)
	}

	return resp, nil
}
