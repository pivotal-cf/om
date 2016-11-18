package api_test

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/network"
)

var _ = Describe("AWSIaasConfigurationService", func() {
	var (
		server *ghttp.Server
	)
	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})
	Describe("FConfigure", func() {
		It("makes a request to IaasConfiguration edit", func() {
			//Using unauthenticated client as don't want to verify the setup of WebClient
			client := network.NewUnauthenticatedClient(server.URL(), true, 1800*time.Second)

			pemBytes, _ := ioutil.ReadFile("fixtures/fake.pem")
			pem := string(pemBytes)
			data := url.Values{}
			data.Set("_method", "put")
			data.Set("authenticity_token", "a-token")
			data.Add("iaas_configuration[access_key_id]", "myAccessKey")
			data.Add("iaas_configuration[secret_access_key]", "mySecretKey")
			data.Add("iaas_configuration[iam_instance_profile]", "")
			data.Add("iaas_configuration[vpc_id]", "myVPC")
			data.Add("iaas_configuration[security_group]", "mySG")
			data.Add("iaas_configuration[key_pair_name]", "myKeyPair")
			data.Add("iaas_configuration[ssh_private_key]", pem)
			data.Add("iaas_configuration[region]", "us-east-1")
			data.Add("iaas_configuration[encrypted]", "0")
			payload := data.Encode()

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/infrastructure/iaas_configuration/edit"),
					ghttp.RespondWith(http.StatusOK, "<meta name=\"csrf-token\" content=\"a-token\"/>"),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/infrastructure/iaas_configuration"),
					ghttp.VerifyBody([]byte(payload)),
					ghttp.RespondWith(http.StatusOK, ""),
				),
			)
			service := api.NewAWSIaasConfigurationService(client)
			output, err := service.Configure(api.AWSIaasConfigurationInput{
				AccessKey:       "myAccessKey",
				SecretKey:       "mySecretKey",
				VPCID:           "myVPC",
				SecurityGroupID: "mySG",
				KeyPairName:     "myKeyPair",
				PrivateKey:      pem,
				Region:          "us-east-1",
				Encrypted:       false,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})
})
