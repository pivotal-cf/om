package generator_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"

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
	})

	Context("Generate", func() {
		var (
			gen     *generator.Executor
			pwd, _  = ioutil.TempDir("", "")
			testGen = path.Join(pwd, "_testGen")
			tmpPath = path.Join(testGen, "templates")
		)

		AfterEach(func() {
			err := os.RemoveAll(testGen)
			Expect(err).ToNot(HaveOccurred())
		})

		It("generates files for all the fixutres", func() {
			fixtures, err := filepath.Glob("./fixtures/metadata/*.yml")
			Expect(err).ToNot(HaveOccurred())

			sort.Strings(fixtures)

			for _, fixtureFilename := range fixtures {
				metadataBytes, err := getFileBytes(fixtureFilename)
				Expect(err).ToNot(HaveOccurred())
				gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
				err = gen.Generate()
				Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("expected %s to be a valid fixture", fixtureFilename))
			}
		})

		It("Should generate files for pks", func() {
			metadataBytes, err := getFileBytes("./fixtures/metadata/pks.yml")
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
			metadataBytes, err := getFileBytes("./fixtures/metadata/nsx-t.yml")
			Expect(err).ToNot(HaveOccurred())
			gen = generator.NewExecutor(metadataBytes, tmpPath, false, true)
			err = gen.Generate()
			Expect(err).ToNot(HaveOccurred())
			template, err := unmarshalProduct(path.Join(tmpPath, "VMware-NSX-T", "2.2.1.9149087", "product.yml"))
			Expect(err).ToNot(HaveOccurred())
			Expect(template.NetworkProperties).Should(BeNil())
			Expect(template.ResourceConfig).Should(BeNil())
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
