package network

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/pivotal-cf/om/progress"
)

//go:generate counterfeiter -o ./fakes/progress_bar.go --fake-name ProgressBar . progressBar
type progressBar interface {
	Start()
	Finish()
	SetTotal64(int64)
	NewProxyReader(io.Reader) io.ReadCloser
}

//go:generate counterfeiter -o ./fakes/livewriter.go --fake-name LiveWriter . liveWriter
type liveWriter interface {
	io.Writer
	Start()
	Stop()
}

type ProgressClient struct {
	client      httpClient
	progressBar progressBar
	liveWriter  liveWriter
}

func NewProgressClient(client httpClient, progressBar progressBar, liveWriter liveWriter) ProgressClient {
	return ProgressClient{
		client:      client,
		progressBar: progressBar,
		liveWriter:  liveWriter,
	}
}

func (pc ProgressClient) Do(req *http.Request) (*http.Response, error) {
	duration, ok := req.Context().Value("polling-interval").(time.Duration)
	if !ok {
		duration = time.Second
	}

	tl := progress.NewTickingLogger(log.New(pc.liveWriter, "", 0), duration)

	switch req.Method {
	case "POST", "PUT":
		req.Body = progress.NewReadCloser(req.Body, pc.progressBar, tl.Start)
		pc.progressBar.SetTotal64(req.ContentLength)
	case "GET":
		tl.Start()
	}

	resp, err := pc.client.Do(req)
	tl.Stop()

	if req.Method == "GET" {
		resp.Body = progress.NewReadCloser(resp.Body, pc.progressBar, nil)
		pc.progressBar.SetTotal64(resp.ContentLength)
	}

	return resp, err
}
