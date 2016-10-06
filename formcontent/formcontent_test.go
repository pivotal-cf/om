package formcontent_test

import (
	"io/ioutil"
	"os"

	"github.com/pivotal-cf/om/formcontent"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Form", func() {
	Describe("Create", func() {
		var fileWithContent string

		BeforeEach(func() {
			handle, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = handle.WriteString("some content")
			Expect(err).NotTo(HaveOccurred())

			fileWithContent = handle.Name()
		})

		AfterEach(func() {
			os.Remove(fileWithContent)
		})

		It("assembles a ContentSubmission out of the provided file", func() {
			form := formcontent.NewForm("something[file]")

			submission, err := form.Create(fileWithContent)
			Expect(err).NotTo(HaveOccurred())

			Expect(submission.Length).To(Equal(int64(264)))
			Expect(submission.ContentType).NotTo(BeEmpty())

			content, err := ioutil.ReadAll(submission.Content)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("some content"))
		})

		Context("when the file provided is empty", func() {
			It("returns an error", func() {
				emptyFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				form := formcontent.NewForm("something[file]")

				_, err = form.Create(emptyFile.Name())
				Expect(err).To(MatchError("file provided has no content"))
			})
		})

		Context("when an error occurs", func() {
			Context("when the original file cannot be read", func() {
				It("returns an error", func() {
					err := os.Remove(fileWithContent)
					Expect(err).NotTo(HaveOccurred())

					form := formcontent.NewForm("something[file]")

					_, err = form.Create(fileWithContent)
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})
	})
})
