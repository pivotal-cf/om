package sha256sum

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

type FileSummer struct {
}

func NewFileSummer() *FileSummer {
	return &FileSummer{}
}

func (f FileSummer) SumFile(filepath string) (string, error) {
	fileToSum, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer fileToSum.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, fileToSum)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
