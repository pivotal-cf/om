package acceptance

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	pathToMain string
	minio      *gexec.Session
)

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	omPath, err := gexec.Build("../main.go", "-ldflags", "-X main.applySleepDurationString=1ms")
	Expect(err).ToNot(HaveOccurred())

	minioPath, _ := exec.LookPath("minio")
	fmt.Fprintf(GinkgoWriter, "Minio PATH : %s", minioPath)
	Expect(minioPath).ToNot(BeEmpty(), "Minio not found")
	if minioPath != "" && 1 == 2 {
		dataDir, err := ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())
		command := exec.Command("minio", "server", "--config-dir", dataDir, "--address", ":9001", dataDir)
		command.Env = []string{
			fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
			"MINIO_ROOT_USER=minio",
			"MINIO_ROOT_PASSWORD=password",
			"MINIO_BROWSER=off",
			"TERM=xterm-256color",
		}
		minio, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(minio.Out, "10s").Should(gbytes.Say("API:"))
		runCommand("mc", "--debug", "config", "host", "add", "testing", "http://minio:9001", minioUser, minioPassword)
	}
	return []byte(omPath)
}, func(data []byte) {
	pathToMain = string(data)

	// Clear any OM env vars so as to not pollute the tests
	re := regexp.MustCompile(`OM_*`)
	for _, pair := range os.Environ() {
		split := strings.Split(pair, "=")
		if re.MatchString(split[0]) {
			Expect(os.Unsetenv(split[0])).NotTo(HaveOccurred())
		}
	}
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	if minio != nil {
		minio.Kill()
	}
	gexec.CleanupBuildArtifacts()
})

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "acceptance")
}

func runCommand(args ...string) {
	fmt.Fprintf(GinkgoWriter, "cmd: %s\n", args)
	command := exec.Command(args[0], args[1:]...)
	configure, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(configure, "20s").Should(gexec.Exit(0))
}

func runCommand2(ignoreError bool, args ...string) {
	fmt.Fprintf(GinkgoWriter, "cmd: %s\n", args)
	command := exec.Command(args[0], args[1:]...)
	configure, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	if !ignoreError {
		Expect(err).ToNot(HaveOccurred())
		Eventually(configure, "20s").Should(gexec.Exit(0))
	}
}

func createTLSServer() *ghttp.Server {
	server := ghttp.NewTLSServer()
	server.RouteToHandler("POST", "/uaa/oauth/token",
		ghttp.CombineHandlers(
			ghttp.RespondWith(http.StatusOK, `{
				"access_token": "some-opsman-token",
				"token_type": "bearer",
				"expires_in": 3600
			}`, map[string][]string{
				"Content-Type": {"application/json"},
			}),
		),
	)

	return server
}

type ensureHandler struct {
	handlers []http.HandlerFunc
}

func (e *ensureHandler) Ensure(funs ...http.HandlerFunc) []http.HandlerFunc {

	for _, fun := range funs {
		e.handlers = append(e.handlers, func(writer http.ResponseWriter, request *http.Request) {
			fun(writer, request)
			e.handlers = e.handlers[1:]
		})
	}

	return e.handlers
}

func (e *ensureHandler) Handlers() []http.HandlerFunc {
	return e.handlers
}
