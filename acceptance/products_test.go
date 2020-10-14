package acceptance

import (
	"github.com/onsi/gomega/ghttp"
	"net/http"
	"os/exec"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("products command", func() {
	const defaultTableOutput = `+----------------+-----------+--------+----------+
|      NAME      | AVAILABLE | STAGED | DEPLOYED |
+----------------+-----------+--------+----------+
| acme-product-1 | 1.13.1    | 1.13.2 | 1.13.3   |
|                | 1.13.2    |        |          |
| acme-product-2 | 1.8.0     | 1.8.1  |          |
+----------------+-----------+--------+----------+
`

	const availableTableOutput = `+----------------+-----------+
|      NAME      | AVAILABLE |
+----------------+-----------+
| acme-product-1 | 1.13.1    |
|                | 1.13.2    |
| acme-product-2 | 1.8.0     |
+----------------+-----------+
`

	const stagedTableOutput = `+----------------+--------+
|      NAME      | STAGED |
+----------------+--------+
| acme-product-1 | 1.13.2 |
| acme-product-2 | 1.8.1  |
| p-bosh         | 1.9.1  |
+----------------+--------+
`

	const deployedTableOutput = `+----------------+----------+
|      NAME      | DEPLOYED |
+----------------+----------+
| acme-product-1 | 1.13.3   |
| p-bosh         | 1.9.1    |
+----------------+----------+
`

	const defaultJsonOutput = `[
		{"name":"acme-product-1","available":["1.13.1","1.13.2"],"staged":"1.13.2","deployed":"1.13.3"},
		{"name":"acme-product-2","available":["1.8.0"],"staged":"1.8.1"},
		{"name":"p-bosh","staged":"1.9.1","deployed":"1.9.1"}
	]`

	const availableJsonOutput = `[
		{"name":"acme-product-1","available":["1.13.1","1.13.2"]},
		{"name":"acme-product-2","available":["1.8.0"]}
	]`

	const stagedJsonOutput = `[
		{"name":"acme-product-1","staged":"1.13.2"},
		{"name":"acme-product-2","staged":"1.8.1"},
		{"name":"p-bosh","staged":"1.9.1"}
	]`

	const deployedJsonOutput = `[
		{"name":"acme-product-1","deployed":"1.13.3"},
		{"name":"p-bosh","deployed":"1.9.1"}
	]`

	const diagnosticReport = `{
		"added_products": {
			"staged": [
				{"name":"acme-product-1","version":"1.13.2"},
				{"name":"acme-product-2","version":"1.8.1"},
				{"name":"p-bosh","version":"1.9.1"}
			],
			"deployed": [
				{"name":"acme-product-1","version":"1.13.3"},
				{"name":"p-bosh","version":"1.9.1"}
			]
		}
	}`

	const availableProducts = `[{
		"name": "acme-product-1",
		"product_version": "1.13.1"
	}, {
		"name": "acme-product-1",
		"product_version": "1.13.2"
	}, {
		"name":"acme-product-2",
		"product_version":"1.8.0"
	}]`

	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = createTLSServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
				ghttp.RespondWith(http.StatusOK, diagnosticReport),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/available_products"),
				ghttp.RespondWith(http.StatusOK, availableProducts),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	It("lists all products on the Ops Manager as well as their available, staged, and deployed versions", func() {
		command := exec.Command(pathToMain,
			"--target", server.URL(),
			"--username", "some-username",
			"--password", "some-password",
			"--skip-ssl-validation",
			"products",
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(defaultTableOutput))
	})

	When("json format is requested", func() {
		It("lists the available, staged, and deployed products on Ops Manager", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"products",
				"--format", "json",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(MatchJSON(defaultJsonOutput))
		})
	})

	When("--available flag is passed", func() {
		It("lists all available products on the Ops Manager as well as their versions", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"products",
				"--available",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(Equal(availableTableOutput))
		})

		When("json format is requested", func() {
			It("lists the available products on Ops Manager", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"products",
					"--available",
					"--format", "json",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))
				Expect(string(session.Out.Contents())).To(MatchJSON(availableJsonOutput))
			})
		})
	})

	When("--staged flag is passed", func() {
		It("lists all staged products on the Ops Manager as well as their versions", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"products",
				"--staged",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(Equal(stagedTableOutput))
		})

		When("json format is requested", func() {
			It("lists the staged products on Ops Manager", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"products",
					"--staged",
					"--format", "json",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))
				Expect(string(session.Out.Contents())).To(MatchJSON(stagedJsonOutput))
			})
		})
	})

	When("--deployed flag is passed", func() {
		It("lists all deployed products on the Ops Manager as well as their versions", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"products",
				"--deployed",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(Equal(deployedTableOutput))
		})

		When("json format is requested", func() {
			It("lists the deployed products on Ops Manager", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"products",
					"--deployed",
					"--format", "json",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))
				Expect(string(session.Out.Contents())).To(MatchJSON(deployedJsonOutput))
			})
		})
	})

	When("--available, --staged, and --deployed flags are passed", func() {
		It("lists all products on the Ops Manager as well as their available, staged, and deployed versions", func() {
			command := exec.Command(pathToMain,
				"--target", server.URL(),
				"--username", "some-username",
				"--password", "some-password",
				"--skip-ssl-validation",
				"products",
				"--available",
				"--staged",
				"--deployed",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(Equal(defaultTableOutput))
		})

		When("json format is requested", func() {
			It("lists the staged products on Ops Manager", func() {
				command := exec.Command(pathToMain,
					"--target", server.URL(),
					"--username", "some-username",
					"--password", "some-password",
					"--skip-ssl-validation",
					"products",
					"--available",
					"--staged",
					"--deployed",
					"--format", "json",
				)

				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))
				Expect(string(session.Out.Contents())).To(MatchJSON(defaultJsonOutput))
			})
		})
	})
})
