package presenters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPresenters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "presenters")
}
