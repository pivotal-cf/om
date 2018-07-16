package commands_test

import (
	"io/ioutil"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var templateNoParameters = `hello: world`
var templateWithParameters = `hello: ((hello))`
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
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		command = commands.NewInterpolate(logger)
	})

	Describe("Execute", func() {
		var (
			inputFile string
			varsFile  string
			varsFile2 string
			opsFile   string
		)

		BeforeEach(func() {
			inputFile = path.Join(os.TempDir(), "input.yml")
			varsFile = path.Join(os.TempDir(), "vars.yml")
			varsFile2 = path.Join(os.TempDir(), "vars2.yml")
			opsFile = path.Join(os.TempDir(), "ops.yml")
		})

		AfterEach(func() {
			os.Remove(inputFile)
			os.Remove(varsFile)
			os.Remove(varsFile2)
			os.Remove(opsFile)
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
				Expect(err.Error()).Should(ContainSubstring("Expected to find variables: hello"))
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

		Context("Failure cases", func() {
			Context("when there is no input file", func() {
				It("returns an error", func() {
					err := command.Execute([]string{
						"--config", "foo.yml",
					})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).Should(ContainSubstring("no such file or directory"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewInterpolate(nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "Interpolates variables into a manifest",
				ShortDescription: "Interpolates variables into a manifest",
				Flags:            command.Options,
			}))
		})
	})
})
