package download_clients

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
