package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/configtemplate/generator"
)

var _ = Describe("Errand Config", func() {
	When("both pre and post deployment errands are present", func() {
		It("populates the errand state with the correct pre / post deployment state", func() {
			metadata := &generator.Metadata{
				PostDeployErrands: []generator.ErrandMetadata{
					generator.ErrandMetadata{
						Name: "post-1",
					},
					generator.ErrandMetadata{
						Name: "post-2",
					},
				},
				PreDeleteErrands: []generator.ErrandMetadata{
					generator.ErrandMetadata{
						Name: "post-1",
					},
					generator.ErrandMetadata{
						Name: "pre-1",
					},
					generator.ErrandMetadata{
						Name: "pre-2",
					},
				},
			}

			errandConfig := generator.CreateErrandConfig(metadata)
			Expect(errandConfig).To(Equal(map[string]generator.Errand{
				"post-1": {
					PostDeployState: "((post-1_post_deploy_state))",
					PreDeleteState:  "((post-1_pre_delete_state))",
				},
				"post-2": {
					PostDeployState: "((post-2_post_deploy_state))",
					PreDeleteState:  "",
				},
				"pre-1": {
					PostDeployState: "",
					PreDeleteState:  "((pre-1_pre_delete_state))",
				},
				"pre-2": {
					PostDeployState: "",
					PreDeleteState:  "((pre-2_pre_delete_state))",
				},
			}))

			errandVarsConfig := generator.CreateErrandVars(metadata)
			Expect(errandVarsConfig).To(Equal(map[string]interface{}{
				"pre-1_pre_delete_state":   "default",
				"pre-2_pre_delete_state":   "default",
				"post-1_post_deploy_state": "default",
				"post-1_pre_delete_state":  "default",
				"post-2_post_deploy_state": "default",
			}))
		})
	})
})
