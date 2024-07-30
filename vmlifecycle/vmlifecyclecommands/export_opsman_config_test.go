package vmlifecyclecommands_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-cf/om/vmlifecycle/configfetchers"
	"github.com/pivotal-cf/om/vmlifecycle/configfetchers/fakes"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/pivotal-cf/om/vmlifecycle/vmlifecyclecommands"
)

var _ = Describe("Export Opsman Config", func() {
	var (
		fakeService          = &fakes.OpsmanConfigFetcherService{}
		outWriter, errWriter *gbytes.Buffer
	)

	BeforeEach(func() {
		outWriter = gbytes.NewBuffer()
		errWriter = gbytes.NewBuffer()
	})

	When("the state file exists", func() {
		It("creates an opsman.yml file", func() {
			fakeService.FetchConfigStub = func() (*vmmanagers.OpsmanConfigFilePayload, error) {
				return &vmmanagers.OpsmanConfigFilePayload{}, nil
			}

			stateFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			configFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			command := exportOpsmanConfigCommand(outWriter, errWriter, fakeService, "")
			command.StateFile = stateFile.Name()
			command.ConfigFile = configFile.Name()

			err = command.Execute(nil)
			Expect(err).ToNot(HaveOccurred())

			config, err := ioutil.ReadFile(configFile.Name())
			Expect(err).ToNot(HaveOccurred())

			Expect(outWriter).To(gbytes.Say(fmt.Sprintf("successfully wrote the Ops Manager config file to: %s", configFile.Name())))

			Expect(string(config)).To(ContainSubstring("opsman-configuration"))
		})
	})

	When("the state file does not exist", func() {
		It("returns an error", func() {
			command := exportOpsmanConfigCommand(outWriter, errWriter, fakeService, "")
			command.StateFile = "not-exist.yml"

			err := command.Execute(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not read state file"))
		})
	})

	When("the state file is invalid", func() {
		It("returns an error", func() {
			stateFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			_, err = stateFile.Write([]byte("invalid-state-file-content"))
			Expect(err).ToNot(HaveOccurred())

			command := exportOpsmanConfigCommand(outWriter, errWriter, fakeService, "")
			command.StateFile = stateFile.Name()

			err = command.Execute(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("could not load state file (%s)", stateFile.Name())))
		})
	})

	When("the init service fails to initialize", func() {
		It("returns an error", func() {
			stateFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			configFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			command := exportOpsmanConfigCommand(outWriter, errWriter, fakeService, "failed to init")
			command.StateFile = stateFile.Name()
			command.ConfigFile = configFile.Name()

			err = command.Execute(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to init"))
		})
	})

	When("configFetcherService fails to get the config", func() {
		It("returns an error", func() {
			fakeService.FetchConfigStub = func() (*vmmanagers.OpsmanConfigFilePayload, error) {
				return &vmmanagers.OpsmanConfigFilePayload{}, errors.New("fetch failed")
			}

			stateFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			configFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			command := exportOpsmanConfigCommand(outWriter, errWriter, fakeService, "")
			command.StateFile = stateFile.Name()
			command.ConfigFile = configFile.Name()

			err = command.Execute(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fetch failed"))
		})
	})
})

func exportOpsmanConfigCommand(
	stdWriter,
	errWriter io.Writer,
	fakeService *fakes.OpsmanConfigFetcherService,
	serviceErrorMessage string,
) ExportOpsmanConfig {
	initFunc := func(_ *vmmanagers.StateInfo, _ *configfetchers.Credentials) (configfetchers.OpsmanConfigFetcherService, error) {
		if serviceErrorMessage != "" {
			return nil, errors.New(serviceErrorMessage)
		}

		return fakeService, nil
	}

	command := NewExportOpsmanConfigCommand(
		stdWriter,
		errWriter,
		initFunc,
	)

	return command
}
