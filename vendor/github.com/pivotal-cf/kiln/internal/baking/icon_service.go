package baking

import (
	"encoding/base64"
	"io/ioutil"
)

type IconService struct {
	logger logger
}

func NewIconService(logger logger) IconService {
	return IconService{
		logger: logger,
	}
}

func (is IconService) Encode(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	is.logger.Println("Encoding icon...")

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(contents), nil
}
