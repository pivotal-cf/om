package api_test

import (
	"net/http"
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
	Describe("Configure", func() {
		It("makes a request to IaasConfiguration edit", func() {
			//Using unauthenticated client as don't want to verify the setup of WebClient
			client := network.NewUnauthenticatedClient(server.URL(), true, 1800*time.Second)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/infrastructure/iaas_configuration/edit"),
					ghttp.RespondWith(http.StatusOK, "<meta name=\"csrf-token\" content=\"a-token\"/>"),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/infrastructure/iaas_configuration"),
					ghttp.VerifyBody([]byte("_method=put&authenticity_token=a-token&iaas_configuration%5Baccess_key_id%5D=myAccessKey&iaas_configuration%5Bencrypted%5D=0&iaas_configuration%5Biam_instance_profile%5D=&iaas_configuration%5Bkey_pair_name%5D=myKeyPair&iaas_configuration%5Bregion%5D=us-east-1&iaas_configuration%5Bsecret_access_key%5D=mySecretKey&iaas_configuration%5Bsecurity_group%5D=mySG&iaas_configuration%5Bssh_private_key%5D=%0A%09-----BEGIN+RSA+PRIVATE+KEY-----%0A%09MIIEowIBAAKCAQEAj8gGr9rfF82ff6WqhfA60V29H3R8TmXAoJi5Tpftc2OHlzRi%2FeG3rQO6oF%2Fa%0A%09zwiG5PzOquYgDymTuopMdlThA0N5Zebyy5%2FGFBkgOERYE3pg0fyVt4mdQSKwoWmePmk24uzQTH6g%0A%09LciSgXE%2BE5XNi8baZ%2FdR%2B9Y8TkzIM5IWoYjlXvmoR7BToclopyAkB%2FynDurrKbLiy1uzLgXi01kD%0A%09JASk1AOLonNZK%2FsTxuwyz8K6quc0JUuHW3XV%2FL5sYbZ2yRXV7DMyP6tnzShaSsJmHbpOmAZuEdE0%0A%09jlBcZ055rl%2BVNyoQRFHwbY%2BA6%2FWQGs3z8EAO391Su2qk0VziMB%2FGEwIDAQABAoIBABDlK0v8xxxP%0A%098D8ao3gLq42wmymYEYdQ05rLd3LxzVgyyNsOr9cvb9rnxkVUHi0XzmoX%2Ftz8H6ZId3CDEcsR3EL6%0A%096cf8pGkjBC6WaT%2FTYBoh1kC2ko03oX%2Fm6KR%2B%2BbZRvL5bn3ptorreh8CoIuhAQR%2F8pOt0l3ClUv7M%0A%09bPYKY5hlhH31eNfdkkPkT21YtO03RvIrqn3sWENm%2FLgZQH6gjk%2FGUX%2FEQLFp2OQMQ2qOPRvaMEp%2F%0A%094hXon4FQHCRrquFUcTNGUkARgUeAq4PP6SUvopX2xYgLAozGM1KoyDCSTd77NO2lLqSRfWMzws7i%0A%09poc4UjpCUw96qM3tf0Y6%2B1HDL5kCgYEA7%2B0RrTAbrxNWd%2FT33Ccgqs7RsZGFfUzD7AmqdverqNLK%0A%09LLPGPwtjfS5T%2F5VhwP5lRDvfz8I21wK0Xe%2F2jzrU271oMxhmVz0UHqwARN9wdyRqe2ltsLB4NYmT%0A%09lUH0edAxRIsEJUCAKGRMpHPRziNH%2FMBqpZr%2BtBKE3I%2BcPY3%2BDecCgYEAmWn%2B5DXIwzlLxSBigy94%0A%09jUWyN8eepbGtnThvvvfddOZ3SzZrn7UP2WbClpONDHVoGG84%2BGwqamggZOr6v3%2FtUpYmH1cuta25%0A%09hliXP283eueU62OAJ2XG9XuK2lCmaL0AnBq99Zgkjnhzwq9XEfxZNNUQPVy%2FOr8o9WA6MUHgyPUC%0A%09gYA5msrOsSlEbLkrDfbgtchDGmsAXjcVsXOs3Vk%2FPRHK8%2Bk0uGkVw%2B88I%2F5o8%2F3Hb4zyyAlhgXjX%0A%09QL%2F7edzR4McwhxZYhjg0I%2BcLwjJCVv9Sq7yhKtv6OzRxbjmv8Wj4QkNB%2BLqjEwxyJjq8lU4%2FVvs2%0A%09tSAl6MPUikm6BwT1Rn1D6wKBgGNUlt2p2UhV07JZyo8H8HT%2F%2BGlXTWgZB8ExJmEuWWv0QK8pGDv6%0A%093r0zZLBb7spvRivz754hYsEslDTjU%2BEj8kQzxZErQKoPRn0u5RcEapagVPKnpPVdV5ngGMJLz8Mn%0A%09BLsOMYpPrPO2F7WpE6YojpW%2Fklk4sPRXiyx81pDIB8P1AoGBAJUMYOiXj%2BE6ylnMXE%2B0WaYqp3L5%0A%09IKCffCKjb9YTjxKmi1Ti04%2B2Km%2F54f0YyWGkZZlHnmZ%2BK1el4TI48Xdwg5%2FVF1uU2vOk5uJ2TbD1%0A%09X7cDWZplTkSE6YpynNjENoTihobwnjXEcebuREjtf%2FClnDt06foe%2BUxFWCQSx2CEJxCp%0A%09-----END+RSA+PRIVATE+KEY-----&iaas_configuration%5Bvpc_id%5D=myVPC")),
					ghttp.RespondWith(http.StatusOK, ""),
				),
			)
			service := api.NewAWSIaasConfigurationService(client)
			pem := `
	-----BEGIN RSA PRIVATE KEY-----
	MIIEowIBAAKCAQEAj8gGr9rfF82ff6WqhfA60V29H3R8TmXAoJi5Tpftc2OHlzRi/eG3rQO6oF/a
	zwiG5PzOquYgDymTuopMdlThA0N5Zebyy5/GFBkgOERYE3pg0fyVt4mdQSKwoWmePmk24uzQTH6g
	LciSgXE+E5XNi8baZ/dR+9Y8TkzIM5IWoYjlXvmoR7BToclopyAkB/ynDurrKbLiy1uzLgXi01kD
	JASk1AOLonNZK/sTxuwyz8K6quc0JUuHW3XV/L5sYbZ2yRXV7DMyP6tnzShaSsJmHbpOmAZuEdE0
	jlBcZ055rl+VNyoQRFHwbY+A6/WQGs3z8EAO391Su2qk0VziMB/GEwIDAQABAoIBABDlK0v8xxxP
	8D8ao3gLq42wmymYEYdQ05rLd3LxzVgyyNsOr9cvb9rnxkVUHi0XzmoX/tz8H6ZId3CDEcsR3EL6
	6cf8pGkjBC6WaT/TYBoh1kC2ko03oX/m6KR++bZRvL5bn3ptorreh8CoIuhAQR/8pOt0l3ClUv7M
	bPYKY5hlhH31eNfdkkPkT21YtO03RvIrqn3sWENm/LgZQH6gjk/GUX/EQLFp2OQMQ2qOPRvaMEp/
	4hXon4FQHCRrquFUcTNGUkARgUeAq4PP6SUvopX2xYgLAozGM1KoyDCSTd77NO2lLqSRfWMzws7i
	poc4UjpCUw96qM3tf0Y6+1HDL5kCgYEA7+0RrTAbrxNWd/T33Ccgqs7RsZGFfUzD7AmqdverqNLK
	LLPGPwtjfS5T/5VhwP5lRDvfz8I21wK0Xe/2jzrU271oMxhmVz0UHqwARN9wdyRqe2ltsLB4NYmT
	lUH0edAxRIsEJUCAKGRMpHPRziNH/MBqpZr+tBKE3I+cPY3+DecCgYEAmWn+5DXIwzlLxSBigy94
	jUWyN8eepbGtnThvvvfddOZ3SzZrn7UP2WbClpONDHVoGG84+GwqamggZOr6v3/tUpYmH1cuta25
	hliXP283eueU62OAJ2XG9XuK2lCmaL0AnBq99Zgkjnhzwq9XEfxZNNUQPVy/Or8o9WA6MUHgyPUC
	gYA5msrOsSlEbLkrDfbgtchDGmsAXjcVsXOs3Vk/PRHK8+k0uGkVw+88I/5o8/3Hb4zyyAlhgXjX
	QL/7edzR4McwhxZYhjg0I+cLwjJCVv9Sq7yhKtv6OzRxbjmv8Wj4QkNB+LqjEwxyJjq8lU4/Vvs2
	tSAl6MPUikm6BwT1Rn1D6wKBgGNUlt2p2UhV07JZyo8H8HT/+GlXTWgZB8ExJmEuWWv0QK8pGDv6
	3r0zZLBb7spvRivz754hYsEslDTjU+Ej8kQzxZErQKoPRn0u5RcEapagVPKnpPVdV5ngGMJLz8Mn
	BLsOMYpPrPO2F7WpE6YojpW/klk4sPRXiyx81pDIB8P1AoGBAJUMYOiXj+E6ylnMXE+0WaYqp3L5
	IKCffCKjb9YTjxKmi1Ti04+2Km/54f0YyWGkZZlHnmZ+K1el4TI48Xdwg5/VF1uU2vOk5uJ2TbD1
	X7cDWZplTkSE6YpynNjENoTihobwnjXEcebuREjtf/ClnDt06foe+UxFWCQSx2CEJxCp
	-----END RSA PRIVATE KEY-----`
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
