package commands_test

import (
	"github.com/fatih/color"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"os"
	"regexp"
	"strings"
	"testing"
)

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "commands")
}

func writeTestConfigFile(contents string) string {
	file, err := os.CreateTemp("", "config-*.yml")
	Expect(err).ToNot(HaveOccurred())

	err = os.WriteFile(file.Name(), []byte(contents), 0777)
	Expect(err).ToNot(HaveOccurred())
	return file.Name()
}

var _ = BeforeSuite(func() {
	//enable color for this suite, so that colors are tested even in parallel
	//(the color library detects non-tty terminals,
	//which ginkgo uses when running in parallel,
	//so we have to override it)
	color.NoColor = false

	// Clear any OM env vars so as to not pollute the tests
	re := regexp.MustCompile(`OM_*`)
	for _, pair := range os.Environ() {
		split := strings.Split(pair, "=")
		if re.MatchString(split[0]) {
			Expect(os.Unsetenv(split[0])).NotTo(HaveOccurred())
		}
	}
})
