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

	file, err := zip.OpenReader(t.Options.ProductPath)
	if err != nil {
		return fmt.Errorf("failed to open product file: %s", err)
	}
	defer file.Close()

	for _, f := range file.File {
		matched, err := regexp.MatchString("metadata/.*", f.Name)
		if err != nil {
			return fmt.Errorf("failed to match file name regex: %s", err)
		}

		if matched {
			meta, err := f.Open()
			if err != nil {
				return fmt.Errorf("failed to open metadata file: %s", err)
			}

			type DecodedFile struct {
				ProductName    string `yaml:"name"`
				ProductVersion string `yaml:"product_version"`
			}
			var v DecodedFile
			err = yaml.NewDecoder(meta).Decode(&v)
			if err != nil {
				return fmt.Errorf("failed to decode metadata file: %s", err)
			}

			if t.Options.ProductName {
				t.stdout.Println(v.ProductName)
			}

			if t.Options.ProductVersion {
				t.stdout.Println(v.ProductVersion)
			}

			return nil
		}
	}

	return errors.New("failed to find metadata file")
}

func (t TileMetadata) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command prints metadata about the given tile",
		ShortDescription: "prints tile metadata",
		Flags:            t.Options,
	}
}
