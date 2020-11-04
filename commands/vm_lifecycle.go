package commands

import (
	"github.com/pivotal-cf/om/vmlifecycle/configfetchers"
	"github.com/pivotal-cf/om/vmlifecycle/runner"
	"github.com/pivotal-cf/om/vmlifecycle/taskmodifier"
	"github.com/pivotal-cf/om/vmlifecycle/vmlifecyclecommands"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"io"
	"os"
	"time"
)

type Automator struct {
	CreateVM                vmlifecyclecommands.CreateVM                `command:"create-vm" description:"Create VM for Ops Manager to a given IaaS"`
	DeleteVM                vmlifecyclecommands.DeleteVM                `command:"delete-vm" description:"Delete VM from a given IaaS"`
	UpgradeOpsman           vmlifecyclecommands.UpgradeOpsman           `command:"upgrade-opsman" description:"Deletes the old opsman vm given an exported installation, bring up a new vm, and import that installation"`
	ExportOpsmanConfig      vmlifecyclecommands.ExportOpsmanConfig      `command:"export-opsman-config" description:"Exports an opsman.yml for an existing Ops Manager VM"`
	PrepareTasksWithSecrets vmlifecyclecommands.PrepareTasksWithSecrets `command:"prepare-tasks-with-secrets" description:"Modifies task files to include config secrets as environment variables"`
}

func NewAutomator(stdout io.Writer, stderr io.Writer) *Automator {
	currentPathToBinaryBeingRun, _ := os.Executable()
	omRunner, _ := runner.NewRunner(currentPathToBinaryBeingRun, stdout, stderr)

	taskModifier := taskmodifier.NewTaskModifier()

	createVM := vmlifecyclecommands.NewCreateVMCommand(stdout, stderr, vmmanagers.NewCreateVMManager)
	deleteVM := vmlifecyclecommands.NewDeleteVMCommand(stdout, stderr, vmmanagers.NewDeleteVMManager)
	pollingInterval := 10 * time.Second

	return &Automator{
		CreateVM:                createVM,
		DeleteVM:                deleteVM,
		UpgradeOpsman:           *vmlifecyclecommands.NewUpgradeOpsman(stdout, stderr, &createVM, &deleteVM, omRunner, pollingInterval, 8*time.Minute),
		ExportOpsmanConfig:      vmlifecyclecommands.NewExportOpsmanConfigCommand(stdout, stderr, configfetchers.NewOpsmanConfigFetcher),
		PrepareTasksWithSecrets: vmlifecyclecommands.NewSecretsModifierCommand(stdout, stderr, taskModifier),
	}
}

func (*Automator) Execute(args []string) error {
	return nil
}
