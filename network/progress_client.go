package network

import (
	"gopkg.in/cheggaaa/pb.v1"
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
	bar := pb.New64(0)
	bar.Output = pc.stderr
	bar.AutoStat = true
	bar.ShowPercent = true
	bar.ShowElapsedTime = true
	bar.SetUnits(pb.U_BYTES)

	switch req.Method {
	case http.MethodPost, http.MethodPut:
		req.Body = bar.NewProxyReader(req.Body)
		bar.SetTotal64(req.ContentLength)
	}

	bar.Start()

	resp, err := pc.client.Do(req)
	if err != nil {
		return nil, err
	}

	if req.Method == http.MethodGet {
		resp.Body = bar.NewProxyReader(resp.Body)
		bar.SetTotal64(resp.ContentLength)
	}

	return resp, nil
}
