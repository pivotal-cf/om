package taskmodifier_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestModify(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Modify Suite")
}

func writeFile(filename string, contents string) {
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, 0777)
	Expect(err).NotTo(HaveOccurred())

	err = os.WriteFile(filename, []byte(contents), 0644)
	Expect(err).NotTo(HaveOccurred())
}

func readFile(filename string) string {
	contents, err := os.ReadFile(filename)
	Expect(err).NotTo(HaveOccurred())

	return string(contents)
}
