package docs_test

import (
	"fmt"
	"github.com/onsi/gomega/gexec"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDocstest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Docstest Suite")
}

var pathToMain string

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	omPath, err := gexec.Build("../main.go", "-ldflags", "-X main.applySleepDurationString=1ms -X github.com/pivotal-cf/om/commands.pivnetHost=http://example.com")
	Expect(err).NotTo(HaveOccurred())

	return []byte(omPath)
}, func(data []byte) {
	pathToMain = string(data)
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})

func runCommand(args ...string) {
	fmt.Fprintf(GinkgoWriter, "cmd: %s\n", args)
	command := exec.Command(args[0], args[1:]...)
	configure, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(configure, "10s").Should(gexec.Exit(0))
}