package presenters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPresenters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "presenters")
}
