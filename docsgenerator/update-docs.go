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
	currentPath, err := os.Getwd()
	if err != nil {
		fmt.Printf("could not determine working directory: %s", err)
		os.Exit(1)
	}

	omPath, err := gexec.Build("github.com/pivotal-cf/om", "-ldflags", "-X main.applySleepDurationString=1ms -X github.com/pivotal-cf/om/commands.pivnetHost=http://example.com")
	if err != nil {
		fmt.Printf("could not build binary: %s\n", err)
		os.Exit(1)
	}

	templateDir := filepath.Join(currentPath, "docsgenerator", "templates")
	docsDir := filepath.Join(currentPath, "docs")

	gen := generator.NewGenerator(templateDir, docsDir, executor.NewExecutor(omPath), os.Stdout)

	err = gen.GenerateDocs()
	if err != nil {
		fmt.Printf("could not generate docs: %s\n", err)
		os.Exit(1)
	}
}
