package extractor

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"regexp"

	"gopkg.in/yaml.v2"
)

var metadataRegexp = regexp.MustCompile(`^(.*\/)?metadata/.*\.yml`)

// maxMetadataSize caps how many decompressed bytes are read from a tile's
// metadata entry. It guards against decompression bombs: a tiny, highly
// compressed zip entry that would otherwise expand to gigabytes in memory.
// Real tile metadata is a few MB at most, so 128 MiB leaves ample headroom.
const maxMetadataSize = 128 << 20 // 128 MiB

func (me *MetadataExtractor) ExtractFromFile(productPath string) (*Metadata, error) {
	zipReader, err := zip.OpenReader(productPath)
	if err != nil {
		return nil, err
	}

	defer zipReader.Close()

	return fromZipFiles(zipReader.File)
}

func fromZipFiles(files []*zip.File) (*Metadata, error) {
	for _, file := range files {
		matched := metadataRegexp.MatchString(file.Name)

		if matched {
			metadata, err := captureMetadata(file)
			if err != nil {
				return nil, err
			}

			return metadata, nil
		}
	}

	return nil, errors.New("no metadata file was found in provided .pivotal")
}

func captureMetadata(file *zip.File) (*Metadata, error) {
	metadataFile, err := file.Open()
	if err != nil {
		return nil, err
	}

	defer metadataFile.Close()

	// Read at most maxMetadataSize+1 bytes. Reading one extra byte lets us tell
	// "exactly at the limit" (allowed) from "over the limit" (rejected), and
	// stops decompression early instead of inflating the whole entry.
	contents, err := io.ReadAll(io.LimitReader(metadataFile, maxMetadataSize+1))
	if err != nil {
		return nil, err
	}

	if len(contents) > maxMetadataSize {
		return nil, fmt.Errorf("metadata file %q exceeds maximum allowed size of %d bytes", file.Name, maxMetadataSize)
	}

	metadata := &Metadata{Raw: contents}
	err = yaml.Unmarshal(contents, &metadata)
	if err != nil {
		return nil, fmt.Errorf("could not extract product metadata: %s", err)
	}

	if metadata.Name == "" || metadata.Version == "" {
		return nil, errors.New("could not extract product metadata: could not find product details in metadata file")
	}

	return metadata, nil
}
