package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
)

var _ = Describe("S3Client", func() {
	Context("GetAllProductVersions", func() {
		It("returns versions matching the slug", func() {
			itemsList := []mockItem{
				newMockItem("product-slug-1.0.0-alpha.preview+123.github_somefile-0.0.1.zip"),
				newMockItem("product-slug-1.1.1_somefile-0.0.2.zip"),
				newMockItem("another-slug-1.2.3_somefile-0.0.3.zip"),
				newMockItem("another-slug-1.1.1_somefile-0.0.4.zip"),
			}
			stower := newMockStower(itemsList)
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}

			client, err := commands.NewS3Client(stower, config)
			Expect(err).ToNot(HaveOccurred())

			versions, err := client.GetAllProductVersions("product-slug")
			Expect(err).ToNot(HaveOccurred())

			Expect(versions).To(Equal([]string{
				"1.0.0-alpha.preview+123.github",
				"1.1.1",
			}))
		})

		It("does not include multiple copies of the same version", func() {
			itemsList := []mockItem{
				newMockItem("product-slug-1.0.0-alpha.preview+123.github_somefile-0.0.1.zip"),
				newMockItem("product-slug-1.1.1_somefile-0.0.2.zip"),
				newMockItem("product-slug-1.1.1_someotherfile-0.0.2.zip"),
				newMockItem("another-slug-1.2.3_somefile-0.0.3.zip"),
				newMockItem("another-slug-1.1.1_somefile-0.0.4.zip"),
			}

			stower := newMockStower(itemsList)
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}

			client, err := commands.NewS3Client(stower, config)
			Expect(err).ToNot(HaveOccurred())

			versions, err := client.GetAllProductVersions("product-slug")
			Expect(err).ToNot(HaveOccurred())

			Expect(versions).To(Equal([]string{
				"1.0.0-alpha.preview+123.github",
				"1.1.1",
			}))
		})
	})

	Context("GetLatestProductFile", func() {
		It("returns a file artifact", func() {
			itemsList := []mockItem{
				newMockItem("product-slug-1.0.0-pcf-vsphere-2.1-build.341.ova"),
				newMockItem("product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList)
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}

			client, err := commands.NewS3Client(stower, config)
			Expect(err).ToNot(HaveOccurred())

			fileArtifact, err := client.GetLatestProductFile("product-slug", "1.1.1", "*vsphere*ova")
			Expect(err).ToNot(HaveOccurred())
			Expect(fileArtifact.Name).To(Equal("product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova"))
		})

		It("errors when two files match the same glob", func() {
			itemsList := []mockItem{
				newMockItem("product-slug-1.0.0-pcf-vsphere-2.1-build.341.ova"),
				newMockItem("product-slug-1.1.1-pcf-vsphere-2.1-build.345.ova"),
				newMockItem("product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList)
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}

			client, err := commands.NewS3Client(stower, config)
			Expect(err).ToNot(HaveOccurred())

			_, err = client.GetLatestProductFile("product-slug", "1.1.1", "*vsphere*ova")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the glob '*vsphere*ova' matches multiple files. Write your glob to match exactly one of the following"))
		})

		It("errors when zero files match the same glob", func() {
			itemsList := []mockItem{
				newMockItem("product-slug-1.0.0-pcf-vsphere-2.1-build.341.ova"),
				newMockItem("product-slug-1.1.1-pcf-vsphere-2.1-build.345.ova"),
				newMockItem("product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList)
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}

			client, err := commands.NewS3Client(stower, config)
			Expect(err).ToNot(HaveOccurred())

			_, err = client.GetLatestProductFile("product-slug", "1.1.1", "*.zip")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the glob '*.zip' matches no file"))
		})
	})

	Context("DownloadProductToFile", func() {
		var file *os.File
		var fileContents = "hello world"

		BeforeEach(func() {
			var err error
			file, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = file.WriteString(fileContents)
			Expect(err).NotTo(HaveOccurred())

			err = file.Close()
		})

		AfterEach(func() {
			err := os.Remove(file.Name())
			Expect(err).ToNot(HaveOccurred())
		})

		Context("DownloadFile", func() {
			It("receives the contents of a file if the file exists", func() {
				stower := &fakes.Stower{}
				location := &fakes.Location{}
				container := &fakes.Container{}
				stower.DialReturns(location, nil)
				location.ContainerReturns(container, nil)

				item := newMockItem(file.Name())
				item.fakeFileName = file.Name()
				container.ItemReturns(item, nil)

				config := commands.S3Configuration{
					Bucket:          "bucket",
					AccessKeyID:     "access-key-id",
					SecretAccessKey: "secret-access-key",
					RegionName:      "region",
					Endpoint:        "endpoint",
				}
				client, err := commands.NewS3Client(stower, config)
				Expect(err).ToNot(HaveOccurred())

				readCloser, err := client.DownloadFile(file.Name())
				Expect(err).ToNot(HaveOccurred())

				b, err := ioutil.ReadAll(readCloser)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(Equal(fileContents))
			})

			It("errors when cannot open file", func() {
				stower := &fakes.Stower{}
				location := &fakes.Location{}
				container := &fakes.Container{}
				stower.DialReturns(location, nil)
				location.ContainerReturns(container, nil)

				item := newMockItem(file.Name())
				item.fileError = errors.New("could not open file")
				container.ItemReturns(item, nil)

				config := commands.S3Configuration{
					Bucket:          "bucket",
					AccessKeyID:     "access-key-id",
					SecretAccessKey: "secret-access-key",
					RegionName:      "region",
					Endpoint:        "endpoint",
				}
				client, err := commands.NewS3Client(stower, config)
				Expect(err).ToNot(HaveOccurred())

				_, err = client.DownloadFile(file.Name())
				Expect(err).To(HaveOccurred())
			})
		})

		It("writes to a file when the file exists", func() {
			stower := &fakes.Stower{}
			location := &fakes.Location{}
			container := &fakes.Container{}
			stower.DialReturns(location, nil)
			location.ContainerReturns(container, nil)

			item := newMockItem(file.Name())
			item.fakeFileName = file.Name()
			container.ItemReturns(item, nil)

			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}
			client, err := commands.NewS3Client(stower, config)
			Expect(err).ToNot(HaveOccurred())

			file, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			err = client.DownloadProductToFile(&commands.FileArtifact{Name: "don't care"}, file)
			Expect(err).ToNot(HaveOccurred())

			contents, err := ioutil.ReadFile(file.Name())
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(Equal([]byte(fileContents)))
		})
	})

	Context("Property Validation", func() {
		DescribeTable("lists missing required properties", func(param string) {
			stower := &fakes.Stower{}
			config := commands.S3Configuration{}
			_, err := commands.NewS3Client(stower, config)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("Field validation for '%s' failed on the 'required' tag", param))
		},
			Entry("requires Bucket", "Bucket"),
			Entry("requires AccessKeyID", "AccessKeyID"),
			Entry("requires SecretAccessKey", "SecretAccessKey"),
			Entry("requires RegionName", "RegionName"),
			Entry("requires Endpoint", "Endpoint"),
		)

		It("defaults optional properties", func() {
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}
			stower := &fakes.Stower{}
			client, err := commands.NewS3Client(stower, config)
			Expect(err).ToNot(HaveOccurred())

			Expect(client.Config.DisableSSL).To(BeFalse())
			Expect(client.Config.SkipSSLVerification).To(BeFalse())
			Expect(client.Config.UseV2Signing).To(BeFalse())
		})
	})

	It("returns an error on storer failure", func() {
		dialError := errors.New("dial error")
		itemsList := []mockItem{}
		stower := newMockStower(itemsList)
		stower.dialError = dialError

		config := commands.S3Configuration{
			Bucket:          "bucket",
			AccessKeyID:     "access-key-id",
			SecretAccessKey: "secret-access-key",
			RegionName:      "region",
			Endpoint:        "endpoint",
		}

		client, err := commands.NewS3Client(stower, config)
		Expect(err).ToNot(HaveOccurred())

		_, err = client.GetAllProductVersions("product-slug")
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(dialError))
	})
})

type mockStower struct {
	itemsList      []mockItem
	dialError      error
	containerError error
	itemError      error
}

func newMockStower(itemsList []mockItem) mockStower {
	return mockStower{
		itemsList: itemsList,
	}
}

func (s mockStower) Dial(kind string, config commands.Config) (commands.Location, error) {
	if s.dialError != nil {
		return nil, s.dialError
	}

	return &fakes.Location{}, nil
}

func (s mockStower) Container(id string) (commands.Container, error) {
	if s.containerError != nil {
		return nil, s.containerError
	}
	return &fakes.Container{}, nil
}

func (s mockStower) Item(id string) (commands.Item, error) {
	if s.itemError != nil {
		return nil, s.itemError
	}
	return &fakes.Item{}, nil
}

func (s mockStower) Walk(container commands.Container, prefix string, pageSize int, fn commands.WalkFunc) error {
	for _, item := range s.itemsList {
		fn(item, nil)
	}

	return nil
}

type mockItem struct {
	idString     string
	fakeFileName string
	fileError    error
}

func newMockItem(idString string) mockItem {
	return mockItem{
		idString: idString,
	}
}

func (m mockItem) Open() (io.ReadCloser, error) {
	if m.fileError != nil {
		return nil, m.fileError
	}

	if m.fakeFileName != "" {
		reader, err := os.Open(m.fakeFileName)
		Expect(err).ToNot(HaveOccurred())
		return ioutil.NopCloser(reader), nil
	}

	return nil, nil
}

func (m mockItem) ID() string {
	return m.idString
}
