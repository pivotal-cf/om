package cmd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jessevdk/go-flags"
)

var _ = Describe("validateCommandFlags", func() {
	type uploadProductOptions struct {
		Product         string   `long:"product" short:"p" description:"path to product" required:"true"`
		PollingInterval int      `long:"polling-interval" short:"i" description:"interval (in seconds) at which to print status" default:"1"`
		Shasum          string   `long:"shasum" description:"shasum of the provided product file to be used for validation"`
		Version         string   `long:"product-version" description:"version of the provided product file to be used for validation"`
		Config          string   `long:"config" short:"c" description:"path to yml file for configuration"`
		VarsEnv         string   `long:"vars-env" description:"load variables from environment variables matching the provided prefix"`
		VarsFile        []string `long:"vars-file" short:"l" description:"load variables from a YAML file"`
		Var             []string `long:"var" short:"v" description:"load variable from the command line. Format: VAR=VAL"`
	}

	var parser *flags.Parser

	BeforeEach(func() {
		parser = flags.NewParser(nil, flags.Default)
		parser.AddCommand("upload-product", "desc", "long desc", &uploadProductOptions{})
	})

	DescribeTable("flag validation",
		func(args []string, wantErr bool, errMsg string) {
			err := validateCommandFlags(parser, args)
			if wantErr {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errMsg))
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		},
		Entry("no args", []string{}, false, ""),
		Entry("valid flags", []string{"upload-product", "--product", "file.pivotal", "--polling-interval", "5", "--shasum", "abc123", "--product-version", "2.3.4"}, false, ""),
		Entry("valid short flags", []string{"upload-product", "-p", "file.pivotal", "-i", "5"}, false, ""),
		Entry("all config interpolation flags", []string{"upload-product", "-c", "config.yml", "--vars-env", "MY", "-l", "vars1.yml", "-l", "vars2.yml", "-v", "FOO=bar", "-v", "BAZ=qux", "-p", "file.pivotal"}, false, ""),
		Entry("mix config and main flags", []string{"upload-product", "-p", "file.pivotal", "-c", "config.yml", "--vars-env", "MY", "--shasum", "abc123", "-l", "vars.yml", "-v", "FOO=bar"}, false, ""),
		Entry("unknown flag with config flags", []string{"upload-product", "-p", "file.pivotal", "-c", "config.yml", "--notaflag"}, true, "unknown flag(s)"),
		Entry("unknown flag", []string{"upload-product", "--notaflag"}, true, "unknown flag(s)"),
		Entry("multiple unknown flags", []string{"upload-product", "--foo", "--bar"}, true, "unknown flag(s)"),
		Entry("flag value looks like flag", []string{"upload-product", "--product", "--notaflag"}, false, ""),
		Entry("unknown short flags", []string{"upload-product", "-p", "file.pivotal", "-x", "18000", "-z", "18000"}, true, "unknown flag(s)"),
	)
})
