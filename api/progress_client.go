package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

type ProgressClient struct {
	client          httpClient
	progress        progress
	liveWriter      liveWriter
	pollingInterval int
}

func NewProgressClient(client httpClient, progress progress, liveWriter liveWriter, pollingInterval int) ProgressClient {
	return ProgressClient{
		client:          client,
		progress:        progress,
		liveWriter:      liveWriter,
		pollingInterval: pollingInterval,
	}
}

func (pc ProgressClient) Do(req *http.Request) (*http.Response, error) {
	pc.progress.SetTotal(req.ContentLength)
	req.Body = pc.progress.NewBarReader(req.Body)

	requestComplete := make(chan bool)
	progressComplete := make(chan bool)

	go pc.trackProgress(progressComplete, requestComplete)

	resp, err := pc.client.Do(req)
	requestComplete <- true
	<-progressComplete

	return resp, err
}

func (pc ProgressClient) trackProgress(progressComplete, requestComplete chan bool) {
	if err := pc.showProgress(progressComplete, requestComplete); err != nil {
		return
	}

	pc.logElapsed(progressComplete, requestComplete)
}

func (pc ProgressClient) showProgress(progressComplete, requestComplete chan bool) error {
	fmt.Println("showProgress")

	fmt.Println("showProgress: progress.Kickoff")
	pc.progress.Kickoff()
	for {
		fmt.Println("showProgress: select")
		select {
		case <-requestComplete:
			fmt.Println("showProgress: select: requestComplete: progress.End")
			pc.progress.End()

			fmt.Println("showProgress: select: requestComplete: send progressComplete")
			progressComplete <- true

			fmt.Println("showProgress: select: requestComplete: return")
			return errors.New("request ended early")

		default:
			fmt.Println("showProgress: select: default: progress.GetCurrent")
			current := pc.progress.GetCurrent()

			fmt.Println("showProgress: select: default: progress.GetTotal")
			total := pc.progress.GetTotal()

			fmt.Printf("showProgress: select: default: current (%d) == total (%d)\n", current, total)
			if current != total {
				fmt.Println("showProgress: select: default: sleep 1")
				time.Sleep(time.Second)
				continue
			}

			fmt.Println("showProgress: select: default: progress.End")
			pc.progress.End()
			return nil
		}
	}
	return nil
}

func (pc ProgressClient) logElapsed(progressComplete, requestComplete chan bool) {
	fmt.Println("liveWriter.Start")
	pc.liveWriter.Start()

	fmt.Println("log.New")
	liveLog := log.New(pc.liveWriter, "", 0)

	fmt.Println("startTime")
	startTime := time.Now().Round(time.Second)

	fmt.Println("ticker")
	ticker := time.NewTicker(time.Duration(pc.pollingInterval) * time.Second)

	fmt.Println("Loop 2")
	for {
		select {
		case <-requestComplete:
			fmt.Println("Loop 2: select: requestComplete: ticker.Stop")
			ticker.Stop()

			fmt.Println("Loop 2: select: requestComplete: liveWriter.Stop")
			pc.liveWriter.Stop()

			fmt.Println("Loop 2: select: requestComplete: send progressComplete")
			progressComplete <- true

			fmt.Println("Loop 2: select: requestComplete: return")
			return

		case now := <-ticker.C:
			fmt.Println("Loop 2: select: ticker.C: liveLog.Printf")
			liveLog.Printf("%s elapsed, waiting for response from Ops Manager...\r", now.Round(time.Second).Sub(startTime).String())
		}
	}
}
