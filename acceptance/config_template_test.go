package acceptance

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = FDescribe("config-template command", func() {
	var (
		product *os.File
	)

	BeforeEach(func() {
		var err error
		product, err = ioutil.TempFile("", "config-template.pivotal")
		Expect(err).NotTo(HaveOccurred())

		zipWriter := zip.NewWriter(product)

		metadata, err := zipWriter.Create("./metadata/metadata.yml")
		Expect(err).NotTo(HaveOccurred())

		fixtureMetadata, err := os.Open("fixtures/metadata.yml")
		Expect(err).NotTo(HaveOccurred())

		_, err = io.Copy(metadata, fixtureMetadata)
		Expect(err).NotTo(HaveOccurred())

		Expect(zipWriter.Close()).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.Remove(product.Name())).To(Succeed())
	})

	It("outputs a configuration template based on the given product file", func() {
		command := exec.Command(pathToMain,
			"config-template",
			"--product", product.Name(),
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "10s").Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(MatchYAML(`---
product-properties:
  .properties.with_default:
    value: some-default
  .properties.without_default:
    value: null
  .some-instance-group.with_default:
    value: some-default
  .some-instance-group.without_default:
    value: null
  .properties.with_named_manifest:
    value: enable
  .properties.with_named_manifest_without_default:
    value: null
`))
	})
})
