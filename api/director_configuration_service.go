package api

import (
	"net/http"
	"net/url"
	"strings"
)

type DirectorConfigurationService struct {
	client httpClient
}

type DirectorConfigurationInput struct {
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
	S3AccessKey                  string
	S3SecretKey                  string
	S3SignatureVersion           string
	DatabaseHost                 string
	DatabasePort                 string
	DatabaseUser                 string
	DatabasePassword             string
	Database                     string
	MaxThreads                   string
	DirectorHostname             string
}

type DirectorConfigurationOutput struct{}

func NewDirectorConfigurationService(client httpClient) DirectorConfigurationService {
	return DirectorConfigurationService{
		client: client,
	}
}

func (s DirectorConfigurationService) Configure(input DirectorConfigurationInput) (DirectorConfigurationOutput, error) {
	if token, err := s.getPage(); err == nil {
		return s.updatePage(token, input)
	} else {
		return DirectorConfigurationOutput{}, err
	}
}

func (s DirectorConfigurationService) getPage() (string, error) {
	var err error
	var req *http.Request
	var resp *http.Response
	var token string
	if req, err = http.NewRequest("GET", "/infrastructure/director_configuration/edit", nil); err == nil {
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

func (s DirectorConfigurationService) updatePage(token string, input DirectorConfigurationInput) (DirectorConfigurationOutput, error) {
	var err error
	var req *http.Request
	var resp *http.Response
	data := url.Values{}
	data.Set("_method", "put")
	data.Set("authenticity_token", token)
	data.Set("director_configuration[ntp_servers_string]", input.NTPServers)
	data.Set("director_configuration[metrics_ip]", input.MetricsIP)
	data.Set("director_configuration[resurrector_enabled]", GetBooleanAsFormValue(input.EnableResurrector))
	data.Set("director_configuration[post_deploy_enabled]", GetBooleanAsFormValue(input.EnablePostDeployScripts))
	data.Set("director_configuration[bosh_recreate_on_next_deploy]", GetBooleanAsFormValue(input.EnableBoshRecreate))
	data.Set("director_configuration[retry_bosh_deploys]", GetBooleanAsFormValue(input.EnableBoshRetryDeploys))
	data.Set("director_configuration[hm_pager_duty_options][enabled]", GetBooleanAsFormValue(input.EnableHealthManagerPagerDuty))
	data.Set("director_configuration[hm_emailer_options][enabled]", GetBooleanAsFormValue(input.EnableHealthManagerEmail))
	if input.S3Endpoint == "" {
		data.Set("director_configuration[blobstore_type]", "internal")
	} else {
		data.Set("director_configuration[blobstore_type]", "s3")
	}
	data.Set("director_configuration[s3_blobstore_options][endpoint]", input.S3Endpoint)
	data.Set("director_configuration[s3_blobstore_options][bucket_name]", input.S3BucketName)
	data.Set("director_configuration[s3_blobstore_options][access_key]", input.S3AccessKey)
	data.Set("director_configuration[s3_blobstore_options][secret_key]", input.S3SecretKey)
	data.Set("director_configuration[s3_blobstore_options][signature_version]", input.S3SignatureVersion)
	if input.DatabaseHost == "" {
		data.Set("director_configuration[database_type]", "internal")
	} else {
		data.Set("director_configuration[database_type]", "external")
	}
	data.Set("director_configuration[external_database_options][host]", input.DatabaseHost)
	data.Set("director_configuration[external_database_options][port]", input.DatabasePort)
	data.Set("director_configuration[external_database_options][user]", input.DatabaseUser)
	data.Set("director_configuration[external_database_options][password]", input.DatabasePassword)
	data.Set("director_configuration[external_database_options][database]", input.Database)
	data.Set("director_configuration[max_threads]", input.MaxThreads)
	data.Set("director_configuration[director_hostname]", input.DirectorHostname)
	if req, err = http.NewRequest("POST", "/infrastructure/director_configuration", strings.NewReader(data.Encode())); err == nil {
		if resp, err = s.client.Do(req); err == nil {
			if err = HandleResponse(resp, http.StatusOK); err == nil {
				return DirectorConfigurationOutput{}, nil
			}
		}
	}
	return DirectorConfigurationOutput{}, err
}
