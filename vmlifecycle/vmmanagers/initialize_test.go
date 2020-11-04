package vmmanagers_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
)

var _ = Describe("select the correct vmmanager instance given the correct config", func() {
	Context("valid vsphere config", func() {
		It("returns the vsphere vmmanager instance", func() {
			configContent := &vmmanagers.OpsmanConfigFilePayload{
				OpsmanConfig: vmmanagers.OpsmanConfig{GCP: nil, Vsphere: &vmmanagers.VsphereConfig{}, AWS: nil, Azure: nil, Openstack: nil},
			}
			create, err := vmmanagers.NewCreateVMManager(configContent, "", vmmanagers.StateInfo{}, gbytes.NewBuffer(), gbytes.NewBuffer())
			Expect(err).ToNot(HaveOccurred())
			Expect(reflect.TypeOf(create)).To(Equal(reflect.TypeOf(&vmmanagers.VsphereVMManager{})))

			delete, err := vmmanagers.NewDeleteVMManager(configContent, "", vmmanagers.StateInfo{}, gbytes.NewBuffer(), gbytes.NewBuffer())
			Expect(err).ToNot(HaveOccurred())
			Expect(reflect.TypeOf(delete)).To(Equal(reflect.TypeOf(&vmmanagers.VsphereVMManager{})))
		})
	})

	Context("valid gcp config", func() {
		It("returns the gcp vmmanager instance", func() {
			configContent := &vmmanagers.OpsmanConfigFilePayload{
				OpsmanConfig: vmmanagers.OpsmanConfig{GCP: &vmmanagers.GCPConfig{}, Vsphere: nil, AWS: nil, Azure: nil, Openstack: nil},
			}
			create, err := vmmanagers.NewCreateVMManager(configContent, "", vmmanagers.StateInfo{}, gbytes.NewBuffer(), gbytes.NewBuffer())
			Expect(err).ToNot(HaveOccurred())
			Expect(reflect.TypeOf(create)).To(Equal(reflect.TypeOf(&vmmanagers.GCPVMManager{})))

			delete, err := vmmanagers.NewDeleteVMManager(configContent, "", vmmanagers.StateInfo{}, gbytes.NewBuffer(), gbytes.NewBuffer())
			Expect(err).ToNot(HaveOccurred())
			Expect(reflect.TypeOf(delete)).To(Equal(reflect.TypeOf(&vmmanagers.GCPVMManager{})))
		})
	})

	Context("valid aws config", func() {
		It("returns the aws vmmanager instance", func() {
			configContent := &vmmanagers.OpsmanConfigFilePayload{
				OpsmanConfig: vmmanagers.OpsmanConfig{GCP: nil, Vsphere: nil, AWS: &vmmanagers.AWSConfig{}, Azure: nil, Openstack: nil},
			}
			create, err := vmmanagers.NewCreateVMManager(configContent, "", vmmanagers.StateInfo{}, gbytes.NewBuffer(), gbytes.NewBuffer())
			Expect(err).ToNot(HaveOccurred())
			Expect(reflect.TypeOf(create)).To(Equal(reflect.TypeOf(&vmmanagers.AWSVMManager{})))

			delete, err := vmmanagers.NewDeleteVMManager(configContent, "", vmmanagers.StateInfo{}, gbytes.NewBuffer(), gbytes.NewBuffer())
			Expect(err).ToNot(HaveOccurred())
			Expect(reflect.TypeOf(delete)).To(Equal(reflect.TypeOf(&vmmanagers.AWSVMManager{})))
		})
	})

	Context("valid azure config", func() {
		It("returns the azure vmmanager instance", func() {
			configContent := &vmmanagers.OpsmanConfigFilePayload{
				OpsmanConfig: vmmanagers.OpsmanConfig{GCP: nil, Vsphere: nil, AWS: nil, Azure: &vmmanagers.AzureConfig{}, Openstack: nil},
			}
			create, err := vmmanagers.NewCreateVMManager(configContent, "", vmmanagers.StateInfo{}, gbytes.NewBuffer(), gbytes.NewBuffer())
			Expect(err).ToNot(HaveOccurred())
			Expect(reflect.TypeOf(create)).To(Equal(reflect.TypeOf(&vmmanagers.AzureVMManager{})))

			delete, err := vmmanagers.NewDeleteVMManager(configContent, "", vmmanagers.StateInfo{}, gbytes.NewBuffer(), gbytes.NewBuffer())
			Expect(err).ToNot(HaveOccurred())
			Expect(reflect.TypeOf(delete)).To(Equal(reflect.TypeOf(&vmmanagers.AzureVMManager{})))
		})
	})

	Context("valid openstack config", func() {
		It("returns the openstack vmmanager instance", func() {
			configContent := &vmmanagers.OpsmanConfigFilePayload{
				OpsmanConfig: vmmanagers.OpsmanConfig{GCP: nil, Vsphere: nil, AWS: nil, Azure: nil, Openstack: &vmmanagers.OpenstackConfig{}},
			}
			create, err := vmmanagers.NewCreateVMManager(configContent, "", vmmanagers.StateInfo{}, gbytes.NewBuffer(), gbytes.NewBuffer())
			Expect(err).ToNot(HaveOccurred())
			Expect(reflect.TypeOf(create)).To(Equal(reflect.TypeOf(&vmmanagers.OpenstackVMManager{})))

			delete, err := vmmanagers.NewDeleteVMManager(configContent, "", vmmanagers.StateInfo{}, gbytes.NewBuffer(), gbytes.NewBuffer())
			Expect(err).ToNot(HaveOccurred())
			Expect(reflect.TypeOf(delete)).To(Equal(reflect.TypeOf(&vmmanagers.OpenstackVMManager{})))
		})
	})

	Describe("failure cases", func() {
		When("there are multiple iaas", func() {
			It("returns an error", func() {
				configContent := &vmmanagers.OpsmanConfigFilePayload{
					OpsmanConfig: struct {
						Vsphere   *vmmanagers.VsphereConfig   `yaml:"vsphere,omitempty"`
						GCP       *vmmanagers.GCPConfig       `yaml:"gcp,omitempty"`
						AWS       *vmmanagers.AWSConfig       `yaml:"aws,omitempty"`
						Azure     *vmmanagers.AzureConfig     `yaml:"azure,omitempty"`
						Openstack *vmmanagers.OpenstackConfig `yaml:"openstack,omitempty"`
						Unknown   map[string]interface{}      `yaml:",inline"`
					}{GCP: nil, Vsphere: nil, AWS: nil, Azure: &vmmanagers.AzureConfig{}, Openstack: &vmmanagers.OpenstackConfig{}},
				}
				_, err := vmmanagers.NewCreateVMManager(configContent, "", vmmanagers.StateInfo{}, gbytes.NewBuffer(), gbytes.NewBuffer())
				Expect(err).To(MatchError("more than one iaas matched, only one in config allowed"))

				_, err = vmmanagers.NewDeleteVMManager(configContent, "", vmmanagers.StateInfo{}, gbytes.NewBuffer(), gbytes.NewBuffer())
				Expect(err).To(MatchError("more than one iaas matched, only one in config allowed"))
			})
		})
	})
})
