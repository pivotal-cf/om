package formcontent

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type Form struct {
	contentType string
	boundary    string
	length      int64
	pr          *io.PipeReader
	pw          *io.PipeWriter
	formFields  *bytes.Buffer
	formWriter  *multipart.Writer
	files       []string
	fileKeys    []*bytes.Buffer
}

type ContentSubmission struct {
	Content       io.Reader
	ContentType   string
	ContentLength int64
}

func NewForm() (*Form) {
	buf := &bytes.Buffer{}

	pr, pw := io.Pipe()

	formWriter := multipart.NewWriter(buf)

	return &Form{
		contentType: formWriter.FormDataContentType(),
		boundary:    formWriter.Boundary(),
		pr:          pr,
		pw:          pw,
		formFields:  buf,
		formWriter:  formWriter,
	}
}

func (f *Form) AddField(key string, value string) error {
	fieldWriter, err := f.formWriter.CreateFormField(key)
	if err != nil {
		return err
	}

	_, err = fieldWriter.Write([]byte(value))
	return err
}

func (f *Form) AddFile(key string, path string) error {
	fileLength, err := verifyFile(path)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	fileKey := multipart.NewWriter(buf)
	fileKey.SetBoundary(f.boundary)

	_, err = fileKey.CreateFormFile(key, filepath.Base(path))
	if err != nil {
		return err
	}

	// add the length of form fields, including trailing boundary
	f.length += fileLength
	f.length += int64(buf.Len())

	f.files = append(f.files, path)
	f.fileKeys = append(f.fileKeys, buf)

	return nil
}

func (f *Form) Finalize() (ContentSubmission) {
	f.formWriter.Close()

	// add the length of form fields, including trailing boundary
	f.length += int64(f.formFields.Len())

	// add the length of `\r\n` between fields
	if len(f.files) > 0 {
		f.length += int64(2 * (len(f.files) - 1))
		if f.formFields.Len() > len(f.boundary)+8 {
			f.length += 2
		}
	}

	go f.writeToPipe()

	return ContentSubmission{
		ContentLength: f.length,
		Content:       f.pr,
		ContentType:   f.contentType,
	}
}

func verifyFile(path string) (int64, error) {
	fileContent, err := os.Open(path)
	if err != nil {
		return 0, err
	}

	defer fileContent.Close()

	stats, err := fileContent.Stat()
	if err != nil {
		return 0, err
	}

	if stats.Size() == 0 {
		return 0, errors.New("file provided has no content")
	}

	return stats.Size(), nil
}

func (f *Form) writeToPipe() {
	var err error
	separate := false

	// write files
	for i, key := range f.fileKeys {
		if separate {
			_, err = f.pw.Write([]byte("\r\n"))
			if err != nil {
				f.pw.CloseWithError(err)
				return
			}
		}

		_, err = io.Copy(f.pw, key)
		if err != nil {
			f.pw.CloseWithError(err)
			return
		}

		fileName := f.files[i]
		err = writeFileToPipe(fileName, f.pw)
		if err != nil {
			f.pw.CloseWithError(err)
			return
		}

		separate = true
	}

	// write fields
	if separate && f.formFields.Len() > len(f.boundary)+8 { // boundary+8 =>format: \r\n--boundary-words--\r\n
		_, err = f.pw.Write([]byte("\r\n"))
		if err != nil {
			f.pw.CloseWithError(err)
			return
		}
	}

	_, err = io.Copy(f.pw, f.formFields)
	if err != nil {
		f.pw.CloseWithError(err)
		return
	}

	f.pw.Close()
	return
}

func writeFileToPipe(fileName string, writer *io.PipeWriter) error {
	fileContent, err := os.Open(fileName)
	if err != nil {
		return err
	}

	defer fileContent.Close()

	_, err = io.Copy(writer, fileContent)
	if err != nil {
		return err
	}

	return nil
}
