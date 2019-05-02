package acceptance

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("interpolate command", func() {
	Context("when given a valid YAML file", func() {
		createFile := func(contents string) *os.File {
			file, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())
			_, err = file.WriteString(contents)
			Expect(err).ToNot(HaveOccurred())
			return file
		}

		It("outputs a YAML file", func() {
			yamlFile := createFile("---\nname: bob\nage: 100")
			command := exec.Command(pathToMain,
				"interpolate", "--config", yamlFile.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 5).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(MatchYAML(`
age: 100
name: bob
`))
		})

		Context("with vars defined in the manifest", func() {
			It("successfully replaces the vars", func() {
				varsFile := createFile("---\nname1: moe\nage1: 500")
				yamlFile := createFile("---\nname: ((name1))\nage: ((age1))")
				command := exec.Command(pathToMain,
					"interpolate",
					"--config", yamlFile.Name(),
					"--vars-file", varsFile.Name(),
				)
				defer varsFile.Close()
				defer yamlFile.Close()

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(MatchYAML(`
age: 500
name: moe
`))
			})

			It("replaces the vars based on the order precedence of the vars file", func() {
				vars1File := createFile("---\nname1: moe\nage1: 500")
				vars2File := createFile("---\nname1: bob")
				yamlFile := createFile("---\nname: ((name1))\nage: ((age1))")
				command := exec.Command(pathToMain,
					"interpolate",
					"--config", yamlFile.Name(),
					"--vars-file", vars1File.Name(),
					"--vars-file", vars2File.Name(),
				)
				defer vars1File.Close()
				defer vars2File.Close()
				defer yamlFile.Close()

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(MatchYAML(`
age: 500
name: bob
`))
			})

		})

		Context("with vars defined in the environment", func() {
			It("successfully replaces the vars", func() {
				yamlFile := createFile("---\nname: ((name1))\nage: ((age1))\nhas_pet: ((has_pet1))")
				defer yamlFile.Close()
				command := exec.Command(pathToMain,
					"interpolate",
					"--config", yamlFile.Name(),
					"--vars-env", "OM_VAR",
				)
				command.Env = append(command.Env, "OM_VAR_age1=500")
				command.Env = append(command.Env, "OM_VAR_name1=moe")
				command.Env = append(command.Env, "OM_VAR_has_pet1=true")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(MatchYAML(`
age: 500
name: moe
has_pet: true
`))
			})

			It("replaces the vars based on the order precedence of the vars environment variable", func() {
				yamlFile := createFile("---\nname: ((name1))\nage: ((age1))")
				defer yamlFile.Close()

				command := exec.Command(pathToMain,
					"interpolate",
					"--config", yamlFile.Name(),
					"--vars-env", "OM_VAR_1",
					"--vars-env", "OM_VAR_2",
				)

				command.Env = append(command.Env, "OM_VAR_1_age1=500")
				command.Env = append(command.Env, "OM_VAR_1_name1=moe")
				command.Env = append(command.Env, "OM_VAR_2_name1=bob")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(MatchYAML(`
age: 500
name: bob
`))
			})

			It("vars-file variables take precedence over var-envs variables", func() {
				yamlFile := createFile("---\nname: ((name1))\nage: ((age1))")
				defer yamlFile.Close()
				varsFile := createFile("---\nage1: 1000")
				defer varsFile.Close()

				command := exec.Command(pathToMain,
					"interpolate",
					"--config", yamlFile.Name(),
					"--vars-env", "OM_VAR",
					"--vars-file", varsFile.Name(),
				)

				command.Env = append(command.Env, "OM_VAR_age1=500")
				command.Env = append(command.Env, "OM_VAR_name1=moe")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(MatchYAML(`
age: 1000
name: moe
`))
			})

			It("handles multi-line environment variables such as certificates", func() {
				yamlFile := createFile("---\nname: ((multi_line_value))")
				defer yamlFile.Close()

				command := exec.Command(pathToMain,
					"interpolate",
					"--config", yamlFile.Name(),
					"--vars-env", "OM_VAR",
				)

				command.Env = append(command.Env, "OM_VAR_multi_line_value=some\nmulti\nline\nvalue")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(MatchYAML(`
name: |-
  some
  multi
  line
  value
`))
			})

			It("handles hash environment variables", func() {
				yamlFile := createFile("---\nhash: ((hash))")
				defer yamlFile.Close()

				command := exec.Command(pathToMain,
					"interpolate",
					"--config", yamlFile.Name(),
					"--vars-env", "OM_VAR",
				)

				hashContents := `---
some-key:
  some-multiline-value: |-
    some
    multi
    line
    value
`
				command.Env = append(command.Env, fmt.Sprintf("OM_VAR_hash=%s", hashContents))

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(MatchYAML(`
hash:
  some-key:
    some-multiline-value: |-
      some
      multi
      line
      value
`))
			})
		})

		Context("when no vars are provided", func() {
			It("errors", func() {
				yamlFile := createFile("---\nname: ((name1))\nage: ((age1))")
				command := exec.Command(pathToMain,
					"interpolate",
					"--config", yamlFile.Name(),
				)
				defer yamlFile.Close()

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(1))
				Eventually(session.Err, 5).Should(gbytes.Say("Expected to find variables: age1"))
			})
		})
	})

	Context("when given standard input", func() {
		It("parses the stdin and returns the value", func() {
			command := exec.Command(pathToMain,
				"interpolate",
			)
			command.Stdin = strings.NewReader("---\nname: bob\nage: 100")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 5).Should(gexec.Exit(0))
			Expect(session.Out.Contents()).To(MatchYAML(`
age: 100
name: bob
`))
		})

		When("the path flag is used", func() {
			It("returns the designated value", func() {
				command := exec.Command(pathToMain,
					"interpolate",
					"--path", "/name",
				)
				command.Stdin = strings.NewReader("---\nname: bob\nage: 100")

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(0))
				Expect(session.Out.Contents()).To(MatchYAML(`
bob
`))
			})
		})
	})
})
