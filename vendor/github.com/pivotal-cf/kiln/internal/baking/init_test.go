package baking_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBaking(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/baking")
}
