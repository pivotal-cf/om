package interpolate_test

import (
	"github.com/pivotal-cf/om/interpolate"
	"io/ioutil"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInterpolate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Interpolate Suite")
}

var _ = Describe("Execute", func() {
	It("errors when the template file does not exist", func() {
		_, err := interpolate.Execute(interpolate.Options{
			TemplateFile: "unknown.txt",
		})
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError("could not read file (unknown.txt): open unknown.txt: no such file or directory"))
	})

	When("path is specified", func() {
		It("returns that specific value", func() {
			contents, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: Bob}`),
				Path:         "/name",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(string(contents)).To(Equal("Bob\n"))
		})

		It("errors with an invalid path", func() {
			_, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: Bob}`),
				Path:         "/age",
			})
			Expect(err).To(HaveOccurred())
		})
	})

	When("environment variables are used", func() {
		It("supports variables with a prefix", func() {
			contents, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: ((name))}`),
				VarsEnvs:     []string{"PREFIX"},
				EnvironFunc: func() []string {
					return []string{"PREFIX_name=Bob"}
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(MatchYAML(`name: Bob`))
		})

		It("errors with an invalid environment variable definition", func() {
			_, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: ((name))}`),
				VarsEnvs:     []string{"PREFIX"},
				EnvironFunc: func() []string {
					return []string{"PREFIX_name"}
				},
			})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Expected environment variable to be key-value pair"))
		})

		It("errors when the environment variable is invalid YAML", func() {
			_, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: ((name))}`),
				VarsEnvs:     []string{"PREFIX"},
				EnvironFunc: func() []string {
					return []string{"PREFIX_name={]"}
				},
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`Deserializing YAML from environment variable 'PREFIX_name'`))
		})

		It("modifies a number if it has been quoted", func() {
			contents, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{age: ((age))}`),
				VarsEnvs:     []string{"PREFIX"},
				EnvironFunc: func() []string {
					return []string{`PREFIX_age="123"`}
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(MatchYAML(`age: "123"`))
		})

		It("handles multiline environment variables", func() {
			contents, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: ((name))}`),
				VarsEnvs:     []string{"PREFIX"},
				EnvironFunc: func() []string {
					return []string{"PREFIX_name=some\nmulti\nline\nvalue"}
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(MatchYAML(`name: "some\nmulti\nline\nvalue"`))
		})

		It("handles environment variables that are objects", func() {
			contents, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: ((name))}`),
				VarsEnvs:     []string{"PREFIX"},
				EnvironFunc: func() []string {
					return []string{"PREFIX_name={some: value}"}
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(MatchYAML(`name: {some: value}`))
		})

		It("allows vars files to override environment variables", func() {
			contents, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{age: ((age))}`),
				VarsFiles:    []string{writeFile(`age: 456`)},
				VarsEnvs:     []string{"PREFIX"},
				EnvironFunc: func() []string {
					return []string{`PREFIX_age="123"`}
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(MatchYAML(`age: 456`))
		})
	})

	When("vars files are specified", func() {
		It("supports loading vars from files", func() {
			contents, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: ((username))}`),
				VarsFiles:    []string{writeFile(`username: Bob`)},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(MatchYAML(`name: Bob`))
		})

		It("applies vars files on left to right precedence", func() {
			contents, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: ((username))}`),
				VarsFiles: []string{
					writeFile(`username: Bob`),
					writeFile(`username: Susie`),
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(MatchYAML(`name: Susie`))
		})

		It("allows individual vars to override vars files", func() {
			contents, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: ((username))}`),
				VarsFiles:    []string{writeFile(`username: Bob`)},
				Vars:         []string{"username=Susie"},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(MatchYAML(`name: Susie`))
		})
	})

	When("vars are specified", func() {
		It("supports loading individual vars", func() {
			contents, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: ((username))}`),
				Vars:         []string{`username=Bob`},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(MatchYAML(`name: Bob`))
		})

		It("supports YAML in vars", func() {
			contents, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: ((username))}`),
				Vars:         []string{`username={foo: bar}`},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(MatchYAML(`name: {foo: bar}`))
		})

		It("handles multiline variables", func() {
			contents, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: ((name))}`),
				Vars:         []string{"name=some\nmulti\nline\nvalue"},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(MatchYAML(`name: "some\nmulti\nline\nvalue"`))
		})

		It("returns an error if a var does not have '='", func() {
			_, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: ((username))}`),
				Vars:         []string{`username`},
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected var 'username' to be in format 'name=value'"))
		})
	})

	When("ops files are specified", func() {
		It("supports ops file modifications", func() {
			contents, err := interpolate.Execute(interpolate.Options{
				TemplateFile: writeFile(`{name: Susie}`),
				OpsFiles:     []string{writeFile(`[{type: replace, path: /name, value: Bob}]`)},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(MatchYAML(`name: Bob`))
		})
	})

	It("fails when no variables are provided", func() {
		_, err := interpolate.Execute(interpolate.Options{
			TemplateFile:  writeFile(`{name: ((username))}`),
			ExpectAllKeys: true,
		})
		Expect(err).To(MatchError(ContainSubstring("Expected to find variables: username")))
	})
})

func writeFile(contents string) string {
	file, err := ioutil.TempFile("", "")
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(file.Name(), []byte(contents), 0777)
	Expect(err).ToNot(HaveOccurred())
	return file.Name()
}
