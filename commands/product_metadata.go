package commands

import (
	"errors"
	"fmt"
	"github.com/pivotal-cf/om/extractor"
)

type ProductMetadata struct {
	stdout  logger
	Options struct {
		ProductPath    string `long:"product-path" short:"p"   required:"true" description:"path to product file"`
		ProductName    bool   `long:"product-name"  description:"show product name"`
		ProductVersion bool   `long:"product-version"  description:"show product version"`
	}
}

func NewProductMetadata(stdout logger) *ProductMetadata {
	return newProductMetadata(stdout)
}

func newProductMetadata(stdout logger) *ProductMetadata {
	return &ProductMetadata{stdout: stdout}
}

func (t ProductMetadata) Execute(args []string) error {
	if !t.Options.ProductName && !t.Options.ProductVersion {
		return errors.New("you must specify product-name and/or product-version")
	}

	metadataExtractor := extractor.NewMetadataExtractor()
	metadata, err := metadataExtractor.ExtractFromFile(t.Options.ProductPath)
	if err != nil {
		return fmt.Errorf("failed to getting metadata: %s", err)
	}

	if t.Options.ProductName {
		t.stdout.Println(metadata.Name)
	}

	if t.Options.ProductVersion {
		t.stdout.Println(metadata.Version)
	}

	return nil
}
