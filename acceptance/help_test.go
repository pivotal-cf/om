package acceptance

import (
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const GLOBAL_USAGE = `ॐ
om helps you interact with an Ops Manager

Usage: om [options] <command> [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Commands:
  apply-changes             triggers an install on the Ops Manager targeted
  configure-authentication  configures Ops Manager with an internal userstore and admin user account
  configure-product         configures a staged product
  curl                      issues an authenticated API request
  export-installation       exports the installation of the target Ops Manager
  help                      prints this usage information
  import-installation       imports a given installation to the Ops Manager targeted
  stage-product             stages a given product in the Ops Manager targeted
  upload-product            uploads a given product to the Ops Manager targeted
  upload-stemcell           uploads a given stemcell to the Ops Manager targeted
  version                   prints the om release version
`

const CONFIGURE_AUTHENTICATION_USAGE = `ॐ  configure-authentication
This unauthenticated command helps setup the authentication mechanism for your Ops Manager.
The "internal" userstore mechanism is the only currently supported option.

Usage: om [options] configure-authentication [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Command Arguments:
  -u, --username                string  admin username
  -p, --password                string  admin password
  -dp, --decryption-passphrase  string  passphrase used to encrypt the installation
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
			command := exec.Command(pathToMain, "help", "configure-authentication")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(CONFIGURE_AUTHENTICATION_USAGE))
		})
	})
})
