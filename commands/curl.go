package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/formcontent"
)

//counterfeiter:generate -o ./fakes/curl_service.go --fake-name CurlService . curlService
type curlService interface {
	Curl(api.RequestServiceCurlInput) (api.RequestServiceCurlOutput, error)
}

type Curl struct {
	formcontent formcontent.Form
	service     curlService
	stdout      logger
	stderr      logger
	Options     struct {
		Path    string   `long:"path"    short:"p" required:"true" description:"path to api endpoint"`
		Method  string   `long:"request" short:"x"                 description:"http verb (defaults to GET, POST when 'data' specified"`
		Data    string   `long:"data"    short:"d"                 description:"api request payload (prefix with @ to read file contents)"`
		Silent  bool     `long:"silent"  short:"s"                 description:"only write response headers to stderr if response status is 4XX or 5XX"`
		Headers []string `long:"header"  short:"H"                 description:"used to specify custom headers with your command" default:"Content-Type: application/json"`
	}
}

func NewCurl(service curlService, stdout logger, stderr logger) *Curl {
	return &Curl{service: service, stdout: stdout, stderr: stderr, formcontent: *formcontent.NewForm()}
}

func (c Curl) Execute(args []string) error {
	requestHeaders := make(http.Header)
	for _, h := range c.Options.Headers {
		split := strings.Split(h, " ")
		requestHeaders.Set(strings.TrimSuffix(split[0], ":"), split[1])
	}

	var data io.Reader = strings.NewReader(c.Options.Data)
	if strings.HasPrefix(c.Options.Data, "@") {
		fname := strings.TrimPrefix(c.Options.Data, "@")
		f, err := os.Open(fname)
		if err != nil {
			return fmt.Errorf("couldn't open %s: %w", fname, err)
		}
		data = f
		defer f.Close()
	}

	// adding support for the data if it contains file with form multipart data
	if strings.Contains(c.Options.Data, "=@") {
		splitVals := strings.Split(c.Options.Data, "=@")
		if len(splitVals) == 2 {
			fileKey, fileName := splitVals[0], splitVals[1]
			err := c.formcontent.AddFile(fileKey, fileName)
			if err != nil {
				return fmt.Errorf("failed to add form content %v", err)
			}
			submission := c.formcontent.Finalize()
			data = submission.Content
			requestHeaders.Set("Content-Type", submission.ContentType)
		}
	}

	input := api.RequestServiceCurlInput{
		Path:    c.Options.Path,
		Method:  c.Options.Method,
		Data:    data,
		Headers: requestHeaders,
	}

	if c.Options.Method == "" {
		input.Method = "GET"

		if c.Options.Data != "" {
			input.Method = "POST"
		}
	}

	output, err := c.service.Curl(input)
	if err != nil {
		return fmt.Errorf("failed to make api request: %s", err)
	}

	writeHeadersToStderr := !c.Options.Silent || output.StatusCode >= 400

	if writeHeadersToStderr {
		c.stderr.Printf("Status: %d %s", output.StatusCode, http.StatusText(output.StatusCode))
	}

	headers := bytes.NewBuffer([]byte{})
	err = output.Headers.Write(headers)
	if err != nil {
		return fmt.Errorf("failed to write api response headers: %s", err)
	}

	if writeHeadersToStderr {
		c.stderr.Printf(headers.String())
	}

	body, err := io.ReadAll(output.Body)
	if err != nil {
		return fmt.Errorf("failed to read api response body: %s", err)
	}
	defer output.Body.Close()

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

	c.stdout.Println(string(body))

	if output.StatusCode >= 400 {
		return fmt.Errorf("server responded with a %d error", output.StatusCode)
	}

	return nil
}
