package acceptance

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("global flags", func() {
	When("provided an unknown global flag", func() {
		It("prints the usage", func() {
			cmd := exec.Command(pathToMain, "-?")

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err.Contents()).Should(ContainSubstring("unknown flag `?'"))
		})
	})

	When("not provided a target flag", func() {
		It("does not return an error if the command is version", func() {
			cmd := exec.Command(pathToMain, "version")

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
		})
	})

	When("a ca cert is required to communicate with the OpsMan", func() {
		It("supports a file from --ca-cert", func() {
			server := testServer(true)
			cert, err := x509.ParseCertificate(server.TLS.Certificates[0].Certificate[0])
			Expect(err).ToNot(HaveOccurred())
			pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})

			command := exec.Command(pathToMain,
				"--username", "some-env-provided-username",
				"--password", "some-env-provided-password",
				"--target", server.URL,
				"--ca-cert", string(pemCert),
				"curl",
				"-p", "/api/v0/available_products",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
		})

		It("supports a string from --ca-cert", func() {
			server := testServer(true)
			cert, err := x509.ParseCertificate(server.TLS.Certificates[0].Certificate[0])
			Expect(err).ToNot(HaveOccurred())
			pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
			caCertFilename := writeFile(string(pemCert))

			command := exec.Command(pathToMain,
				"--username", "some-env-provided-username",
				"--password", "some-env-provided-password",
				"--target", server.URL,
				"--ca-cert", caCertFilename,
				"curl",
				"-p", "/api/v0/available_products",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
		})
	})

	It("takes precedence over the env var values", func() {
		server := testServer(true)

		command := exec.Command(pathToMain,
			"--username", "some-env-provided-username",
			"--password", "some-env-provided-password",
			"--skip-ssl-validation",
			"curl",
			"-p", "/api/v0/available_products",
		)
		command.Env = append(command.Env, "OM_TARGET="+server.URL)
		command.Env = append(command.Env, "PASSWORD=bogus")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(MatchJSON(`[ { "name": "p-bosh", "product_version": "999.99" } ]`))
	})

	It("takes precedence over env file", func() {
		server := testServer(true)

		command := exec.Command(pathToMain,
			"--env", writeFile(`target: incorrect-server-url`),
			"--username", "some-env-provided-username",
			"--password", "some-env-provided-password",
			"--target", server.URL,
			"--skip-ssl-validation",
			"curl",
			"-p", "/api/v0/available_products",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(MatchJSON(`[ { "name": "p-bosh", "product_version": "999.99" } ]`))
	})
})

func writeFile(contents string) string {
	file, err := os.CreateTemp("", "")
	Expect(err).ToNot(HaveOccurred())

	err = os.WriteFile(file.Name(), []byte(contents), 0777)
	Expect(err).ToNot(HaveOccurred())
	return file.Name()
}
