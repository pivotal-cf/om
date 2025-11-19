package md5sum_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMD5(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MD5 Suite")
}
