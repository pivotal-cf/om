package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	_, executable := filepath.Split(os.Args[0])

	if os.Getenv("STUB_EFFECTIVE_NAME") == "" || os.Getenv("STUB_EFFECTIVE_NAME") == executable {
		if val := os.Getenv("STUB_ERROR_CODE"); val != "" {
			var msg = executable + " error!!"
			if errorMsg := os.Getenv("STUB_ERROR_MSG"); errorMsg != "" {
				msg = errorMsg
			}

			code, err := strconv.Atoi(val)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Fprintln(os.Stderr, msg)
			os.Exit(code)
		}
	}

	if os.Getenv("STUB_OUTPUT") != "" {
		fmt.Fprintln(os.Stdout, os.Getenv("STUB_OUTPUT"))
	}
	fmt.Fprintf(os.Stderr, "%s %v\n", executable, strings.Join(os.Args[1:], " "))
	fmt.Fprintf(os.Stderr, "env: %v\n", strings.Join(os.Environ(), " "))

	os.Exit(0)
}
