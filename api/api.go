package api

//go:generate counterfeiter -o ./fakes/logger.go --fake-name Logger . logger
type logger interface {
	Println(v ...interface{})
}

type Api struct {
	client                 httpClient
	unauthedClient         httpClient
	progressClient         httpClient
	unauthedProgressClient httpClient
	logger                 logger
}

type ApiInput struct {
	Client                 httpClient
	UnauthedClient         httpClient
	ProgressClient         httpClient
	UnauthedProgressClient httpClient
	Logger                 logger
}

func New(input ApiInput) Api {
	return Api{
		client:                 input.Client,
		unauthedClient:         input.UnauthedClient,
		progressClient:         input.ProgressClient,
		unauthedProgressClient: input.UnauthedProgressClient,
		logger:                 input.Logger,
	}
}
