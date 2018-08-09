package helper

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

type FileMD5SumCalculator struct{}

func NewFileMD5SumCalculator() FileMD5SumCalculator {
	return FileMD5SumCalculator{}
}

func (c FileMD5SumCalculator) Checksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileMD5 := md5.New()

	_, err = io.Copy(fileMD5, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", fileMD5.Sum(nil)), nil
}
