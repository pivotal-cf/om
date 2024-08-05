package vmlifecyclecommands_test

import (
	"os"
	"path/filepath"
	"testing"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
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

	err = ioutil.WriteFile(filename, []byte(contents), 0644)
	Expect(err).NotTo(HaveOccurred())
}

func writeFile(contents string) string {
	tempfile, err := ioutil.TempFile("", "2.2-build.296.yml")
	Expect(err).ToNot(HaveOccurred())
	_, err = tempfile.WriteString(contents)
	Expect(err).ToNot(HaveOccurred())
	err = tempfile.Close()
	Expect(err).ToNot(HaveOccurred())

	return tempfile.Name()
}

func readFile(filename string) string {
	contents, err := ioutil.ReadFile(filename)
	Expect(err).ToNot(HaveOccurred())
	return string(contents)
}
