package acceptance

import (
	"fmt"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	pathToMain string
	minio      *gexec.Session
)

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	omPath, err := gexec.Build("../main.go", "-ldflags", "-X main.applySleepDurationString=1ms -X github.com/pivotal-cf/om/commands.pivnetHost=http://example.com")
	Expect(err).NotTo(HaveOccurred())

	minioPath, _ := exec.LookPath("minio")
	if minioPath != "" {
		dataDir, err := ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())
		command := exec.Command("minio", "server", "--config-dir", dataDir, "--address", ":9001", dataDir)
		command.Env = []string{
			"MINIO_ACCESS_KEY=minio",
			"MINIO_SECRET_KEY=password",
			"MINIO_BROWSER=off",
			"TERM=xterm-256color",
		}
		minio, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(minio.Out, "10s").Should(gbytes.Say("Endpoint:"))
		runCommand("mc", "--debug", "config", "host", "add", "testing", "http://127.0.0.1:9001", "minio", "password")
	}
	return []byte(omPath)
}, func(data []byte) {
	pathToMain = string(data)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	if minio != nil {
		minio.Kill()
	}
	gexec.CleanupBuildArtifacts()
})

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "acceptance")
}

func runCommand(args ...string) {
	fmt.Fprintf(GinkgoWriter, "cmd: %s\n", args)
	command := exec.Command(args[0], args[1:]...)
	configure, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(configure, "10s").Should(gexec.Exit(0))
}


