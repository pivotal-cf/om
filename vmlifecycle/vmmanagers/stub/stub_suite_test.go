package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStub(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stub Suite")
}
