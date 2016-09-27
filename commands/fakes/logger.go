package fakes

import "fmt"

type Logger struct {
	Lines []string
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.Lines = append(l.Lines, fmt.Sprintf(format, v...))
}
