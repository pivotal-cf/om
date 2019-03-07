package acceptance

import (
	"github.com/onsi/gomega/gbytes"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = PDescribe("config-template command", func() {

	BeforeEach(func() {
		var err error
		var fakePivnetMetadataResponse []byte

		fixtureMetadata, err := os.Open("fixtures/metadata.yml")
		defer fixtureMetadata.Close()

		Expect(err).NotTo(HaveOccurred())

		_, err = fixtureMetadata.Read(fakePivnetMetadataResponse)
		Expect(err).NotTo(HaveOccurred())
	})

	It("writes a config template subdir for the product in the output directory", func() {
		command := exec.Command(pathToMain,
			"config-template",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, "10s").Should(gexec.Exit(0))

		Expect(string(session.Out.Contents())).To(gbytes.Say(`I wrote you some stuff yo`))
	})
})
