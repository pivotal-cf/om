package builder

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type ReleaseManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	File    string `json:"file"`
	SHA1    string `json: "sha1"`
}

type ReleaseManifestReader struct{}

func NewReleaseManifestReader() ReleaseManifestReader {
	return ReleaseManifestReader{}
}

func (r ReleaseManifestReader) Read(releaseTarball string) (Part, error) {
	file, err := os.Open(releaseTarball)
	if err != nil {
		return Part{}, err
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return Part{}, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return Part{}, fmt.Errorf("could not find release.MF in %q", releaseTarball)
			}

			return Part{}, fmt.Errorf("error while reading %q: %s", releaseTarball, err)
		}

		if filepath.Base(header.Name) == "release.MF" {
			break
		}
	}

	var releaseManifest ReleaseManifest
	releaseManifestContents, err := ioutil.ReadAll(tr)
	if err != nil {
		return Part{}, err // NOTE: cannot replicate this error scenario in a test
	}

	err = yaml.Unmarshal(releaseManifestContents, &releaseManifest)
	if err != nil {
		return Part{}, err
	}

	releaseManifest.File = filepath.Base(releaseTarball)

	_, err = file.Seek(0, 0)
	if err != nil {
		return Part{}, err // NOTE: cannot replicate this error scenario in a test
	}

	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return Part{}, err // NOTE: cannot replicate this error scenario in a test
	}

	releaseManifest.SHA1 = fmt.Sprintf("%x", hash.Sum(nil))

	return Part{
		Name:     releaseManifest.Name,
		Metadata: releaseManifest,
	}, nil
}
