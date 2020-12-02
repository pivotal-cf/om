package taskmodifier

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/pivotal-cf/om/interpolate"
	"gopkg.in/yaml.v2"
)

type TaskModifier struct {
	taskDir     string
	configPaths []string
	varsPaths   []string
	envPrefix   string
	stderr      io.Writer
}

func NewTaskModifier() *TaskModifier {
	return &TaskModifier{
		envPrefix: "OM",
	}
}

type taskFile struct {
	Fields map[string]interface{} `yaml:",inline"`
	Params map[string]string      `yaml:"params"`
}

func (c *TaskModifier) ModifyTasksWithSecrets(stderr io.Writer, taskDir string, configPaths []string, varsPaths []string) error {
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		return fmt.Errorf("task directory '%s' does not exist: %s", taskDir, err)
	}

	for _, configDir := range configPaths {
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			return fmt.Errorf("config directory '%s' does not exist: %s", configDir, err)
		}
	}

	c.taskDir = taskDir
	c.configPaths = configPaths
	c.varsPaths = varsPaths
	c.stderr = stderr

	taskFilenames, err := doublestar.Glob(filepath.Join(c.taskDir, "**", "*.{yml,yaml}"))
	if err != nil {
		return fmt.Errorf("could not find tasks files: %s", err)
	}

	varsFiles, err := c.getVarsFilenames()
	if err != nil {
		return err
	}

	missingValues, err := c.getMissingValues(varsFiles)
	if err != nil {
		return err
	}

	for _, taskFilename := range taskFilenames {
		err = c.updateTaskWithSecrets(taskFilename, missingValues)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *TaskModifier) updateTaskWithSecrets(taskFilename string, secretNames []string) error {
	var task taskFile
	contents, err := ioutil.ReadFile(taskFilename)
	if err != nil {
		return fmt.Errorf("could not read task file '%s': %s", taskFilename, err)
	}

	err = yaml.Unmarshal(contents, &task)
	if err != nil {
		return fmt.Errorf("could not parse task file '%s': %s", taskFilename, err)
	}

	if task.Params == nil {
		task.Params = map[string]string{}
	}

	if len(secretNames) > 0 {
		task.Params[fmt.Sprintf("%s_VARS_ENV", c.envPrefix)] = fmt.Sprintf("%s_VAR", c.envPrefix)

		for _, secretName := range secretNames {
			paramName := fmt.Sprintf("%s_VAR_%s", c.envPrefix, secretName)
			if _, ok := task.Params[paramName]; !ok {
				task.Params[paramName] = fmt.Sprintf("((%s))", secretName)
			}
		}
	}

	fmt.Fprintf(c.stderr, "modifying task file %s\n", taskFilename)

	contents, err = yaml.Marshal(task)
	if err != nil {
		return fmt.Errorf("could not marshal modified task '%s': %s", taskFilename, err)
	}

	err = ioutil.WriteFile(taskFilename, contents, 0666)
	if err != nil {
		return fmt.Errorf("could write modified task '%s': %s", taskFilename, err)
	}

	return nil
}

func (c *TaskModifier) getVarsFilenames() ([]string, error) {
	var (
		varsFiles []string
	)

	for _, varsPath := range c.varsPaths {
		varsFilenames := []string{varsPath}

		info, err := os.Stat(varsPath)
		if err != nil {
			return nil, fmt.Errorf("vars directory '%s' does not exist: %s", varsPath, err)
		}

		if info.IsDir() {
			varsFilenames, err = doublestar.Glob(filepath.Join(varsPath, "**", "*.{yml,yaml}"))
			if err != nil {
				return nil, fmt.Errorf("could not find vars files in '%s': %s", varsPath, err)
			}
		}

		varsFiles = append(varsFiles, varsFilenames...)
	}

	return varsFiles, nil
}

func (c *TaskModifier) getMissingValues(varsFiles []string) ([]string, error) {
	var missingValues []string
	for _, configPath := range c.configPaths {
		configFilenames := []string{configPath}

		if info, _ := os.Stat(configPath); info.IsDir() {
			configFilenames, _ = doublestar.Glob(filepath.Join(configPath, "**", "*.{yml,yaml}"))
		}

		for _, configFilename := range configFilenames {
			_, err := interpolate.Execute(interpolate.Options{
				TemplateFile:  configFilename,
				VarsFiles:     varsFiles,
				ExpectAllKeys: true,
			})
			if err != nil {
				if !strings.Contains(err.Error(), "Expected to find variables") {
					return nil, fmt.Errorf("could not interpolate vars from '%s' into config file '%s': %s", c.varsPaths, configFilename, err)
				}

				splitErr := strings.Split(err.Error(), ": ")
				if len(splitErr) > 1 {
					fmt.Fprintf(c.stderr, "found secrets in %s\n", configFilename)

					missingValues = append(missingValues, strings.Split(splitErr[1], "\n")...)
				}
			}
		}
	}

	return missingValues, nil
}
