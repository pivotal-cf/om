package extractor

import "net/http"

type Option func(*MetadataExtractor)

type MetadataExtractor struct{
	httpClient httpClient
}

func NewMetadataExtractor(options ...Option) *MetadataExtractor {
	me := &MetadataExtractor{
		httpClient: http.DefaultClient,
	}

	for _, o := range options {
		o(me)
	}

	return me
}
