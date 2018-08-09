package cargo_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCargo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/cargo")
}
