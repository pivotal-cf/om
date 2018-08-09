package fakes

import "fmt"

type Logger struct {
	PrintfCall struct {
		Receives struct {
			LogLines []string
		}
	}

	PrintlnCall struct {
		Receives struct {
			LogLines []string
		}
	}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.PrintfCall.Receives.LogLines = append(l.PrintfCall.Receives.LogLines, fmt.Sprintf(format, v...))
}

func (l *Logger) Println(v ...interface{}) {
	l.PrintlnCall.Receives.LogLines = append(l.PrintlnCall.Receives.LogLines, fmt.Sprintf("%s", v...))
}
