package renderers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/renderers"
	"github.com/pivotal-cf/om/renderers/fakes"
)

var _ = Describe("Factory", func() {
	Describe("Create", func() {
		var (
			factory   renderers.Factory
			envGetter *fakes.EnvGetter
		)
		Context("WhenPSModulePathSet", func() {
			BeforeEach(func() {
				envGetter = &fakes.EnvGetter{}
				factory = renderers.NewFactory(envGetter)
			})
			It("creates powershell renderer", func() {
				envGetter.GetReturns("anything")
				shellType := ""
				renderer, err := factory.Create(shellType)
				Expect(err).To(BeNil())
				Expect(renderer.Type()).To(Equal(renderers.ShellTypePowershell))
			})
		})
		Context("WhenPSModulePathUnset", func() {
			BeforeEach(func() {
				envGetter := &fakes.EnvGetter{}
				factory = renderers.NewFactory(envGetter)
			})
			It("creates posix renderer", func() {
				shellType := ""
				renderer, err := factory.Create(shellType)
				Expect(err).To(BeNil())
				Expect(renderer.Type()).To(Equal(renderers.ShellTypePosix))
			})
		})
	})
})
