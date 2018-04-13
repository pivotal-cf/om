package progress

import (
	"io"
	"os"

	"gopkg.in/cheggaaa/pb.v1"
)

type Bar struct {
	bar *pb.ProgressBar
}

func NewBar() Bar {
	bar := pb.New(0)
	bar.SetUnits(pb.U_BYTES)
	bar.Width = 80
	bar.Output = os.Stderr
	return Bar{bar}
}

func (b Bar) NewProxyReader(r io.Reader) io.ReadCloser {
	return b.bar.NewProxyReader(r)
}

func (b Bar) Start() {
	b.bar.Start()
}

func (b Bar) Finish() {
	b.bar.Finish()
}

func (b Bar) SetTotal64(size int64) {
	b.bar.Total = size
}
