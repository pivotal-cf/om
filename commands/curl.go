package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/common"
	"github.com/pivotal-cf/om/flags"
)

//go:generate counterfeiter -o ./fakes/request_service.go --fake-name RequestService . requestService
type requestService interface {
	Invoke(api.RequestServiceInvokeInput) (api.RequestServiceInvokeOutput, error)
}

type Curl struct {
	requestService requestService
	stdout         common.Logger
	stderr         common.Logger
	Options        struct {
		Path   string `short:"p" long:"path"    description:"path to api endpoint"`
		Method string `short:"x" long:"request" description:"http verb"`
		Data   string `short:"d" long:"data"    description:"api request payload"`
	}
}

func NewCurl(rs requestService, stdout common.Logger, stderr common.Logger) Curl {
	return Curl{requestService: rs, stdout: stdout, stderr: stderr}
}

func (c Curl) Execute(args []string) error {
	_, err := flags.Parse(&c.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse curl flags: %s", err)
	}

	input := api.RequestServiceInvokeInput{
		Path:   c.Options.Path,
		Method: c.Options.Method,
		Data:   strings.NewReader(c.Options.Data),
	}

	output, err := c.requestService.Invoke(input)
	if err != nil {
		return fmt.Errorf("failed to make api request: %s", err)
	}

	c.stderr.Printf("Status: %d %s", output.StatusCode, http.StatusText(output.StatusCode))

	headers := bytes.NewBuffer([]byte{})
	err = output.Headers.Write(headers)
	if err != nil {
		return fmt.Errorf("failed to write api response headers: %s", err)
	}

	c.stderr.Printf(headers.String())

	body, err := ioutil.ReadAll(output.Body)
	if err != nil {
		return fmt.Errorf("failed to read api response body: %s", err)
	}

	c.stdout.Printf(string(body))

	return nil
}

func (c Curl) Usage() Usage {
	return Usage{
		Description:      "This command issues an authenticated API request as defined in the arguments",
		ShortDescription: "issues an authenticated API request",
		Flags:            c.Options,
	}
}
