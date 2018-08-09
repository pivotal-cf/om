package builder

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Zipper struct {
	writer *zip.Writer
}

func NewZipper() Zipper {
	return Zipper{}
}

func (z *Zipper) SetWriter(writer io.Writer) {
	z.writer = zip.NewWriter(writer)
}

func (z Zipper) Add(path string, file io.Reader) error {
	if z.writer == nil {
		return errors.New("zipper path must be set")
	}

	return z.add(&zip.FileHeader{
		Name:   path,
		Method: zip.Deflate,
	}, file)
}

func (z Zipper) AddWithMode(path string, file io.Reader, mode os.FileMode) error {
	if z.writer == nil {
		return errors.New("zipper path must be set")
	}

	fh := &zip.FileHeader{
		Name:   path,
		Method: zip.Deflate,
	}
	fh.SetMode(mode)

	return z.add(fh, file)
}

func (z Zipper) add(fh *zip.FileHeader, file io.Reader) error {
	f, err := z.writer.CreateHeader(fh)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, file)
	if err != nil {
		return err
	}

	return nil
}

func (z Zipper) Close() error {
	return z.writer.Close()
}

func (z Zipper) CreateFolder(path string) error {
	if z.writer == nil {
		return errors.New("zipper path must be set")
	}

	path = fmt.Sprintf("%s%c", filepath.Clean(path), filepath.Separator)

	_, err := z.writer.Create(path)
	if err != nil {
		return err
	}

	return err
}
