package presenters

import (
	"encoding/json"
	"io"

	"github.com/pivotal-cf/om/models"
)

type JSONPresenter struct {
	stdout io.Writer
}

func NewJSONPresenter(stdout io.Writer) JSONPresenter {
	return JSONPresenter{
		stdout: stdout,
	}
}

func (j JSONPresenter) PresentInstallations(installations []models.Installation) {
	encoder := json.NewEncoder(j.stdout)
	encoder.Encode(&installations)
}

func (j JSONPresenter) PresentAvailableProducts(products []models.Product) {
	encoder := json.NewEncoder(j.stdout)
	encoder.Encode(&products)
}
