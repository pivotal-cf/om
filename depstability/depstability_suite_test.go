package depstability_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDepstability(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Depstability Suite")
}
