package progress

import (
	"io"
)

//go:generate counterfeiter -o ./fakes/progress_bar.go --fake-name ProgressBar . progressBar
type progressBar interface {
	Start()
	Finish()
	NewProxyReader(io.Reader) io.ReadCloser
}

type ReadCloser struct {
	reader      io.Reader
	progressBar progressBar
	callback    func()

	started bool
	closed  bool
}

func NewReadCloser(reader io.Reader, progressBar progressBar, callback func()) *ReadCloser {
	return &ReadCloser{
		reader:      progressBar.NewProxyReader(reader),
		progressBar: progressBar,
		callback:    callback,
	}
}

func (rc *ReadCloser) Read(b []byte) (int, error) {
	if !rc.started {
		rc.started = true
		rc.progressBar.Start()
	}

	result, err := rc.reader.Read(b)
	if err == io.EOF {
		_ = rc.Close()
	}
	return result, err
}

func (rc *ReadCloser) Close() error {
	if rc.closed {
		return nil
	}
	rc.closed = true

	rc.finish()

	if closer, ok := rc.reader.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

func (rc *ReadCloser) finish() {
	rc.progressBar.Finish()

	if rc.callback != nil {
		rc.callback()
	}
}
