package baking

import (
	"os"
	"path/filepath"
	"regexp"
)

type ReleasesService struct {
	logger logger
	reader partReader
}

func NewReleasesService(logger logger, reader partReader) ReleasesService {
	return ReleasesService{
		logger: logger,
		reader: reader,
	}
}

func (s ReleasesService) FromDirectories(directories []string) (map[string]interface{}, error) {
	s.logger.Println("Reading release manifests...")

	var tarballs []string
	for _, directory := range directories {
		err := filepath.Walk(directory, filepath.WalkFunc(func(path string, _ os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if match, _ := regexp.MatchString("tgz$|tar.gz$", path); match {
				tarballs = append(tarballs, path)
			}

			return nil
		}))

		if err != nil {
			return nil, err
		}
	}

	manifests := map[string]interface{}{}
	for _, tarball := range tarballs {
		manifest, err := s.reader.Read(tarball)
		if err != nil {
			return nil, err
		}

		manifests[manifest.Name] = manifest.Metadata
	}

	return manifests, nil
}
