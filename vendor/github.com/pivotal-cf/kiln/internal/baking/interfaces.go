package baking

import "github.com/pivotal-cf/kiln/builder"

//go:generate counterfeiter -o ./fakes/logger.go --fake-name Logger . logger
type logger interface {
	Println(v ...interface{})
}

//go:generate counterfeiter -o ./fakes/part_reader.go --fake-name PartReader . partReader
type partReader interface {
	Read(path string) (builder.Part, error)
}

//go:generate counterfeiter -o ./fakes/directory_reader.go --fake-name DirectoryReader . directoryReader
type directoryReader interface {
	Read(path string) ([]builder.Part, error)
}
