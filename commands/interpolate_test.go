package commands_test

import (
	"io/ioutil"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var templateNoParameters = `hello: world`
var templateNoParametersOverStdin = `hello: from standard input`
var templateWithParameters = `hello: ((hello))`
var templateWithMultipleParameters = `
hello: ((hello))
world: ((world))
`
var varsFileParameter = `hello: world`
var varsFileParameter2 = `hello: new world`
var opsFileParameter = `- type: replace
  path: /foo?
  value: bar
`

var _ = Describe("Interpolate", func() {
	var (
		command commands.Interpolate
		logger  *fakes.Logger
		stdin   *os.File
	)

	BeforeEach(func() {
		var err error
		stdin, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())
		ioutil.WriteFile(stdin.Name(), []byte(templateNoParametersOverStdin), os.ModeCharDevice|0755) // mimic a character device so it'll be picked up in the conditional
		logger = &fakes.Logger{}
		command = commands.NewInterpolate(func() []string { return nil }, logger, stdin)
	})

	AfterEach(func() {
		err := os.Remove(stdin.Name())
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Execute", func() {
		var (
			inputFile string
			varsFile  string
			varsFile2 string
			opsFile   string
		)

		BeforeEach(func() {
			tmpFile, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			inputFile = tmpFile.Name()

			tmpFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			varsFile = tmpFile.Name()

			tmpFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			varsFile2 = tmpFile.Name()

			tmpFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			opsFile = tmpFile.Name()
		})

		AfterEach(func() {
			err := os.Remove(inputFile)
			Expect(err).NotTo(HaveOccurred())
			err = os.Remove(varsFile)
			Expect(err).NotTo(HaveOccurred())
			err = os.Remove(varsFile2)
			Expect(err).NotTo(HaveOccurred())
			err = os.Remove(opsFile)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("no vars or ops file inputs", func() {
			It("succeeds", func() {
				err := ioutil.WriteFile(inputFile, []byte(templateNoParameters), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = command.Execute([]string{
					"--config", inputFile,
				})
				Expect(err).NotTo(HaveOccurred())

				content := logger.PrintlnArgsForCall(0)
				Expect(content[0].(string)).To(MatchYAML("hello: world"))
			})

			It("fails when all parameters are not specified", func() {
				err := ioutil.WriteFile(inputFile, []byte(templateWithParameters), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = command.Execute([]string{
					"--config", inputFile,
				})
				Expect(err).To(HaveOccurred())
				splitErr := strings.Split(err.Error(), "\n")
				Expect(splitErr).To(ConsistOf("Expected to find variables:", "hello"))
			})
		})

		Context("with vars file input", func() {
			It("succeeds", func() {
				err := ioutil.WriteFile(inputFile, []byte(templateNoParameters), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = ioutil.WriteFile(varsFile, []byte(varsFileParameter), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = command.Execute([]string{
					"--config", inputFile,
					"--vars-file", varsFile,
				})
				Expect(err).NotTo(HaveOccurred())

				content := logger.PrintlnArgsForCall(0)
				Expect(content[0].(string)).To(MatchYAML("hello: world"))
			})

			It("succeeds when multiple vars files", func() {
				err := ioutil.WriteFile(inputFile, []byte(templateWithParameters), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = ioutil.WriteFile(varsFile, []byte(varsFileParameter), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = ioutil.WriteFile(varsFile2, []byte(varsFileParameter2), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = command.Execute([]string{
					"--config", inputFile,
					"--vars-file", varsFile,
					"--vars-file", varsFile2,
				})
				Expect(err).NotTo(HaveOccurred())

				content := logger.PrintlnArgsForCall(0)
				Expect(content[0].(string)).To(MatchYAML("hello: new world"))
			})
		})

		Context("with vars input", func() {
			It("succeeds", func() {
				err := ioutil.WriteFile(inputFile, []byte(templateWithParameters), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = command.Execute([]string{
					"--config", inputFile,
					"--var", "hello=world",
				})
				Expect(err).NotTo(HaveOccurred())

				content := logger.PrintlnArgsForCall(0)
				Expect(content[0].(string)).To(MatchYAML("hello: world"))
			})

			It("succeeds with multiple vars inputs", func() {
				err := ioutil.WriteFile(inputFile, []byte(templateWithMultipleParameters), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = command.Execute([]string{
					"--config", inputFile,
					"--var", "hello=world",
					"--var", "world=hello",
				})
				Expect(err).NotTo(HaveOccurred())

				content := logger.PrintlnArgsForCall(0)
				Expect(content[0].(string)).To(MatchYAML("hello: world\nworld: hello"))
			})

			It("takes the last value if there are duplicate vars", func() {
				err := ioutil.WriteFile(inputFile, []byte(templateWithMultipleParameters), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = command.Execute([]string{
					"--config", inputFile,
					"--var", "hello=world",
					"--var", "world=hello",
					"--var", "hello=otherWorld",
				})
				Expect(err).NotTo(HaveOccurred())

				content := logger.PrintlnArgsForCall(0)
				Expect(content[0].(string)).To(MatchYAML("hello: otherWorld\nworld: hello"))
			})
		})

		Context("with ops file input", func() {
			It("succeeds", func() {
				err := ioutil.WriteFile(inputFile, []byte(templateNoParameters), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = ioutil.WriteFile(opsFile, []byte(opsFileParameter), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = command.Execute([]string{
					"--config", inputFile,
					"--ops-file", opsFile,
				})
				Expect(err).NotTo(HaveOccurred())

				content := logger.PrintlnArgsForCall(0)
				Expect(content[0].(string)).To(MatchYAML(`foo: bar
hello: world`))
			})
		})

		When("path flag is set", func() {
			It("returns a value from the interpolated file", func() {
				err := ioutil.WriteFile(inputFile, []byte(`{"a": "((interpolated-value))", "c":"d" }`), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = ioutil.WriteFile(varsFile, []byte(`{"interpolated-value": "b"}`), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = command.Execute([]string{
					"--config", inputFile,
					"--vars-file", varsFile,
					"--path", "/a",
				})
				Expect(err).NotTo(HaveOccurred())

				content := logger.PrintlnArgsForCall(0)
				Expect(content[0].(string)).To(MatchYAML(`b`))
			})
		})

		When("the skip-missing flag is set", func() {
			When("there are missing parameters", func() {
				It("succeeds", func() {
					err := ioutil.WriteFile(inputFile, []byte(templateWithParameters), 0755)
					Expect(err).NotTo(HaveOccurred())
					err = command.Execute([]string{
						"--config", inputFile,
						"--skip-missing",
					})
					Expect(err).NotTo(HaveOccurred())

					content := logger.PrintlnArgsForCall(0)
					Expect(content[0].(string)).To(MatchYAML(templateWithParameters))
				})
			})
		})

		When("no flags are set and no stdin provided", func() {
			It("errors", func() {
				command = commands.NewInterpolate(func() []string { return nil }, logger, os.Stdin)
				err := command.Execute([]string{})
				Expect(err).To(MatchError(ContainSubstring("no file or STDIN input provided.")))
			})
		})

		When("no stdin provided and --config -", func() {
			It("errors", func() {
				command = commands.NewInterpolate(func() []string { return nil }, logger, os.Stdin)
				err := command.Execute([]string{"--config", "-"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no file or STDIN input provided."))
			})
		})

		When("the config is passed via stdin with no config flag", func() {
			It("uses stdin", func() {
				err := command.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())
				content := logger.PrintlnArgsForCall(0)
				Expect(content[0].(string)).To(MatchYAML("hello: from standard input"))
			})
		})

		When("the config is passed via stdin with --config -", func() {
			It("uses stdin", func() {
				err := command.Execute([]string{"--config", "-"})
				Expect(err).NotTo(HaveOccurred())
				content := logger.PrintlnArgsForCall(0)
				Expect(content[0].(string)).To(MatchYAML("hello: from standard input"))
			})
		})

		When("the config is passed via stdin and a config file", func() {
			It("uses the config file", func() {
				err := ioutil.WriteFile(inputFile, []byte(templateNoParameters), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = command.Execute([]string{
					"--config", inputFile,
				})
				Expect(err).NotTo(HaveOccurred())

				content := logger.PrintlnArgsForCall(0)
				Expect(content[0].(string)).To(MatchYAML("hello: world"))
			})
		})
	})
})
