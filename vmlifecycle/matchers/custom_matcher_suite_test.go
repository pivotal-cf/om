package matchers_test

import (
	"testing"

	"github.com/pivotal-cf/om/vmlifecycle/matchers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCustomMatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Matchers Suite")
}

var _ = Describe("Order Consist Of", func() {
	It("fails when the slices don't match length", func() {
		Expect([]string{"a", "b", "c"}).ToNot(matchers.OrderedConsistOf("a", "b"))
	})

	It("fails when the array don't match length", func() {
		Expect([3]string{"a", "b", "c"}).ToNot(matchers.OrderedConsistOf("a", "b"))
	})

	Context("with anything else", func() {
		It("should error", func() {
			failures := InterceptGomegaFailures(func() {
				Expect("foo").Should(matchers.OrderedConsistOf("f", "o", "o"))
			})

			Expect(failures).Should(HaveLen(1))
		})
	})

	When("the slices match length", func() {
		It("defaults an equals matcher for expected elements", func() {
			Expect([]string{"a", "b", "c"}).To(matchers.OrderedConsistOf("a", "b", "c"))
			Expect([3]string{"a", "b", "c"}).To(matchers.OrderedConsistOf("a", "b", "c"))
		})
		When("passed matchers", func() {
			It("should pass if matchers pass", func() {
				Expect([]string{"a", "aabaa", "c"}).To(matchers.OrderedConsistOf(
					Equal("a"),
					MatchRegexp("b"),
					Not(Equal("d")),
				))
				Expect([3]string{"a", "aabaa", "c"}).To(matchers.OrderedConsistOf(
					Equal("a"),
					MatchRegexp("b"),
					Not(Equal("d")),
				))
			})

			It("depends on the order of the matchers", func() {
				Expect([]string{"a", "aabaa", "c"}).ToNot(matchers.OrderedConsistOf(
					MatchRegexp("b"),
					Equal("a"),
					Not(Equal("d")),
				))
				Expect([3]string{"a", "aabaa", "c"}).ToNot(matchers.OrderedConsistOf(
					MatchRegexp("b"),
					Equal("a"),
					Not(Equal("d")),
				))
			})

			When("a matcher errors", func() {
				It("should not pass", func() {
					Expect([]string{"foo", "bar", "baz"}).ShouldNot(matchers.OrderedConsistOf(BeFalse(), "foo", "bar"))
					Expect([3]string{"foo", "bar", "baz"}).ShouldNot(matchers.OrderedConsistOf(BeFalse(), "foo", "bar"))
					Expect([]interface{}{"foo", "bar", false}).ShouldNot(matchers.OrderedConsistOf(BeFalse(), ContainSubstring("foo"), "bar"))
				})
			})
		})

		When("passed exactly one argument, and that argument is a slice", func() {
			It("should match against the elements of that argument", func() {
				Expect([]string{"foo", "bar", "baz"}).Should(matchers.OrderedConsistOf([]string{"foo", "bar", "baz"}))
				Expect([3]string{"foo", "bar", "baz"}).Should(matchers.OrderedConsistOf([]string{"foo", "bar", "baz"}))
			})
		})
	})
})
