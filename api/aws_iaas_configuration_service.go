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
	AccessKey       string
	SecretKey       string
	InstanceProfile string
	VPCID           string
	SecurityGroupID string
	KeyPairName     string
	PrivateKey      string
	Region          string
	Encrypted       bool
}

type AWSIaasConfigurationOutput struct{}

func NewAWSIaasConfigurationService(client httpClient) AWSIaasConfigurationService {
	return AWSIaasConfigurationService{
		client: client,
	}
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
	data.Add("iaas_configuration[iam_instance_profile]", input.InstanceProfile)
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
