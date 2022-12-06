package commands_test

import (
	"log"
	"os"
	"path/filepath"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/commands"
)

var _ = FDescribe(reflect.TypeOf(commands.ReplaceRelease{}).Name(), func() {
	var (
		stdout *log.Logger
		tmp    string

		cmd *commands.ReplaceRelease
	)

	BeforeEach(func() {
		stdout = log.New(GinkgoWriter, "", 0)
		tmp = os.TempDir()

		cmd = commands.NewReplaceRelease(stdout)
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmp)
	})

	When("provided all the required inputs", func() {
		BeforeEach(func() {
			cmd.Options.Product = filepath.Join("testdata", "replace-release", "tile-0.1.2.pivotal")
			cmd.Options.ProductOutput = filepath.Join(tmp, "updated-tile.pivotal")
			cmd.Options.NewVersion = "0.1.3-banana.1"
			cmd.Options.ExistingRelease = "hello-release/0.1.4"
			cmd.Options.NewRelease = filepath.Join("testdata", "replace-release", "hello-release-0.1.5.tgz")
			cmd.Options.Verbose = true
		})

		It("does not return an error", func() {
			err := cmd.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
