package commands

//go:generate counterfeiter -o ./fakes/other_logger.go --fake-name OtherLogger . logger
type logger interface {
	Printf(format string, v ...interface{})
}
