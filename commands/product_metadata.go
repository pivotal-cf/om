package commands

import (
	"errors"
	"fmt"
	"github.com/pivotal-cf/om/configtemplate/generator"
	"github.com/pivotal-cf/om/configtemplate/metadata"
)

type ProductMetadata struct {
	buildProvider productMetadataBuildProvider
	stdout        logger
	Options       struct {
		InterpolateOptions interpolateConfigFileOptions `group:"config file interpolation"`

		ProductPath string `long:"product-path" short:"p" description:"path to product file"`

		ProductName    bool `long:"product-name"    description:"show product name"`
		ProductVersion bool `long:"product-version" description:"show product version"`

		PivnetApiToken       string `long:"pivnet-api-token"`
		PivnetProductSlug    string `long:"pivnet-product-slug"    description:"the product name in pivnet"`
		PivnetProductVersion string `long:"pivnet-product-version" description:"the version of the product from which to generate a template"`
		PivnetHost           string `long:"pivnet-host"            description:"the API endpoint for Pivotal Network"                        default:"https://network.pivotal.io"`
		FileGlob             string `long:"file-glob" short:"f"    description:"a glob to match exactly one file in the pivnet product slug" default:"*.pivotal"`
		PivnetDisableSSL     bool   `long:"pivnet-disable-ssl"     description:"whether to disable ssl validation when contacting the Pivotal Network"`
	}
}

var DefaultProductMetadataProvider = func() func(c *ProductMetadata) MetadataProvider {
	return func(c *ProductMetadata) MetadataProvider {
		options := c.Options
		if options.ProductPath != "" {
			return metadata.NewFileProvider(options.ProductPath)
		}
		return metadata.NewPivnetProvider(options.PivnetHost, options.PivnetApiToken, options.PivnetProductSlug, options.PivnetProductVersion, options.FileGlob, options.PivnetDisableSSL)
	}
}

type productMetadataBuildProvider func(*ProductMetadata) MetadataProvider

func NewProductMetadata(bp productMetadataBuildProvider, stdout logger) *ProductMetadata {
	return newProductMetadata(bp, stdout)
}

func newProductMetadata(bp productMetadataBuildProvider, stdout logger) *ProductMetadata {
	return &ProductMetadata{
		buildProvider: bp,
		stdout: stdout,
	}
}

func (t ProductMetadata) Execute(args []string) error {
	err := t.Validate()
	if err != nil {
		return err
	}

	metadataSource := t.newMetadataSource()
	metadataBytes, err := metadataSource.MetadataBytes()
	if err != nil {
		return fmt.Errorf("error getting metadata for %s at version %s: %s", t.Options.PivnetProductSlug, t.Options.PivnetProductVersion, err)
	}

	meta, err := generator.NewMetadata(metadataBytes)
	if err != nil {
		return err
	}

	if t.Options.ProductName {
		t.stdout.Println(meta.Name)
	}

	if t.Options.ProductVersion {
		t.stdout.Println(meta.Version)
	}

	return nil
}

func (t *ProductMetadata) newMetadataSource() (metadataSource MetadataProvider) {
	return t.buildProvider(t)
}

func (t *ProductMetadata) Validate() error {
	if !t.Options.ProductName && !t.Options.ProductVersion {
		return errors.New("you must specify product-name and/or product-version")
	}

	if t.Options.PivnetApiToken != "" && t.Options.PivnetProductSlug != "" && t.Options.PivnetProductVersion != "" && t.Options.ProductPath == "" {
		return nil
	}

	if t.Options.PivnetApiToken == "" && t.Options.PivnetProductSlug == "" && t.Options.PivnetProductVersion == "" && t.Options.ProductPath != "" {
		return nil
	}

	return errors.New("cannot load tile metadata: please provide either pivnet flags OR product-path")
}
