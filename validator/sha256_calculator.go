package validator

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

type FileSHA256HashCalculator struct{}

func NewSHA256Calculator() FileSHA256HashCalculator {
	return FileSHA256HashCalculator{}
}

func (c FileSHA256HashCalculator) Checksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	digest := sha256.New()
	_, err = io.Copy(digest, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}
