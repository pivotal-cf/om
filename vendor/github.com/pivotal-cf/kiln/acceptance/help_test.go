package acceptance

import (
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const GLOBAL_USAGE = `kiln
kiln helps you build ops manager compatible tiles

Usage: kiln [options] <command> [<args>]
  --help, -h     bool  prints this usage information (default: false)
  --version, -v  bool  prints the kiln release version (default: false)

Commands:
  bake     bakes a tile
  help     prints this usage information
  version  prints the kiln release version
`

const BAKE_USAGE = `kiln bake
Bakes tile metadata, stemcell, releases, and migrations into a format that can be consumed by OpsManager.

Usage: kiln [options] bake [<args>]
  --help, -h     bool  prints this usage information (default: false)
  --version, -v  bool  prints the kiln release version (default: false)

Command Arguments:
  --bosh-variables-directory, -vd    string (variadic)            path to a directory containing BOSH variables
  --embed, -e                        string (variadic)            path to files to include in the tile /embed directory
  --forms-directory, -f              string (variadic)            path to a directory containing forms
  --icon, -i                         string                       path to icon file
  --instance-groups-directory, -ig   string (variadic)            path to a directory containing instance groups
  --jobs-directory, -j               string (variadic)            path to a directory containing jobs
  --metadata, -m                     string (required)            path to the metadata file
  --migrations-directory, -md        string (variadic)            path to a directory containing migrations
  --output-file, -o                  string (required)            path to where the tile will be output
  --properties-directory, -pd        string (variadic)            path to a directory containing property blueprints
  --releases-directory, -rd          string (required, variadic)  path to a directory containing release tarballs
  --runtime-configs-directory, -rcd  string (variadic)            path to a directory containing runtime configs
  --stemcell-tarball, -st            string                       path to a stemcell tarball
  --stub-releases, -sr               bool                         skips importing release tarballs into the tile
  --variable, -vr                    string (variadic)            key value pairs of variables to interpolate
  --variables-file, -vf              string (variadic)            path to a file containing variables to interpolate
  --version, -v                      string                       version of the tile
`

var _ = Describe("help", func() {
	Context("when given no command at all", func() {
		It("prints the global usage", func() {
			command := exec.Command(pathToMain)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(GLOBAL_USAGE))
		})
	})

	Context("when given the -h short flag", func() {
		It("prints the global usage", func() {
			command := exec.Command(pathToMain, "-h")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(GLOBAL_USAGE))
		})
	})

	Context("when given the --help long flag", func() {
		It("prints the global usage", func() {
			command := exec.Command(pathToMain, "--help")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(GLOBAL_USAGE))
		})
	})

	Context("when given the help command", func() {
		It("prints the global usage", func() {
			command := exec.Command(pathToMain, "help")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(GLOBAL_USAGE))
		})
	})

	Context("when given a command", func() {
		It("prints the usage for that command", func() {
			command := exec.Command(pathToMain, "help", "bake")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(BAKE_USAGE))
		})
	})
})
