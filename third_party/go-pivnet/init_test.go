package pivnet_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	apiPrefix   = "/api/v2"
	productSlug = "some-product-name"
)

func TestPivnetClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pivnet Client Suite")
}
