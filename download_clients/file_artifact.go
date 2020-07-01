package download_clients

import "github.com/pivotal-cf/go-pivnet/v4"

type PivnetFileArtifact struct {
	slug          string
	releaseID     int
	productFile   pivnet.ProductFile
}

func (f PivnetFileArtifact) Name() string {
	return f.productFile.AWSObjectKey
}

func (f PivnetFileArtifact) SHA256() string {
	return f.productFile.SHA256
}

type stowFileArtifact struct {
	name   string
	sha256 string
}

func (f stowFileArtifact) Name() string {
	return f.name
}

func (f stowFileArtifact) SHA256() string {
	return f.sha256
}
