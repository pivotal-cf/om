package builder_test

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"time"

	"github.com/pivotal-cf/kiln/builder"
	"github.com/pivotal-cf/kiln/builder/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StemcellManifestReader", func() {
	var (
		filesystem *fakes.Filesystem
		reader     builder.StemcellManifestReader
	)

	BeforeEach(func() {
		filesystem = &fakes.Filesystem{}
		reader = builder.NewStemcellManifestReader(filesystem)

		tarball := NewBuffer(bytes.NewBuffer([]byte{}))
		gw := gzip.NewWriter(tarball)
		tw := tar.NewWriter(gw)

		stemcellManifest := bytes.NewBuffer([]byte(`---
version: '9999'
operating_system: centOS
`))

		header := &tar.Header{
			Name:    "./stemcell.MF",
			Size:    int64(stemcellManifest.Len()),
			Mode:    int64(0644),
			ModTime: time.Now(),
		}

		err := tw.WriteHeader(header)
		Expect(err).NotTo(HaveOccurred())

		_, err = io.Copy(tw, stemcellManifest)
		Expect(err).NotTo(HaveOccurred())

		err = tw.Close()
		Expect(err).NotTo(HaveOccurred())

		err = gw.Close()
		Expect(err).NotTo(HaveOccurred())

		filesystem.OpenReturns(tarball, nil)
	})

	Describe("Read", func() {
		It("extracts the stemcell manifest information from the tarball", func() {
			stemcellManifest, err := reader.Read("/path/to/stemcell/tarball")
			Expect(err).NotTo(HaveOccurred())
			Expect(stemcellManifest).To(Equal(builder.Part{
				Metadata: builder.StemcellManifest{
					Version:         "9999",
					OperatingSystem: "centOS",
				},
			}))

			Expect(filesystem.OpenArgsForCall(0)).To(Equal("/path/to/stemcell/tarball"))
		})

		Context("failure cases", func() {
			Context("when the tarball cannot be opened", func() {
				It("returns an error", func() {
					filesystem.OpenReturns(nil, errors.New("failed to open tarball"))

					_, err := reader.Read("/path/to/stemcell/tarball")
					Expect(err).To(MatchError("failed to open tarball"))
				})
			})

			Context("when reading from the tarball errors", func() {
				It("returns an error", func() {
					erroringReader := &fakes.ReadCloser{}
					erroringReader.ReadReturns(0, errors.New("cannot read tarball"))
					filesystem.OpenStub = func(name string) (io.ReadCloser, error) {
						return erroringReader, nil
					}

					_, err := reader.Read("/path/to/stemcell/tarball")
					Expect(err).To(MatchError("cannot read tarball"))
					Expect(erroringReader.CloseCallCount()).To(Equal(1))
				})
			})

			Context("when the input is not a valid gzip", func() {
				It("returns an error", func() {
					filesystem.OpenReturns(NewBuffer(bytes.NewBuffer([]byte("I am a banana!"))), nil)

					_, err := reader.Read("/path/to/stemcell/tarball")
					Expect(err).To(MatchError("gzip: invalid header"))
				})
			})

			Context("when the header file is corrupt", func() {
				It("returns an error", func() {
					tarball := NewBuffer(bytes.NewBuffer([]byte{}))
					gw := gzip.NewWriter(tarball)
					tw := tar.NewWriter(gw)

					err := tw.Close()
					Expect(err).NotTo(HaveOccurred())

					err = gw.Close()
					Expect(err).NotTo(HaveOccurred())
					filesystem.OpenReturns(tarball, nil)

					_, err = reader.Read("/path/to/stemcell/tarball")
					Expect(err).To(MatchError("could not find stemcell.MF in \"/path/to/stemcell/tarball\""))
				})
			})

			Context("when there is no stemcell.MF", func() {
				It("returns an error", func() {
					tarball := NewBuffer(bytes.NewBuffer([]byte{}))
					gw := gzip.NewWriter(tarball)
					tw := tar.NewWriter(gw)

					stemcellManifest := bytes.NewBuffer([]byte(`---
version: '9999'
operating_system: centOS
`))

					header := &tar.Header{
						Name:    "./someotherfile.MF",
						Size:    int64(stemcellManifest.Len()),
						Mode:    int64(0644),
						ModTime: time.Now(),
					}

					err := tw.WriteHeader(header)
					Expect(err).NotTo(HaveOccurred())

					_, err = io.Copy(tw, stemcellManifest)
					Expect(err).NotTo(HaveOccurred())

					err = tw.Close()
					Expect(err).NotTo(HaveOccurred())

					err = gw.Close()
					Expect(err).NotTo(HaveOccurred())

					filesystem.OpenReturns(tarball, nil)
					_, err = reader.Read("/path/to/stemcell/tarball")
					Expect(err).To(MatchError("could not find stemcell.MF in \"/path/to/stemcell/tarball\""))
				})
			})

			Context("when the tarball is corrupt", func() {
				It("returns an error", func() {
					tarball := NewBuffer(bytes.NewBuffer([]byte{}))
					gw := gzip.NewWriter(tarball)
					tw := bufio.NewWriter(gw)

					_, err := tw.WriteString("I am a banana!")
					Expect(err).NotTo(HaveOccurred())

					err = tw.Flush()
					Expect(err).NotTo(HaveOccurred())

					err = gw.Close()
					Expect(err).NotTo(HaveOccurred())

					filesystem.OpenReturns(tarball, nil)
					_, err = reader.Read("/path/to/stemcell/tarball")
					Expect(err).To(MatchError("error while reading \"/path/to/stemcell/tarball\": unexpected EOF"))
				})
			})

			Context("when the stemcell manifest is not YAML", func() {
				It("returns an error", func() {
					tarball := NewBuffer(bytes.NewBuffer([]byte{}))
					gw := gzip.NewWriter(tarball)
					tw := tar.NewWriter(gw)

					stemcellManifest := bytes.NewBuffer([]byte(`%%%%%`))

					header := &tar.Header{
						Name:    "./stemcell.MF",
						Size:    int64(stemcellManifest.Len()),
						Mode:    int64(0644),
						ModTime: time.Now(),
					}

					err := tw.WriteHeader(header)
					Expect(err).NotTo(HaveOccurred())

					_, err = io.Copy(tw, stemcellManifest)
					Expect(err).NotTo(HaveOccurred())

					err = tw.Close()
					Expect(err).NotTo(HaveOccurred())

					err = gw.Close()
					Expect(err).NotTo(HaveOccurred())

					filesystem.OpenReturns(tarball, nil)

					_, err = reader.Read("/path/to/stemcell/tarball")
					Expect(err).To(MatchError("yaml: could not find expected directive name"))
				})
			})
		})
	})
})
