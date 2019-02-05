package commands_test

import (
	"bytes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"io/ioutil"
)

var _ = FDescribe("S3Client", func() {
	Context("GetAllProductVersions", func() {
		It("returns versions matching the slug", func() {
			storer := &fakes.BlobStorer{}
			client := commands.NewS3Client(storer)
			storer.ListFilesReturns(
				[]string{
					"product-slug-1.0.0-alpha.preview+123.github_somefile-0.0.1.zip",
					"product-slug-1.1.1_somefile-0.0.2.zip",
					"another-slug-1.2.3_somefile-0.0.3.zip",
					"another-slug-1.1.1_somefile-0.0.4.zip",
				},
				nil)

			versions, err := client.GetAllProductVersions("product-slug")
			Expect(err).ToNot(HaveOccurred())

			Expect(versions).To(Equal([]string{
				"1.0.0-alpha.preview+123.github",
				"1.1.1",
			}))
		})

		It("does not include multiple copies of the same version", func() {
			storer := &fakes.BlobStorer{}
			client := commands.NewS3Client(storer)
			storer.ListFilesReturns(
				[]string{
					"product-slug-1.0.0-alpha.preview+123.github_somefile-0.0.1.zip",
					"product-slug-1.1.1_somefile-0.0.2.zip",
					"product-slug-1.1.1_someotherfile-0.0.2.zip",
					"another-slug-1.2.3_somefile-0.0.3.zip",
					"another-slug-1.1.1_somefile-0.0.4.zip",
				},
				nil)

			versions, err := client.GetAllProductVersions("product-slug")
			Expect(err).ToNot(HaveOccurred())

			Expect(versions).To(Equal([]string{
				"1.0.0-alpha.preview+123.github",
				"1.1.1",
			}))
		})

		It("returns an error on storer failure", func() {
			storer := &fakes.BlobStorer{}
			client := commands.NewS3Client(storer)
			storer.ListFilesReturns(
				nil,
				errors.New("some error"),
				)

			_, err := client.GetAllProductVersions("product-slug")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("GetLatestProductFile", func() {
		It("returns a file artifact", func() {
			storer := &fakes.BlobStorer{}
			client := commands.NewS3Client(storer)
			storer.ListFilesReturns(
				[]string{
					"product-slug-1.0.0-pcf-vsphere-2.1-build.341.ova",
					"product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova",
				},
				nil,
			)

			fileArtifact, err := client.GetLatestProductFile("product-slug", "1.1.1", "*vsphere*ova")
			Expect(err).ToNot(HaveOccurred())
			Expect(fileArtifact.Name).To(Equal("product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova"))
		})

		It("errors when two files match the same glob", func() {
			storer := &fakes.BlobStorer{}
			client := commands.NewS3Client(storer)
			storer.ListFilesReturns(
				[]string{
					"product-slug-1.0.0-pcf-vsphere-2.1-build.341.ova",
					"product-slug-1.1.1-pcf-vsphere-2.1-build.345.ova",
					"product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova",
				},
				nil,
			)

			_, err := client.GetLatestProductFile("product-slug", "1.1.1", "*vsphere*ova")
			Expect(err).To(HaveOccurred())
		})

		It("errors when zero files match the same glob", func() {
			storer := &fakes.BlobStorer{}
			client := commands.NewS3Client(storer)
			storer.ListFilesReturns(
				[]string{
					"product-slug-1.0.0-pcf-vsphere-2.1-build.341.ova",
					"product-slug-1.1.1-pcf-vsphere-2.1-build.345.ova",
					"product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova",
				},
				nil,
			)

			_, err := client.GetLatestProductFile("product-slug", "1.1.1", "*.zip")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("DownloadProductToFile", func() {
		It("writes to a file when the file exists", func() {
			storer := &fakes.BlobStorer{}
			client := commands.NewS3Client(storer)

			blob := ioutil.NopCloser(bytes.NewReader([]byte("hello world")))
			storer.DownloadFileReturns(blob, nil)

			file, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			err = client.DownloadProductToFile(&commands.FileArtifact{Name: "don't care"}, file)
			Expect(err).ToNot(HaveOccurred())

			contents, err := ioutil.ReadFile(file.Name())
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(Equal([]byte("hello world")))
		})

		It("errors when the file does not exist", func() {
			storer := &fakes.BlobStorer{}
			client := commands.NewS3Client(storer)

			storer.DownloadFileReturns(nil, errors.New("some error"))

			file, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			err = client.DownloadProductToFile(&commands.FileArtifact{Name: "don't care"}, file)
			Expect(err).To(HaveOccurred())
		})
	})
})
