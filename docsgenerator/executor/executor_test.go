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

	FIt("retrieves the command names from om", func() {
		os.Setenv("STUB_OUTPUT", `
om helps you interact with an Ops Manager

Usage: 
  om [options] <command> [<args>]

Commands:
  activate-certificate-authority  activates a certificate authority on the Ops Manager
  apply-changes                   triggers an install on the Ops Manager targeted
  assign-multi-stemcell           assigns multiple uploaded stemcells to a product in the targeted Ops Manager 2.6+
  errands                         list errands for a product
  interpolate                     interpolates variables into a manifest

Global Flags:
  --ca-cert, OM_CA_CERT                                  string  OpsManager CA certificate path or value
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
		os.Setenv("STUB_OUTPUT", `ॐ  activate-certificate-authority
This authenticated command activates an existing certificate authority on the Ops Manager

Usage: om [options] activate-certificate-authority [<args>]
  --ca-cert, OM_CA_CERT                                  string  OpsManager CA certificate path or value
`)

		ex := executor.NewExecutor(pathToStub)
		output, err := ex.GetDescription("activate-certificate-authority")
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(Equal("This authenticated command activates an existing certificate authority on the Ops Manager"))
	})

	It("retrieves the command help from om", func() {
		helpText := `ॐ  activate-certificate-authority
This authenticated command activates an existing certificate authority on the Ops Manager

Usage: om [options] activate-certificate-authority [<args>]
  --ca-cert, OM_CA_CERT                                  string  OpsManager CA certificate path or value`

		os.Setenv("STUB_OUTPUT", helpText)

		ex := executor.NewExecutor(pathToStub)
		output, err := ex.GetCommandHelp("activate-certificate-authority")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(ContainSubstring(helpText))
	})
})
