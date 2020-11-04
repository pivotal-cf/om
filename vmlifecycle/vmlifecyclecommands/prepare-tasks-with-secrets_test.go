package vmlifecyclecommands_test

import (
	"errors"
	"fmt"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/vmlifecycle/vmlifecyclecommands"
	"github.com/pivotal-cf/om/vmlifecycle/vmlifecyclecommands/fakes"
	"io"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PrepareTasksWithSecrets", func() {
	var (
		fakeService          *fakes.TaskModifierService
		taskDir              string
		configDir            string
		configDir2           string
		varsDir              string
		varsDir2             string
		outWriter, errWriter *gbytes.Buffer
	)

	BeforeEach(func() {
		fakeService = &fakes.TaskModifierService{}
		outWriter = gbytes.NewBuffer()
		errWriter = gbytes.NewBuffer()

		var err error
		taskDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		configDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		configDir2, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		varsDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		varsDir2, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	It("calls task modifier with the appropriate arguments", func() {
		fakeService.ModifyTasksWithSecretsStub = func(writer io.Writer, tasksDir string, configDirs []string, varsDirs []string) error {
			return nil
		}

		command := secretsModifierCommand(outWriter, errWriter, fakeService)
		command.TaskDir = taskDir
		command.ConfigDir = []string{configDir, configDir2}
		command.VarDir = []string{varsDir, varsDir2}

		err := command.Execute(nil)
		Expect(err).ToNot(HaveOccurred())

		Expect(fakeService.ModifyTasksWithSecretsCallCount()).To(Equal(1))
		_, tasksDir, configPaths, varsPaths := fakeService.ModifyTasksWithSecretsArgsForCall(0)
		Expect(tasksDir).To(Equal(taskDir))
		Expect(configPaths).To(Equal([]string{configDir, configDir2}))
		Expect(varsPaths).To(Equal([]string{varsDir, varsDir2}))

		Expect(outWriter).To(gbytes.Say(fmt.Sprintf("successfully added secrets to provided tasks")))
	})

	It("returns an error if the service fails", func() {
		fakeService.ModifyTasksWithSecretsStub = func(writer io.Writer, tasksDir string, configDirs []string, varsDirs []string) error {
			return errors.New("some error modifying")
		}

		configFile1 := filepath.Join(configDir, "config-with-secrets.yml")
		configFile2 := filepath.Join(configDir, "config-without-secrets.yml")

		writeSpecifiedFile(filepath.Join(taskDir, "with-params.yml"), `params: {}`)
		writeSpecifiedFile(filepath.Join(taskDir, "without-params.yml"), `not-params: {}`)
		writeSpecifiedFile(configFile1, `some: ((secret))`)
		writeSpecifiedFile(configFile2, `another: non-secret`)

		command := secretsModifierCommand(outWriter, errWriter, fakeService)
		command.TaskDir = taskDir
		command.ConfigDir = []string{configDir}

		err := command.Execute(nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("could not modify tasks: some error modifying"))
	})
})

func secretsModifierCommand(
	stdWriter,
	errWriter io.Writer,
	fakeService *fakes.TaskModifierService,
) vmlifecyclecommands.PrepareTasksWithSecrets {
	command := vmlifecyclecommands.NewSecretsModifierCommand(
		stdWriter,
		errWriter,
		fakeService,
	)

	return command
}
