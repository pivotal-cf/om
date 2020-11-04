package runner_test

import (
	"fmt"
	"github.com/fatih/color"
	"os"
	"testing"

	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/vmlifecycle/runner"
)

func TestRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Runner Suite")
}

var _ = Describe("Runner", func() {
	BeforeEach(func() {
		color.NoColor = true
	})

	Context("Execute", func() {
		It("invokes a CLI with arguments", func() {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			r, err := runner.NewRunner("echo", stdout, stderr)
			Expect(err).ToNot(HaveOccurred())
			o, e, err := r.Execute([]interface{}{"Hello,", "World"})

			Expect(err).ToNot(HaveOccurred())

			Expect(stdout.String()).To(ContainSubstring(`[stdout]: Hello, World`))
			Expect(o.String()).To(Equal("Hello, World\n"))
			Expect(stderr.String()).To(ContainSubstring(`Executing: "echo Hello, World"`))
			Expect(e.String()).To(Equal(""))
		})

		It("Redacts specified values from the output", func() {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			r, err := runner.NewRunner("echo", stdout, stderr)
			Expect(err).ToNot(HaveOccurred())
			o, e, err := r.Execute([]interface{}{"Hello,", runner.Redact("World")})

			Expect(err).ToNot(HaveOccurred())

			Expect(stdout.String()).To(ContainSubstring(`[stdout]: Hello, World`))
			Expect(o.String()).To(Equal("Hello, World\n"))
			Expect(stderr.String()).To(ContainSubstring(`Executing: "echo Hello, <REDACTED>"`))
			Expect(e.String()).To(Equal(""))
		})
	})

	Context("ExecuteWithEnvVars", func() {
		It("invokes a CLI with arguments and adds specified environment variables", func() {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			r, err := runner.NewRunner("bash", stdout, stderr)
			Expect(err).ToNot(HaveOccurred())
			o, _, err := r.ExecuteWithEnvVars([]string{"NAME=123"}, []interface{}{"-c", "echo NAME=$NAME"})

			Expect(err).ToNot(HaveOccurred())

			Expect(stdout.String()).To(ContainSubstring(`[stdout]: NAME=123`))
			Expect(o.String()).To(Equal("NAME=123\n"))
		})

		It("invokes a CLI with arguments and includes all environment variables from the process context", func() {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			r, err := runner.NewRunner("bash", stdout, stderr)
			Expect(err).ToNot(HaveOccurred())
			_, _, err = r.ExecuteWithEnvVars([]string{"key1=value", "key2=value"}, []interface{}{"-c", "env"})

			Expect(err).ToNot(HaveOccurred())

			Expect(stdout.String()).To(MatchRegexp(fmt.Sprintf("PATH=%s", os.Getenv("PATH"))))
		})

		It("invokes a CLI with arguments and overrides vars from the environment with those specified", func() {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			Expect(os.Setenv("blahblah", "123")).ToNot(HaveOccurred())
			defer os.Unsetenv("blahblah")

			r, err := runner.NewRunner("bash", stdout, stderr)
			Expect(err).ToNot(HaveOccurred())
			_, _, err = r.ExecuteWithEnvVars([]string{"blahblah=abc"}, []interface{}{"-c", "env"})

			Expect(err).ToNot(HaveOccurred())

			Expect(stdout.String()).ToNot(MatchRegexp("blahblah=123"))
			Expect(stdout.String()).To(MatchRegexp("blahblah=abc"))
		})
	})

	When("the cli does not exist", func() {
		It("returns an error", func() {
			_, err := runner.NewRunner("does-not-exist-cli", nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the cli 'does-not-exist-cli' is not available in PATH"))
		})
	})
})
