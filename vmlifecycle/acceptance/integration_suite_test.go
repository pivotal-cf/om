package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var (
	pathToMain string
	pathToStub string
)

var _ = SynchronizedBeforeSuite(func() []byte {
	pathToMain, err := gexec.Build("github.com/pivotal-cf/om",
		"--ldflags",
		"-X main.version=9.9.9-test")
	Expect(err).ToNot(HaveOccurred())

	pathToStub, err := gexec.Build("github.com/pivotal-cf/om/vmlifecycle/vmmanagers/stub")
	Expect(err).ToNot(HaveOccurred())

	tmpDir := filepath.Dir(pathToStub)
	omPath := tmpDir + "/om"
	err = os.Link(pathToStub, omPath)
	Expect(err).ToNot(HaveOccurred())

	gcloudPath := tmpDir + "/gcloud"
	err = os.Link(pathToStub, gcloudPath)
	Expect(err).ToNot(HaveOccurred())

	return []byte(strings.Join([]string{pathToMain, pathToStub}, "|"))
}, func(data []byte) {
	parts := strings.Split(string(data), "|")
	pathToMain = parts[0]
	pathToStub = parts[1]

	tmpDir := filepath.Dir(pathToStub)
	err := os.Setenv("PATH", tmpDir+":"+os.Getenv("PATH"))
	Expect(err).ToNot(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})

func writeFile(contents string) string {
	tempfile, err := os.CreateTemp("", "some*.yaml")
	Expect(err).ToNot(HaveOccurred())
	_, err = tempfile.WriteString(contents)
	Expect(err).ToNot(HaveOccurred())
	err = tempfile.Close()
	Expect(err).ToNot(HaveOccurred())

	return tempfile.Name()
}
