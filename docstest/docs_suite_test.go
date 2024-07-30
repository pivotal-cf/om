package docs_test

import (
	"testing"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
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
	Expect(err).ToNot(HaveOccurred())

	return []byte(omPath)
}, func(data []byte) {
	pathToMain = string(data)
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})
