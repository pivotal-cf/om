package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/tile-config-generator/generator"
	"github.com/pivotalservices/tile-config-generator/generator/fakes"
	"gopkg.in/yaml.v2"
)

var withPersistentDisk = `instance_type:
  id: ((myjob_instance_type))
instances: ((myjob_instances))
persistent_disk:
  size_mb: ((myjob_persistent_disk_size))`

var withoutPersistentDisk = `instance_type:
  id: ((myjob_instance_type))
instances: ((myjob_instances))`

var _ = Describe("Resource Config", func() {
	Context("CreateResourceConfig", func() {
		It("Should return new resource config", func() {
			metadata := &generator.Metadata{
				JobTypes: []generator.JobType{
					generator.JobType{
						Name:               "job1",
						InstanceDefinition: generator.InstanceDefinition{Configurable: true, Default: 0},
					},
					generator.JobType{
						Name:               "job2",
						InstanceDefinition: generator.InstanceDefinition{Configurable: true, Default: 1},
					},
				},
			}
			resourceConfig := generator.CreateResourceConfig(metadata)
			Expect(resourceConfig).ShouldNot(BeNil())
			Expect(len(resourceConfig)).Should(Equal(2))
			Expect(resourceConfig).Should(HaveKey("job1"))
			Expect(resourceConfig).Should(HaveKey("job2"))
		})
	})
	Context("CreateResource", func() {
		var (
			jobType *fakes.FakeJobType
		)
		BeforeEach(func() {
			jobType = new(fakes.FakeJobType)
		})
		It("Should resource with persistent disk", func() {
			jobType.HasPersistentDiskReturns(true)
			jobType.InstanceDefinitionConfigurableReturns(true)
			resource := generator.CreateResource("my-job", jobType)
			Expect(resource).ShouldNot(BeNil())
			Expect(resource.PersistentDisk).ShouldNot(BeNil())
		})

		It("Should marshall to yaml without persistent disk", func() {
			jobType.HasPersistentDiskReturns(false)
			jobType.InstanceDefinitionConfigurableReturns(true)
			resource := generator.CreateResource("my-job", jobType)
			Expect(resource).ShouldNot(BeNil())
			Expect(resource.PersistentDisk).Should(BeNil())
		})

		It("Should marshall to yaml with persistent disk", func() {
			jobType.HasPersistentDiskReturns(true)
			jobType.InstanceDefinitionConfigurableReturns(true)
			resource := generator.CreateResource("myjob", jobType)
			Expect(resource).ShouldNot(BeNil())
			yml, err := yaml.Marshal(resource)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(withPersistentDisk))
		})

		It("Should resource without persistent disk", func() {
			jobType.HasPersistentDiskReturns(false)
			jobType.InstanceDefinitionConfigurableReturns(true)
			resource := generator.CreateResource("myjob", jobType)
			Expect(resource).ShouldNot(BeNil())
			yml, err := yaml.Marshal(resource)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(withoutPersistentDisk))
		})

	})
})
