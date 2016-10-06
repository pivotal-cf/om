package api

type DiagnosticService struct {
	client httpClient
}

type DiagnosticReport struct {
	Stemcells []string
}

func NewDiagnosticService(client httpClient) DiagnosticService {
	return DiagnosticService{
		client: client,
	}
}

func (ds DiagnosticService) Report() (DiagnosticReport, error) {
	return DiagnosticReport{}, nil
}
