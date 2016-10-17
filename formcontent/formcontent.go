package formcontent

import (
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
)

type Form struct {
	multipartWriter *multipart.Writer
	body            *os.File
}

type ContentSubmission struct {
	Length      int64
	Content     io.Reader
	ContentType string
}

func NewForm() (Form, error) {
	body, err := ioutil.TempFile("", "")
	if err != nil {
		return Form{}, err
	}

	return Form{
		multipartWriter: multipart.NewWriter(body),
		body:            body,
	}, nil
}

func (f Form) Finalize() (ContentSubmission, error) {
	err := f.multipartWriter.Close()
	if err != nil {
		return ContentSubmission{}, err
	}

	_, err = f.body.Seek(0, 0)
	if err != nil {
		return ContentSubmission{}, err
	}

	stats, err := f.body.Stat()
	if err != nil {
		return ContentSubmission{}, err
	}

	return ContentSubmission{
		Length:      stats.Size(),
		Content:     f.body,
		ContentType: f.multipartWriter.FormDataContentType(),
	}, nil
}

func (f Form) AddFile(key, path string) error {
	originalContent, err := os.Open(path)
	if err != nil {
		return err
	}

	defer originalContent.Close()

	stats, err := originalContent.Stat()
	if err != nil {
		return err
	}

	if stats.Size() == 0 {
		return errors.New("file provided has no content")
	}

	formFile, err := f.multipartWriter.CreateFormFile(key, filepath.Base(path))
	if err != nil {
		return err
	}

	_, err = io.Copy(formFile, originalContent)
	if err != nil {
		return err
	}

	return nil
}

func (f Form) AddField(key, value string) error {
	fieldWriter, err := f.multipartWriter.CreateFormField(key)
	if err != nil {
		return err
	}

	fieldWriter.Write([]byte(value))
	return nil
}
