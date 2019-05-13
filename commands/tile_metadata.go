package commands

import (
	"errors"
	"fmt"

	"archive/zip"

	"regexp"

	"github.com/pivotal-cf/jhanda"
	"gopkg.in/yaml.v2"
)

type TileMetadata struct {
	stdout  logger
	Options struct {
		ProductPath    string `long:"product-path" short:"p"   required:"true" description:"path to product file"`
		ProductName    bool   `long:"product-name"  description:"show product name"`
		ProductVersion bool   `long:"product-version"  description:"show product version"`
	}
}

func NewTileMetadata(stdout logger) TileMetadata {
	return TileMetadata{stdout: stdout}
}

func (t TileMetadata) Execute(args []string) error {
	if _, err := jhanda.Parse(&t.Options, args); err != nil {
		return fmt.Errorf("could not parse tile-metadata flags: %s", err)
	}

	if !t.Options.ProductName && !t.Options.ProductVersion {
		return errors.New("you must specify product-name and/or product-version")
	}

	metadata, err := getTileMetadata(t.Options.ProductPath)
	if err != nil {
		return fmt.Errorf("failed to getting metadata: %s", err)
	}

	if t.Options.ProductName {
		t.stdout.Println(metadata.ProductName)
	}

	if t.Options.ProductVersion {
		t.stdout.Println(metadata.ProductVersion)
	}

	return nil
}

func (t TileMetadata) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command prints metadata about the given tile",
		ShortDescription: "prints tile metadata",
		Flags:            t.Options,
	}
}

type metadataPayload struct {
	ProductName    string `yaml:"name"`
	ProductVersion string `yaml:"product_version"`
}

func getTileMetadata(filename string) (*metadataPayload, error) {
	file, err := zip.OpenReader(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open product file '%s': %s", filename, err)
	}
	defer file.Close()

	for _, f := range file.File {
		matched, err := regexp.MatchString(`metadata/.+\.yml`, f.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to match file name regex: %s", err)
		}

		if matched {
			meta, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open metadata file: %s", err)
			}

			var v metadataPayload
			err = yaml.NewDecoder(meta).Decode(&v)
			if err != nil {
				return nil, fmt.Errorf("failed to decode metadata file: %s", err)
			}

			return &v, nil
		}
	}

	return nil, errors.New("failed to find metadata file")
}
