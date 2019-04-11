package download_clients

type PivnetFileArtifact struct {
	name          string
	sha256        string
	slug          string
	releaseID     int
	productFileID int
}

func (f PivnetFileArtifact) Name() string {
	return f.name
}

func (f PivnetFileArtifact) SHA256() string {
	return f.sha256
}

type s3FileArtifact struct {
	name   string
	sha256 string
}

func (f s3FileArtifact) Name() string {
	return f.name
}

func (f s3FileArtifact) SHA256() string {
	return f.sha256
}

type stemcell struct {
	slug    string
	version string
}

func (s stemcell) Slug() string {
	return s.slug
}

func (s stemcell) Version() string {
	return s.version
}
