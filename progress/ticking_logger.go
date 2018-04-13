package progress

import (
	"time"
)

//go:generate counterfeiter -o ./fakes/logger.go --fake-name Logger . logger
type logger interface {
	Printf(format string, v ...interface{})
}

type TickingLogger struct {
	logger   logger
	ticker   *time.Ticker
	duration time.Duration
	done     chan struct{}
}

func NewTickingLogger(logger logger, duration time.Duration) *TickingLogger {
	return &TickingLogger{
		logger:   logger,
		duration: duration,
		done:     make(chan struct{}),
	}
}

func (tl *TickingLogger) Start() {
	if tl.ticker == nil {
		tl.ticker = time.NewTicker(tl.duration)

		go func() {
			startTime := time.Now()

			for {
				select {
				case <-tl.ticker.C:
					duration := time.Now().Sub(startTime).Round(time.Second).String()
					tl.logger.Printf("%s elapsed, waiting for response from Ops Manager...\r", duration)

				case <-tl.done:
					tl.ticker.Stop()
					close(tl.done)
					return
				}
			}
		}()
	}
}

func (tl *TickingLogger) Stop() {
	tl.done <- struct{}{}
	<-tl.done
}
