package commands

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type CommonConfiguration struct {
	AuthenticityToken string `url:"authenticity_token"`
	Method            string `url:"_method"`
	Commit            string `url:"commit,omitempty"`
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

type VSphereIaaSConfiguration struct {
	VCenterHost              string `url:"iaas_configuration[vcenter_host],omitempty" json:"vcenter_host"`
	VCenterUserName          string `url:"iaas_configuration[vcenter_username],omitempty" json:"vcenter_username"`
	VCenterPassword          string `url:"iaas_configuration[vcenter_password],omitempty" json:"vcenter_password"`
	DatacenterName           string `url:"iaas_configuration[datacenter],omitempty" json:"datacenter"`
	VirtualDiskType          string `url:"iaas_configuration[disk_type],omitempty" json:"disk_type"`
	EphemeralDatastoreNames  string `url:"iaas_configuration[ephemeral_datastores_string],omitempty" json:"ephemeral_datastores_string"`
	PersistentDatastoreNames string `url:"iaas_configuration[persistent_datastores_string],omitempty" json:"persistent_datastores_string"`
	VMFolder                 string `url:"iaas_configuration[bosh_vm_folder],omitempty" json:"bosh_vm_folder"`
	TemplateFolder           string `url:"iaas_configuration[bosh_template_folder],omitempty" json:"bosh_template_folder"`
	DiskPathFolder           string `url:"iaas_configuration[bosh_disk_path],omitempty" json:"bosh_disk_path"`
}

type CommonIaaSConfiguration struct {
	SSHPrivateKey string `url:"iaas_configuration[ssh_private_key],omitempty" json:"ssh_private_key"`
}

type IaaSConfiguration struct {
	GCPIaaSConfiguration
	AzureIaaSConfiguration
	AWSIaaSConfiguration
	VSphereIaaSConfiguration
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

type AvailabilityZonesConfiguration struct {
	AvailabilityZones []AvailabilityZone `json:"availability_zones"`
	Names             []string           `url:"availability_zones[availability_zones][][iaas_identifier],omitempty"`
	VSphereNames      []string           `url:"availability_zones[availability_zones][][name],omitempty"`
	Clusters          []string           `url:"availability_zones[availability_zones][][cluster],omitempty"`
	ResourcePools     []string           `url:"availability_zones[availability_zones][][resource_pool],omitempty"`
}

type AvailabilityZone struct {
	Name         string `json:"name"`
	Cluster      string `json:"cluster"`
	ResourcePool string `json:"resource_pool"`
}

type SecurityConfiguration struct {
	TrustedCertificates string `url:"security_tokens[trusted_certificates],omitempty" json:"trusted_certificates"`
	VMPasswordType      string `url:"security_tokens[vm_password_type],omitempty" json:"vm_password_type"`
}

type NetworkAssignment struct {
	UserProvidedNetworkName string `json:"network" url:"-"`
	UserProvidedAZName      string `json:"singleton_availability_zone" url:"-"`
	NetworkGUID             string `url:"bosh_product[network_reference],omitempty"`
	AZGUID                  string `url:"bosh_product[singleton_availability_zone_reference],omitempty"`
}

type BoshConfiguration struct {
	IaaSConfiguration
	DirectorConfiguration
	AvailabilityZonesConfiguration
	SecurityConfiguration
	NetworkAssignment
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
	Recipients string `url:"recipients[value],omitempty" json:"recipients"`
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

type NetworksConfiguration struct {
	ICMP     bool `url:"infrastructure[icmp_checks_enabled],int" json:"icmp_checks_enabled"`
	Networks Networks
	CommonConfiguration
}

type Networks []NetworkConfiguration

func (n Networks) EncodeValues(key string, v *url.Values) error {
	var (
		networkingFields []reflect.Type
		networkingValues []reflect.Value
	)

	for index, config := range n {
		networkingFields = append(networkingFields, reflect.TypeOf(config))
		networkingValues = append(networkingValues, reflect.ValueOf(config))

		numNetworks := strconv.Itoa(index)

		for i, subnet := range config.Subnets {
			networkingFields = append(networkingFields, reflect.TypeOf(subnet))
			networkingValues = append(networkingValues, reflect.ValueOf(subnet))

			numSubnets := strconv.Itoa(i)

			assignIndex(networkingFields, networkingValues, numNetworks, numSubnets, v)
		}
	}

	return nil
}

func assignIndex(fields []reflect.Type, values []reflect.Value, numNetworks string, numSubnets string, urlValues *url.Values) {
	for index, v := range values {
		for i := 0; i < v.NumField(); i++ {
			field := fields[index].Field(i)
			value := v.Field(i)
			tag := field.Tag.Get("url")

			if tag == "" {
				continue
			}

			newTag := strings.Replace(tag, "**subnet**", numSubnets, -1)
			finalTag := strings.Replace(newTag, "**network**", numNetworks, -1)

			switch value.Kind() {
			case reflect.Int:
				urlValues.Set(finalTag, "0")
			case reflect.String:
				urlValues.Set(finalTag, value.String())
			case reflect.Bool:
				boolString := "0"
				if value.Bool() {
					boolString = "1"
				}
				urlValues.Set(finalTag, boolString)
			case reflect.Slice:
				temp := *urlValues
				temp[finalTag] = value.Interface().([]string)
			}
		}
	}
}

type NetworkConfiguration struct {
	GUID           int      `url:"network_collection[networks_attributes][**network**][guid]"`
	Name           string   `url:"network_collection[networks_attributes][**network**][name]" json:"name"`
	ServiceNetwork bool     `url:"network_collection[networks_attributes][**network**][service_network]" json:"service_network"`
	IAASIdentifier string   `url:"network_collection[networks_attributes][**network**][subnets][**subnet**][iaas_identifier]" json:"iaas_identifier"`
	Subnets        []Subnet `json:"subnets"`
}

type Subnet struct {
	CIDR                  string   `url:"network_collection[networks_attributes][**network**][subnets][**subnet**][cidr]" json:"cidr"`
	ReservedIPRanges      string   `url:"network_collection[networks_attributes][**network**][subnets][**subnet**][reserved_ip_ranges]" json:"reserved_ip_ranges"`
	DNS                   string   `url:"network_collection[networks_attributes][**network**][subnets][**subnet**][dns]" json:"dns"`
	Gateway               string   `url:"network_collection[networks_attributes][**network**][subnets][**subnet**][gateway]" json:"gateway"`
	AvailabilityZones     []string `json:"availability_zones,omitempty"`
	AvailabilityZoneGUIDs []string `url:"network_collection[networks_attributes][**network**][subnets][**subnet**][availability_zone_references][]"`
}
