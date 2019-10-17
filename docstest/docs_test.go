package docs_test

import (
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Documentation coverage", func() {
	Context("Commands doc", func() {
		It("Mentions every command we distribute", func() {
			commands := getCommandNames()
			taskDoc := readFile("../docs/README.md")

			for _, command := range commands {
				Expect(taskDoc).To(
					ContainSubstring(command),
					fmt.Sprintf("docs/README.md should have %s\n", command),
				)
			}
		})
	})
})

func readFile(docName string) (docContents string) {
	docPath, err := filepath.Abs(docName)
	Expect(err).NotTo(HaveOccurred())

	docContentsBytes, err := ioutil.ReadFile(docPath)
	docContents = string(docContentsBytes)
	Expect(err).NotTo(HaveOccurred())

	return docContents
}

func getCommandNames() []string {
	command := exec.Command(pathToMain, "--help")

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	session.Wait()

	output := strings.Split(string(session.Out.Contents()), "\n")

	var isCommand bool
	var commands []string
	for _, commandLine := range output {
		if strings.Contains(commandLine, "Commands:") && !isCommand {
			isCommand = true
			continue
		}

		if isCommand && commandLine != "" {
			splitCommandLine := strings.Fields(commandLine)
			commands = append(commands, splitCommandLine[0])
		}
	}

	return commands
}
