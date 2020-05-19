package generator_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/configtemplate/generator"
)

type Template struct {
	NetworkProperties interface{} `yaml:"network-properties"`
	ProductProperties interface{} `yaml:"product-properties"`
	ResourceConfig    interface{} `yaml:"resource-config,omitempty"`
	ErrandConfig      interface{} `yaml:"errand-config,omitempty"`
}

var _ = Describe("Executor", func() {
	Context("CreateTemplate", func() {
		var (
			gen      *generator.Executor
			metadata *generator.Metadata
		)
		BeforeEach(func() {
			gen = &generator.Executor{}
			metadata = getMetadata("fixtures/metadata/p_healthwatch.yml")
		})

		It("Should create output template with network properties", func() {
			template, err := gen.CreateTemplate(metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(template).ToNot(BeNil())
			Expect(template.NetworkProperties).ToNot(BeNil())
		})
		It("Should create output template with product properties", func() {
			template, err := gen.CreateTemplate(metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(template).ToNot(BeNil())
			Expect(template.ProductProperties).ToNot(BeNil())
		})
		It("Should create output template with resource config properties", func() {
			template, err := gen.CreateTemplate(metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(template).ToNot(BeNil())
			Expect(template.ResourceConfig).ToNot(BeNil())
		})
		It("Should not create output template with syslog config properties", func() {
			template, err := gen.CreateTemplate(metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(template).ToNot(BeNil())
			Expect(template.SyslogProperties).To(BeNil())
		})
	})

	It("Adds syslog section if supported", func() {
		gen := &generator.Executor{}
		metadata := getMetadata("fixtures/metadata/p-spring-cloud-services.yml")
		template, err := gen.CreateTemplate(metadata)
		Expect(err).ToNot(HaveOccurred())
		Expect(template).ToNot(BeNil())
		Expect(template.SyslogProperties).ToNot(BeNil())
	})

	Context("Generate", func() {
		var (
			pwd, _  = ioutil.TempDir("", "")
			testGen = path.Join(pwd, "_testGen")
			tmpPath = path.Join(testGen, "templates")
		)

		AfterEach(func() {
			err := os.RemoveAll(testGen)
			Expect(err).ToNot(HaveOccurred())
		})

		It("generates files for all the fixtures", func() {
			fixtures, err := filepath.Glob("./fixtures/metadata/*.yml")
			Expect(err).ToNot(HaveOccurred())

			sort.Strings(fixtures)

			for _, fixtureFilename := range fixtures {
				metadataBytes, err := getFileBytes(fixtureFilename)
				Expect(err).ToNot(HaveOccurred())
				gen := generator.NewExecutor(metadataBytes, tmpPath, false, true, 10, false)
				err = gen.Generate()
				Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("expected %s to be a valid fixture", fixtureFilename))
			}
		})

		It("Should generate files for pks", func() {
			By("successfully executing the generator")

			metadataBytes, err := getFileBytes("./fixtures/metadata/pks.yml")
			Expect(err).ToNot(HaveOccurred())
			gen := generator.NewExecutor(metadataBytes, tmpPath, false, true, 10, false)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())

			By("generating files in directories")

			productPath := path.Join(tmpPath, "pivotal-container-service", "1.1.3-build.11")
			files := listFilesInDirectory(productPath)

			Expect(files).To(ConsistOf([]string{
				"/default-vars.yml",
				"/errand-vars.yml",
				"/features/cloud_provider-gcp.yml",
				"/features/cloud_provider-vsphere.yml",
				"/features/network_selector-nsx.yml",
				"/features/pks-vrli-enabled.yml",
				"/features/plan2_selector-active.yml",
				"/features/plan2_selector-inactive.yml",
				"/features/plan3_selector-active.yml",
				"/features/plan3_selector-inactive.yml",
				"/features/proxy_selector-enabled.yml",
				"/features/syslog_migration_selector-enabled.yml",
				"/features/telemetry_selector-disabled.yml",
				"/features/telemetry_selector-enabled.yml",
				"/features/uaa-ldap.yml",
				"/features/wavefront-enabled.yml",
				"/network/2-az-configuration.yml",
				"/network/3-az-configuration.yml",
				"/optional/add-network_selector-nsx-nsx-t-ca-cert.yml",
				"/optional/add-pks-vrli-enabled-ca_cert.yml",
				"/optional/add-plan1_selector-active-addons_spec.yml",
				"/optional/add-plan1_selector-active-allow_privileged_containers.yml",
				"/optional/add-plan1_selector-active-disable_deny_escalating_exec.yml",
				"/optional/add-plan1_selector-active-errand_vm_type.yml",
				"/optional/add-plan1_selector-active-master_persistent_disk_type.yml",
				"/optional/add-plan1_selector-active-master_vm_type.yml",
				"/optional/add-plan1_selector-active-worker_persistent_disk_type.yml",
				"/optional/add-plan1_selector-active-worker_vm_type.yml",
				"/optional/add-plan2_selector-active-addons_spec.yml",
				"/optional/add-plan2_selector-active-allow_privileged_containers.yml",
				"/optional/add-plan2_selector-active-disable_deny_escalating_exec.yml",
				"/optional/add-plan2_selector-active-errand_vm_type.yml",
				"/optional/add-plan2_selector-active-master_persistent_disk_type.yml",
				"/optional/add-plan2_selector-active-master_vm_type.yml",
				"/optional/add-plan2_selector-active-worker_persistent_disk_type.yml",
				"/optional/add-plan2_selector-active-worker_vm_type.yml",
				"/optional/add-plan3_selector-active-addons_spec.yml",
				"/optional/add-plan3_selector-active-allow_privileged_containers.yml",
				"/optional/add-plan3_selector-active-disable_deny_escalating_exec.yml",
				"/optional/add-plan3_selector-active-errand_vm_type.yml",
				"/optional/add-plan3_selector-active-master_persistent_disk_type.yml",
				"/optional/add-plan3_selector-active-master_vm_type.yml",
				"/optional/add-plan3_selector-active-worker_persistent_disk_type.yml",
				"/optional/add-plan3_selector-active-worker_vm_type.yml",
				"/optional/add-proxy_selector-enabled-http_proxy_credentials.yml",
				"/optional/add-proxy_selector-enabled-http_proxy_url.yml",
				"/optional/add-proxy_selector-enabled-https_proxy_credentials.yml",
				"/optional/add-proxy_selector-enabled-https_proxy_url.yml",
				"/optional/add-proxy_selector-enabled-no_proxy.yml",
				"/optional/add-syslog_migration_selector-enabled-ca_cert.yml",
				"/optional/add-syslog_migration_selector-enabled-permitted_peer.yml",
				"/optional/add-uaa-ldap-email_domains.yml",
				"/optional/add-uaa-ldap-first_name_attribute.yml",
				"/optional/add-uaa-ldap-group_search_base.yml",
				"/optional/add-uaa-ldap-last_name_attribute.yml",
				"/optional/add-uaa-ldap-server_ssl_cert.yml",
				"/optional/add-uaa-ldap-server_ssl_cert_alias.yml",
				"/optional/add-vm_extensions.yml",
				"/optional/add-wavefront-enabled-wavefront_alert_targets.yml",
				"/product.yml",
				"/required-vars.yml",
				"/resource-vars.yml",
				"/resource/pivotal-container-service_additional_vm_extensions.yml",
				"/resource/pivotal-container-service_elb_names.yml",
				"/resource/pivotal-container-service_internet_connected.yml",
				"/resource/pivotal-container-service_nsx_security_groups.yml",
			}))

			By("being able to unmarshal a template")

			template, err := unmarshalProduct(path.Join(tmpPath, "pivotal-container-service", "1.1.3-build.11", "product.yml"))
			Expect(err).ToNot(HaveOccurred())
			Expect(template.NetworkProperties).ToNot(BeNil())
			Expect(template.ResourceConfig).ToNot(BeNil())
		})
	})
})

func listFilesInDirectory(productPath string) []string {
	files := []string{}

	filepath.Walk(productPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, strings.Replace(path, productPath, "", 1))
		}
		return nil
	})

	sort.Strings(files)

	return files
}

func getFileBytes(metadataFile string) ([]byte, error) {
	return ioutil.ReadFile(metadataFile)
}

func unmarshalProduct(targetFile string) (*Template, error) {
	template := &Template{}
	yamlFile, err := ioutil.ReadFile(targetFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, template)
	if err != nil {
		return nil, err
	}
	return template, nil
}
