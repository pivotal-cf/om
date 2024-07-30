package configfetchers_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfigFetchers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ConfigFetchers Suite")
}
