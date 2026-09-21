package vmmanagers_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/fatih/color"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
)

func TestVMManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VMManagers Suite")
}

var _ = BeforeSuite(func() {
	log.SetOutput(GinkgoWriter)
	pathToStub, err := gexec.Build("github.com/pivotal-cf/om/vmlifecycle/vmmanagers/stub")
	Expect(err).ToNot(HaveOccurred())

	tmpDir := filepath.Dir(pathToStub)
	os.Setenv("PATH", tmpDir+":"+os.Getenv("PATH"))

	govcPath := tmpDir + "/govc"
	gcloudPath := tmpDir + "/gcloud"
	omPath := tmpDir + "/om"
	err = os.Link(pathToStub, govcPath)
	Expect(err).ToNot(HaveOccurred())

	err = os.Link(pathToStub, omPath)
	Expect(err).ToNot(HaveOccurred())

	err = os.Link(pathToStub, gcloudPath)
	Expect(err).ToNot(HaveOccurred())

	color.NoColor = true
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func testIAASForPropertiesInExampleFile(iaas string) {
	It("has an example file the represents all the correct fields", func() {
		filename := fmt.Sprintf("../../../docs-platform-automation/docs/examples/opsman-config/%s.yml", strings.ToLower(iaas))
		exampleFile, err := os.ReadFile(filename)
		Expect(err).ToNot(HaveOccurred())

		isolateCommentedParamRegex := regexp.MustCompile(`(?m)^(\s+)# ([\w-]+: )`)
		exampleFile = isolateCommentedParamRegex.ReplaceAll(exampleFile, []byte("$1$2"))

		config := vmmanagers.OpsmanConfigFilePayload{}
		err = yaml.UnmarshalStrict(exampleFile, &config)
		Expect(err).ToNot(HaveOccurred())

		configStruct := reflect.ValueOf(config.OpsmanConfig)
		iaasPtrStruct := configStruct.FieldByName(iaas)
		iaasStruct := iaasPtrStruct.Elem()

		Expect(iaasStruct.NumField()).To(BeNumerically(">", 0))

		testPropertiesExist(iaasStruct, filename)
	})
}

func testPropertiesExist(vst reflect.Value, filename string) {
	tst := vst.Type()
	for i := 0; i < vst.NumField(); i++ {
		errorMsg := fmt.Sprintf("field %s does not exist or is an empty value in the iaas example config %q", tst.Field(i).Name, filename)
		field := vst.Field(i)
		switch field.Kind() {
		case reflect.Struct:
			testPropertiesExist(vst.Field(i), filename)
		case reflect.Bool:
			if tst.Field(i).Name != "UseUnmanagedDiskDEPRECATED" && tst.Field(i).Name != "UseInstanceProfileDEPRECATED" && tst.Field(i).Name != "Encrypted" {
				Expect(field.Bool()).ToNot(Equal(false), errorMsg)
			}
		case reflect.String:
			ignoredFields := []string{"KmsKeyId", "AvailabilityZoneId", "Affinity", "GroupName", "HostId", "Tenancy", "HostResourceGroupArn", "GroupId", "AvailabilityZone"}

			if !slices.Contains(ignoredFields[:], tst.Field(i).Name) {
				Expect(field.String()).ToNot(Equal(""), errorMsg)
			}
		case reflect.Int:
			Expect(field.Int()).ToNot(Equal(0), errorMsg)
		case reflect.Slice:
			Expect(field.Slice(0, 0)).ToNot(Equal(""), errorMsg)
		case reflect.Map:
			Expect(field.MapKeys()).ToNot(HaveLen(0), errorMsg)
		default:
			Fail(fmt.Sprintf("unexpected type: '%s' in the iaas config", field.Kind()))
		}
	}
}

func writePDFFile(contents string) string {
	tempfile, err := os.CreateTemp("", "some*.pdf")
	Expect(err).ToNot(HaveOccurred())
	_, err = tempfile.WriteString(contents)
	Expect(err).ToNot(HaveOccurred())
	err = tempfile.Close()
	Expect(err).ToNot(HaveOccurred())

	return tempfile.Name()
}
