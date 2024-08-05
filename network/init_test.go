package network_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNetwork(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "network")
}

func writeFile(contents string) string {
	file, err := os.CreateTemp("", "")
	Expect(err).ToNot(HaveOccurred())

	err = os.WriteFile(file.Name(), []byte(contents), 0777)
	Expect(err).ToNot(HaveOccurred())
	return file.Name()
}
