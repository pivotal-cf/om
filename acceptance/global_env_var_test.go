package acceptance

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("env var creds", func() {
	It("authenticates with OM_TARGET env var", func() {
		server := testServer(true)

		command := exec.Command(pathToMain,
			"--username", "some-env-provided-username",
			"--password", "some-env-provided-password",
			"--skip-ssl-validation",
			"curl",
			"-p", "/api/v0/available_products",
		)
		command.Env = append(command.Env, "OM_TARGET="+server.URL)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(MatchJSON(`[ { "name": "p-bosh", "product_version": "999.99" } ]`))
	})

	It("authenticates with OM_USERNAME and OM_PASSWORD env vars", func() {
		server := testServer(true)

		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--skip-ssl-validation",
			"curl",
			"-p", "/api/v0/available_products",
		)
		command.Env = append(command.Env, "OM_USERNAME=some-env-provided-username")
		command.Env = append(command.Env, "OM_PASSWORD=some-env-provided-password")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(MatchJSON(`[ { "name": "p-bosh", "product_version": "999.99" } ]`))
	})

	It("authenticates with OM_CLIENT_ID and OM_CLIENT_SECRET env vars", func() {
		server := testServer(false)
		command := exec.Command(pathToMain,
			"--target", server.URL,
			"--client-id", "some-client-id",
			"--skip-ssl-validation",
			"curl",
			"-p", "/api/v0/available_products",
		)
		command.Env = append(command.Env, "OM_CLIENT_ID=some-client-id")
		command.Env = append(command.Env, "OM_CLIENT_SECRET=shhh-secret")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(MatchJSON(`[ { "name": "p-bosh", "product_version": "999.99" } ]`))
	})

	It("overwrites the env file", func() {
		var err error
		var configFile *os.File
		configContent := `
---
password: some-env-provided-password
username: some-env-provided-username
target: %s
skip-ssl-validation: true
connect-timeout: 10
`

		server := testServer(true)

		configFile, err = ioutil.TempFile("", "config.yml")
		Expect(err).NotTo(HaveOccurred())

		_, err = configFile.WriteString(fmt.Sprintf(configContent, server.URL))
		Expect(err).NotTo(HaveOccurred())

		err = configFile.Close()
		Expect(err).NotTo(HaveOccurred())

		command := exec.Command(pathToMain,
			"--env", configFile.Name(),
			"curl",
			"-p", "/api/v0/available_products",
		)

		command.Env = append(command.Env, "OM_USERNAME=invalid-username")
		command.Env = append(command.Env, "OM_PASSWORD=invalid-password")

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
	})
})

func testServer(useUsernamePasswordAuth bool) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch req.URL.Path {
		case "/uaa/oauth/token":
			req.ParseForm()

			if useUsernamePasswordAuth {
				if req.PostForm.Get("username") != "some-env-provided-username" || req.PostForm.Get("password") != "some-env-provided-password" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			} else {
				if username, password, ok := req.BasicAuth(); !ok || username != "some-client-id" || password != "shhh-secret" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}

			w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
		case "/api/v0/available_products":
			auth := req.Header.Get("Authorization")

			if auth != "Bearer some-opsman-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.Write([]byte(`[ { "name": "p-bosh", "product_version": "999.99" } ]`))
		default:
			_, err := httputil.DumpRequest(req, true)
			Expect(err).NotTo(HaveOccurred())

			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
