package download_clients

import (
	"github.com/pivotal-cf/om/extractor"
	"os"
)

//counterfeiter:generate -o ./fakes/file_artifacter.go --fake-name FileArtifacter . FileArtifacter
type FileArtifacter interface {
	Name() string
	SHA256() string
	ProductMetadata() (*extractor.Metadata, error)
}

//counterfeiter:generate -o ./fakes/stemcell_artifacter.go --fake-name StemcellArtifacter . StemcellArtifacter
type StemcellArtifacter interface {
	Slug() string
	Version() string
}

//counterfeiter:generate -o ./fakes/product_downloader_service.go --fake-name ProductDownloader . ProductDownloader
type ProductDownloader interface {
	Name() string
	GetAllProductVersions(slug string) ([]string, error)
	GetLatestProductFile(slug, version, glob string) (FileArtifacter, error)
	DownloadProductToFile(fa FileArtifacter, file *os.File) error
	GetLatestStemcellForProduct(fa FileArtifacter, downloadedProductFileName string) (StemcellArtifacter, error)
}
