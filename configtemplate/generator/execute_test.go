package generator_test

import (
	"io/ioutil"
	"os"
	"path"

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
			fileData, err := ioutil.ReadFile("./fixtures/p_healthwatch.yml")
			Expect(err).ToNot(HaveOccurred())
			metadata, err = generator.NewMetadata(fileData)
			Expect(err).ToNot(HaveOccurred())
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
	})

	Context("Generate", func() {
		var (
			gen     *generator.Executor
			pwd, _  = ioutil.TempDir("", "")
			testGen = path.Join(pwd, "_testGen")
			tmpPath = path.Join(testGen, "templates")
		)
		BeforeEach(func() {

		})
		AfterEach(func() {
			err := os.RemoveAll(testGen)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should generate files for p-healthwatch", func() {
			metadataBytes, err := getFileBytes("./fixtures/p_healthwatch.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should generate files for pas", func() {
			metadataBytes, err := getFileBytes("./fixtures/pas.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should generate files for pas 2.2", func() {
			metadataBytes, err := getFileBytes("./fixtures/pas_2_2.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should generate files for mysql_v2", func() {
			metadataBytes, err := getFileBytes("./fixtures/mysql_v2.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should generate files for scs", func() {
			metadataBytes, err := getFileBytes("./fixtures/scs.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should generate files for srt", func() {
			metadataBytes, err := getFileBytes("./fixtures/srt.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, true, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should generate files for push notifications", func() {
			metadataBytes, err := getFileBytes("./fixtures/p_push_notifications.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should generate files for pivotal cloud cache", func() {
			metadataBytes, err := getFileBytes("./fixtures/cloudcache.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should generate files for rabbitmq", func() {
			metadataBytes, err := getFileBytes("./fixtures/rabbit-mq.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should generate files for rabbitmq 1.4", func() {
			metadataBytes, err := getFileBytes("./fixtures/rabbit-mq-1.4.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should generate files for redis", func() {
			metadataBytes, err := getFileBytes("./fixtures/p-redis.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should generate files for apigee", func() {
			metadataBytes, err := getFileBytes("./fixtures/apigee.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should generate files for pks", func() {
			metadataBytes, err := getFileBytes("./fixtures/pks.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
			template, err := unmarshalProduct(path.Join(tmpPath, "pivotal-container-service", "1.1.3-build.11", "product.yml"))
			Expect(err).ToNot(HaveOccurred())
			Expect(template.NetworkProperties).ToNot(BeNil())
			Expect(template.ResourceConfig).ToNot(BeNil())
		})
		It("Should generate files for nsx-t", func() {
			metadataBytes, err := getFileBytes("./fixtures/nsx-t.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
			template, err := unmarshalProduct(path.Join(tmpPath, "VMware-NSX-T", "2.2.1.9149087", "product.yml"))
			Expect(err).ToNot(HaveOccurred())
			Expect(template.NetworkProperties).Should(BeNil())
			Expect(template.ResourceConfig).Should(BeNil())
		})

		It("Should generate files for aws-services", func() {
			metadataBytes, err := getFileBytes("./fixtures/aws-services.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should generate files for a9s postgres", func() {
			metadataBytes, err := getFileBytes("./fixtures/a9s_postgres.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

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
