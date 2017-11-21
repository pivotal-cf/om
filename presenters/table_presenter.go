package presenters

import "github.com/pivotal-cf/om/models"

//go:generate counterfeiter -o ./fakes/table_writer.go --fake-name TableWriter . tableWriter

type tableWriter interface {
	SetHeader([]string)
	Append([]string)
	SetAlignment(int)
	Render()
	SetAutoFormatHeaders(bool)
	SetAutoWrapText(bool)
}

type TablePresenter struct {
	tableWriter tableWriter
}

func NewTablePresenter(tableWriter tableWriter) TablePresenter {
	return TablePresenter{
		tableWriter: tableWriter,
	}
}

func (t TablePresenter) PresentInstallations(installations []models.Installation) {
	t.tableWriter.SetHeader([]string{"ID", "User", "Status", "Started At", "Finished At"})

	for _, installation := range installations {
		t.tableWriter.Append([]string{
			installation.Id,
			installation.User,
			installation.Status,
			installation.StartedAt,
			installation.FinishedAt,
		})
	}

	t.tableWriter.Render()
}
