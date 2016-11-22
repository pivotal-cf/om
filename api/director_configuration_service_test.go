package api_test

import (
	"net/http"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/network"
)

var _ = Describe("DirectorConfigurationService", func() {
	var (
		server *ghttp.Server
	)
	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})
	Describe("Configure", func() {
		It("makes a request to DirectorConfiguration edit with external blob store and database", func() {
			var (
				controlNtpServers       = "0.amazon.pool.ntp.org"
				controlAccessKey        = "myAccessKey"
				controlSecretKey        = "mySecretKey"
				controlResurrector      = true
				controlS3Endpoint       = "https://s3.amazonaws.com"
				controlS3BucketName     = "sandbox-pcf-bosh"
				controlS3Version        = "2"
				controlDatabaseHost     = "sandbox-pcf.us-east-1.rds.amazonaws.com"
				controlDatabasePort     = "3306"
				controlDatabaseUser     = "bosh"
				controlDatabasePassword = "boshbosh"
				controlDatabase         = "bosh"
			)
			//Using unauthenticated client as don't want to verify the setup of WebClient
			client := network.NewUnauthenticatedClient(server.URL(), true, 1800*time.Second)

			data := url.Values{}
			data.Set("_method", "put")
			data.Set("authenticity_token", "a-token")
			data.Set("director_configuration[ntp_servers_string]", controlNtpServers)
			data.Set("director_configuration[metrics_ip]", "")
			data.Set("director_configuration[resurrector_enabled]", api.GetBooleanAsFormValue(controlResurrector))
			data.Set("director_configuration[post_deploy_enabled]", "0")
			data.Set("director_configuration[bosh_recreate_on_next_deploy]", "0")
			data.Set("director_configuration[retry_bosh_deploys]", "0")
			data.Set("director_configuration[hm_pager_duty_options][enabled]", "0")
			data.Set("director_configuration[hm_emailer_options][enabled]", "0")
			data.Set("director_configuration[blobstore_type]", "s3")
			data.Set("director_configuration[s3_blobstore_options][endpoint]", controlS3Endpoint)
			data.Set("director_configuration[s3_blobstore_options][bucket_name]", controlS3BucketName)
			data.Set("director_configuration[s3_blobstore_options][access_key]", controlAccessKey)
			data.Set("director_configuration[s3_blobstore_options][secret_key]", controlSecretKey)
			data.Set("director_configuration[s3_blobstore_options][signature_version]", controlS3Version)
			data.Set("director_configuration[database_type]", "external")
			data.Set("director_configuration[external_database_options][host]", controlDatabaseHost)
			data.Set("director_configuration[external_database_options][port]", controlDatabasePort)
			data.Set("director_configuration[external_database_options][user]", controlDatabaseUser)
			data.Set("director_configuration[external_database_options][password]", controlDatabasePassword)
			data.Set("director_configuration[external_database_options][database]", controlDatabase)
			data.Set("director_configuration[max_threads]", "")
			data.Set("director_configuration[director_hostname]", "")
			payload := data.Encode()

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/infrastructure/director_configuration/edit"),
					ghttp.RespondWith(http.StatusOK, "<meta name=\"csrf-token\" content=\"a-token\"/>"),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/infrastructure/director_configuration"),
					ghttp.VerifyBody([]byte(payload)),
					ghttp.RespondWith(http.StatusOK, ""),
				),
			)
			service := api.NewDirectorConfigurationService(client)
			output, err := service.Configure(api.DirectorConfigurationInput{
				NTPServers:         controlNtpServers,
				S3AccessKey:        controlAccessKey,
				S3SecretKey:        controlSecretKey,
				S3Endpoint:         controlS3Endpoint,
				S3BucketName:       controlS3BucketName,
				S3SignatureVersion: controlS3Version,
				EnableResurrector:  controlResurrector,
				DatabaseHost:       controlDatabaseHost,
				DatabasePort:       controlDatabasePort,
				DatabaseUser:       controlDatabaseUser,
				DatabasePassword:   controlDatabasePassword,
				Database:           controlDatabase,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})

		It("makes a request to DirectorConfiguration edit with internal blob store and database", func() {
			var (
				controlNtpServers  = "0.amazon.pool.ntp.org"
				controlResurrector = true
			)
			//Using unauthenticated client as don't want to verify the setup of WebClient
			client := network.NewUnauthenticatedClient(server.URL(), true, 1800*time.Second)

			data := url.Values{}
			data.Set("_method", "put")
			data.Set("authenticity_token", "a-token")
			data.Set("director_configuration[ntp_servers_string]", controlNtpServers)
			data.Set("director_configuration[metrics_ip]", "")
			data.Set("director_configuration[resurrector_enabled]", api.GetBooleanAsFormValue(controlResurrector))
			data.Set("director_configuration[post_deploy_enabled]", "0")
			data.Set("director_configuration[bosh_recreate_on_next_deploy]", "0")
			data.Set("director_configuration[retry_bosh_deploys]", "0")
			data.Set("director_configuration[hm_pager_duty_options][enabled]", "0")
			data.Set("director_configuration[hm_emailer_options][enabled]", "0")
			data.Set("director_configuration[blobstore_type]", "internal")
			data.Set("director_configuration[s3_blobstore_options][endpoint]", "")
			data.Set("director_configuration[s3_blobstore_options][bucket_name]", "")
			data.Set("director_configuration[s3_blobstore_options][access_key]", "")
			data.Set("director_configuration[s3_blobstore_options][secret_key]", "")
			data.Set("director_configuration[s3_blobstore_options][signature_version]", "")
			data.Set("director_configuration[database_type]", "internal")
			data.Set("director_configuration[external_database_options][host]", "")
			data.Set("director_configuration[external_database_options][port]", "")
			data.Set("director_configuration[external_database_options][user]", "")
			data.Set("director_configuration[external_database_options][password]", "")
			data.Set("director_configuration[external_database_options][database]", "")
			data.Set("director_configuration[max_threads]", "")
			data.Set("director_configuration[director_hostname]", "")
			payload := data.Encode()

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/infrastructure/director_configuration/edit"),
					ghttp.RespondWith(http.StatusOK, "<meta name=\"csrf-token\" content=\"a-token\"/>"),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/infrastructure/director_configuration"),
					ghttp.VerifyBody([]byte(payload)),
					ghttp.RespondWith(http.StatusOK, ""),
				),
			)
			service := api.NewDirectorConfigurationService(client)
			output, err := service.Configure(api.DirectorConfigurationInput{
				NTPServers:        controlNtpServers,
				EnableResurrector: controlResurrector,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})
})
