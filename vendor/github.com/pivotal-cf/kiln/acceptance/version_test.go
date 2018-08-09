package acceptance

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("version command", func() {
	var version string

	BeforeEach(func() {
		version = fmt.Sprintf("v0.0.0-dev.%d", time.Now().Unix())

		var err error
		pathToMain, err = gexec.Build("github.com/pivotal-cf/kiln",
			"--ldflags", fmt.Sprintf("-X main.version=%s", version))
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when given the version command", func() {
		It("prints the version number", func() {
			command := exec.Command(pathToMain, "version")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(fmt.Sprintf("kiln version %s", version)))
		})
	})

	Context("when given the -v short flag", func() {
		It("returns the binary version", func() {
			command := exec.Command(pathToMain, "-v")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(fmt.Sprintf("kiln version %s", version)))
		})
	})

	Context("when given the --version long flag", func() {
		It("returns the binary version", func() {
			command := exec.Command(pathToMain, "--version")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(fmt.Sprintf("kiln version %s", version)))
		})
	})
})
