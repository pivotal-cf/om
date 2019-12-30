package generator_test

import (
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/docsgenerator/fakes"
	"github.com/pivotal-cf/om/docsgenerator/generator"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
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

	AfterEach(func() {
		err := os.RemoveAll(templatesDir)
		Expect(err).ToNot(HaveOccurred())

		err = os.RemoveAll(docsDir)
		Expect(err).ToNot(HaveOccurred())
	})

	When("template files exist", func() {
		const (
			commandOneAdditional  = "command-one additional"
			commandTwoAdditional  = "command-two additional"
			commandOneDescription = "command-one description"
			commandTwoDescription = "command-two description"
			readmeBefore          = "before content"
			readmeAfter           = "after content"
		)

		BeforeEach(func() {
			commandOneTemplateDir := filepath.Join(templatesDir, "command-one")
			err := os.MkdirAll(commandOneTemplateDir, 0755)
			Expect(err).ToNot(HaveOccurred())

			createFile(commandOneTemplateDir, generator.AdditionalInfoFileName, commandOneAdditional)
			createFile(commandOneTemplateDir, generator.DescriptionFileName, commandOneDescription)

			commandTwoTemplateDir := filepath.Join(templatesDir, "command-two")
			err = os.MkdirAll(commandTwoTemplateDir, 0755)
			Expect(err).ToNot(HaveOccurred())

			createFile(commandTwoTemplateDir, generator.AdditionalInfoFileName, commandTwoAdditional)
			createFile(commandTwoTemplateDir, generator.DescriptionFileName, commandTwoDescription)

			createFile(templatesDir, generator.ReadmeBeforeFileName, readmeBefore)
			createFile(templatesDir, generator.ReadmeAfterFileName, readmeAfter)
		})

		It("creates a readme for each om command with data from the template files", func() {
			commandDescriptions := map[string]string{
				"command-one": "command-one description",
				"command-two": "command-two description",
			}
			ex.GetCommandNamesAndDescriptionsStub = func() (map[string]string, error) {
				return commandDescriptions, nil
			}

			ex.GetCommandHelpStub = func(commandName string) ([]byte, error) {
				return []byte(fmt.Sprintf("%s help", commandName)), nil
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex, GinkgoWriter)
			err := gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			docsFolders, err := ioutil.ReadDir(docsDir)
			Expect(err).ToNot(HaveOccurred())

			var docsFolderNames []string
			for _, docsFolder := range docsFolders {
				if docsFolder.IsDir() {
					docsFolderNames = append(docsFolderNames, docsFolder.Name())
				}
			}

			var commandNames []string
			for command := range commandDescriptions {
				commandNames = append(commandNames, command)
			}
			sort.Strings(commandNames)

			Expect(docsFolderNames).To(Equal(commandNames))

			for _, commandName := range commandNames {
				checkCommandReadmeContent(filepath.Join(docsDir, commandName), true)
			}
		})

		It("does not overwrite the template files", func() {
			commandDescriptions := map[string]string{
				"command-one": "command-one description",
				"command-two": "command-two description",
			}
			ex.GetCommandNamesAndDescriptionsStub = func() (map[string]string, error) {
				return commandDescriptions, nil
			}

			ex.GetCommandHelpStub = func(commandName string) ([]byte, error) {
				return []byte(fmt.Sprintf("%s help", commandName)), nil
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex, GinkgoWriter)
			err := gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			checkFileEquals(filepath.Join(templatesDir, "command-one", generator.AdditionalInfoFileName), commandOneAdditional)
			checkFileEquals(filepath.Join(templatesDir, "command-one", generator.DescriptionFileName), commandOneDescription)
			checkFileEquals(filepath.Join(templatesDir, "command-two", generator.AdditionalInfoFileName), commandTwoAdditional)
			checkFileEquals(filepath.Join(templatesDir, "command-two", generator.DescriptionFileName), commandTwoDescription)
			checkFileEquals(filepath.Join(templatesDir, generator.ReadmeBeforeFileName), readmeBefore)
			checkFileEquals(filepath.Join(templatesDir, generator.ReadmeAfterFileName), readmeAfter)
		})

		It("uses the readme templates to create the final readme", func() {
			commandDescriptions := map[string]string{
				"command-one": "command-one description om",
				"command-two": "command-two description om",
			}

			ex.GetCommandNamesAndDescriptionsStub = func() (strings map[string]string, e error) {
				return commandDescriptions, nil
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex, GinkgoWriter)
			err := gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			readmeContent, err := ioutil.ReadFile(filepath.Join(docsDir, generator.ReadmeFileName))
			Expect(err).ToNot(HaveOccurred())

			Expect(string(readmeContent)).To(Equal(fmt.Sprintf(
				generator.ReadmeTemplate,
				"before content",
				"| Command | Description |\n| ------------- | ------------- |\n| [command-one](command-one/README.md) | command-one description om |\n| [command-two](command-two/README.md) | command-two description om |",
				"after content",
			)))
		})
	})

	When("there are no template files for a command", func() {
		It("creates template files based on the list of commands from om", func() {
			commandDescriptions := map[string]string{
				"command-one": "command-one description",
				"command-two": "command-two description",
			}
			ex.GetCommandNamesAndDescriptionsStub = func() (strings map[string]string, e error) {
				return commandDescriptions, nil
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex, GinkgoWriter)
			err := gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			checkTemplateFiles(filepath.Join(templatesDir, "command-one"))
			checkTemplateFiles(filepath.Join(templatesDir, "command-two"))
		})

		It("uses the description from om for each command missing a template", func() {
			commandName := "command-missing-template"
			ex.GetCommandNamesAndDescriptionsStub = func() (map[string]string, error) {
				return map[string]string{commandName: "missing-command"}, nil
			}

			ex.GetCommandHelpStub = func(commandName string) ([]byte, error) {
				return []byte(fmt.Sprintf("%s help", commandName)), nil
			}

			ex.GetDescriptionStub = func(commandName string) (string, error) {
				return fmt.Sprintf("%s description", commandName), nil
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex, GinkgoWriter)
			err := gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			docsFolders, err := ioutil.ReadDir(docsDir)
			Expect(err).ToNot(HaveOccurred())

			var docsFolderNames []string
			for _, docsFolder := range docsFolders {
				if docsFolder.IsDir() {
					docsFolderNames = append(docsFolderNames, docsFolder.Name())
				}
			}

			Expect(docsFolderNames).To(Equal([]string{commandName}))

			checkCommandReadmeContent(filepath.Join(docsDir, commandName), false)
		})

		It("creates template files for the base readme file", func() {
			gen := generator.NewGenerator(templatesDir, docsDir, ex, GinkgoWriter)
			err := gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			checkReadmeTemplateFiles(templatesDir)
		})

		It("doesn't print the template text in the outputted files", func() {
			commandDescriptions := map[string]string{
				"command-one": "command-one description",
				"command-two": "command-two description",
			}
			ex.GetCommandNamesAndDescriptionsStub = func() (strings map[string]string, e error) {
				return commandDescriptions, nil
			}

			ex.GetCommandHelpStub = func(commandName string) ([]byte, error) {
				return []byte(fmt.Sprintf("%s help", commandName)), nil
			}

			ex.GetDescriptionStub = func(commandName string) (string, error) {
				return fmt.Sprintf("%s description", commandName), nil
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex, GinkgoWriter)
			err := gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			checkFileDoesNotContain(filepath.Join(docsDir, "command-one", generator.ReadmeFileName), fmt.Sprintf(generator.AdditionalInfoTemplate, "command-one"))
			checkFileDoesNotContain(filepath.Join(docsDir, "command-one", generator.ReadmeFileName), fmt.Sprintf(generator.DescriptionTemplate, "command-one"))
			checkFileDoesNotContain(filepath.Join(docsDir, "command-two", generator.ReadmeFileName), fmt.Sprintf(generator.AdditionalInfoTemplate, "command-two"))
			checkFileDoesNotContain(filepath.Join(docsDir, "command-two", generator.ReadmeFileName), fmt.Sprintf(generator.DescriptionTemplate, "command-two"))
			checkFileDoesNotContain(filepath.Join(docsDir, generator.ReadmeFileName), generator.ReadmeBeforeTemplate)
			checkFileDoesNotContain(filepath.Join(docsDir, generator.ReadmeFileName), generator.ReadmeAfterTemplate)
		})
	})

	When("there are extra template folders", func() {
		It("removes template folders if the command is no longer in om", func() {
			missingCommandPath := filepath.Join(templatesDir, "missing-command")
			err := os.MkdirAll(missingCommandPath, 0755)
			Expect(err).ToNot(HaveOccurred())

			commandDescriptions := map[string]string{
				"command-one": "command-one description",
				"command-two": "command-two description",
			}
			ex.GetCommandNamesAndDescriptionsStub = func() (strings map[string]string, e error) {
				return commandDescriptions, nil
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex, GinkgoWriter)
			err = gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			commandFolders, err := ioutil.ReadDir(templatesDir)
			Expect(err).ToNot(HaveOccurred())

			var commandFolderNames []string
			for _, commandFolder := range commandFolders {
				if commandFolder.IsDir() {
					commandFolderNames = append(commandFolderNames, commandFolder.Name())
				}
			}

			var commandNames []string
			for command := range commandDescriptions {
				commandNames = append(commandNames, command)
			}
			sort.Strings(commandNames)

			Expect(commandFolderNames).To(Equal(commandNames))

			_, err = os.Stat(missingCommandPath)
			Expect(err).To(HaveOccurred())
		})

		It("doesn't remove the base template dir", func() {
			gen := generator.NewGenerator(templatesDir, docsDir, ex, GinkgoWriter)
			err := gen.GenerateDocs()
			Expect(err).ToNot(HaveOccurred())

			_, err = os.Stat(templatesDir)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("om can't return the command names and descriptions", func() {
		It("returns an error", func() {
			ex.GetCommandNamesAndDescriptionsStub = func() (map[string]string, error) {
				return nil, errors.New("om commandNames error")
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex, GinkgoWriter)
			err := gen.GenerateDocs()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("om commandNames error"))
		})
	})

	When("the templates dir doesn't exist", func() {
		It("returns an error", func() {
			gen := generator.NewGenerator("doesnt-exist", docsDir, ex, GinkgoWriter)
			err := gen.GenerateDocs()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no such file or directory"))
		})
	})

	When("om can't return the command description", func() {
		It("returns an error", func() {
			ex.GetCommandNamesAndDescriptionsStub = func() (map[string]string, error) {
				return map[string]string{"command-one": "command-one description"}, nil
			}

			ex.GetDescriptionStub = func(commandName string) (string, error) {
				return "", errors.New("om description error")
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex, GinkgoWriter)
			err := gen.GenerateDocs()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("om description error"))
		})
	})

	When("om can't return the command help", func() {
		It("returns an error", func() {
			ex.GetCommandNamesAndDescriptionsStub = func() (map[string]string, error) {
				return map[string]string{"command-one": "command-one description"}, nil
			}

			ex.GetDescriptionStub = func(commandName string) (string, error) {
				return "some description", nil
			}

			ex.GetCommandHelpStub = func(commandName string) ([]byte, error) {
				return nil, errors.New("om help error")
			}

			gen := generator.NewGenerator(templatesDir, docsDir, ex, GinkgoWriter)
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

func checkReadmeTemplateFiles(containingDir string) {
	descriptionContents, err := ioutil.ReadFile(filepath.Join(containingDir, generator.ReadmeBeforeFileName))
	Expect(err).ToNot(HaveOccurred())
	Expect(string(descriptionContents)).To(Equal(generator.ReadmeBeforeTemplate))

	additionalInfoContents, err := ioutil.ReadFile(filepath.Join(containingDir, generator.ReadmeAfterFileName))
	Expect(err).ToNot(HaveOccurred())
	Expect(string(additionalInfoContents)).To(Equal(generator.ReadmeAfterTemplate))
}

func checkFileDoesNotContain(filePath string, lines ...string) {
	content, err := ioutil.ReadFile(filePath)
	Expect(err).ToNot(HaveOccurred())

	for _, line := range lines {
		Expect(string(content)).ToNot(ContainSubstring(line))
	}
}

func checkFileEquals(filePath string, expected string) {
	content, err := ioutil.ReadFile(filePath)
	Expect(err).ToNot(HaveOccurred())

	Expect(string(content)).To(Equal(expected))
}

func createFile(dir string, name string, contents string) {
	f, err := os.Create(filepath.Join(dir, name))
	Expect(err).ToNot(HaveOccurred())

	_, err = f.Write([]byte(contents))
	Expect(err).ToNot(HaveOccurred())

	err = f.Close()
	Expect(err).ToNot(HaveOccurred())
}
