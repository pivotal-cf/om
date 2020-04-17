package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/configtemplate/generator"
	"github.com/pivotal-cf/om/configtemplate/generator/fakes"
	"gopkg.in/yaml.v2"
)

var withPersistentDisk = `instance_type:
  id: ((resource-var-myjob_instance_type))
instances: ((resource-var-myjob_instances))
persistent_disk:
  size_mb: ((resource-var-myjob_persistent_disk_size))
max_in_flight: ((resource-var-myjob_max_in_flight))`

var withoutPersistentDisk = `instance_type:
  id: ((resource-var-myjob_instance_type))
instances: ((resource-var-myjob_instances))
max_in_flight: ((resource-var-myjob_max_in_flight))`

var _ = Describe("Resource Config", func() {
	Context("CreateResourceConfig", func() {
		It("Should return new resource config", func() {
			metadata := &generator.Metadata{
				JobTypes: []generator.JobType{
					{
						Name:               "job1",
						InstanceDefinition: generator.InstanceDefinition{Configurable: true, Default: 0},
					},
					{
						Name:               "job2",
						InstanceDefinition: generator.InstanceDefinition{Configurable: true, Default: 1},
					},
				},
			}
			resourceConfig := generator.CreateResourceConfig(metadata)
			Expect(resourceConfig).ToNot(BeNil())
			Expect(len(resourceConfig)).Should(Equal(2))
			Expect(resourceConfig).Should(HaveKey("job1"))
			Expect(resourceConfig["job1"].MaxInFlight).ToNot(BeNil())
			Expect(resourceConfig).Should(HaveKey("job2"))
			Expect(resourceConfig["job2"].MaxInFlight).ToNot(BeNil())
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
			Expect(resource).ToNot(BeNil())
			Expect(resource.PersistentDisk).ToNot(BeNil())
		})

		It("Should marshall to yaml without persistent disk", func() {
			jobType.HasPersistentDiskReturns(false)
			jobType.InstanceDefinitionConfigurableReturns(true)
			resource := generator.CreateResource("my-job", jobType)
			Expect(resource).ToNot(BeNil())
			Expect(resource.PersistentDisk).Should(BeNil())
		})

		It("Should marshall to yaml with persistent disk", func() {
			jobType.HasPersistentDiskReturns(true)
			jobType.InstanceDefinitionConfigurableReturns(true)
			resource := generator.CreateResource("myjob", jobType)
			Expect(resource).ToNot(BeNil())
			yml, err := yaml.Marshal(resource)
			Expect(err).ToNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(withPersistentDisk))
		})

		It("Should resource without persistent disk", func() {
			jobType.HasPersistentDiskReturns(false)
			jobType.InstanceDefinitionConfigurableReturns(true)
			resource := generator.CreateResource("myjob", jobType)
			Expect(resource).ToNot(BeNil())
			yml, err := yaml.Marshal(resource)
			Expect(err).ToNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(withoutPersistentDisk))
		})
	})

	Context("CreateResourceVars", func() {
		It("Should include instances when configurable", func() {
			metadata := &generator.Metadata{
				JobTypes: []generator.JobType{
					{
						Name:                "job1",
						InstanceDefinition:  generator.InstanceDefinition{Configurable: true, Default: 1},
						ResourceDefinitions: []generator.ResourceDefinition{{Name: "persistent_disk", Configurable: false}},
					},
				},
			}
			vars := generator.CreateResourceVars(metadata)
			Expect(vars).To(HaveKey("resource-var-job1_instances"))
		})

		It("Should not include instances when not configurable", func() {
			metadata := &generator.Metadata{
				JobTypes: []generator.JobType{
					{
						Name:                "job1",
						InstanceDefinition:  generator.InstanceDefinition{Configurable: false, Default: 1},
						ResourceDefinitions: []generator.ResourceDefinition{{Name: "persistent_disk", Configurable: false}},
					},
				},
			}
			vars := generator.CreateResourceVars(metadata)
			Expect(vars).ToNot(HaveKey("resource-var-job1_instances"))
		})

		It("Should include a disk variable with persistent disk", func() {
			metadata := &generator.Metadata{
				JobTypes: []generator.JobType{
					{
						Name:                "job1",
						InstanceDefinition:  generator.InstanceDefinition{Configurable: false, Default: 1},
						ResourceDefinitions: []generator.ResourceDefinition{{Name: "persistent_disk", Configurable: true}},
					},
				},
			}
			vars := generator.CreateResourceVars(metadata)
			Expect(vars).To(HaveKey("resource-var-job1_persistent_disk_size"))
		})

		It("Should not include a disk variable without persistent disk", func() {
			metadata := &generator.Metadata{
				JobTypes: []generator.JobType{
					{
						Name:                "job1",
						InstanceDefinition:  generator.InstanceDefinition{Configurable: false, Default: 1},
						ResourceDefinitions: []generator.ResourceDefinition{{Name: "persistent_disk", Configurable: false}},
					},
				},
			}
			vars := generator.CreateResourceVars(metadata)
			Expect(vars).ToNot(HaveKey("resource-var-job1_persistent_disk_size"))
		})

		It("Should include max_in_flight", func() {
			metadata := &generator.Metadata{
				JobTypes: []generator.JobType{
					{
						Name:                "job1",
						InstanceDefinition:  generator.InstanceDefinition{Configurable: false, Default: 1},
						ResourceDefinitions: []generator.ResourceDefinition{{Name: "persistent_disk", Configurable: false}},
					},
				},
			}
			vars := generator.CreateResourceVars(metadata)
			Expect(vars).To(HaveKey("resource-var-job1_max_in_flight"))
		})
	})
})
