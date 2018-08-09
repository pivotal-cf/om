package builder_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pivotal-cf/kiln/builder"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MetadataPartsDirectoryReader", func() {
	var (
		reader  builder.MetadataPartsDirectoryReader
		tempDir string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "metadata-parts-directory-reader")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Read", func() {
		BeforeEach(func() {
			err := ioutil.WriteFile(filepath.Join(tempDir, "vars-file-1.yml"), []byte(`---
- name: variable-1
  type: certificate
- name: variable-2
  alias: variable-2-alias
  type: user
`), 0755)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(filepath.Join(tempDir, "vars-file-2.yml"), []byte(`---
name: variable-3
type: password
`), 0755)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(filepath.Join(tempDir, "ignores.any-other-extension"), []byte("not-yaml"), 0755)
			Expect(err).ToNot(HaveOccurred())

			reader = builder.NewMetadataPartsDirectoryReader()
		})

		It("reads the contents of each yml file in the directory", func() {
			vars, err := reader.Read(tempDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(vars).To(Equal([]builder.Part{
				{
					File: "vars-file-1.yml",
					Name: "variable-1",
					Metadata: map[interface{}]interface{}{
						"name": "variable-1",
						"type": "certificate",
					},
				},
				{
					File: "vars-file-1.yml",
					Name: "variable-2-alias",
					Metadata: map[interface{}]interface{}{
						"name": "variable-2",
						"type": "user",
					},
				},
				{
					File: "vars-file-2.yml",
					Name: "variable-3",
					Metadata: map[interface{}]interface{}{
						"name": "variable-3",
						"type": "password",
					},
				},
			}))
		})

		Context("when the directory does not exist", func() {
			It("returns an error", func() {
				_, err := reader.Read("/dir/that/does/not/exist")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("/dir/that/does/not/exist"))
			})
		})

		Context("when there is an error reading from a file", func() {
			It("errors", func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "unreadable-file.yml"), []byte(`unused`), 0000)
				Expect(err).ToNot(HaveOccurred())
				_, err = reader.Read(tempDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unreadable-file.yml"))
			})
		})

		Context("when a yaml file is malformed", func() {
			It("errors", func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "not-valid-yaml.yml"), []byte(`{{`), 0755)
				Expect(err).ToNot(HaveOccurred())

				_, err = reader.Read(tempDir)
				Expect(err).To(MatchError(ContainSubstring("cannot unmarshal")))
			})
		})

		Context("when file contains an array item without a name", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "vars-file-1.yml"), []byte(`[{foo: bar}]`), 0755)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := reader.Read(tempDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("name"))
			})
		})

		Context("when variable file contains a map item without a name", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "vars-file-1.yml"), []byte(`{foo: bar}`), 0755)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := reader.Read(tempDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("name"))
			})
		})
	})

	Context("when a top-level key is specified", func() {
		BeforeEach(func() {
			err := ioutil.WriteFile(filepath.Join(tempDir, "_order.yml"), []byte(`---
variable_order:
- variable-3
- variable-2
- variable-1
`), 0755)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(filepath.Join(tempDir, "vars-file-1.yml"), []byte(`---
variables:
- name: variable-1
  type: certificate
- name: variable-2
  type: user
`), 0755)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(filepath.Join(tempDir, "vars-file-2.yml"), []byte(`---
variables:
- name: variable-3
  type: password
`), 0755)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(filepath.Join(tempDir, "ignores.any-other-extension"), []byte("not-yaml"), 0755)
			Expect(err).ToNot(HaveOccurred())

			reader = builder.NewMetadataPartsDirectoryReaderWithTopLevelKey("variables")
		})

		Describe("Read", func() {
			It("reads the contents of each yml file in the directory", func() {
				vars, err := reader.Read(tempDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).To(Equal([]builder.Part{
					{
						File: "vars-file-1.yml",
						Name: "variable-1",
						Metadata: map[interface{}]interface{}{
							"name": "variable-1",
							"type": "certificate",
						},
					},
					{
						File: "vars-file-1.yml",
						Name: "variable-2",
						Metadata: map[interface{}]interface{}{
							"name": "variable-2",
							"type": "user",
						},
					},
					{
						File: "vars-file-2.yml",
						Name: "variable-3",
						Metadata: map[interface{}]interface{}{
							"name": "variable-3",
							"type": "password",
						},
					},
				}))
			})
		})

		Context("failure cases", func() {
			Context("when a yaml file does not contain the top-level key", func() {
				It("errors", func() {
					err := ioutil.WriteFile(filepath.Join(tempDir, "not-a-vars-file.yml"), []byte(`constants: []`), 0755)
					Expect(err).ToNot(HaveOccurred())

					_, err = reader.Read(tempDir)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(MatchRegexp(`not a variables file: .*not-a-vars-file\.yml`))
				})
			})
		})

		Context("when variable file is neither a slice or a map", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "vars-file-1.yml"), []byte(`variables: foo`), 0755)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := reader.Read(tempDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("either slice or map"))
			})
		})

		Context("when variable file contains an invalid item", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "vars-file-1.yml"), []byte(`variables: [foo]`), 0755)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := reader.Read(tempDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("must be a map"))
			})
		})

		Context("when variable file contains an array item without a name", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "vars-file-1.yml"), []byte(`variables: [{foo: bar}]`), 0755)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := reader.Read(tempDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("name"))
			})
		})

		Context("when variable file contains a map item without a name", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "vars-file-1.yml"), []byte(`variables: {foo: bar}`), 0755)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := reader.Read(tempDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("name"))
			})
		})

		Context("when specifying an Order key", func() {
			BeforeEach(func() {
				reader = builder.NewMetadataPartsDirectoryReaderWithOrder("variables", "variable_order")
			})

			It("returns the contents of the files in the directory sorted by _order.yml", func() {
				vars, err := reader.Read(tempDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).To(Equal([]builder.Part{
					{
						File: "vars-file-2.yml",
						Name: "variable-3",
						Metadata: map[interface{}]interface{}{
							"name": "variable-3",
							"type": "password",
						},
					},
					{
						File: "vars-file-1.yml",
						Name: "variable-2",
						Metadata: map[interface{}]interface{}{
							"name": "variable-2",
							"type": "user",
						},
					},
					{
						File: "vars-file-1.yml",
						Name: "variable-1",
						Metadata: map[interface{}]interface{}{
							"name": "variable-1",
							"type": "certificate",
						},
					},
				},
				))
			})

			Context("failure cases", func() {
				Context("when _order.yml file cannot be read", func() {
					BeforeEach(func() {
						err := os.RemoveAll(filepath.Join(tempDir, "_order.yml"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns an error", func() {
						_, err := reader.Read(tempDir)
						Expect(err.Error()).To(ContainSubstring("_order.yml"))
					})
				})

				Context("when _order.yml file is not in valid format", func() {
					BeforeEach(func() {
						err := ioutil.WriteFile(filepath.Join(tempDir, "_order.yml"), []byte(`variable_order: foo`), 0755)
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns an error", func() {
						_, err := reader.Read(tempDir)
						Expect(err.Error()).To(ContainSubstring("Invalid format"))
					})
				})

				Context("when _order.yml file does not have the specified orderKey", func() {
					BeforeEach(func() {
						reader = builder.NewMetadataPartsDirectoryReaderWithOrder("variables", "bad_order_key")
					})

					It("returns an error", func() {
						_, err := reader.Read(tempDir)
						Expect(err.Error()).To(ContainSubstring("bad_order_key"))
					})
				})

				Context("when _order.yml file contains a name that does not exist", func() {
					BeforeEach(func() {
						err := ioutil.WriteFile(filepath.Join(tempDir, "_order.yml"), []byte(`variable_order: [some-file]`), 0755)
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns an error", func() {
						_, err := reader.Read(tempDir)
						Expect(err.Error()).To(ContainSubstring("some-file"))
					})
				})
			})
		})
	})
})
