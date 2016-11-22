package api

import (
	"net/http"
	"net/url"
	"strings"
)

type AWSIaasConfigurationService struct {
	client httpClient
}

type AWSIaasConfigurationInput struct {
	AccessKey                    string `json:"accessKey,omitempty"`
	SecretKey                    string `json:"secretKey,omitempty"`
	VPCID                        string
	SecurityGroupID              string
	KeyPairName                  string
	PrivateKey                   string `json:"privateKey,omitempty"`
	Region                       string
	Encrypted                    bool
	NTPServers                   string
	MetricsIP                    string
	EnableResurrector            bool
	EnablePostDeployScripts      bool
	EnableBoshRecreate           bool
	EnableBoshRetryDeploys       bool
	EnableHealthManagerPagerDuty bool
	EnableHealthManagerEmail     bool
	S3Endpoint                   string
	S3BucketName                 string
	S3SignatureVersion           string
	DatabaseHost                 string
	DatabasePort                 string
	DatabaseUser                 string
	DatabasePassword             string `json:"databasePassword,omitempty"`
	Database                     string
	MaxThreads                   string
	DirectorHostname             string
	AZNames                      []string
	Networks                     []Network
	EnableICMPChecks             bool
	SingletonAZ                  string
	SingletonNetwork             string
}

type AWSIaasConfigurationOutput struct{}

func NewAWSIaasConfigurationService(client httpClient) AWSIaasConfigurationService {
	return AWSIaasConfigurationService{
		client: client,
	}
}

func (s AWSIaasConfigurationService) Invoke(input AWSIaasConfigurationInput) (AWSIaasConfigurationOutput, error) {
	if _, err := s.Configure(input); err != nil {
		return AWSIaasConfigurationOutput{}, err
	}
	directorInput := DirectorConfigurationInput{
		S3AccessKey:                  input.AccessKey,
		S3SecretKey:                  input.SecretKey,
		NTPServers:                   input.NTPServers,
		MetricsIP:                    input.MetricsIP,
		EnableResurrector:            input.EnableResurrector,
		EnablePostDeployScripts:      input.EnablePostDeployScripts,
		EnableBoshRecreate:           input.EnableBoshRecreate,
		EnableBoshRetryDeploys:       input.EnableBoshRetryDeploys,
		EnableHealthManagerPagerDuty: input.EnableHealthManagerPagerDuty,
		EnableHealthManagerEmail:     input.EnableHealthManagerEmail,
		S3Endpoint:                   input.S3Endpoint,
		S3BucketName:                 input.S3BucketName,
		S3SignatureVersion:           input.S3SignatureVersion,
		DatabaseHost:                 input.DatabaseHost,
		DatabasePort:                 input.DatabasePort,
		DatabaseUser:                 input.DatabaseUser,
		DatabasePassword:             input.DatabasePassword,
		Database:                     input.Database,
		MaxThreads:                   input.MaxThreads,
		DirectorHostname:             input.DirectorHostname,
	}
	if _, err := NewDirectorConfigurationService(s.client).Configure(directorInput); err != nil {
		return AWSIaasConfigurationOutput{}, err
	}
	azInput := AZConfigurationInput{
		AZNames: input.AZNames,
	}
	if azOutput, err := NewAZConfigurationService(s.client).Configure(azInput); err == nil {
		networksInput := NetworkConfigurationInput{
			AZs:              azOutput.AZs,
			Networks:         input.Networks,
			EnableICMPChecks: input.EnableICMPChecks,
		}
		if networksOutput, err := NewNetworkConfigurationService(s.client).Configure(networksInput); err == nil {
			azNetworkInput := AZNetworkConfigurationInput{
				AZs:              azOutput.AZs,
				Networks:         networksOutput.Networks,
				SingletonAZ:      input.SingletonAZ,
				SingletonNetwork: input.SingletonNetwork,
			}
			if _, err := NewAZNetworkConfigurationService(s.client).Configure(azNetworkInput); err != nil {
				return AWSIaasConfigurationOutput{}, err
			}
		}
	}
	return AWSIaasConfigurationOutput{}, nil
}

func (s AWSIaasConfigurationService) Configure(input AWSIaasConfigurationInput) (AWSIaasConfigurationOutput, error) {
	if token, err := s.getPage(); err == nil {
		return s.updatePage(token, input)
	} else {
		return AWSIaasConfigurationOutput{}, err
	}
}

func (s AWSIaasConfigurationService) getPage() (string, error) {
	var err error
	var req *http.Request
	var resp *http.Response
	var token string
	if req, err = http.NewRequest("GET", "/infrastructure/iaas_configuration/edit", nil); err == nil {
		if resp, err = s.client.Do(req); err == nil {
			defer resp.Body.Close()
			if err = HandleResponse(resp, http.StatusOK); err == nil {
				token, err = GetCSRFToken(resp)
				return token, err
			}
		}
	}
	return "", err
}

func (s AWSIaasConfigurationService) updatePage(token string, input AWSIaasConfigurationInput) (AWSIaasConfigurationOutput, error) {
	var err error
	var req *http.Request
	var resp *http.Response
	data := url.Values{}
	data.Set("_method", "put")
	data.Set("authenticity_token", token)
	data.Add("iaas_configuration[access_key_id]", input.AccessKey)
	data.Add("iaas_configuration[secret_access_key]", input.SecretKey)
	data.Add("iaas_configuration[iam_instance_profile]", "")
	data.Add("iaas_configuration[vpc_id]", input.VPCID)
	data.Add("iaas_configuration[security_group]", input.SecurityGroupID)
	data.Add("iaas_configuration[key_pair_name]", input.KeyPairName)
	data.Add("iaas_configuration[ssh_private_key]", input.PrivateKey)
	data.Add("iaas_configuration[region]", input.Region)
	data.Add("iaas_configuration[encrypted]", GetBooleanAsFormValue(input.Encrypted))
	if req, err = http.NewRequest("POST", "/infrastructure/iaas_configuration", strings.NewReader(data.Encode())); err == nil {
		if resp, err = s.client.Do(req); err == nil {
			if err = HandleResponse(resp, http.StatusOK); err == nil {
				return AWSIaasConfigurationOutput{}, nil
			}
		}
	}
	return AWSIaasConfigurationOutput{}, err
}
