package network

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

//go:generate counterfeiter -o ./fakes/progress.go --fake-name Progress . progress
type progress interface {
	SetTotal(int64)
	NewBarReader(io.Reader) io.ReadCloser
	Kickoff()
	End()
	GetTotal() int64
	GetCurrent() int64
}

//go:generate counterfeiter -o ./fakes/livewriter.go --fake-name LiveWriter . liveWriter
type liveWriter interface {
	io.Writer
	Start()
	Stop()
}

type ProgressClient struct {
	client     httpClient
	progress   progress
	liveWriter liveWriter
}

func NewProgressClient(client httpClient, progress progress, liveWriter liveWriter) ProgressClient {
	return ProgressClient{
		client:     client,
		progress:   progress,
		liveWriter: liveWriter,
	}
}

func (pc ProgressClient) Do(req *http.Request) (*http.Response, error) {
	pc.progress.SetTotal(req.ContentLength)
	req.Body = pc.progress.NewBarReader(req.Body)

	pollingInterval, ok := req.Context().Value("polling-interval").(time.Duration)
	if !ok {
		pollingInterval = time.Second
	}

	requestComplete := make(chan bool)
	progressComplete := make(chan bool)

	go pc.trackProgress(progressComplete, requestComplete, pollingInterval)

	resp, err := pc.client.Do(req)
	requestComplete <- true
	<-progressComplete

	return resp, err
}

func (pc ProgressClient) trackProgress(progressComplete, requestComplete chan bool, pollingInterval time.Duration) {
	if err := pc.showProgress(progressComplete, requestComplete); err != nil {
		return
	}

	pc.logElapsed(progressComplete, requestComplete, pollingInterval)
}

func (pc ProgressClient) showProgress(progressComplete, requestComplete chan bool) error {
	pc.progress.Kickoff()
	for {
		select {
		case <-requestComplete:
			pc.progress.End()
			progressComplete <- true
			return errors.New("request ended early")

		default:
			if pc.progress.GetCurrent() != pc.progress.GetTotal() {
				time.Sleep(time.Second)
				continue
			}

			pc.progress.End()
			return nil
		}
	}
	return nil
}

func (pc ProgressClient) logElapsed(progressComplete, requestComplete chan bool, pollingInterval time.Duration) {
	pc.liveWriter.Start()
	liveLog := log.New(pc.liveWriter, "", 0)
	startTime := time.Now().Round(time.Second)
	ticker := time.NewTicker(pollingInterval)

	for {
		select {
		case <-requestComplete:
			ticker.Stop()
			pc.liveWriter.Stop()
			progressComplete <- true
			return

		case now := <-ticker.C:
			liveLog.Printf("%s elapsed, waiting for response from Ops Manager...\r", now.Round(time.Second).Sub(startTime).String())
		}
	}
}
