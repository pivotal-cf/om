package ui

type Ui struct {
	client httpClient
}

type UiInput struct {
	Client httpClient
}

func New(input UiInput) Ui {
	return Ui{
		client: input.Client,
	}
}
