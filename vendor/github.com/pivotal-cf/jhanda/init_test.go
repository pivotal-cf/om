package jhanda_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestJhanda(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "jhanda")
}
