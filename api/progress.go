package api

import "io"

//go:generate counterfeiter -o ./fakes/progress.go --fake-name Progress . progress
type progress interface {
	SetTotal(int64)
	NewBarReader(io.Reader) io.Reader
	Kickoff()
	End()
	GetTotal() int64
	GetCurrent() int64
}
