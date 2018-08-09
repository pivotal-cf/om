package helper

import (
	"io"
	"os"
	"path/filepath"
)

var OpenFile = os.OpenFile

type Filesystem struct{}

func NewFilesystem() Filesystem {
	return Filesystem{}
}

func (f Filesystem) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

func (f Filesystem) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (f Filesystem) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

func (f Filesystem) Remove(path string) error {
	return os.Remove(path)
}
