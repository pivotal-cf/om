package main_test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var (
	pathToMain string
)

var _ = BeforeSuite(func() {
	var err error
	pathToMain, err = gexec.Build("github.com/pivotal-cf/om/vmlifecycle/vmmanagers/stub")
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

var _ = Describe("stub", func() {
	When("environment variables are set", func() {
		It("prints them out to standard out", func() {
			command := exec.Command(pathToMain)
			command.Env = []string{
				"HELLO=WORLD",
				"TESTING=123",
			}
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			Eventually(session.Err).Should(gbytes.Say(`env: HELLO=WORLD TESTING=123`))
		})
	})

	Context("called with command line arguments", func() {
		It("prints them to standard out", func() {
			command := exec.Command(pathToMain, "'some command'", "--another", "-f", "argument")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			Eventually(session.Err).Should(gbytes.Say(`stub 'some command' --another -f argument`))
		})

		It("reports its own name", func() {
			tmpExec := filepath.Dir(pathToMain) + "/test-stub"
			err := os.Link(pathToMain, tmpExec)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(tmpExec, "'some command'", "--another", "-f", "argument")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			Eventually(session.Err).Should(gbytes.Say(`test-stub 'some command' --another -f argument`))
		})
	})

	When("STUB_ERROR_CODE is set", func() {
		It("returns that error code", func() {
			os.Setenv("STUB_ERROR_CODE", "123")
			defer os.Unsetenv("STUB_ERROR_CODE")
			command := exec.Command(pathToMain)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(123))
		})

		It("prints an error message", func() {
			os.Setenv("STUB_ERROR_CODE", "123")
			defer os.Unsetenv("STUB_ERROR_CODE")
			command := exec.Command(pathToMain)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session.Err).Should(gbytes.Say(`stub error!!`))
		})
	})

	When("STUB_ERROR_MSG and STUB_ERROR_CODE is set", func() {
		It("returns that error msg with correct error code", func() {
			os.Setenv("STUB_ERROR_CODE", "12")
			os.Setenv("STUB_ERROR_MSG", "some error")

			defer os.Unsetenv("STUB_ERROR_CODE")
			defer os.Unsetenv("STUB_ERROR_MSG")
			command := exec.Command(pathToMain)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(12))
			Eventually(session.Err).Should(gbytes.Say("some error"))
		})
	})

	When("STUB_ERROR_MSG and STUB_ERROR_CODE is NOT set", func() {
		It("does nothing", func() {
			os.Setenv("STUB_ERROR_MSG", "some error")

			defer os.Unsetenv("STUB_ERROR_MSG")
			command := exec.Command(pathToMain)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Consistently(session).ShouldNot(gexec.Exit(12))
		})
	})

	When("STUB_OUTPUT is set and matches STUB_EFFECTIVE_NAME", func() {
		It("should return the output", func() {
			os.Setenv("STUB_OUTPUT", "response!")
			defer os.Unsetenv("STUB_OUTPUT")
			command := exec.Command(pathToMain)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session.Out).Should(gbytes.Say(`response!`))
		})
	})

})
