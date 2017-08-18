package acceptance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("revert-staged-changes command", func() {
	var (
		server          *httptest.Server
		receivedCookies []*http.Cookie
		Forms           []url.Values
	)

	BeforeEach(func() {
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/uaa/oauth/token":
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`))
			case "/":
				http.SetCookie(w, &http.Cookie{
					Name:  "somecookie",
					Value: "somevalue",
					Path:  "/",
				})

				w.Write([]byte(`<html>
				<body>
					<form action="/some-other-form" method="post">
						<input name="_method" value="fakemethod2" />
						<input name="authenticity_token" value="fake_authenticity2" />
					</form>
					<form action="/installation" method="post">
						<input name="_method" value="fakemethod" />
						<input name="authenticity_token" value="fake_authenticity" />
					</form>
					<form action="/install" method="post">
						<input name="_method" value="fakemethod" />
						<input name="authenticity_token" value="fake_authenticity" />
					</form>
					</body>
				</html>`))
			case "/installation":
				receivedCookies = req.Cookies()
				req.ParseForm()
				Forms = append(Forms, req.Form)
			default:
				out, err := httputil.DumpRequest(req, true)
				Expect(err).NotTo(HaveOccurred())
				Fail(fmt.Sprintf("unexpected request: %s", out))
			}
		}))
	})

	AfterEach(func() {
		Forms = []url.Values{}
	})

	var (
		command *exec.Cmd
	)

	BeforeEach(func() {
		command = exec.Command(pathToMain,
			"--target", server.URL,
			"--username", "fake-username",
			"--password", "fake-password",
			"--skip-ssl-validation",
			"revert-staged-changes",
		)
	})

	It("reverts staged changes on the installation dashboard", func() {
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		Expect(session.Out).To(gbytes.Say("reverting staged changes on the targeted Ops Manager"))
		Expect(receivedCookies).To(HaveLen(1))
		Expect(receivedCookies[0].Name).To(Equal("somecookie"))

		Expect(Forms[0].Get("authenticity_token")).To(Equal("fake_authenticity"))
		Expect(Forms[0].Get("_method")).To(Equal("delete"))
		Expect(Forms[0].Get("commit")).To(Equal("Confirm"))
	})
})
