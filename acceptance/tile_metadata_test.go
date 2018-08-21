package acceptance

import (
	"archive/zip"
	"io/ioutil"
	"os"

	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("tile-metadata command", func() {
	var (
		productFile *os.File
		err         error
	)

	BeforeEach(func() {
		productFile, err = ioutil.TempFile("", "fake-tile")
		z := zip.NewWriter(productFile)

		f, err := z.Create("metadata/fake-tile.yml")
		Expect(err).NotTo(HaveOccurred())

		_, err = f.Write([]byte(`
name: fake-tile
product_version: 1.2.3
`))
		Expect(err).NotTo(HaveOccurred())

		Expect(z.Close()).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(productFile.Name())).To(Succeed())
	})

	It("shows product name from tile metadata file", func() {
		command := exec.Command(pathToMain,
			"tile-metadata",
			"-p", productFile.Name(),
			"--product-name",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out.Contents()).To(ContainSubstring("fake-tile"))
	})

	It("shows product version from tile metadata file", func() {
		command := exec.Command(pathToMain,
			"tile-metadata",
			"-p", productFile.Name(),
			"--product-version",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(session.Out.Contents()).To(ContainSubstring("1.2.3"))
	})
})
