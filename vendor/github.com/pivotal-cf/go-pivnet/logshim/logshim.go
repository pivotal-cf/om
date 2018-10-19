package logshim

import (
	"fmt"
	"log"

	"github.com/pivotal-cf/go-pivnet/logger"
)

type LogShim struct {
	infoLogger  *log.Logger
	debugLogger *log.Logger
	verbose     bool
}

func NewLogShim(
	infoLogger *log.Logger,
	debugLogger *log.Logger,
	verbose bool,
) *LogShim {
	return &LogShim{
		infoLogger:  infoLogger,
		debugLogger: debugLogger,
		verbose:     verbose,
	}
}

func (l LogShim) Debug(action string, data ...logger.Data) {
	if l.verbose {
		l.debugLogger.Println(fmt.Sprintf("%s%s", action, appendString(data...)))
	}
}

func (l LogShim) Info(action string, data ...logger.Data) {
	l.infoLogger.Println(fmt.Sprintf("%s%s", action, appendString(data...)))
}

func appendString(data ...logger.Data) string {
	if len(data) > 0 {
		return fmt.Sprintf(" - %+v", data)
	}
	return ""
}
