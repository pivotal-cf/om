package api

import (
	"io"
	"io/ioutil"
)

func SetReadAll(f func(io.Reader) ([]byte, error)) {
	readAll = f
}

func ResetReadAll() {
	readAll = ioutil.ReadAll
}
