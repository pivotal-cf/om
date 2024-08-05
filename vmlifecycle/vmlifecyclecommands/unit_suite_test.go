package vmlifecyclecommands_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCreateVM(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VM Lifecycle Suite")
}

func writeSpecifiedFile(filename string, contents string) {
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, 0777)
	Expect(err).NotTo(HaveOccurred())

	err = os.WriteFile(filename, []byte(contents), 0644)
	Expect(err).NotTo(HaveOccurred())
}

func writeFile(contents string) string {
	tempfile, err := os.CreateTemp("", "2.2-build.296.yml")
	Expect(err).ToNot(HaveOccurred())
	_, err = tempfile.WriteString(contents)
	Expect(err).ToNot(HaveOccurred())
	err = tempfile.Close()
	Expect(err).ToNot(HaveOccurred())

	return tempfile.Name()
}

func readFile(filename string) string {
	contents, err := os.ReadFile(filename)
	Expect(err).ToNot(HaveOccurred())
	return string(contents)
}
