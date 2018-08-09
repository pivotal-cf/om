package acceptance

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var pathToMain string

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "acceptance")
}

var _ = BeforeSuite(func() {
	var err error
	pathToMain, err = gexec.Build("github.com/pivotal-cf/kiln")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func createTarball(dir, name, manifestName, manifest string) (string, error) {
	tarball, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return "", err
	}

	gw := gzip.NewWriter(tarball)
	tw := tar.NewWriter(gw)

	err = ioutil.WriteFile(filepath.Join(dir, manifestName), []byte(manifest), 0644)
	if err != nil {
		return "", err
	}

	manifestFile, err := os.Open(filepath.Join(dir, manifestName))
	if err != nil {
		return "", err
	}

	stat, err := manifestFile.Stat()
	if err != nil {
		return "", err
	}

	header := &tar.Header{
		Name:    fmt.Sprintf("./%s", manifestName),
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	err = tw.WriteHeader(header)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(tw, manifestFile)
	if err != nil {
		return "", err
	}

	err = tw.Close()
	if err != nil {
		return "", err
	}

	err = gw.Close()
	if err != nil {
		return "", err
	}

	err = manifestFile.Close()
	if err != nil {
		return "", err
	}

	err = os.RemoveAll(manifestFile.Name())
	if err != nil {
		return "", err
	}

	return tarball.Name(), nil
}
