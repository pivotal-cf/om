package progress

import (
	"fmt"
	"io"
	"time"
)

//go:generate counterfeiter -o ./fakes/live_writer.go --fake-name LiveWriter . liveWriter
type liveWriter interface {
	io.Writer
	Start()
	Stop()
	Flush() error
}

type TickingLogger struct {
	liveWriter liveWriter
	ticker     *time.Ticker
	duration   time.Duration
	done       chan struct{}
}

func NewTickingLogger(liveWriter liveWriter, duration time.Duration) *TickingLogger {
	return &TickingLogger{
		liveWriter: liveWriter,
		duration:   duration,
		done:       make(chan struct{}),
	}
}

func (tl *TickingLogger) Start() {
	if tl.ticker == nil {
		tl.liveWriter.Start()
		tl.ticker = time.NewTicker(tl.duration)

		go func() {
			startTime := time.Now()

			for {
				select {
				case <-tl.ticker.C:
					duration := time.Now().Sub(startTime).Round(tl.duration).String()
					fmt.Fprintf(tl.liveWriter, "%s elapsed, waiting for response from Ops Manager...\r", duration)

				case <-tl.done:
					tl.ticker.Stop()
					fmt.Fprintln(tl.liveWriter)
					tl.liveWriter.Stop()
					tl.liveWriter.Flush()
					close(tl.done)
					return
				}
			}
		}()
	}
}

func (tl *TickingLogger) Stop() {
	if tl.ticker == nil {
		return
	}

	tl.done <- struct{}{}
	<-tl.done
}
