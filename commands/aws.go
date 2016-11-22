package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

//go:generate counterfeiter -o ./fakes/aws_service.go --fake-name AWSService . awsService
type awsService interface {
	Invoke(api.AWSIaasConfigurationInput) (api.AWSIaasConfigurationOutput, error)
}

type AWS struct {
	awsService awsService
	stdout     logger
	stderr     logger
	Options    struct {
		AccessKey        string `short:"a" long:"accessKey"    description:"aws access key"`
		SecretKey        string `short:"s" long:"secretKey"    description:"aws secret key"`
		PrivateKey       string `short:"p" long:"privateKey"    description:"pem encoded private key"`
		DatabasePassword string `short:"d" long:"databasePassword"    description:"password for external database (optional)"`
		ProvideTemplate  bool   `short:"t" long:"provideTemplate"    description:"provide template json file aws.json"`
		Config           string `short:"c" long:"config"    description:"configuration json"`
	}
}

func NewAWS(aws awsService, stdout logger, stderr logger) AWS {
	return AWS{awsService: aws, stdout: stdout, stderr: stderr}
}

func (a AWS) Execute(args []string) error {
	_, err := flags.Parse(&a.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse aws flags: %s", err)
	}
	if a.Options.ProvideTemplate {
		input := &api.AWSIaasConfigurationInput{
			AZNames: []string{"us-east-1a", "us-east-1b", "us-east-1d"},
			Networks: []api.Network{
				api.Network{
					Subnets: []api.Subnet{
						api.Subnet{},
					},
				},
			},
		}
		if data, err := json.Marshal(input); err == nil {
			return ioutil.WriteFile("./aws.json", data, 0755)
		}
	}
	input := &api.AWSIaasConfigurationInput{}
	if err = json.Unmarshal([]byte(a.Options.Config), input); err != nil {
		return fmt.Errorf("could not parse json: %s", err)
	}
	input.AccessKey = a.Options.AccessKey
	input.SecretKey = a.Options.SecretKey
	input.DatabasePassword = a.Options.DatabasePassword
	input.PrivateKey = a.Options.PrivateKey

	_, err = a.awsService.Invoke(*input)
	return err
}

func (a AWS) Usage() Usage {
	return Usage{
		Description:      "This command configures diretor tile for AWS",
		ShortDescription: "AWS director tile",
		Flags:            a.Options,
	}
}
