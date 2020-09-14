package docs_test

import (
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os"
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

			var missing []string
			for _, command := range commands {
				if !strings.Contains(taskDoc, command) {
					missing = append(missing, command)
				}
			}

			Expect(missing).To(HaveLen(0), fmt.Sprintf("docs/README.md should have: \n%s\n run `go run docsgenerator/update-docs.go` to fix", strings.Join(missing, "\n")))
		})

		It("contains a docs readme for each command", func() {
			commands := getCommandNames()
			docsPath := filepath.Join("..", "docs")

			var missing []string
			for _, command := range commands {
				if _, err := os.Stat(filepath.Join(docsPath, command, "README.md")); err != nil {
					missing = append(missing, command)
				}
			}

			Expect(missing).To(HaveLen(0), fmt.Sprintf("docs should have a readme for each command. Missing readmes for: \n%s\n run `go run docsgenerator/update-docs.go` to fix", strings.Join(missing, "\n")))
		})
	})
})

func readFile(docName string) (docContents string) {
	docPath, err := filepath.Abs(docName)
	Expect(err).ToNot(HaveOccurred())

	docContentsBytes, err := ioutil.ReadFile(docPath)
	docContents = string(docContentsBytes)
	Expect(err).ToNot(HaveOccurred())

	return docContents
}

func getCommandNames() []string {
	command := exec.Command(pathToMain, "--help")

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	session.Wait()

	output := strings.Split(string(session.Out.Contents()), "\n")

	var inTheCommandZone bool
	var commands []string
	for _, commandLine := range output {
		if strings.Contains(commandLine, "Commands:") && !inTheCommandZone {
			inTheCommandZone = true
			continue
		}

		if strings.Contains(commandLine, "Global Flags:") {
			break
		}

		if inTheCommandZone && commandLine != "" {
			splitCommandLine := strings.Fields(commandLine)
			commands = append(commands, splitCommandLine[0])
		}
	}

	return commands
}
