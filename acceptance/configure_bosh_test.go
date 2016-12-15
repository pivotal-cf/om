package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("configure-bosh command", func() {
	var (
		server          *httptest.Server
		receivedCookies []*http.Cookie
		Forms           []url.Values
	)

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/uaa/oauth/token":
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
			case "/infrastructure/iaas_configuration/edit":
				http.SetCookie(w, &http.Cookie{
					Name:  "somecookie",
					Value: "somevalue",
					Path:  "/",
				})

				w.Write([]byte(`<html>
				<body>
					<form action="/some-form" method="post">
						<input name="_method" value="fakemethod" />
						<input name="authenticity_token" value="fake_authenticity" />
					</form>
					</body>
				</html>`))
			case "/infrastructure/director_configuration/edit":
				w.Write([]byte(`<html>
				<body>
					<form action="/some-form" method="post">
						<input name="_method" value="fakemethod" />
						<input name="authenticity_token" value="fake_authenticity" />
					</form>
					</body>
				</html>`))
			case "/infrastructure/networks/edit":
				w.Write([]byte(`<html>
				<body>
					<form action="/some-form" method="post">
						<input name="_method" value="fakemethod" />
						<input name="authenticity_token" value="fake_authenticity" />
						<input name="network_collection[networks_attributes][][guid]" type="hidden" value="some-network-guid" \>
						<input name="network_collection[networks_attributes][][name]" value="some-network-name" \>
					</form>
					</body>
				</html>`))
			case "/infrastructure/availability_zones/edit":
				w.Write([]byte(`<html>
				<body>
					<form action="/some-form" method="post">
						<input name="_method" value="fakemethod" />
						<input name="authenticity_token" value="fake_authenticity" />
						<input name="availability_zones[availability_zones][][iaas_identifier]" type="hidden" value="some-az-1" \>
						<input name="availability_zones[availability_zones][][iaas_identifier]" type="hidden" value="some-other-az-2" \>
						<input name="availability_zones[availability_zones][][guid]" type="hidden" value="my-az-guid1" \>
						<input name="availability_zones[availability_zones][][guid]" type="hidden" value="my-az-guid2" \>
					</form>
					</body>
				</html>`))
			case "/infrastructure/director/az_and_network_assignment/edit":
				w.Write([]byte(`<html>
				<body>
					<form action="/some-form" method="some-method">
						<input name="_method" value="fakemethod" />
						<input name="authenticity_token" value="fake_authenticity" />
						<select name="bosh_product[network_reference]" id="bosh_product_network_reference">
							<option value=""></option>
							<option value="some-network-guid">some-network</option>
						</select>
					</form>
				</body>
			</html>`))
			case "/infrastructure/security_tokens/edit":
				w.Write([]byte(`<html>
				<body>
					<form action="/some-form" method="post">
						<input name="_method" value="fakemethod" />
						<input name="authenticity_token" value="fake_authenticity" />
					</form>
					</body>
				</html>`))
			case "/some-form":
				receivedCookies = req.Cookies()
				req.ParseForm()
				Forms = append(Forms, req.Form)
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	AfterEach(func() {
		Forms = []url.Values{}
	})

	Context("GCP", func() {
		var (
			command *exec.Cmd
		)

		BeforeEach(func() {
			iaasConfiguration := `{
				"project": "my-project",
				"default_deployment_tag": "my-vms",
				"auth_json": "{\"some-auth-field\": \"some-value\",\"some-private-key\": \"some-private-key\"}"
			}`

			directorConfiguration := `{
				"ntp_servers_string": "some-ntp-servers-string",
				"metrics_ip": "some-metrics-ip",
				"hm_pager_duty_options": {
					"enabled": true
				}
			}`

			availabilityZonesConfiguration := `{
			  "availability_zones": ["some-az-1", "some-other-az-2"]
			}`

			securityConfiguration := `{
				"trusted_certificates": "some-trusted-certificates",
				"vm_password_type": "some-vm-password-type"
			}`

			networkConfiguration := `{
				"icmp_checks_enabled": true,
				"networks": [{
					"name": "some-network",
					"service_network": true,
					"iaas_identifier": "some-iaas-identifier",
					"subnets": [
						{
							"cidr": "10.0.1.0/24",
							"reserved_ip_ranges": "10.0.1.0-10.0.1.4",
							"dns": "8.8.8.8",
							"gateway": "10.0.1.1",
							"availability_zones": [
								"some-az-1",
								"some-other-az-2"
							]
						}
					]
				}]
			}`

			networkAssignment := `{
				  "singleton_availability_zone": "some-az-1",
					"network": "some-network"
			}`

			command = exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "fake-username",
				"--password", "fake-password",
				"--skip-ssl-validation",
				"configure-bosh",
				"--iaas-configuration", iaasConfiguration,
				"--director-configuration", directorConfiguration,
				"--security-configuration", securityConfiguration,
				"--az-configuration", availabilityZonesConfiguration,
				"--networks-configuration", networkConfiguration,
				"--network-assignment", networkAssignment)
		})

		It("configures the bosh tile with the provided bosh configuration", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(session.Out).To(gbytes.Say("configuring iaas specific options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("configuring director options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("configuring availability zones for bosh tile"))
			Expect(session.Out).To(gbytes.Say("configuring network options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("configuring security options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("finished configuring bosh tile"))

			Expect(receivedCookies).To(HaveLen(1))
			Expect(receivedCookies[0].Name).To(Equal("somecookie"))

			Expect(Forms[0].Get("iaas_configuration[project]")).To(Equal("my-project"))
			Expect(Forms[0].Get("iaas_configuration[default_deployment_tag]")).To(Equal("my-vms"))
			Expect(Forms[0].Get("iaas_configuration[auth_json]")).To(Equal(`{"some-auth-field": "some-value","some-private-key": "some-private-key"}`))
			Expect(Forms[0].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[0].Get("_method")).To(Equal("fakemethod"))

			Expect(Forms[1].Get("director_configuration[ntp_servers_string]")).To(Equal("some-ntp-servers-string"))
			Expect(Forms[1].Get("director_configuration[metrics_ip]")).To(Equal("some-metrics-ip"))
			Expect(Forms[1].Get("director_configuration[hm_pager_duty_options][enabled]")).To(Equal("true"))
			Expect(Forms[1].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[1].Get("_method")).To(Equal("fakemethod"))

			Expect(Forms[2]["availability_zones[availability_zones][][iaas_identifier]"]).To(Equal([]string{"some-az-1", "some-other-az-2"}))
			Expect(Forms[2].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[2].Get("_method")).To(Equal("fakemethod"))

			Expect(Forms[3].Get("infrastructure[icmp_checks_enabled]")).To(Equal("1"))
			Expect(Forms[3].Get("network_collection[networks_attributes][0][name]")).To(Equal("some-network"))
			Expect(Forms[3].Get("network_collection[networks_attributes][0][service_network]")).To(Equal("1"))
			Expect(Forms[3].Get("network_collection[networks_attributes][0][subnets][0][iaas_identifier]")).To(Equal("some-iaas-identifier"))
			Expect(Forms[3].Get("network_collection[networks_attributes][0][subnets][0][cidr]")).To(Equal("10.0.1.0/24"))
			Expect(Forms[3].Get("network_collection[networks_attributes][0][subnets][0][reserved_ip_ranges]")).To(Equal("10.0.1.0-10.0.1.4"))
			Expect(Forms[3].Get("network_collection[networks_attributes][0][subnets][0][dns]")).To(Equal("8.8.8.8"))
			Expect(Forms[3].Get("network_collection[networks_attributes][0][subnets][0][gateway]")).To(Equal("10.0.1.1"))
			Expect(Forms[3]["network_collection[networks_attributes][0][subnets][0][availability_zone_references][]"]).To(ConsistOf([]string{"my-az-guid1", "my-az-guid2"}))
			Expect(Forms[3].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[3].Get("_method")).To(Equal("fakemethod"))

			Expect(Forms[4].Get("bosh_product[singleton_availability_zone_reference]")).To(Equal("my-az-guid1"))
			Expect(Forms[4].Get("bosh_product[network_reference]")).To(Equal("some-network-guid"))
			Expect(Forms[4].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[4].Get("_method")).To(Equal("fakemethod"))

			Expect(Forms[5].Get("security_tokens[trusted_certificates]")).To(Equal("some-trusted-certificates"))
			Expect(Forms[5].Get("security_tokens[vm_password_type]")).To(Equal("some-vm-password-type"))
			Expect(Forms[5].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[5].Get("_method")).To(Equal("fakemethod"))
		})

		It("does not configure keys that are not part of input", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			_, ok := Forms[0]["iaas_configuration[subscription_id]"]
			Expect(ok).To(BeFalse())
		})
	})

	Context("Azure", func() {
		var command *exec.Cmd

		BeforeEach(func() {
			iaasConfiguration := `{
				"subscription_id": "my-subscription",
				"tenant_id": "my-tenant",
				"client_id": "my-client",
				"client_secret": "my-client-secret",
				"resource_group_name": "my-group",
				"bosh_storage_account_name": "my-storage-account",
				"deployments_storage_account_name": "my-deployments-storage-account",
				"default_security_group": "my-security-group",
				"ssh_public_key": "my-public-key",
				"ssh_private_key": "my-private-key"
			}`

			command = exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "fake-username",
				"--password", "fake-password",
				"--skip-ssl-validation",
				"configure-bosh",
				"--iaas-configuration", iaasConfiguration)
		})

		It("configures the bosh tile with the provided iaas configuration", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(session.Out).To(gbytes.Say("configuring iaas specific options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("finished configuring bosh tile"))

			Expect(Forms[0].Get("iaas_configuration[subscription_id]")).To(Equal("my-subscription"))
			Expect(Forms[0].Get("iaas_configuration[tenant_id]")).To(Equal("my-tenant"))
			Expect(Forms[0].Get("iaas_configuration[client_id]")).To(Equal("my-client"))
			Expect(Forms[0].Get("iaas_configuration[client_secret]")).To(Equal("my-client-secret"))
			Expect(Forms[0].Get("iaas_configuration[resource_group_name]")).To(Equal("my-group"))
			Expect(Forms[0].Get("iaas_configuration[bosh_storage_account_name]")).To(Equal("my-storage-account"))
			Expect(Forms[0].Get("iaas_configuration[deployments_storage_account_name]")).To(Equal("my-deployments-storage-account"))
			Expect(Forms[0].Get("iaas_configuration[default_security_group]")).To(Equal("my-security-group"))
			Expect(Forms[0].Get("iaas_configuration[ssh_public_key]")).To(Equal("my-public-key"))
			Expect(Forms[0].Get("iaas_configuration[ssh_private_key]")).To(Equal("my-private-key"))

			Expect(Forms[0].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[0].Get("_method")).To(Equal("fakemethod"))
		})

		It("does not configure keys that are not part of input", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			_, ok := Forms[0]["iaas_configuration[project]"]
			Expect(ok).To(BeFalse())
		})
	})

	Context("AWS", func() {
		var command *exec.Cmd

		BeforeEach(func() {
			iaasConfiguration := `{
				"access_key_id": "my-access-key",
				"secret_access_key": "my-secret-key",
				"vpc_id": "my-vpc",
				"security_group": "my-security-group",
				"key_pair_name": "my-key-pair",
				"ssh_private_key": "my-private-ssh-key",
				"region": "some-region",
				"encrypted": true
			}`

			command = exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "fake-username",
				"--password", "fake-password",
				"--skip-ssl-validation",
				"configure-bosh",
				"--iaas-configuration", iaasConfiguration)
		})

		It("configures the bosh tile with the provided iaas configuration", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(session.Out).To(gbytes.Say("configuring iaas specific options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("finished configuring bosh tile"))

			Expect(Forms[0].Get("iaas_configuration[access_key_id]")).To(Equal("my-access-key"))
			Expect(Forms[0].Get("iaas_configuration[secret_access_key]")).To(Equal("my-secret-key"))
			Expect(Forms[0].Get("iaas_configuration[vpc_id]")).To(Equal("my-vpc"))
			Expect(Forms[0].Get("iaas_configuration[security_group]")).To(Equal("my-security-group"))
			Expect(Forms[0].Get("iaas_configuration[key_pair_name]")).To(Equal("my-key-pair"))
			Expect(Forms[0].Get("iaas_configuration[ssh_private_key]")).To(Equal("my-private-ssh-key"))
			Expect(Forms[0].Get("iaas_configuration[region]")).To(Equal("some-region"))
			Expect(Forms[0].Get("iaas_configuration[encrypted]")).To(Equal("true"))

			Expect(Forms[0].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[0].Get("_method")).To(Equal("fakemethod"))
		})

		It("does not configure keys that are not part of input", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			_, ok := Forms[0]["iaas_configuration[subscription_id]"]
			Expect(ok).To(BeFalse())
		})
	})

	Context("vSphere", func() {
		var command *exec.Cmd

		BeforeEach(func() {
			iaasConfiguration := `{
				"vcenter_host": "some-vcenter-host",
				"vcenter_username": "my-vcenter-username",
				"vcenter_password": "my-vcenter-password",
				"datacenter": "some-datacenter-name",
				"disk_type": "some-virtual-disk-type",
				"ephemeral_datastores_string": "some-ephemeral-datastores",
				"persistent_datastores_string": "some-persistent-datastores",
				"bosh_vm_folder": "some-vm-folder",
				"bosh_template_folder": "some-template-folder",
				"bosh_disk_path": "some-disk-path"
			}`

			command = exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "fake-username",
				"--password", "fake-password",
				"--skip-ssl-validation",
				"configure-bosh",
				"--iaas-configuration", iaasConfiguration)
		})

		It("configures the bosh tile with the provided iaas configuration", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(session.Out).To(gbytes.Say("configuring iaas specific options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("finished configuring bosh tile"))

			Expect(Forms[0].Get("iaas_configuration[vcenter_host]")).To(Equal("some-vcenter-host"))
			Expect(Forms[0].Get("iaas_configuration[vcenter_username]")).To(Equal("my-vcenter-username"))
			Expect(Forms[0].Get("iaas_configuration[vcenter_password]")).To(Equal("my-vcenter-password"))
			Expect(Forms[0].Get("iaas_configuration[datacenter]")).To(Equal("some-datacenter-name"))
			Expect(Forms[0].Get("iaas_configuration[disk_type]")).To(Equal("some-virtual-disk-type"))
			Expect(Forms[0].Get("iaas_configuration[ephemeral_datastores_string]")).To(Equal("some-ephemeral-datastores"))
			Expect(Forms[0].Get("iaas_configuration[persistent_datastores_string]")).To(Equal("some-persistent-datastores"))
			Expect(Forms[0].Get("iaas_configuration[bosh_vm_folder]")).To(Equal("some-vm-folder"))
			Expect(Forms[0].Get("iaas_configuration[bosh_template_folder]")).To(Equal("some-template-folder"))
			Expect(Forms[0].Get("iaas_configuration[bosh_disk_path]")).To(Equal("some-disk-path"))

			Expect(Forms[0].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[0].Get("_method")).To(Equal("fakemethod"))
		})

		It("does not configure keys that are not part of input", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			_, ok := Forms[0]["iaas_configuration[subscription_id]"]
			Expect(ok).To(BeFalse())
		})
	})
})
