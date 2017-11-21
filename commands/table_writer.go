package commands

//go:generate counterfeiter -o ./fakes/table_writer.go --fake-name TableWriter . tableWriter

type tableWriter interface {
	SetHeader([]string)
	Append([]string)
	SetAlignment(int)
	Render()
	SetAutoFormatHeaders(bool)
	SetAutoWrapText(bool)
}
