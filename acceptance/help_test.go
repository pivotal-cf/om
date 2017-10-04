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
  -c, --client-id            string  Client ID for the Ops Manager VM (not required for unauthenticated commands)
  -s, --client-secret        string  Client Secret for the Ops Manager VM (not required for unauthenticated commands)
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Commands:
  activate-certificate-authority    activates a certificate authority on the Ops Manager
  apply-changes                     triggers an install on the Ops Manager targeted
  available-products                list available products
  certificate-authorities           lists certificates managed by Ops Manager
  configure-authentication          configures Ops Manager with an internal userstore and admin user account
  configure-bosh                    configures Ops Manager deployed bosh director
  configure-product                 configures a staged product
  create-certificate-authority      creates a certificate authority on the Ops Manager
  credential-references             list credential references for a deployed product
  credentials                       fetch credentials for a deployed product
  curl                              issues an authenticated API request
  delete-certificate-authority      deletes a certificate authority on the Ops Manager
  delete-installation               deletes all the products on the Ops Manager targeted
  delete-product                    deletes a product from the Ops Manager
  delete-unused-products            deletes unused products on the Ops Manager targeted
  deployed-products                 lists deployed products
  errands                           list errands for a product
  export-installation               exports the installation of the target Ops Manager
  generate-certificate-authority    generates a certificate authority on the Opsman
  help                              prints this usage information
  import-installation               imports a given installation to the Ops Manager targeted
  installation-log                  output installation logs
  installations                     list recent installation events
  pending-changes                   lists pending changes
  regenerate-certificate-authority  regenerates a certificate authority on the Opsman
  revert-staged-changes             reverts staged changes on the Ops Manager targeted
  set-errand-state                  sets state for a product's errand
  stage-product                     stages a given product in the Ops Manager targeted
  staged-products                   lists staged products
  unstage-product                   unstages a given product from the Ops Manager targeted
  upload-product                    uploads a given product to the Ops Manager targeted
  upload-stemcell                   uploads a given stemcell to the Ops Manager targeted
  version                           prints the om release version
`

const CONFIGURE_AUTHENTICATION_USAGE = `ॐ  configure-authentication
This unauthenticated command helps setup the authentication mechanism for your Ops Manager.
The "internal" userstore mechanism is the only currently supported option.

Usage: om [options] configure-authentication [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -c, --client-id            string  Client ID for the Ops Manager VM (not required for unauthenticated commands)
  -s, --client-secret        string  Client Secret for the Ops Manager VM (not required for unauthenticated commands)
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Command Arguments:
  -u, --username                string  admin username
  -p, --password                string  admin password
  -dp, --decryption-passphrase  string  passphrase used to encrypt the installation
  --http-proxy-url              string  proxy for outbound HTTP network traffic
  --https-proxy-url             string  proxy for outbound HTTPS network traffic
  --no-proxy                    string  comma-separated list of hosts that do not go through the proxy
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
