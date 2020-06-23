package metadata

import (
	"archive/zip"
	"errors"
	"io"
	"regexp"
)

var metadataRegexp = regexp.MustCompile(`metadata/.*\.yml`)

func ExtractMetadataFromZip(zipreader *zip.Reader) ([]byte, error) {
	for _, file := range zipreader.File {
		if metadataRegexp.MatchString(file.Name) {
			data := make([]byte, file.UncompressedSize64)
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			_, err = io.ReadFull(rc, data)
			if err != nil {
				return nil, err
			}
			rc.Close()
			return data, nil
		}
	}
	return nil, errors.New("no metadata file was found in provided .pivotal")
}
