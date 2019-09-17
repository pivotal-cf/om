package download_clients_test

import (
	"archive/zip"
	"github.com/graymeta/stow"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/download_clients"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"text/template"
)

var _ = Describe("S3Client", func() {
	Describe("GetAllProductVersions", func() {
		When("configuring s3", func() {
			It("can support v2 signing", func() {
				itemsList := []mockItem{
					newMockItem("[product-slug,1.1.1]somefile-0.0.2.zip"),
				}
				stower := newMockStower(itemsList)
				config := download_clients.S3Configuration{
					Bucket:          "bucket",
					AccessKeyID:     "access-key-id",
					SecretAccessKey: "secret-access-key",
					RegionName:      "region",
					Endpoint:        "endpoint",
					EnableV2Signing: true,
				}

				client, err := download_clients.NewS3Client(stower, config, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				_, err = client.GetAllProductVersions("product-slug")
				Expect(err).ToNot(HaveOccurred())

				Expect(stower.config).ToNot(BeNil())
				actualValue, _ := stower.config.Config("v2_signing")
				Expect(actualValue).To(Equal("true"))
			})
		})
	})

	Describe("property validation and defaults", func() {
		DescribeTable("required property validation", func(param string) {
			stower := &mockStower{}
			config := download_clients.S3Configuration{}
			_, err := download_clients.NewS3Client(stower, config, GinkgoWriter)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("Field validation for '%s' failed on the 'required' tag", param))
		},
			Entry("requires Bucket", "Bucket"),
			Entry("requires RegionName", "RegionName"),
		)

		It("defaults optional properties", func() {
			config := download_clients.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}
			stower := &mockStower{itemsList: []mockItem{}}
			client, err := download_clients.NewS3Client(stower, config, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			retrievedDisableSSLValue, retrievedValuePresence := client.Config.Config("disable_ssl")
			Expect(retrievedValuePresence).To(Equal(true))
			Expect(retrievedDisableSSLValue).To(Equal("false"))

			retrievedAuthTypeValue, retrievedValuePresence := client.Config.Config("auth_type")
			Expect(retrievedValuePresence).To(Equal(true))
			Expect(retrievedAuthTypeValue).To(Equal("accesskey"))
		})

		When("both region and endpoint are given", func() {
			It("returns an error if they do not match", func() {
				config := download_clients.S3Configuration{
					Bucket:          "bucket",
					AccessKeyID:     "access-key-id",
					SecretAccessKey: "secret-access-key",
					RegionName:      "wrongRegion",
					Endpoint:        "endpoint",
				}
				stower := &mockStower{itemsList: []mockItem{}}
				_, err := download_clients.NewS3Client(stower, config, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("AuthType is set", func() {
			var config download_clients.S3Configuration
			BeforeEach(func() {
				config = download_clients.S3Configuration{
					Bucket:          "bucket",
					AccessKeyID:     "access-key-id",
					SecretAccessKey: "secret-access-key",
					RegionName:      "wrongRegion",
					Endpoint:        "endpoint",
					AuthType:        "fakeAuthType",
				}
			})

			It("passes the auth_type down to stow", func() {
				stower := &mockStower{itemsList: []mockItem{}}
				client, err := download_clients.NewS3Client(stower, config, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				retrievedAuthTypeValue, retrievedValuePresence := client.Config.Config("auth_type")
				Expect(retrievedValuePresence).To(Equal(true))
				Expect(retrievedAuthTypeValue).To(Equal("fakeAuthType"))

			})

			When("AuthType is 'iam' and the id/secret are not provided", func() {
				BeforeEach(func() {
					config.AuthType = "iam"
					config.AccessKeyID = ""
					config.SecretAccessKey = ""
				})

				It("does not raise a validation error", func() {
					stower := &mockStower{itemsList: []mockItem{}}
					_, err := download_clients.NewS3Client(stower, config, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())
				})
			})
			When("AuthType is accesskey/default and the id/secret are not provided", func() {
				BeforeEach(func() {
					config.AuthType = "accesskey"
					config.AccessKeyID = ""
					config.SecretAccessKey = ""

				})

				It("raises a validation error", func() {
					stower := &mockStower{itemsList: []mockItem{}}
					_, err := download_clients.NewS3Client(stower, config, GinkgoWriter)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	It("returns an error on stower failure", func() {
		dialError := errors.New("dial error")
		itemsList := []mockItem{{}}
		stower := newMockStower(itemsList)
		stower.dialError = dialError

		config := download_clients.S3Configuration{
			Bucket:          "bucket",
			AccessKeyID:     "access-key-id",
			SecretAccessKey: "secret-access-key",
			RegionName:      "region",
			Endpoint:        "endpoint",
		}

		client, err := download_clients.NewS3Client(stower, config, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		_, err = client.GetAllProductVersions("product-slug")
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(dialError))
	})
})

type mockStower struct {
	itemsList     []mockItem
	location      mockLocation
	dialCallCount int
	dialError     error
	config        download_clients.StowConfiger
}

func newMockStower(itemsList []mockItem) *mockStower {
	return &mockStower{
		itemsList: itemsList,
	}
}

func (s *mockStower) Dial(kind string, config download_clients.StowConfiger) (stow.Location, error) {
	s.config = config
	s.dialCallCount++
	if s.dialError != nil {
		return nil, s.dialError
	}

	return s.location, nil
}

func (s *mockStower) Walk(container stow.Container, prefix string, pageSize int, fn stow.WalkFunc) error {
	for _, item := range s.itemsList {
		_ = fn(item, nil)
	}

	return nil
}

type mockLocation struct {
	io.Closer
	container      *mockContainer
	containerError error
}

func (m mockLocation) CreateContainer(name string) (stow.Container, error) {
	return mockContainer{}, nil
}
func (m mockLocation) Containers(prefix string, cursor string, count int) ([]stow.Container, string, error) {
	return []stow.Container{mockContainer{}}, "", nil
}
func (m mockLocation) Container(id string) (stow.Container, error) {
	if m.containerError != nil {
		return nil, m.containerError
	}
	return m.container, nil
}
func (m mockLocation) RemoveContainer(id string) error {
	return nil
}
func (m mockLocation) ItemByURL(url *url.URL) (stow.Item, error) {
	return mockItem{}, nil
}

type mockContainer struct {
	item mockItem
}

func (m mockContainer) ID() string {
	return ""
}
func (m mockContainer) Name() string {
	return ""
}
func (m mockContainer) Item(id string) (stow.Item, error) {
	return m.item, nil
}
func (m mockContainer) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	return []stow.Item{mockItem{}}, "", nil
}
func (m mockContainer) RemoveItem(id string) error {
	return nil
}
func (m mockContainer) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
	return mockItem{}, nil
}

type mockItem struct {
	stow.Item
	idString  string
	fileError error
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

	return ioutil.NopCloser(strings.NewReader("hello world")), nil
}

func (m mockItem) ID() string {
	return m.idString
}

func (m mockItem) Size() (int64, error) {
	return 0, nil
}

func createPivotalFile(productFileName, stemcellName, stemcellVersion string) string {
	tempfile, err := ioutil.TempFile("", productFileName)
	Expect(err).NotTo(HaveOccurred())

	zipper := zip.NewWriter(tempfile)
	file, err := zipper.Create("metadata/props.yml")
	Expect(err).NotTo(HaveOccurred())

	contents, err := ioutil.ReadFile("./fixtures/example-product-metadata.yml")
	Expect(err).NotTo(HaveOccurred())

	context := struct {
		StemcellName    string
		StemcellVersion string
	}{
		StemcellName:    stemcellName,
		StemcellVersion: stemcellVersion,
	}

	tmpl, err := template.New("example-product").Parse(string(contents))
	Expect(err).NotTo(HaveOccurred())

	err = tmpl.Execute(file, context)
	Expect(err).NotTo(HaveOccurred())

	Expect(zipper.Close()).NotTo(HaveOccurred())
	return tempfile.Name()
}
