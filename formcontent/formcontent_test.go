package formcontent_test

import (
	"io"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/formcontent"
)

var _ = Describe("Formcontent", func() {
	var form *formcontent.Form

	Describe("Reset", func() {
		var fileWithContent string

		BeforeEach(func() {
			handle, err := os.CreateTemp("", "")
			Expect(err).ToNot(HaveOccurred())

			_, err = handle.WriteString("some content")
			Expect(err).ToNot(HaveOccurred())

			fileWithContent = handle.Name()

			form = formcontent.NewForm()
			err = form.AddFile("something[original-file]", fileWithContent)
			Expect(err).ToNot(HaveOccurred())
			err = form.AddField("key", "value")
			Expect(err).ToNot(HaveOccurred())

			submission := form.Finalize()
			Expect(submission.ContentLength).ToNot(BeZero())

			// drain the reader to force pipe to be closed
			_, err = io.ReadAll(submission.Content)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := os.Remove(fileWithContent)
			Expect(err).ToNot(HaveOccurred())
		})

		It("resets the fields on the form", func() {
			form.Reset()

			err := form.AddFile("something[other-file]", fileWithContent)
			Expect(err).ToNot(HaveOccurred())

			submission := form.Finalize()
			Expect(submission.ContentLength).ToNot(BeZero())

			contents, err := io.ReadAll(submission.Content)
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(ContainSubstring("other-file"))
			Expect(contents).ToNot(ContainSubstring("original-file"))
		})
	})

	Describe("AddFile", func() {
		var (
			fileWithContent1 string
			fileWithContent2 string
		)

		BeforeEach(func() {
			handle1, err := os.CreateTemp("", "")
			Expect(err).ToNot(HaveOccurred())

			_, err = handle1.WriteString("some content")
			Expect(err).ToNot(HaveOccurred())

			fileWithContent1 = handle1.Name()

			handle2, err := os.CreateTemp("", "")
			Expect(err).ToNot(HaveOccurred())

			_, err = handle2.WriteString("some more content")
			Expect(err).ToNot(HaveOccurred())

			fileWithContent2 = handle2.Name()

			form = formcontent.NewForm()
		})

		AfterEach(func() {
			os.Remove(fileWithContent1)
			os.Remove(fileWithContent2)
		})

		It("writes out the provided file as a multipart form using the writer", func() {
			err := form.AddFile("something[file1]", fileWithContent1)
			Expect(err).ToNot(HaveOccurred())

			err = form.AddFile("something[file2]", fileWithContent2)
			Expect(err).ToNot(HaveOccurred())

			submission := form.Finalize()

			content, err := io.ReadAll(submission.Content)
			Expect(err).ToNot(HaveOccurred())

			Expect(string(content)).To(MatchRegexp(`^--\w+\r\nContent-Disposition: form-data; name=\"something\[file1\]\"; filename=\"\w+\"\r\n` +
				`Content-Type: application/octet-stream\r\n\r\n` +
				`some content` +
				`\r\n--\w+\r\nContent-Disposition: form-data; name=\"something\[file2\]\"; filename=\"\w+\"\r\n` +
				`Content-Type: application/octet-stream\r\n\r\n` +
				`some more content` +
				`\r\n--\w+--\r\n$`))
		})

		When("the file provided is empty", func() {
			It("returns an error", func() {
				emptyFile, err := os.CreateTemp("", "")
				Expect(err).ToNot(HaveOccurred())

				form := formcontent.NewForm()

				err = form.AddFile("foo", emptyFile.Name())
				Expect(err).To(MatchError("file provided has no content"))
			})
		})

		When("an error occurs", func() {
			When("the original file cannot be read", func() {
				It("returns an error", func() {
					form := formcontent.NewForm()

					err := form.AddFile("foo", "/file/does/not/exist")
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})
	})

	Describe("AddField", func() {
		BeforeEach(func() {
			form = formcontent.NewForm()
		})

		It("writes out the provided fields into the multipart form using the writer", func() {
			err := form.AddField("key1", "value1")
			Expect(err).ToNot(HaveOccurred())

			err = form.AddField("key2", "value2")
			Expect(err).ToNot(HaveOccurred())

			submission := form.Finalize()

			content, err := io.ReadAll(submission.Content)
			Expect(err).ToNot(HaveOccurred())

			Expect(string(content)).To(MatchRegexp(`^--\w+\r\nContent-Disposition: form-data; name="key1"\r\n\r\nvalue1` +
				`\r\n--\w+\r\nContent-Disposition: form-data; name="key2"\r\n\r\nvalue2` +
				`\r\n--\w+--\r\n$`))
		})
	})

	Describe("AddCombined", func() {
		var fileWithContent1 string

		BeforeEach(func() {
			var err error

			handle1, err := os.CreateTemp("", "")
			Expect(err).ToNot(HaveOccurred())

			_, err = handle1.WriteString("some content")
			Expect(err).ToNot(HaveOccurred())

			fileWithContent1 = handle1.Name()

			form = formcontent.NewForm()
		})

		AfterEach(func() {
			os.Remove(fileWithContent1)
		})

		It("writes out the provided fields into the multipart form using the writer", func() {
			err := form.AddField("key1", "value1")
			Expect(err).ToNot(HaveOccurred())

			err = form.AddFile("file1", fileWithContent1)
			Expect(err).ToNot(HaveOccurred())

			submission := form.Finalize()

			content, err := io.ReadAll(submission.Content)
			Expect(err).ToNot(HaveOccurred())

			Expect(submission.ContentLength).To(Or(Equal(int64(372)), Equal(int64(373)), Equal(int64(374))))
			Expect(string(content)).To(MatchRegexp(`^--\w+\r\nContent-Disposition: form-data; name=\"file1\"; filename=\"\w+\"\r\n` +
				`Content-Type: application/octet-stream\r\n\r\n` +
				`some content` +
				`\r\n--\w+\r\nContent-Disposition: form-data; name="key1"\r\n\r\nvalue1` +
				`\r\n--\w+--\r\n$`))
		})
	})

	Describe("Finalize", func() {
		var form *formcontent.Form

		BeforeEach(func() {
			form = formcontent.NewForm()
		})

		It("returns a content submission which includes the correct length and content type", func() {
			err := form.AddField("key1", "value1")
			Expect(err).ToNot(HaveOccurred())

			submission := form.Finalize()

			Expect(submission.ContentLength).To(Equal(int64(185)))
			Expect(submission.ContentType).To(ContainSubstring("multipart/form-data"))
		})

	})
})
