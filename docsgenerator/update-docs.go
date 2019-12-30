package main

import (
	"fmt"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/om/docsgenerator/executor"
	"github.com/pivotal-cf/om/docsgenerator/generator"
	"os"
	"path/filepath"
)

func main() {
	omPath, err := gexec.Build("../main.go", "-ldflags", "-X main.applySleepDurationString=1ms -X github.com/pivotal-cf/om/commands.pivnetHost=http://example.com")
	if err != nil {
		fmt.Printf("could not build binary: %s\n", err)
		os.Exit(1)
	}

	templateDir, err := filepath.Abs("templates/")
	if err != nil {
		fmt.Printf("failed to get absolute path of templates directory: %s\n", err)
		os.Exit(1)
	}

	docsDir, err := filepath.Abs("../docs/")
	if err != nil {
		fmt.Printf("failed to get absolute path of docs directory: %s\n", err)
		os.Exit(1)
	}

	gen := generator.NewGenerator(templateDir, docsDir, executor.NewExecutor(omPath), os.Stdout)

	err = gen.GenerateDocs()
	if err != nil {
		fmt.Printf("could not generate docs: %s\n", err)
		os.Exit(1)
	}
}
