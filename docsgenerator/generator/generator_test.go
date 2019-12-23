package generator_test

import (
	"errors"
	"fmt"
	"github.com/pivotal-cf/om/docsgenerator/fakes"
	"github.com/pivotal-cf/om/docsgenerator/generator"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generator", func() {
	var (
		ex           *fakes.Executor
		templatesDir string
		docsDir      string
	)

	BeforeEach(func() {
		var err error
		ex = &fakes.Executor{}

		templatesDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		docsDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())
	})

	When("template files exist", func() {
		BeforeEach(func() {
			commandOneTemplateDir := filepath.Join(templatesDir, "command-one")
			err := os.MkdirAll(commandOneTemplateDir, 0755)
			Expect(err).ToNot(HaveOccurred())

			createFile(commandOneTemplateDir, generator.AdditionalInfoFileName, "command-one additional")
			createFile(commandOneTemplateDir, generator.DescriptionFileName, "command-one description")

			commandTwoTemplateDir := filepath.Join(templatesDir, "command-two")
			err = os.MkdirAll(commandTwoTemplateDir, 0755)
			Expect(err).ToNot(HaveOccurred())

			createFile(commandTwoTemplateDir, generator.AdditionalInfoFileName, "command-two additional")
			createFile(commandTwoTemplateDir, generator.DescriptionFileName, "command-two description")
		})

		It("creates a readme for each om command with data from the template files", func() {
			commandNames := []string{"command-one", "command-two"}
			ex.GetCommandNamesStub = func() ([]string, error) {
				return commandNames, nil
			}

			ex.GetCommandHelpStub = func(commandName string) ([]byte, error) {
				return []byte(fmt.Sprintf("%s help", commandName)), nil
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex)
			err := gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			docsFolders, err := ioutil.ReadDir(docsDir)
			Expect(err).ToNot(HaveOccurred())

			var docsFolderNames []string
			for _, docsFolder := range docsFolders {
				docsFolderNames = append(docsFolderNames, docsFolder.Name())
			}

			Expect(docsFolderNames).To(Equal(commandNames))

			for _, commandName := range commandNames {
				checkCommandReadmeContent(filepath.Join(docsDir, commandName), true)
			}
		})

		It("uses the description from om for each command missing a template", func() {
			commandName := "command-missing-template"
			ex.GetCommandNamesStub = func() ([]string, error) {
				return []string{commandName}, nil
			}

			ex.GetCommandHelpStub = func(commandName string) ([]byte, error) {
				return []byte(fmt.Sprintf("%s help", commandName)), nil
			}

			ex.GetDescriptionStub = func(commandName string) (string, error) {
				return fmt.Sprintf("%s description", commandName), nil
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex)
			err := gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			docsFolders, err := ioutil.ReadDir(docsDir)
			Expect(err).ToNot(HaveOccurred())

			var docsFolderNames []string
			for _, docsFolder := range docsFolders {
				docsFolderNames = append(docsFolderNames, docsFolder.Name())
			}

			Expect(docsFolderNames).To(Equal([]string{commandName}))

			checkCommandReadmeContent(filepath.Join(docsDir, commandName), false)
		})
	})

	When("there are no template files for a command", func() {
		It("creates template files based on the list of commands from om", func() {
			commandNames := []string{"command-one", "command-two"}
			ex.GetCommandNamesStub = func() (strings []string, e error) {
				return commandNames, nil
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex)
			err := gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			checkTemplateFiles(filepath.Join(templatesDir, "command-one"))
			checkTemplateFiles(filepath.Join(templatesDir, "command-two"))
		})
	})

	When("there are extra template folders", func() {
		It("removes template folders if the command is no longer in om", func() {
			missingCommandPath := filepath.Join(templatesDir, "missing-command")
			err := os.MkdirAll(missingCommandPath, 0755)
			Expect(err).ToNot(HaveOccurred())

			commandNames := []string{"command-one", "command-two"}
			ex.GetCommandNamesStub = func() (strings []string, e error) {
				return commandNames, nil
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex)
			err = gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			commandFolders, err := ioutil.ReadDir(templatesDir)
			Expect(err).ToNot(HaveOccurred())

			var commandFolderNames []string
			for _, commandFolder := range commandFolders {
				commandFolderNames = append(commandFolderNames, commandFolder.Name())
			}

			Expect(commandFolderNames).To(Equal(commandNames))

			_, err = os.Stat(missingCommandPath)
			Expect(err).To(HaveOccurred())
		})

		It("doesn't remove the base template dir", func() {
			gen := generator.NewGenerator(templatesDir, docsDir, ex)
			err := gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			_, err = os.Stat(templatesDir)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("om can't return the command names", func() {
		It("returns an error", func() {
			ex.GetCommandNamesStub = func() ([]string, error) {
				return nil, errors.New("om commandNames error")
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex)
			err := gen.GenerateDocs()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("om commandNames error"))
		})
	})

	When("the templates dir doesn't exist", func() {
		It("returns an error", func() {
			gen := generator.NewGenerator("doesnt-exist", docsDir, ex)
			err := gen.GenerateDocs()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no such file or directory"))
		})
	})

	When("om can't return the command description", func() {
		It("returns an error", func() {
			ex.GetCommandNamesStub = func() ([]string, error) {
				return []string{"command-one"}, nil
			}

			ex.GetDescriptionStub = func(commandName string) (string, error) {
				return "", errors.New("om description error")
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex)
			err := gen.GenerateDocs()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("om description error"))
		})
	})

	When("om can't return the command help", func() {
		It("returns an error", func() {
			ex.GetCommandNamesStub = func() ([]string, error) {
				return []string{"command-one"}, nil
			}

			ex.GetDescriptionStub = func(commandName string) (string, error) {
				return "some description", nil
			}

			ex.GetCommandHelpStub = func(commandName string) ([]byte, error) {
				return nil, errors.New("om help error")
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex)
			err := gen.GenerateDocs()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("om help error"))
		})
	})
})

func checkCommandReadmeContent(containingDir string, additional bool) {
	commandName := filepath.Base(containingDir)

	readmeContent, err := ioutil.ReadFile(filepath.Join(containingDir, generator.ReadmeFileName))
	Expect(err).ToNot(HaveOccurred())

	additionalText := ""
	if additional {
		additionalText = fmt.Sprintf("%s additional", commandName)

	}

	Expect(string(readmeContent)).To(Equal(fmt.Sprintf(
		generator.CommandReadmeTemplate,
		commandName,
		commandName,
		fmt.Sprintf("%s description", commandName),
		fmt.Sprintf("%s help", commandName),
		additionalText,
	)))
}

func checkTemplateFiles(containingDir string) {
	commandName := filepath.Base(containingDir)

	descriptionContents, err := ioutil.ReadFile(filepath.Join(containingDir, generator.DescriptionFileName))
	Expect(err).ToNot(HaveOccurred())
	Expect(string(descriptionContents)).To(Equal(fmt.Sprintf(generator.DescriptionTemplate, commandName)))

	additionalInfoContents, err := ioutil.ReadFile(filepath.Join(containingDir, generator.AdditionalInfoFileName))
	Expect(err).ToNot(HaveOccurred())
	Expect(string(additionalInfoContents)).To(Equal(fmt.Sprintf(generator.AdditionalInfoTemplate, commandName)))
}

func createFile(dir string, name string, contents string) {
	f, err := os.Create(filepath.Join(dir, name))
	Expect(err).ToNot(HaveOccurred())

	_, err = f.Write([]byte(contents))
	Expect(err).ToNot(HaveOccurred())

	err = f.Close()
	Expect(err).ToNot(HaveOccurred())
}
