package commands

import "github.com/pivotal-cf/jhanda"

type Version struct {
	logger  logger
	version string
}

func NewVersion(logger logger, version string) Version {
	return Version{
		logger:  logger,
		version: version,
	}
}

func (v Version) Execute(args []string) error {
	v.logger.Printf("kiln version %s\n", v.version)

	return nil
}

func (v Version) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command prints the kiln release version number.",
		ShortDescription: "prints the kiln release version",
	}
}
