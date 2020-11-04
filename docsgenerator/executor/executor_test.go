package executor_test

import (
	"github.com/pivotal-cf/om/docsgenerator/executor"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executor", func() {
	It("executes om with the provided args", func() {
		os.Setenv("STUB_OUTPUT", `1.2.3`)

		ex := executor.NewExecutor(pathToStub)
		output, err := ex.RunOmCommand("version")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(Equal("1.2.3\n"))
	})

	It("retrieves the command names from om", func() {
		os.Setenv("STUB_OUTPUT", `
Usage:
  om [OPTIONS] <command>

Application Options:
      --ca-cert=               OpsManager CA certificate path or value [$OM_CA_CERT]

Help Options:
  -h, --help                   Show this help message

Available commands:
  activate-certificate-authority  activates a certificate authority on the Ops Manager
  apply-changes                   triggers an install on the Ops Manager targeted
  assign-multi-stemcell           assigns multiple uploaded stemcells to a product in the targeted Ops Manager 2.6+
  errands                         list errands for a product
  interpolate                     interpolates variables into a manifest
`)

		ex := executor.NewExecutor(pathToStub)
		output, err := ex.GetCommandNamesAndDescriptions()
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(Equal(map[string]string{
			"activate-certificate-authority": "activates a certificate authority on the Ops Manager",
			"apply-changes":                  "triggers an install on the Ops Manager targeted",
			"assign-multi-stemcell":          "assigns multiple uploaded stemcells to a product in the targeted Ops Manager 2.6+",
			"errands":                        "list errands for a product",
			"interpolate":                    "interpolates variables into a manifest",
		}))
	})

	It("retrieves the command description from om", func() {
		os.Setenv("STUB_OUTPUT", `Usage:
  om [OPTIONS] activate-certificate-authority [activate-certificate-authority-OPTIONS]

This authenticated command activates an existing certificate authority on the
Ops Manager

Application Options:
      --ca-cert=               OpsManager CA certificate path or value [$OM_CA_CERT]
`)

		ex := executor.NewExecutor(pathToStub)
		output, err := ex.GetDescription("activate-certificate-authority")
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(ContainSubstring("This authenticated command activates an existing certificate authority on the\nOps Manager"))
	})

	It("retrieves the command help from om", func() {
		helpText := `Usage:
  om [OPTIONS] activate-certificate-authority [activate-certificate-authority-OPTIONS]

This authenticated command activates an existing certificate authority on the Ops Manager

Application Options:
      --ca-cert=               OpsManager CA certificate path or value [$OM_CA_CERT]`

		os.Setenv("STUB_OUTPUT", helpText)

		ex := executor.NewExecutor(pathToStub)
		output, err := ex.GetCommandHelp("activate-certificate-authority")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(ContainSubstring(helpText))
	})
})
