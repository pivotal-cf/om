package cargo_test

import (
	"io/ioutil"
	"os"

	"github.com/pivotal-cf/kiln/internal/cargo"
	"github.com/pivotal-cf/kiln/internal/cargo/opsman"
	"github.com/pivotal-cf/kiln/proofing"
	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("Generator", func() {
	var generator cargo.Generator

	BeforeEach(func() {
		generator = cargo.NewGenerator()
	})

	Describe("Execute", func() {
		It("generates a well-formed manifest", func() {
			f, err := os.Open("fixtures/metadata.yml")
			defer f.Close()
			Expect(err).NotTo(HaveOccurred())

			template, err := proofing.Parse(f)
			Expect(err).NotTo(HaveOccurred())

			manifest := generator.Execute(template, cargo.OpsManagerConfig{
				DeploymentName: "some-product-name",
				AvailabilityZones: []string{
					"some-az-1",
					"some-az-2",
				},
				Stemcells: []opsman.Stemcell{
					{
						Name:    "some-stemcell-name",
						Version: "some-stemcell-version",
						OS:      "some-stemcell-os",
					},
					{
						Name:    "other-stemcell-name",
						Version: "other-stemcell-version",
						OS:      "other-stemcell-os",
					},
				},
				ResourceConfigs: []opsman.ResourceConfig{
					{
						Name:      "some-job-type-name",
						Instances: opsman.ResourceConfigInstances{Value: 1},
					},
					{
						Name:      "other-job-type-name",
						Instances: opsman.ResourceConfigInstances{Value: -1}, // NOTE: negative value indicates "automatic"
					},
				},
			})

			actualManifest, err := yaml.Marshal(manifest)
			Expect(err).NotTo(HaveOccurred())

			expectedManifest, err := ioutil.ReadFile("fixtures/manifest.yml")
			Expect(err).NotTo(HaveOccurred())

			Expect(actualManifest).To(HelpfullyMatchYAML(string(expectedManifest)))
		})
	})
})
