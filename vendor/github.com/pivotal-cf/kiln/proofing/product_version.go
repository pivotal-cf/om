package proofing

type ProductVersion struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`

	// TODO: validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/product_version.rb#L8-L15
}
