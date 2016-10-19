package commands

import (
	"io"
	"strings"
)

type LogWriter struct {
	writer io.Writer
	offset int64
}

func NewLogWriter(writer io.Writer) *LogWriter {
	return &LogWriter{
		writer: writer,
	}
}

func (lw *LogWriter) Flush(logs string) error {
	reader := strings.NewReader(logs)

	_, err := reader.Seek(lw.offset, 0)
	if err != nil {
		return err
	}

	written, err := io.Copy(lw.writer, reader)
	if err != nil {
		return err
	}

	lw.offset += written

	return nil
}
