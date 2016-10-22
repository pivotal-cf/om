package common

//go:generate counterfeiter -o ./fakes/other_logger.go --fake-name OtherLogger . Logger
type Logger interface {
	Printf(format string, v ...interface{})
}
