package formcontent_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFormcontent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "formcontent")
}
