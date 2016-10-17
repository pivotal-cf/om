package formcontent_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/formcontent"
)

var _ = Describe("Form", func() {

	Describe("AddFile", func() {
		var fileWithContent1 string
		var fileWithContent2 string
		var form formcontent.Form

		BeforeEach(func() {
			handle1, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = handle1.WriteString("some content")
			Expect(err).NotTo(HaveOccurred())

			fileWithContent1 = handle1.Name()

			handle2, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = handle2.WriteString("some more content")
			Expect(err).NotTo(HaveOccurred())

			fileWithContent2 = handle2.Name()

			form, err = formcontent.NewForm()
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.Remove(fileWithContent1)
			os.Remove(fileWithContent2)
		})

		It("writes out the provided file as a multipart form using the writer", func() {
			err := form.AddFile("something[file1]", fileWithContent1)
			Expect(err).NotTo(HaveOccurred())

			err = form.AddFile("something[file2]", fileWithContent2)
			Expect(err).NotTo(HaveOccurred())

			submission, err := form.Finalize()
			Expect(err).NotTo(HaveOccurred())

			content, err := ioutil.ReadAll(submission.Content)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(ContainSubstring("name=\"something[file1]\""))
			Expect(string(content)).To(ContainSubstring("some content"))
			Expect(string(content)).To(ContainSubstring("name=\"something[file2]\""))
			Expect(string(content)).To(ContainSubstring("some more content"))
		})

		Context("when the file provided is empty", func() {
			It("returns an error", func() {
				emptyFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				form, err := formcontent.NewForm()
				Expect(err).NotTo(HaveOccurred())

				err = form.AddFile("foo", emptyFile.Name())
				Expect(err).To(MatchError("file provided has no content"))
			})
		})

		Context("when an error occurs", func() {
			Context("when the original file cannot be read", func() {
				It("returns an error", func() {
					form, err := formcontent.NewForm()
					Expect(err).NotTo(HaveOccurred())

					err = form.AddFile("foo", "/file/does/not/exist")
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})
	})

	Describe("AddField", func() {
		var form formcontent.Form

		BeforeEach(func() {
			var err error
			form, err = formcontent.NewForm()
			Expect(err).NotTo(HaveOccurred())
		})

		It("writes out the provided fields into the multipart form using the writer", func() {
			err := form.AddField("key1", "value1")
			Expect(err).NotTo(HaveOccurred())

			err = form.AddField("key2", "value2")
			Expect(err).NotTo(HaveOccurred())

			err = form.AddField("key3", "value3")
			Expect(err).NotTo(HaveOccurred())

			submission, err := form.Finalize()
			Expect(err).NotTo(HaveOccurred())

			content, err := ioutil.ReadAll(submission.Content)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(ContainSubstring("name=\"key1\""))
			Expect(string(content)).To(ContainSubstring("value1"))
			Expect(string(content)).To(ContainSubstring("name=\"key2\""))
			Expect(string(content)).To(ContainSubstring("value2"))
			Expect(string(content)).To(ContainSubstring("name=\"key3\""))
			Expect(string(content)).To(ContainSubstring("value3"))
		})
	})

	Describe("Finalize", func() {
		var form formcontent.Form

		BeforeEach(func() {
			var err error
			form, err = formcontent.NewForm()
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns a content submission which includes the correct length and content type", func() {
			err := form.AddField("key1", "value1")
			Expect(err).NotTo(HaveOccurred())

			submission, err := form.Finalize()
			Expect(err).NotTo(HaveOccurred())

			Expect(submission.Length).To(Equal(int64(185)))
			Expect(submission.ContentType).To(ContainSubstring("multipart/form-data"))
		})

	})
})
