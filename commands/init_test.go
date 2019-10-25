package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"

	"testing"
)

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "commands")
}

func writeTestConfigFile(contents string) string {
	file, err := ioutil.TempFile("", "config-*.yml")
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(file.Name(), []byte(contents), 0777)
	Expect(err).ToNot(HaveOccurred())
	return file.Name()
}
