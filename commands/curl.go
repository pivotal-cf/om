package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"encoding/json"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

//go:generate counterfeiter -o ./fakes/request_service.go --fake-name RequestService . requestService
type requestService interface {
	Invoke(api.RequestServiceInvokeInput) (api.RequestServiceInvokeOutput, error)
}

type Curl struct {
	requestService requestService
	stdout         logger
	stderr         logger
	Options        struct {
		Path   string `short:"p" long:"path"    description:"path to api endpoint"`
		Method string `short:"x" long:"request" description:"http verb" default:"GET"`
		Data   string `short:"d" long:"data"    description:"api request payload"`
	}
}

func NewCurl(rs requestService, stdout logger, stderr logger) Curl {
	return Curl{requestService: rs, stdout: stdout, stderr: stderr}
}

func (c Curl) Execute(args []string) error {
	_, err := flags.Parse(&c.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse curl flags: %s", err)
	}

	if c.Options.Path == "" {
		return errors.New("could not parse curl flags: -path is a required parameter. Please run `om curl --help` for more info.")
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

	for _, contentType := range output.Headers["Content-Type"] {
		if strings.HasPrefix(contentType, "application/json") {
			var prettyJSON bytes.Buffer
			err := json.Indent(&prettyJSON, body, "", "  ")
			if err != nil {
				panic(err)
			}
			body = prettyJSON.Bytes()

			break
		}
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
