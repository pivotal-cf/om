package builder

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type TileWriter struct {
	filesystem       filesystem
	zipper           zipper
	logger           logger
	md5SumCalculator md5SumCalculator
}

//go:generate counterfeiter -o ./fakes/filesystem.go --fake-name Filesystem . filesystem
type filesystem interface {
	Create(path string) (io.WriteCloser, error)
	Open(path string) (io.ReadCloser, error)
	Walk(root string, walkFn filepath.WalkFunc) error
	Remove(path string) error
}

type md5SumCalculator interface {
	Checksum(path string) (string, error)
}

//go:generate counterfeiter -o ./fakes/zipper.go --fake-name Zipper . zipper

type zipper interface {
	SetWriter(writer io.Writer)
	Add(path string, file io.Reader) error
	AddWithMode(path string, file io.Reader, mode os.FileMode) error
	CreateFolder(path string) error
	Close() error
}

//go:generate counterfeiter -o ./fakes/file_info.go --fake-name FileInfo . fileinfo

type fileinfo interface {
	Name() string
	Size() int64
	Mode() os.FileMode
	ModTime() time.Time
	IsDir() bool
	Sys() interface{}
}

func NewTileWriter(filesystem filesystem, zipper zipper, logger logger, md5SumCalculator md5SumCalculator) TileWriter {
	return TileWriter{
		filesystem:       filesystem,
		zipper:           zipper,
		logger:           logger,
		md5SumCalculator: md5SumCalculator,
	}
}

type WriteInput struct {
	OutputFile           string
	StubReleases         bool
	MigrationDirectories []string
	ReleaseDirectories   []string
	EmbedPaths           []string
}

func (w TileWriter) Write(generatedMetadataContents []byte, input WriteInput) error {
	w.logger.Printf("Building %s...", input.OutputFile)

	f, err := w.filesystem.Create(input.OutputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	w.zipper.SetWriter(f)

	err = w.addToZipper(filepath.Join("metadata", "metadata.yml"), bytes.NewBuffer(generatedMetadataContents), input.OutputFile)
	if err != nil {
		w.removeOutputFile(input.OutputFile)
		return err
	}

	err = w.addMigrations(input.MigrationDirectories, input.OutputFile)
	if err != nil {
		w.removeOutputFile(input.OutputFile)
		return err
	}

	err = w.addReleases(input.ReleaseDirectories, input.StubReleases, input.OutputFile)
	if err != nil {
		w.removeOutputFile(input.OutputFile)
		return err
	}

	err = w.addEmbeddedPaths(input.EmbedPaths, input.OutputFile)
	if err != nil {
		w.removeOutputFile(input.OutputFile)
		return err
	}

	err = w.zipper.Close()
	if err != nil {
		w.removeOutputFile(input.OutputFile)
		return err
	}

	w.logger.Printf("Calculating md5 sum of %s...", input.OutputFile)
	md5Sum, err := w.md5SumCalculator.Checksum(input.OutputFile)
	if err != nil {
		return err
	}

	w.logger.Printf("Calculated md5 sum: %s", md5Sum)

	return nil
}

func (w TileWriter) addReleases(releasesDirs []string, stubReleases bool, outputFile string) error {
	for _, releasesDirectory := range releasesDirs {
		err := w.addReleaseTarballs(releasesDirectory, stubReleases, outputFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w TileWriter) addReleaseTarballs(releasesDir string, stubReleases bool, outputFile string) error {
	return w.filesystem.Walk(releasesDir, func(filePath string, info os.FileInfo, err error) error {
		isTarball, _ := regexp.MatchString("tgz$|tar.gz$", filePath)
		if !isTarball {
			return nil
		}

		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		var file io.ReadCloser = ioutil.NopCloser(strings.NewReader(""))
		if !stubReleases {
			file, err = w.filesystem.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()
		}

		return w.addToZipper(filepath.Join("releases", filepath.Base(filePath)), file, outputFile)
	})
}

func (w TileWriter) addEmbeddedPaths(embedPaths []string, outputFile string) error {
	for _, embedPath := range embedPaths {
		err := w.addEmbeddedPath(embedPath, outputFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w TileWriter) addEmbeddedPath(pathToEmbed, outputFile string) error {
	return w.filesystem.Walk(pathToEmbed, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := w.filesystem.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		relativePath, err := filepath.Rel(pathToEmbed, filePath)
		if err != nil {
			return err //not tested
		}

		entryPath := filepath.Join("embed", filepath.Join(filepath.Base(pathToEmbed), relativePath))
		return w.addToZipperWithMode(entryPath, file, info.Mode(), outputFile)
	})
}

func (w TileWriter) addMigrations(migrationsDir []string, outputFile string) error {
	var found bool

	for _, migrationDir := range migrationsDir {
		err := w.filesystem.Walk(migrationDir, func(filePath string, info os.FileInfo, err error) error {
			isTest, _ := regexp.MatchString(`tests\/`, filePath)
			isJsFile, _ := regexp.MatchString(`.js$`, filePath)
			if isTest || !isJsFile {
				return nil
			}

			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			found = true

			file, err := w.filesystem.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			return w.addToZipper(filepath.Join("migrations", "v1", filepath.Base(filePath)), file, outputFile)
		})

		if err != nil {
			return err
		}
	}

	if !found {
		return w.addEmptyMigrationsDirectory(outputFile)
	}

	return nil
}

func (w TileWriter) addToZipper(path string, contents io.Reader, outputFile string) error {
	w.logger.Printf("Adding %s to %s...", path, outputFile)

	return w.zipper.Add(path, contents)
}

func (w TileWriter) addToZipperWithMode(path string, contents io.Reader, mode os.FileMode, outputFile string) error {
	w.logger.Printf("Adding %s to %s...", path, outputFile)

	return w.zipper.AddWithMode(path, contents, mode)
}

func (w TileWriter) containsMigrations(entries []string) bool {
	migrationsPrefix := filepath.Join("migrations", "v1")
	for _, entry := range entries {
		if strings.HasPrefix(entry, migrationsPrefix) {
			return true
		}
	}
	return false
}

func (w TileWriter) addEmptyMigrationsDirectory(outputFile string) error {
	w.logger.Printf("Creating empty migrations folder in %s...", outputFile)
	err := w.zipper.CreateFolder(filepath.Join("migrations", "v1"))
	if err != nil {
		return err
	}
	return nil
}

func (w TileWriter) removeOutputFile(path string) {
	err := w.filesystem.Remove(path)
	if err != nil {
		w.logger.Printf("failed cleaning up zip %q: %s", path, err.Error())
	}
}
