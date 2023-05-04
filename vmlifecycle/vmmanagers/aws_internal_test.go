package vmmanagers_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers/fakes"
)

// TestAWSConfiguration this a throw away test
func TestAWSConfiguration(t *testing.T) {
	t.Run("access-key", func(t *testing.T) {
		//config := AWSConfig{
		//	AWSCredential: AWSCredential{
		//		AccessKeyId:     "some-id",
		//		SecretAccessKey: "some-key",
		//	},
		//},
		//manager := NewAWSVMManager(io.Discard, io.Discard, &OpsmanConfigFilePayload{
		//	OpsmanConfig: OpsmanConfig{
		//		AWS: &config,
		//	},
		//}, "", StateInfo{}, &fakes.AwsRunner{}, time.Duration(0))
		//
		//env := manager.addEnvVars()
		//manager.ExecuteWithInstanceProfile(nil, nil)
		//config.validateConfig()
	})
	t.Run("assume-role", func(t *testing.T) {
		runner := &fakes.AwsRunner{}
		config := vmmanagers.AWSConfig{
			SecurityGroupIds: []string{"banana"},
			AWSCredential: vmmanagers.AWSCredential{
				Region: "pluto",
				AWSInstanceProfile: vmmanagers.AWSInstanceProfile{
					AssumeRole: "dice",
				},
			},
		}
		manager := vmmanagers.NewAWSVMManager(io.Discard, io.Discard, &vmmanagers.OpsmanConfigFilePayload{
			OpsmanConfig: vmmanagers.OpsmanConfig{
				AWS: &config,
			},
		}, "", vmmanagers.StateInfo{}, runner, time.Duration(0))

		please := gomega.NewWithT(t)
		please.Expect(config.ValidateConfig()).NotTo(gomega.HaveOccurred())

		var configFileContents string
		runner.ExecuteWithEnvVarsStub = func(env []string, _ []interface{}) (*bytes.Buffer, *bytes.Buffer, error) {
			environ := makeMapEnv(env)

			configFilePath, found := environ["AWS_CONFIG_FILE"]
			please.Expect(found).To(gomega.BeTrue())
			contents, err := os.ReadFile(configFilePath)
			please.Expect(err).NotTo(gomega.HaveOccurred())
			configFileContents = string(contents)

			return nil, nil, nil
		}

		_, _, err := manager.ExecuteWithInstanceProfile(manager.AddEnvVars(), nil)
		please.Expect(err).NotTo(gomega.HaveOccurred())

		env, args := runner.ExecuteWithEnvVarsArgsForCall(0)

		fmt.Println(env, args, configFileContents)
	})
	//t.Run("instance-profile", func(t *testing.T) {
	//	config := AWSConfig{
	//		AWSCredential: AWSCredential{
	//			AWSInstanceProfile: AWSInstanceProfile{
	//				AssumeRole:             "some-role",
	//				IAMInstanceProfileName: "om-cli-instance-profile",
	//			},
	//		},
	//		IAMInstanceProfileName: "ops-manager-instance-profile",
	//	}
	//
	//	authType := config.authenticationType()
	//	please := gomega.NewWithT(t)
	//	please.Expect(authType).To(Equal("instance-profile"))
	//})
	//t.Run("assume-role", func(t *testing.T) {
	//	config := AWSConfig{}
	//
	//	authType := config.authenticationType()
	//	please := NewWithT(t)
	//	please.Expect(authType).To(Equal("assume-role"))
	//})

}

func makeMapEnv(values []string) map[string]string {
	m := make(map[string]string, len(values))
	for _, val := range values {
		segments := strings.Split(val, "=")
		m[segments[0]] = segments[1]
	}
	return m
}
