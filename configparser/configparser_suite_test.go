package configparser_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConfigparser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Configparser Suite")
}
