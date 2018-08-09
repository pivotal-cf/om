package builder_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pivotal-cf/kiln/builder"
	"github.com/pivotal-cf/kiln/builder/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("TileWriter", func() {
	var (
		filesystem *fakes.Filesystem
		zipper     *fakes.Zipper
		logger     *fakes.Logger
		md5Calc    *fakes.MD5SumCalculator
		tileWriter builder.TileWriter
		outputFile string

		expectedFile *os.File
	)

	BeforeEach(func() {
		filesystem = &fakes.Filesystem{}
		zipper = &fakes.Zipper{}
		logger = &fakes.Logger{}
		md5Calc = &fakes.MD5SumCalculator{}
		tileWriter = builder.NewTileWriter(filesystem, zipper, logger, md5Calc)
		outputFile = "some-output-dir/cool-product-file-1.2.3-build.4.pivotal"
	})

	Describe("Write", func() {
		BeforeEach(func() {
			expectedFile = &os.File{}
			filesystem.CreateReturns(expectedFile, nil)
		})

		DescribeTable("writes tile to disk", func(stubbed bool, errorWhenAttemptingToOpenRelease error) {
			input := builder.WriteInput{
				ReleaseDirectories:   []string{"/some/path/releases", "/some/other/path/releases"},
				MigrationDirectories: []string{"/some/path/migrations", "/some/other/path/migrations"},
				OutputFile:           outputFile,
				StubReleases:         stubbed,
			}

			dirInfo := &fakes.FileInfo{}
			dirInfo.IsDirReturns(true)

			releaseInfo := &fakes.FileInfo{}
			releaseInfo.IsDirReturns(false)

			migrationInfo := &fakes.FileInfo{}
			migrationInfo.IsDirReturns(false)

			filesystem.WalkStub = func(root string, walkFn filepath.WalkFunc) error {
				switch root {
				case "/some/path/releases":
					walkFn("/some/path/releases", dirInfo, nil)
					walkFn("/some/path/releases/release-1.tgz", releaseInfo, nil)
					walkFn("/some/path/releases/release-2.tgz", releaseInfo, nil)
				case "/some/other/path/releases":
					walkFn("/some/other/path/releases", dirInfo, nil)
					walkFn("/some/other/path/releases/release-3.tgz", releaseInfo, nil)
					walkFn("/some/other/path/releases/release-4.tgz", releaseInfo, nil)
				case "/some/path/migrations":
					walkFn("/some/path/migrations", dirInfo, nil)
					walkFn("/some/path/migrations/migration-1.js", migrationInfo, nil)
					walkFn("/some/path/migrations/migration-2.js", migrationInfo, nil)
					walkFn("/some/path/migrations/tests/migration-2_test.js", migrationInfo, nil)
					walkFn("/some/path/migrations/tests/migration-2.js", migrationInfo, nil)
					walkFn("/some/path/migrations/not-a-js-migration.txt", migrationInfo, nil)
				case "/some/other/path/migrations":
					walkFn("/some/other/path/migrations", dirInfo, nil)
					walkFn("/some/other/path/migrations/other-migration.js", migrationInfo, nil)
				default:
					return nil
				}
				return nil
			}

			filesystem.OpenStub = func(path string) (io.ReadCloser, error) {
				switch path {
				case "/some/path/releases/release-1.tgz":
					return NewBuffer(bytes.NewBuffer([]byte("release-1"))), errorWhenAttemptingToOpenRelease
				case "/some/path/releases/release-2.tgz":
					return NewBuffer(bytes.NewBuffer([]byte("release-2"))), errorWhenAttemptingToOpenRelease
				case "/some/other/path/releases/release-3.tgz":
					return NewBuffer(bytes.NewBuffer([]byte("release-3"))), errorWhenAttemptingToOpenRelease
				case "/some/other/path/releases/release-4.tgz":
					return NewBuffer(bytes.NewBuffer([]byte("release-4"))), errorWhenAttemptingToOpenRelease
				case "/some/path/migrations/migration-1.js":
					return NewBuffer(bytes.NewBuffer([]byte("migration-1"))), nil
				case "/some/path/migrations/migration-2.js":
					return NewBuffer(bytes.NewBuffer([]byte("migration-2"))), nil
				case "/some/path/migrations/tests/migration-2_test.js":
					return NewBuffer(bytes.NewBuffer([]byte("i-am-a-test"))), nil
				case "/some/path/migrations/not-a-js-migration.txt":
					return NewBuffer(bytes.NewBuffer([]byte("i-am-not-a-js-migration"))), nil
				case "/some/path/migrations/tests/migration-2.js":
					return NewBuffer(bytes.NewBuffer([]byte("some-migration"))), nil
				case "/some/other/path/migrations/other-migration.js":
					return NewBuffer(bytes.NewBuffer([]byte("other-migration"))), nil
				default:
					return nil, fmt.Errorf("open %s: no such file or directory", path)
				}
			}

			md5Calc.ChecksumCall.Returns.Sum = "THIS-IS-THE-SUM"

			err := tileWriter.Write([]byte("generated-metadata-contents"), input)
			Expect(err).NotTo(HaveOccurred())

			Expect(zipper.SetWriterCallCount()).To(Equal(1))
			actualFile := zipper.SetWriterArgsForCall(0)
			Expect(actualFile).To(Equal(expectedFile))

			Expect(zipper.AddCallCount()).To(Equal(8))

			path, file := zipper.AddArgsForCall(0)
			Expect(path).To(Equal(filepath.Join("metadata", "metadata.yml")))
			Eventually(gbytes.BufferReader(file)).Should(gbytes.Say("generated-metadata-contents"))

			path, file = zipper.AddArgsForCall(1)
			Expect(path).To(Equal(filepath.Join("migrations", "v1", "migration-1.js")))
			Eventually(gbytes.BufferReader(file)).Should(gbytes.Say("migration-1"))

			path, file = zipper.AddArgsForCall(2)
			Expect(path).To(Equal(filepath.Join("migrations", "v1", "migration-2.js")))
			Eventually(gbytes.BufferReader(file)).Should(gbytes.Say("migration-2"))

			path, file = zipper.AddArgsForCall(3)
			Expect(path).To(Equal(filepath.Join("migrations", "v1", "other-migration.js")))
			Eventually(gbytes.BufferReader(file)).Should(gbytes.Say("other-migration"))

			path, file = zipper.AddArgsForCall(4)
			Expect(path).To(Equal(filepath.Join("releases", "release-1.tgz")))
			checkReleaseFileContent("release-1", stubbed, file)

			path, file = zipper.AddArgsForCall(5)
			Expect(path).To(Equal(filepath.Join("releases", "release-2.tgz")))
			checkReleaseFileContent("release-2", stubbed, file)

			path, file = zipper.AddArgsForCall(6)
			Expect(path).To(Equal(filepath.Join("releases", "release-3.tgz")))
			checkReleaseFileContent("release-3", stubbed, file)

			path, file = zipper.AddArgsForCall(7)
			Expect(path).To(Equal(filepath.Join("releases", "release-4.tgz")))
			checkReleaseFileContent("release-4", stubbed, file)

			Expect(zipper.CloseCallCount()).To(Equal(1))

			Expect(logger.PrintfCall.Receives.LogLines).To(Equal([]string{
				fmt.Sprintf("Building %s...", outputFile),
				fmt.Sprintf("Adding metadata/metadata.yml to %s...", outputFile),
				fmt.Sprintf("Adding migrations/v1/migration-1.js to %s...", outputFile),
				fmt.Sprintf("Adding migrations/v1/migration-2.js to %s...", outputFile),
				fmt.Sprintf("Adding migrations/v1/other-migration.js to %s...", outputFile),
				fmt.Sprintf("Adding releases/release-1.tgz to %s...", outputFile),
				fmt.Sprintf("Adding releases/release-2.tgz to %s...", outputFile),
				fmt.Sprintf("Adding releases/release-3.tgz to %s...", outputFile),
				fmt.Sprintf("Adding releases/release-4.tgz to %s...", outputFile),
				fmt.Sprintf("Calculating md5 sum of %s...", outputFile),
				"Calculated md5 sum: THIS-IS-THE-SUM",
			}))

			Expect(md5Calc.ChecksumCall.CallCount).To(Equal(1))
			Expect(md5Calc.ChecksumCall.Receives.Path).To(Equal("some-output-dir/cool-product-file-1.2.3-build.4.pivotal"))
		},
			Entry("without stubbing releases", false, nil),
			Entry("with stubbed releases", true, errors.New("don't open release")),
		)

		Context("when releases directory is provided", func() {
			BeforeEach(func() {
				dirInfo := &fakes.FileInfo{}
				dirInfo.IsDirReturns(true)

				releaseInfo := &fakes.FileInfo{}
				releaseInfo.IsDirReturns(false)

				filesystem.WalkStub = func(root string, walkFn filepath.WalkFunc) error {
					switch root {
					case "/some/path/releases":
						walkFn("/some/path/releases", dirInfo, nil)
						walkFn("/some/path/releases/release-1.tgz", releaseInfo, nil)
						walkFn("/some/path/releases/release-2.tgz", releaseInfo, nil)
						walkFn("/some/path/releases/not-a-release.txt", releaseInfo, nil)
						walkFn(root, dirInfo, nil)
					case "/some/path/migrations":
						walkFn("/some/path/migrations", dirInfo, nil)
					default:
						return nil
					}

					return nil
				}

				filesystem.OpenStub = func(path string) (io.ReadCloser, error) {
					if path == "/some/path/releases/release-1.tgz" {
						return NewBuffer(bytes.NewBufferString("release-1")), nil
					}

					if path == "/some/path/releases/release-2.tgz" {
						return NewBuffer(bytes.NewBufferString("release-1")), nil
					}

					return nil, nil
				}
			})

			Context("and no migrations are provided", func() {
				It("creates empty migrations/v1 folder", func() {
					input := builder.WriteInput{
						ReleaseDirectories:   []string{"/some/path/releases"},
						MigrationDirectories: []string{},
						OutputFile:           "some-output-dir/cool-product-file-1.2.3-build.4.pivotal",
						StubReleases:         false,
					}

					err := tileWriter.Write([]byte("generated-metadata-contents"), input)
					Expect(err).NotTo(HaveOccurred())

					Expect(logger.PrintfCall.Receives.LogLines).To(Equal([]string{
						fmt.Sprintf("Building %s...", outputFile),
						fmt.Sprintf("Adding metadata/metadata.yml to %s...", outputFile),
						fmt.Sprintf("Creating empty migrations folder in %s...", outputFile),
						fmt.Sprintf("Adding releases/release-1.tgz to %s...", outputFile),
						fmt.Sprintf("Adding releases/release-2.tgz to %s...", outputFile),
						fmt.Sprintf("Calculating md5 sum of %s...", outputFile),
						"Calculated md5 sum: ",
					}))

					Expect(zipper.CreateFolderCallCount()).To(Equal(1))
					Expect(zipper.CreateFolderArgsForCall(0)).To(Equal(filepath.Join("migrations", "v1")))
				})
			})

			Context("and the migrations directory is empty", func() {
				It("creates empty migrations/v1 folder", func() {
					input := builder.WriteInput{
						ReleaseDirectories:   []string{"/some/path/releases"},
						MigrationDirectories: []string{"/some/path/migrations"},
						OutputFile:           "some-output-dir/cool-product-file-1.2.3-build.4.pivotal",
						StubReleases:         false,
					}

					err := tileWriter.Write([]byte("generated-metadata-contents"), input)
					Expect(err).NotTo(HaveOccurred())

					Expect(logger.PrintfCall.Receives.LogLines).To(Equal([]string{
						fmt.Sprintf("Building %s...", outputFile),
						fmt.Sprintf("Adding metadata/metadata.yml to %s...", outputFile),
						fmt.Sprintf("Creating empty migrations folder in %s...", outputFile),
						fmt.Sprintf("Adding releases/release-1.tgz to %s...", outputFile),
						fmt.Sprintf("Adding releases/release-2.tgz to %s...", outputFile),
						fmt.Sprintf("Calculating md5 sum of %s...", outputFile),
						"Calculated md5 sum: ",
					}))

					Expect(zipper.CreateFolderCallCount()).To(Equal(1))
					Expect(zipper.CreateFolderArgsForCall(0)).To(Equal(filepath.Join("migrations", "v1")))
				})
			})
		})

		Context("when a file to embed is provided", func() {
			BeforeEach(func() {
				dirInfo := &fakes.FileInfo{}
				dirInfo.IsDirReturns(true)

				releaseInfo := &fakes.FileInfo{}
				releaseInfo.IsDirReturns(false)

				embedFileInfo := &fakes.FileInfo{}
				embedFileInfo.IsDirReturns(false)
				embedFileInfo.ModeReturns(12345678)

				filesystem.WalkStub = func(root string, walkFn filepath.WalkFunc) error {
					if root == "/some/path/releases" {
						walkFn(root, dirInfo, nil)
						walkFn(filepath.Join(root, "release-1.tgz"), releaseInfo, nil)
						walkFn(filepath.Join(root, "release-2.tgz"), releaseInfo, nil)
					} else if root == "/some/path/to-embed/my-file.txt" {
						walkFn(root, embedFileInfo, nil)
					}
					return nil
				}

				filesystem.OpenStub = func(path string) (io.ReadCloser, error) {
					if path == "/some/path/to-embed/my-file.txt" {
						return NewBuffer(bytes.NewBufferString("contents-of-embedded-file")), nil
					}

					return NewBuffer(bytes.NewBufferString("contents-of-non-embedded-file")), nil
				}
			})

			It("embeds the file in the embed directory", func() {
				input := builder.WriteInput{
					ReleaseDirectories:   []string{"/some/path/releases"},
					MigrationDirectories: []string{},
					EmbedPaths:           []string{"/some/path/to-embed/my-file.txt"},
					OutputFile:           "some-output-dir/cool-product-file-1.2.3-build.4.pivotal",
					StubReleases:         false,
				}

				err := tileWriter.Write([]byte("generated-metadata-contents"), input)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintfCall.Receives.LogLines).To(Equal([]string{
					fmt.Sprintf("Building %s...", outputFile),
					fmt.Sprintf("Adding metadata/metadata.yml to %s...", outputFile),
					fmt.Sprintf("Creating empty migrations folder in %s...", outputFile),
					fmt.Sprintf("Adding releases/release-1.tgz to %s...", outputFile),
					fmt.Sprintf("Adding releases/release-2.tgz to %s...", outputFile),
					fmt.Sprintf("Adding embed/my-file.txt to %s...", outputFile),
					fmt.Sprintf("Calculating md5 sum of %s...", outputFile),
					"Calculated md5 sum: ",
				}))

				Expect(zipper.AddWithModeCallCount()).To(Equal(1))
				path, file, mode := zipper.AddWithModeArgsForCall(0)
				Expect(path).To(Equal(filepath.Join("embed", "my-file.txt")))
				Eventually(gbytes.BufferReader(file)).Should(gbytes.Say("contents-of-embedded-file"))
				Expect(mode).To(Equal(os.FileMode(12345678)))
			})
		})

		Context("when a directory to embed is provided", func() {
			BeforeEach(func() {
				dirInfo := &fakes.FileInfo{}
				dirInfo.IsDirReturns(true)

				releaseInfo := &fakes.FileInfo{}
				releaseInfo.IsDirReturns(false)

				embedFileInfo := &fakes.FileInfo{}
				embedFileInfo.IsDirReturns(false)
				embedFileInfo.ModeReturns(87654)

				filesystem.WalkStub = func(root string, walkFn filepath.WalkFunc) error {
					if root == "/some/path/releases" {
						walkFn(root, dirInfo, nil)
						walkFn(filepath.Join(root, "release-1.tgz"), releaseInfo, nil)
						walkFn(filepath.Join(root, "release-2.tgz"), releaseInfo, nil)
					} else if root == "/some/path/to-embed" {
						walkFn(root, dirInfo, nil)
						walkFn(filepath.Join(root, "my-file-1.txt"), embedFileInfo, nil)
						walkFn(filepath.Join(root, "my-file-2.txt"), embedFileInfo, nil)
					}
					return nil
				}

				filesystem.OpenStub = func(path string) (io.ReadCloser, error) {
					if path == "/some/path/to-embed/my-file-1.txt" {
						return NewBuffer(bytes.NewBufferString("contents-of-embedded-file-1")), nil
					} else if path == "/some/path/to-embed/my-file-2.txt" {
						return NewBuffer(bytes.NewBufferString("contents-of-embedded-file-2")), nil
					}

					return NewBuffer(bytes.NewBufferString("contents-of-non-embedded-file")), nil
				}
			})

			It("embeds the directory in the embed directory", func() {
				input := builder.WriteInput{
					ReleaseDirectories:   []string{"/some/path/releases"},
					MigrationDirectories: []string{},
					EmbedPaths:           []string{"/some/path/to-embed"},
					OutputFile:           "some-output-dir/cool-product-file-1.2.3-build.4.pivotal",
					StubReleases:         false,
				}

				err := tileWriter.Write([]byte("generated-metadata-contents"), input)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintfCall.Receives.LogLines).To(Equal([]string{
					fmt.Sprintf("Building %s...", outputFile),
					fmt.Sprintf("Adding metadata/metadata.yml to %s...", outputFile),
					fmt.Sprintf("Creating empty migrations folder in %s...", outputFile),
					fmt.Sprintf("Adding releases/release-1.tgz to %s...", outputFile),
					fmt.Sprintf("Adding releases/release-2.tgz to %s...", outputFile),
					fmt.Sprintf("Adding embed/to-embed/my-file-1.txt to %s...", outputFile),
					fmt.Sprintf("Adding embed/to-embed/my-file-2.txt to %s...", outputFile),
					fmt.Sprintf("Calculating md5 sum of %s...", outputFile),
					"Calculated md5 sum: ",
				}))

				path, file, mode := zipper.AddWithModeArgsForCall(0)
				Expect(path).To(Equal(filepath.Join("embed", "to-embed", "my-file-1.txt")))
				Eventually(gbytes.BufferReader(file)).Should(gbytes.Say("contents-of-embedded-file-1"))
				Expect(mode).To(Equal(os.FileMode(87654)))

				path, file, mode = zipper.AddWithModeArgsForCall(1)
				Expect(path).To(Equal(filepath.Join("embed", "to-embed", "my-file-2.txt")))
				Eventually(gbytes.BufferReader(file)).Should(gbytes.Say("contents-of-embedded-file-2"))
				Expect(mode).To(Equal(os.FileMode(87654)))
			})
		})

		Context("failure cases", func() {
			Context("when creating the zip file fails", func() {
				BeforeEach(func() {
					filesystem.CreateReturns(nil, errors.New("boom!"))
				})

				It("returns the error", func() {
					input := builder.WriteInput{
						OutputFile: outputFile,
					}

					err := tileWriter.Write([]byte{}, input)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("boom!"))
				})
			})

			Context("when the zipper fails to create migrations folder", func() {
				BeforeEach(func() {
					zipper.CreateFolderReturns(errors.New("failed to create folder"))
				})

				It("returns an error", func() {
					input := builder.WriteInput{
						StubReleases: true,
						OutputFile:   outputFile,
					}

					err := tileWriter.Write([]byte{}, input)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("failed to create folder"))

					Expect(filesystem.RemoveCallCount()).To(Equal(1))
					Expect(filesystem.RemoveArgsForCall(0)).To(Equal(outputFile))
				})

				Context("when removing the zip file fails", func() {
					BeforeEach(func() {
						filesystem.RemoveReturns(errors.New("boom!"))
					})

					It("returns an error", func() {
						input := builder.WriteInput{
							StubReleases: true,
							OutputFile:   outputFile,
						}

						err := tileWriter.Write([]byte{}, input)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("failed to create folder"))

						expectedLogLine := fmt.Sprintf("failed cleaning up zip %q: %s", outputFile, "boom!")
						Expect(logger.PrintfCall.Receives.LogLines).To(
							ContainElement(expectedLogLine),
						)
					})
				})
			})

			Context("when a release file cannot be opened", func() {
				var input builder.WriteInput

				BeforeEach(func() {
					dirInfo := &fakes.FileInfo{}
					dirInfo.IsDirReturns(true)

					releaseInfo := &fakes.FileInfo{}
					releaseInfo.IsDirReturns(false)

					filesystem.WalkStub = func(root string, walkFn filepath.WalkFunc) error {
						walkFn("/some/path/releases", dirInfo, nil)
						err := walkFn("/some/path/releases/release-1.tgz", releaseInfo, nil)

						return err
					}

					filesystem.OpenStub = func(path string) (io.ReadCloser, error) {
						if path == "/some/path/releases/release-1.tgz" {
							return nil, errors.New("failed to open release")
						}

						return nil, nil
					}

					input = builder.WriteInput{
						ReleaseDirectories: []string{"/some/path/releases"},
						OutputFile:         outputFile,
					}
				})

				It("returns an error", func() {
					err := tileWriter.Write([]byte("generated-metadata-contents"), input)
					Expect(err).To(MatchError("failed to open release"))

					Expect(filesystem.RemoveCallCount()).To(Equal(1))
					Expect(filesystem.RemoveArgsForCall(0)).To(Equal(outputFile))
				})

				Context("when removing the zip file fails", func() {
					BeforeEach(func() {
						filesystem.RemoveReturns(errors.New("boom!"))
					})

					It("returns an error", func() {
						err := tileWriter.Write([]byte{}, input)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("failed to open release"))

						expectedLogLine := fmt.Sprintf("failed cleaning up zip %q: %s", outputFile, "boom!")
						Expect(logger.PrintfCall.Receives.LogLines).To(
							ContainElement(expectedLogLine),
						)
					})
				})
			})

			Context("when a migration file cannot be opened", func() {
				var input builder.WriteInput

				BeforeEach(func() {
					dirInfo := &fakes.FileInfo{}
					dirInfo.IsDirReturns(true)

					releaseInfo := &fakes.FileInfo{}
					releaseInfo.IsDirReturns(false)

					migrationInfo := &fakes.FileInfo{}
					migrationInfo.IsDirReturns(false)

					filesystem.WalkStub = func(root string, walkFn filepath.WalkFunc) error {
						walkFn("/some/path/migrations", dirInfo, nil)
						err := walkFn("/some/path/migrations/migration-1.js", migrationInfo, nil)
						if err != nil {
							return err
						}

						walkFn("/some/path/releases", dirInfo, nil)
						err = walkFn("/some/path/releases/release-1.tgz", releaseInfo, nil)

						return err
					}

					filesystem.OpenStub = func(path string) (io.ReadCloser, error) {
						if path == "/some/path/migrations/migration-1.js" {
							return nil, errors.New("failed to open migration")
						}

						if path == "/some/path/releases/release-1.tgz" {
							return NewBuffer(bytes.NewBufferString("release-1")), nil
						}

						return nil, nil
					}

					input = builder.WriteInput{
						ReleaseDirectories:   []string{"/some/path/releases"},
						MigrationDirectories: []string{"/some/path/migrations"},
						StubReleases:         true,
						OutputFile:           outputFile,
					}
				})

				It("returns an error", func() {
					err := tileWriter.Write([]byte{}, input)
					Expect(err).To(MatchError("failed to open migration"))

					Expect(filesystem.RemoveCallCount()).To(Equal(1))
					Expect(filesystem.RemoveArgsForCall(0)).To(Equal(outputFile))
				})

				Context("when removing the zip file fails", func() {
					BeforeEach(func() {
						filesystem.RemoveReturns(errors.New("boom!"))
					})

					It("returns an error", func() {
						err := tileWriter.Write([]byte{}, input)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("failed to open migration"))

						expectedLogLine := fmt.Sprintf("failed cleaning up zip %q: %s", outputFile, "boom!")
						Expect(logger.PrintfCall.Receives.LogLines).To(
							ContainElement(expectedLogLine),
						)
					})
				})
			})

			Context("when an embed file cannot be opened", func() {
				var input builder.WriteInput

				BeforeEach(func() {
					dirInfo := &fakes.FileInfo{}
					dirInfo.IsDirReturns(true)

					embedInfo := &fakes.FileInfo{}
					embedInfo.IsDirReturns(false)

					filesystem.WalkStub = func(root string, walkFn filepath.WalkFunc) error {
						walkFn("/some/path/embed", dirInfo, nil)
						err := walkFn("/some/path/embed/my-file-1.tgz", embedInfo, nil)

						return err
					}

					filesystem.OpenStub = func(path string) (io.ReadCloser, error) {
						if path == "/some/path/embed/my-file-1.tgz" {
							return nil, errors.New("failed to open embed")
						}

						return nil, nil
					}

					input = builder.WriteInput{
						EmbedPaths: []string{"/some/path/embed"},
						OutputFile: outputFile,
					}
				})

				It("returns an error", func() {
					err := tileWriter.Write([]byte("generated-metadata-contents"), input)
					Expect(err).To(MatchError("failed to open embed"))

					Expect(filesystem.RemoveCallCount()).To(Equal(1))
					Expect(filesystem.RemoveArgsForCall(0)).To(Equal(outputFile))
				})

				Context("when removing the zip file fails", func() {
					BeforeEach(func() {
						filesystem.RemoveReturns(errors.New("boom!"))
					})

					It("returns an error", func() {
						err := tileWriter.Write([]byte{}, input)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("failed to open embed"))

						expectedLogLine := fmt.Sprintf("failed cleaning up zip %q: %s", outputFile, "boom!")
						Expect(logger.PrintfCall.Receives.LogLines).To(
							ContainElement(expectedLogLine),
						)
					})
				})
			})

			Context("when the zipper fails to add a file", func() {
				BeforeEach(func() {
					zipper.AddReturns(errors.New("failed to add file to zip"))
				})

				It("returns an error", func() {
					input := builder.WriteInput{
						StubReleases: true,
						OutputFile:   outputFile,
					}

					err := tileWriter.Write([]byte{}, input)
					Expect(err).To(MatchError("failed to add file to zip"))

					Expect(filesystem.RemoveCallCount()).To(Equal(1))
					Expect(filesystem.RemoveArgsForCall(0)).To(Equal(outputFile))
				})

				Context("when removing the zip file fails", func() {
					BeforeEach(func() {
						filesystem.RemoveReturns(errors.New("boom!"))
					})

					It("returns an error", func() {
						input := builder.WriteInput{
							StubReleases: true,
							OutputFile:   outputFile,
						}

						err := tileWriter.Write([]byte{}, input)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("failed to add file to zip"))

						expectedLogLine := fmt.Sprintf("failed cleaning up zip %q: %s", outputFile, "boom!")
						Expect(logger.PrintfCall.Receives.LogLines).To(
							ContainElement(expectedLogLine),
						)
					})
				})
			})

			Context("when the zipper fails to close", func() {
				var input builder.WriteInput

				BeforeEach(func() {
					zipper.CloseReturns(errors.New("failed to close the zip"))

					input = builder.WriteInput{
						StubReleases: true,
						OutputFile:   outputFile,
					}
				})

				It("returns an error", func() {
					err := tileWriter.Write([]byte{}, input)
					Expect(err).To(MatchError("failed to close the zip"))

					Expect(filesystem.RemoveCallCount()).To(Equal(1))
					Expect(filesystem.RemoveArgsForCall(0)).To(Equal(outputFile))
				})

				Context("when removing the zip file fails", func() {
					BeforeEach(func() {
						filesystem.RemoveReturns(errors.New("boom!"))
					})

					It("returns an error", func() {
						err := tileWriter.Write([]byte{}, input)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("failed to close the zip"))

						expectedLogLine := fmt.Sprintf("failed cleaning up zip %q: %s", outputFile, "boom!")
						Expect(logger.PrintfCall.Receives.LogLines).To(
							ContainElement(expectedLogLine),
						)
					})
				})
			})

			Context("when the MD5 cannot be calculated", func() {
				It("returns an error", func() {

					md5Calc.ChecksumCall.Returns.Error = errors.New("MD5 cannot be calculated")

					input := builder.WriteInput{
						StubReleases: true,
					}

					err := tileWriter.Write([]byte{}, input)
					Expect(err).To(MatchError("MD5 cannot be calculated"))
				})
			})
		})
	})
})

func checkReleaseFileContent(releaseContent string, stubbed bool, file io.Reader) {
	if stubbed == false {
		Eventually(gbytes.BufferReader(file)).Should(gbytes.Say(releaseContent))
	} else {
		Eventually(gbytes.BufferReader(file)).Should(gbytes.Say(""))
	}
}
