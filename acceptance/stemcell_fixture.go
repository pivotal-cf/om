package acceptance

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
)

// MinimalStemcellTgz returns a valid minimal stemcell .tgz (gzipped tarball containing stemcell.MF).
// Used by acceptance tests so extractStemcellManifest() succeeds and the full test suite passes
// without -skipPackage acceptance. The manifest uses generic OS/version/infrastructure for compatibility.
func MinimalStemcellTgz() []byte {
	manifestContent := `name: bosh-acceptance-stemcell
version: "621.77"
operating_system: ubuntu-xenial
cloud_properties:
  infrastructure: google-kvm
`
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	_ = tw.WriteHeader(&tar.Header{Name: "stemcell.MF", Size: int64(len(manifestContent))})
	_, _ = tw.Write([]byte(manifestContent))
	_ = tw.Close()
	_ = gw.Close()
	return buf.Bytes()
}
