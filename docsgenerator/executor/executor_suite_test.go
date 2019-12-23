package executor_test

import (
	"github.com/onsi/gomega/gexec"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var pathToStub string

func TestExecutor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Executor Suite")
}

var _ = BeforeSuite(func() {
	var err error
	pathToStub, err = gexec.Build("github.com/pivotal-cf/om/stub")
	Expect(err).ToNot(HaveOccurred())

	tmpDir := filepath.Dir(pathToStub)
	os.Setenv("PATH", tmpDir+":"+os.Getenv("PATH"))

	omPath := tmpDir + "/om"
	err = os.Link(pathToStub, omPath)
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
