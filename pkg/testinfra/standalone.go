//go:build ignore

// Run with: go run ./pkg/testinfra/standalone.go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pivotal-cf/om/pkg/testinfra"
)

func main() {
	fmt.Println("Starting SPNEGO test infrastructure...")

	env, err := testinfra.StartSPNEGOInfraStandalone()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	env.PrintInstructions()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
	env.Close()
}
