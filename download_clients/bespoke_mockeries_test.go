package download_clients_test

import (
	"io"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/graymeta/stow"
	"github.com/pivotal-cf/om/download_clients"
)

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
