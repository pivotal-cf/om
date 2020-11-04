package extractopsmansemver_test

import (
	"fmt"
	"github.com/onsi/ginkgo/extensions/table"
	"github.com/pivotal-cf/om/vmlifecycle/extractopsmansemver"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestExtractOpsmanSemver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ExtractOpsmanSemver Suite")
}

var _ = Describe("Do", func() {
	for _, fileNameFixture := range []string{"OpsManager%sonGCP.yml", "[ops-manager,2.2.3]ops-manager-vsphere-%s.ova"} {
		table.DescribeTable("extracts the version number from the file "+fileNameFixture, func(versionInFilename string, expectedVersion string) {
			filename := fmt.Sprintf(fileNameFixture, versionInFilename)
			version, err := extractopsmansemver.Do(filename)
			Expect(err).ToNot(HaveOccurred())

			Expect(version.String()).To(Equal(expectedVersion))
		},
			table.Entry("semver numbers via patch", "2.5.3", "2.5.3"),
			table.Entry("semver numbers that have build numbers", "2.5.0-build.0", "2.5.0-build.0"),
			table.Entry("build numbers via patch", "2.4-build.193", "2.4.193"),
		)
	}
})
