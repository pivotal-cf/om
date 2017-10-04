package progress

import (
	"io"
	"os"

	"gopkg.in/cheggaaa/pb.v1"
)

type Bar struct {
	*pb.ProgressBar
}

func NewBar() Bar {
	bar := pb.New(0)
	bar.SetUnits(pb.U_BYTES)
	bar.Width = 80
	bar.Output = os.Stdout
	return Bar{bar}
}

func (b Bar) SetTotal(initialSize int64) {
	b.Total = initialSize
}

func (b Bar) GetCurrent() int64 {
	return b.Get()
}

func (b Bar) GetTotal() int64 {
	return b.Total
}

func (b Bar) NewBarReader(r io.Reader) io.Reader {
	return b.NewProxyReader(r)
}

func (b Bar) Kickoff() {
	b.Start()
}

func (b Bar) End() {
	b.Finish()
}
