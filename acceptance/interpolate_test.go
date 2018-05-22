package acceptance

import
(
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os/exec"
	"github.com/onsi/gomega/gexec"
	"os"
	"io/ioutil"
	"github.com/onsi/gomega/gbytes"
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
				"interpolate", "-c", yamlFile.Name(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, 5).Should(gexec.Exit(0))
			Eventually(session.Out, 5).Should(gbytes.Say("age: 100\nname: bob"))
		})

		Context("with vars defined in the manifest", func() {
			It("successfully replaces the vars", func() {
				varsFile := createFile("---\nname1: moe\nage1: 500")
				yamlFile := createFile("---\nname: ((name1))\nage: ((age1))")
				command := exec.Command(pathToMain,
					"interpolate",
					"-c", yamlFile.Name(),
					"-l", varsFile.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(0))
				Eventually(session.Out, 5).Should(gbytes.Say("age: 500\nname: moe"))
			})

			It("errors when no vars are provided", func() {
				yamlFile := createFile("---\nname: ((name1))\nage: ((age1))")
				command := exec.Command(pathToMain,
					"interpolate",
					"-c", yamlFile.Name(),
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session, 5).Should(gexec.Exit(1))
				Eventually(session.Err, 5).Should(gbytes.Say("Expected to find variables: age1"))
			})
		})
	})
})
