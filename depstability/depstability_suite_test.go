package depstability_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDepstability(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Depstability Suite")
}
