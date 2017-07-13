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
			case "/api/v0/diagnostic_report":
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{}`))
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
						<div class="content">
							<input name="availability_zones[availability_zones][][iaas_identifier]" type="hidden" value="some-az-1" \>
							<input name="availability_zones[availability_zones][][iaas_identifier]" type="hidden" value="some-other-az-2" \>
						</div>
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
			case "/infrastructure/director/resources/edit":
				w.Write([]byte(`<html>
				<body>
					<form action="/some-form" method="post">
						<input name="_method" value="fakemethod" />
						<input name="authenticity_token" value="fake_authenticity" />
						<select name="product_resources_form[compilation][vm_type_id]" id="product_resources_form_compilation_vm_type_id">
							<option value="">Automatic: large.cpu (cpu: 4, ram: 4 GB, disk: 16 GB)</option>
							<option value="micro">micro (cpu: 1, ram: 1 GB, disk: 8 GB)</option>
							<option value="micro.cpu">micro.cpu (cpu: 2, ram: 2 GB, disk: 8 GB)</option>
							<option value="small">small (cpu: 1, ram: 2 GB, disk: 8 GB)</option>
							<option value="small.disk">small.disk (cpu: 1, ram: 2 GB, disk: 16 GB)</option>
							<option value="medium">medium (cpu: 2, ram: 4 GB, disk: 8 GB)</option>
							<option value="medium.mem">medium.mem (cpu: 1, ram: 6 GB, disk: 8 GB)</option>
							<option value="medium.disk">medium.disk (cpu: 2, ram: 4 GB, disk: 32 GB)</option>
							<option value="medium.cpu">medium.cpu (cpu: 4, ram: 4 GB, disk: 8 GB)</option>
							<option value="large">large (cpu: 2, ram: 8 GB, disk: 16 GB)</option>
							<option value="large.mem">large.mem (cpu: 2, ram: 12 GB, disk: 16 GB)</option>
							<option value="large.disk">large.disk (cpu: 2, ram: 8 GB, disk: 64 GB)</option>
							<option value="large.cpu">large.cpu (cpu: 4, ram: 4 GB, disk: 16 GB)</option>
							<option value="xlarge">xlarge (cpu: 4, ram: 16 GB, disk: 32 GB)</option>
							<option value="xlarge.mem">xlarge.mem (cpu: 4, ram: 24 GB, disk: 32 GB)</option>
							<option value="xlarge.disk">xlarge.disk (cpu: 4, ram: 16 GB, disk: 128 GB)</option>
							<option value="xlarge.cpu">xlarge.cpu (cpu: 8, ram: 8 GB, disk: 32 GB)</option>
							<option value="2xlarge">2xlarge (cpu: 8, ram: 32 GB, disk: 64 GB)</option>
							<option value="2xlarge.mem">2xlarge.mem (cpu: 8, ram: 48 GB, disk: 64 GB)</option>
							<option value="2xlarge.disk">2xlarge.disk (cpu: 8, ram: 32 GB, disk: 256 GB)</option>
							<option value="2xlarge.cpu">2xlarge.cpu (cpu: 16, ram: 16 GB, disk: 64 GB)</option>
						</select>
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
			"availability_zones": [
			  {"name": "some-az-1"},
			  {"name": "some-other-az-2"}
			]}`

			securityConfiguration := `{
				"trusted_certificates": "some-trusted-certificates",
				"vm_password_type": "some-vm-password-type"
			}`

			networkConfiguration := `{
				"icmp_checks_enabled": true,
				"networks": [{
					"name": "some-network",
					"service_network": true,
					"subnets": [
						{
							"iaas_identifier": "some-iaas-identifier",
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

			resourceConfig := `{
				"director": {
					"instance_type": {
						"id": "m1.medium"
					},
					"persistent_disk": {
						"size_mb": "20480"
					},
					"internet_connected": true,
					"elb_names": ["my-elb"]
				},
				"compilation": {
					"instances": 1,
					"instance_type": {
						"id": "m1.medium"
					},
					"internet_connected": true,
					"elb_names": ["my-elb"]
				}
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
				"--network-assignment", networkAssignment,
				"--resource-configuration", resourceConfig)
		})

		It("configures the bosh tile with the provided bosh configuration", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			Expect(session.Out).To(gbytes.Say("configuring iaas specific options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("configuring director options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("configuring security options for bosh tile"))
			Expect(session.Out).To(gbytes.Say("configuring resources for bosh tile"))
			Expect(session.Out).To(gbytes.Say("configuring availability zones for bosh tile"))
			Expect(session.Out).To(gbytes.Say("configuring network options for bosh tile"))
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

			Expect(Forms[2].Get("security_tokens[trusted_certificates]")).To(Equal("some-trusted-certificates"))
			Expect(Forms[2].Get("security_tokens[vm_password_type]")).To(Equal("some-vm-password-type"))
			Expect(Forms[2].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[2].Get("_method")).To(Equal("fakemethod"))

			Expect(Forms[3].Get("product_resources_form[director][vm_type_id]")).To(Equal("m1.medium"))
			Expect(Forms[3].Get("product_resources_form[director][disk_type_id]")).To(Equal("20480"))
			Expect(Forms[3].Get("product_resources_form[director][internet_connected]")).To(Equal("true"))
			Expect(Forms[3].Get("product_resources_form[director][elb_names]")).To(Equal("my-elb"))
			Expect(Forms[3].Get("product_resources_form[compilation][instances]")).To(Equal("1"))
			Expect(Forms[3].Get("product_resources_form[compilation][vm_type_id]")).To(Equal("m1.medium"))
			Expect(Forms[3].Get("product_resources_form[compilation][internet_connected]")).To(Equal("true"))
			Expect(Forms[3].Get("product_resources_form[compilation][elb_names]")).To(Equal("my-elb"))
			Expect(Forms[3].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[3].Get("_method")).To(Equal("fakemethod"))

			Expect(Forms[4]["availability_zones[availability_zones][][iaas_identifier]"]).To(Equal([]string{"some-az-1", "some-other-az-2"}))
			Expect(Forms[4].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[4].Get("_method")).To(Equal("fakemethod"))

			Expect(Forms[5].Get("infrastructure[icmp_checks_enabled]")).To(Equal("1"))
			Expect(Forms[5].Get("network_collection[networks_attributes][0][name]")).To(Equal("some-network"))
			Expect(Forms[5].Get("network_collection[networks_attributes][0][service_network]")).To(Equal("1"))
			Expect(Forms[5].Get("network_collection[networks_attributes][0][subnets][0][iaas_identifier]")).To(Equal("some-iaas-identifier"))
			Expect(Forms[5].Get("network_collection[networks_attributes][0][subnets][0][cidr]")).To(Equal("10.0.1.0/24"))
			Expect(Forms[5].Get("network_collection[networks_attributes][0][subnets][0][reserved_ip_ranges]")).To(Equal("10.0.1.0-10.0.1.4"))
			Expect(Forms[5].Get("network_collection[networks_attributes][0][subnets][0][dns]")).To(Equal("8.8.8.8"))
			Expect(Forms[5].Get("network_collection[networks_attributes][0][subnets][0][gateway]")).To(Equal("10.0.1.1"))
			Expect(Forms[5]["network_collection[networks_attributes][0][subnets][0][availability_zone_references][]"]).To(ConsistOf([]string{"my-az-guid1", "my-az-guid2"}))
			Expect(Forms[5].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[5].Get("_method")).To(Equal("fakemethod"))

			Expect(Forms[6].Get("bosh_product[singleton_availability_zone_reference]")).To(Equal("my-az-guid1"))
			Expect(Forms[6].Get("bosh_product[network_reference]")).To(Equal("some-network-guid"))
			Expect(Forms[6].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[6].Get("_method")).To(Equal("fakemethod"))
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

	Context("OpenStack", func() {
		var command *exec.Cmd

		BeforeEach(func() {
			iaasConfiguration := `{
				"api_ssl_cert": "os-api-ssl-cert",
				"disable_dhcp": false,
				"openstack_domain": "os-domain",
				"openstack_authentication_url": "https//openstack.com/v2",
				"ignore_server_availability_zone": true,
				"openstack_key_pair_name": "os-key-pair",
				"keystone_version": "v2.0",
				"openstack_password": "os-password",
				"openstack_region": "os-region",
				"openstack_security_group": "os-security-group",
				"openstack_tenant": "os-tenant",
				"openstack_username": "os-user",
				"ssh_private_key": "my-private-ssh-key"
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

			Expect(Forms[0].Get("iaas_configuration[api_ssl_cert]")).To(Equal("os-api-ssl-cert"))
			Expect(Forms[0].Get("iaas_configuration[disable_dhcp]")).To(Equal("false"))
			Expect(Forms[0].Get("iaas_configuration[domain]")).To(Equal("os-domain"))
			Expect(Forms[0].Get("iaas_configuration[identity_endpoint]")).To(Equal("https//openstack.com/v2"))
			Expect(Forms[0].Get("iaas_configuration[ignore_server_availability_zone]")).To(Equal("true"))
			Expect(Forms[0].Get("iaas_configuration[key_pair_name]")).To(Equal("os-key-pair"))
			Expect(Forms[0].Get("iaas_configuration[keystone_version]")).To(Equal("v2.0"))
			Expect(Forms[0].Get("iaas_configuration[password]")).To(Equal("os-password"))
			Expect(Forms[0].Get("iaas_configuration[region]")).To(Equal("os-region"))
			Expect(Forms[0].Get("iaas_configuration[security_group]")).To(Equal("os-security-group"))
			Expect(Forms[0].Get("iaas_configuration[tenant]")).To(Equal("os-tenant"))
			Expect(Forms[0].Get("iaas_configuration[username]")).To(Equal("os-user"))
			Expect(Forms[0].Get("iaas_configuration[ssh_private_key]")).To(Equal("my-private-ssh-key"))

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

			availabilityZonesConfiguration := `{"availability_zones": [
			    {
			      "name": "some-az-1",
			      "cluster": "some-cluster-1",
			      "resource_pool": "some-resource-pool-1"
			    },
			    {
			      "name": "some-other-az-2",
			      "cluster": "some-other-cluster-2",
			      "resource_pool": "some-other-resource-pool-2"
			    }
			  ]
			}`

			command = exec.Command(pathToMain,
				"--target", server.URL,
				"--username", "fake-username",
				"--password", "fake-password",
				"--skip-ssl-validation",
				"configure-bosh",
				"--iaas-configuration", iaasConfiguration,
				"--az-configuration", availabilityZonesConfiguration)
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

			Expect(Forms[1]["availability_zones[availability_zones][][name]"]).To(Equal([]string{"some-az-1", "some-other-az-2"}))
			Expect(Forms[1]["availability_zones[availability_zones][][cluster]"]).To(Equal([]string{"some-cluster-1", "some-other-cluster-2"}))
			Expect(Forms[1]["availability_zones[availability_zones][][resource_pool]"]).To(Equal([]string{"some-resource-pool-1", "some-other-resource-pool-2"}))
			Expect(Forms[1].Get("authenticity_token")).To(Equal("fake_authenticity"))
			Expect(Forms[1].Get("_method")).To(Equal("fakemethod"))
		})

		It("does not configure keys that are not part of input", func() {
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			_, ok := Forms[0]["iaas_configuration[subscription_id]"]
			Expect(ok).To(BeFalse())
		})

		Context("when adding nsx properties", func() {
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
					"bosh_disk_path": "some-disk-path",
					"nsx_networking_enabled": true,
					"nsx_address": "some-nsx-address",
					"nsx_password": "some-password",
					"nsx_username": "some-username",
					"nsx_ca_certificate": "some-nsx-ca-certificate"
				}`

				command = exec.Command(pathToMain,
					"--target", server.URL,
					"--username", "fake-username",
					"--password", "fake-password",
					"--skip-ssl-validation",
					"configure-bosh",
					"--iaas-configuration", iaasConfiguration)
			})

			It("configures the bosh tile with additional nsx properties", func() {
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))

				Expect(Forms[0].Get("iaas_configuration[nsx_networking_enabled]")).To(Equal("true"))
				Expect(Forms[0].Get("iaas_configuration[nsx_address]")).To(Equal("some-nsx-address"))
				Expect(Forms[0].Get("iaas_configuration[nsx_password]")).To(Equal("some-password"))
				Expect(Forms[0].Get("iaas_configuration[nsx_username]")).To(Equal("some-username"))
				Expect(Forms[0].Get("iaas_configuration[nsx_ca_certificate]")).To(Equal("some-nsx-ca-certificate"))
			})
		})
	})
})
