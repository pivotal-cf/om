package network

import (
	"io"
	"net/http"
	"time"

	"github.com/pivotal-cf/om/progress"
)

//go:generate counterfeiter -o ./fakes/progress_bar.go --fake-name ProgressBar . progressBar
type progressBar interface {
	Start()
	Finish()
	SetTotal64(int64)
	Reset()
	NewProxyReader(io.Reader) io.ReadCloser
}

//go:generate counterfeiter -o ./fakes/livewriter.go --fake-name LiveWriter . liveWriter
type liveWriter interface {
	io.Writer
	Start()
	Stop()
	Flush() error
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

	// reset bar in case request is being retried
	pc.progressBar.Reset()

	tl := progress.NewTickingLogger(pc.liveWriter, duration)

	startedTicker := make(chan bool)

	switch req.Method {
	case "POST", "PUT":
		req.Body = progress.NewReadCloser(req.Body, pc.progressBar, func() {
			tl.Start()
			close(startedTicker)
		})
		pc.progressBar.SetTotal64(req.ContentLength)
	case "GET":
		tl.Start()
		close(startedTicker)
	}

	resp, err := pc.client.Do(req)

	// the req.Body is closed asynchronously, but we'll also guard against
	// it never getting closed by continuing after X seconds
	waitForChanWithTimeout(startedTicker, 2*time.Second)

	tl.Stop()
	if err != nil {
		return nil, err
	}

	if req.Method == "GET" {
		resp.Body = progress.NewReadCloser(resp.Body, pc.progressBar, nil)
		pc.progressBar.SetTotal64(resp.ContentLength)
	}

	return resp, nil
}

func waitForChanWithTimeout(waitChan <-chan bool, timeout time.Duration) {
	select {
	case <-waitChan:
	case <-time.After(timeout):
	}
}
