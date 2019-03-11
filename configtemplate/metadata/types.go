package metadata

type Provider interface {
	MetadataBytes() ([]byte, error)
}
