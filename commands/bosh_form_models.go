package commands

type CommonConfiguration struct {
	AuthenticityToken string `url:"authenticity_token"`
	Method            string `url:"_method"`
}

type GCPIaaSConfiguration struct {
	Project              string `url:"iaas_configuration[project],omitempty" json:"project"`
	DefaultDeploymentTag string `url:"iaas_configuration[default_deployment_tag],omitempty" json:"default_deployment_tag"`
	AuthJSON             string `url:"iaas_configuration[auth_json],omitempty" json:"auth_json"`
}

type AzureIaaSConfiguration struct {
	SubscriptionID                string `url:"iaas_configuration[subscription_id],omitempty" json:"subscription_id"`
	TenantID                      string `url:"iaas_configuration[tenant_id],omitempty" json:"tenant_id"`
	ClientID                      string `url:"iaas_configuration[client_id],omitempty" json:"client_id"`
	ClientSecret                  string `url:"iaas_configuration[client_secret],omitempty" json:"client_secret"`
	ResourceGroupName             string `url:"iaas_configuration[resource_group_name],omitempty" json:"resource_group_name"`
	BoshStorageAccountName        string `url:"iaas_configuration[bosh_storage_account_name],omitempty" json:"bosh_storage_account_name"`
	DefaultSecurityGroup          string `url:"iaas_configuration[default_security_group],omitempty" json:"default_security_group"`
	SSHPublicKey                  string `url:"iaas_configuration[ssh_public_key],omitempty" json:"ssh_public_key"`
	DeploymentsStorageAccountName string `url:"iaas_configuration[deployments_storage_account_name],omitempty" json:"deployments_storage_account_name"`
}

type AWSIaaSConfiguration struct {
	AccessKeyID     string `url:"iaas_configuration[access_key_id],omitempty" json:"access_key_id"`
	SecretAccessKey string `url:"iaas_configuration[secret_access_key],omitempty" json:"secret_access_key"`
	VpcID           string `url:"iaas_configuration[vpc_id],omitempty" json:"vpc_id"`
	SecurityGroup   string `url:"iaas_configuration[security_group],omitempty" json:"security_group"`
	KeyPairName     string `url:"iaas_configuration[key_pair_name],omitempty" json:"key_pair_name"`
	Region          string `url:"iaas_configuration[region],omitempty" json:"region"`
	Encrypted       *bool  `url:"iaas_configuration[encrypted],omitempty" json:"encrypted"`
}

type CommonIaaSConfiguration struct {
	SSHPrivateKey string `url:"iaas_configuration[ssh_private_key],omitempty" json:"ssh_private_key"`
}

type IaaSConfiguration struct {
	GCPIaaSConfiguration
	AzureIaaSConfiguration
	AWSIaaSConfiguration
	CommonIaaSConfiguration
}

type DirectorConfiguration struct {
	NTPServers                string                  `url:"director_configuration[ntp_servers_string],omitempty" json:"ntp_servers_string"`
	MetricsIP                 string                  `url:"director_configuration[metrics_ip],omitempty" json:"metrics_ip"`
	EnableVMResurrectorPlugin *bool                   `url:"director_configuration[resurrector_enabled],omitempty" json:"resurrector_enabled"`
	EnablePostDeployScripts   *bool                   `url:"director_configuration[post_deploy_enabled],omitempty" json:"post_deploy_enabled"`
	RecreateAllVMs            *bool                   `url:"director_configuration[bosh_recreate_on_next_deploy],omitempty" json:"bosh_recreate_on_next_deploy"`
	EnableBoshDeployRetries   *bool                   `url:"director_configuration[retry_bosh_deploys],omitempty" json:"retry_bosh_deploys"`
	HMPagerDutyOptions        HMPagerDutyOptions      `url:"director_configuration[hm_pager_duty_options],omitempty" json:"hm_pager_duty_options,omitempty"`
	HMEmailPlugin             HMEmailPlugin           `url:"director_configuration[hm_emailer_options],omitempty" json:"hm_emailer_options,omitempty"`
	BlobStoreType             string                  `url:"director_configuration[blobstore_type],omitempty" json:"blobstore_type"`
	S3BlobstoreOptions        S3BlobstoreOptions      `url:"director_configuration[s3_blobstore_options],omitempty" json:"s3_blobstore_options,omitempty"`
	DatabaseType              string                  `url:"director_configuration[database_type],omitempty" json:"database_type"`
	ExternalDatabaseOptions   ExternalDatabaseOptions `url:"director_configuration[external_database_options],omitempty" json:"external_database_options,omitempty"`
	MaxThreads                *int                    `url:"director_configuration[max_threads],omitempty" json:"max_threads"`
	DirectorHostname          string                  `url:"director_configuration[director_hostname],omitempty" json:"director_hostname"`
}

type BoshConfiguration struct {
	IaaSConfiguration
	DirectorConfiguration
	CommonConfiguration
}

type HMPagerDutyOptions struct {
	Enabled    *bool  `url:"enabled,omitempty" json:"enabled"`
	ServiceKey string `url:"service_key,omitempty" json:"service_key"`
	HTTPProxy  string `url:"http_proxy,omitempty" json:"http_proxy"`
}

type HMEmailPlugin struct {
	Enabled    *bool  `url:"enabled,omitempty" json:"enabled"`
	Host       string `url:"host,omitempty" json:"host"`
	Port       *int   `url:"port,omitempty" json:"port"`
	Domain     string `url:"domain,omitempty" json:"domain"`
	From       string `url:"from,omitempty" json:"from"`
	Recipients string `url:"recipients,omitempty" json:"recipients"`
	Username   string `url:"smtp_user,omitempty" json:"smtp_user"`
	Password   string `url:"smtp_password,omitempty" json:"smtp_password"`
	TLS        *bool  `url:"tls,omitempty" json:"tls"`
}

type S3BlobstoreOptions struct {
	Endpoint         string `url:"endpoint,omitempty" json:"endpoint"`
	BucketName       string `url:"bucket_name,omitempty" json:"bucket_name"`
	Accesskey        string `url:"access_key,omitempty" json:"access_key"`
	SecretKey        string `url:"secret_key,omitempty" json:"secret_key"`
	SignatureVersion string `url:"signature_version,omitempty" json:"signature_version"`
	Region           string `url:"region,omitempty" json:"region"`
}

type ExternalDatabaseOptions struct {
	Host     string `url:"host,omitempty" json:"host"`
	Port     *int   `url:"port,omitempty" json:"port"`
	Username string `url:"user,omitempty" json:"user"`
	Password string `url:"password,omitempty" json:"password"`
	Database string `url:"database,omitempty" json:"database"`
}
