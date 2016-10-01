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
	formFile string
}

type ContentSubmission struct {
	Length      int64
	Content     io.ReadCloser
	ContentType string
}

func NewForm(formFile string) Form {
	return Form{
		formFile: formFile,
	}
}

func (f Form) Create(path string) (ContentSubmission, error) {
	originalContent, err := os.Open(path)
	if err != nil {
		return ContentSubmission{}, err
	}

	defer originalContent.Close()

	stats, err := originalContent.Stat()
	if err != nil {
		return ContentSubmission{}, err
	}

	if stats.Size() == 0 {
		return ContentSubmission{}, errors.New("file provided has no content")
	}

	body, err := ioutil.TempFile("", "")
	if err != nil {
		return ContentSubmission{}, err
	}

	multipartWriter := multipart.NewWriter(body)

	formFile, err := multipartWriter.CreateFormFile(f.formFile, filepath.Base(path))
	if err != nil {
		return ContentSubmission{}, err
	}

	_, err = io.Copy(formFile, originalContent)
	if err != nil {
		return ContentSubmission{}, err
	}

	err = multipartWriter.Close()
	if err != nil {
		return ContentSubmission{}, err
	}

	_, err = body.Seek(0, 0)
	if err != nil {
		return ContentSubmission{}, err
	}

	stats, err = body.Stat()
	if err != nil {
		return ContentSubmission{}, err
	}

	return ContentSubmission{
		Length:      stats.Size(),
		Content:     body,
		ContentType: multipartWriter.FormDataContentType(),
	}, nil
}
