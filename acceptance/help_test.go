package acceptance

import (
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("help", func() {
	When("given no command at all", func() {
		It("prints the global usage", func() {
			command := exec.Command(pathToMain)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(GLOBAL_USAGE_COMMANDS))
			Expect(string(session.Out.Contents())).To(ContainSubstring(GLOBAL_USAGE_FLAGS))
		})
	})

	When("given the -h short flag", func() {
		It("prints the global usage", func() {
			command := exec.Command(pathToMain, "-h")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(GLOBAL_USAGE_COMMANDS))
			Expect(string(session.Out.Contents())).To(ContainSubstring(GLOBAL_USAGE_FLAGS))
		})
	})

	When("given the --help long flag", func() {
		It("prints the global usage", func() {
			command := exec.Command(pathToMain, "--help")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(GLOBAL_USAGE_COMMANDS))
			Expect(string(session.Out.Contents())).To(ContainSubstring(GLOBAL_USAGE_FLAGS))
		})
	})

	When("given the help command", func() {
		It("prints the global usage", func() {
			command := exec.Command(pathToMain, "help")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(GLOBAL_USAGE_COMMANDS))
			Expect(string(session.Out.Contents())).To(ContainSubstring(GLOBAL_USAGE_FLAGS))
		})
	})

	When("given a command", func() {
		It("prints the usage for that command", func() {
			command := exec.Command(pathToMain, "help", "configure-authentication")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(CONFIGURE_AUTHENTICATION_USAGE))
			Expect(string(session.Out.Contents())).To(ContainSubstring(GLOBAL_USAGE_FLAGS))
		})
	})
})

const GLOBAL_USAGE_COMMANDS = `
om helps you interact with an Ops Manager

Usage:
  om [options] <command> [<args>]

Commands:
  activate-certificate-authority  activates a certificate authority on the Ops Manager
`

const GLOBAL_USAGE_FLAGS = `
Global Flags:
  --ca-cert, OM_CA_CERT                                  string  OpsManager CA certificate path or value
  --client-id, -c, OM_CLIENT_ID                          string  Client ID for the Ops Manager VM (not required for unauthenticated commands)
  --client-secret, -s, OM_CLIENT_SECRET                  string  Client Secret for the Ops Manager VM (not required for unauthenticated commands)
  --connect-timeout, -o, OM_CONNECT_TIMEOUT              int     timeout in seconds to make TCP connections (default: 10)
  --decryption-passphrase, -d, OM_DECRYPTION_PASSPHRASE  string  Passphrase to decrypt the installation if the Ops Manager VM has been rebooted (optional for most commands)
  --env, -e                                              string  env file with login credentials
  --help, -h                                             bool    prints this usage information (default: false)
  --password, -p, OM_PASSWORD                            string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  --request-timeout, -r, OM_REQUEST_TIMEOUT              int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)
  --skip-ssl-validation, -k, OM_SKIP_SSL_VALIDATION      bool    skip ssl certificate validation during http requests (default: false)
  --target, -t, OM_TARGET                                string  location of the Ops Manager VM
  --trace, -tr, OM_TRACE                                 bool    prints HTTP requests and response payloads
  --username, -u, OM_USERNAME                            string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  --version, -v                                          bool    prints the om release version (default: false)
  OM_VARS_ENV                                            string  load vars from environment variables by specifying a prefix (e.g.: 'MY' to load MY_var=value)
`

const CONFIGURE_AUTHENTICATION_USAGE = `
This unauthenticated command helps setup the internal userstore authentication mechanism for your Ops Manager.

Usage:
  om [options] configure-authentication [<args>]

Flags:
  --config, -c                                            string             path to yml file for configuration (keys must match the following command line flags)
  --decryption-passphrase, -dp, OM_DECRYPTION_PASSPHRASE  string (required)  passphrase used to encrypt the installation
  --http-proxy-url                                        string             proxy for outbound HTTP network traffic
  --https-proxy-url                                       string             proxy for outbound HTTPS network traffic
  --no-proxy                                              string             comma-separated list of hosts that do not go through the proxy
  --password, -p, OM_PASSWORD                             string (required)  admin password
  --precreated-client-secret                              string             create a UAA client on the Ops Manager vm. The client_secret will be the value provided to this option
  --username, -u, OM_USERNAME                             string (required)  admin username
  --var, -v                                               string (variadic)  load variable from the command line. Format: VAR=VAL
  --vars-env, OM_VARS_ENV                                 string (variadic)  load variables from environment variables matching the provided prefix (e.g.: 'MY' to load MY_var=value)
  --vars-file, -l                                         string (variadic)  load variables from a YAML file
` + GLOBAL_USAGE_FLAGS
