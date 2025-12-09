//go:build integration

package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pivotal-cf/om/pkg/testinfra"
)

// spnegoTestEnv holds the shared test infrastructure for SPNEGO proxy tests.
type spnegoTestEnv struct {
	OmBinary    string
	ProxyURL    string
	KRB5Path    string
	PivnetToken string
}

func setupSPNEGOTestEnv(t *testing.T) (*spnegoTestEnv, func()) {
	t.Helper()

	pivnetToken := os.Getenv("PIVNET_TOKEN")
	if pivnetToken == "" {
		t.Fatal("PIVNET_TOKEN environment variable must be set")
	}

	omBinary := buildOmBinary(t)
	t.Logf("om binary built at: %s", omBinary)

	// Use shared infrastructure from pkg/testinfra
	infraEnv, cleanup := testinfra.StartSPNEGOInfra(t)

	return &spnegoTestEnv{
		OmBinary:    omBinary,
		ProxyURL:    infraEnv.ProxyURL,
		KRB5Path:    infraEnv.KRB5Path,
		PivnetToken: pivnetToken,
	}, cleanup
}

// TestSPNEGOProxy runs full E2E SPNEGO proxy tests with PivNet download.
func TestSPNEGOProxy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	env, cleanup := setupSPNEGOTestEnv(t)
	defer cleanup()

	t.Run("CLIFlags", func(t *testing.T) {
		testSPNEGOWithCLIFlags(t, env)
	})

	t.Run("ConfigFile", func(t *testing.T) {
		testSPNEGOWithConfigFile(t, env)
	})
}

// testSPNEGOWithCLIFlags tests SPNEGO proxy auth using command-line flags.
func testSPNEGOWithCLIFlags(t *testing.T, env *spnegoTestEnv) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "downloads")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	cmd := exec.Command(env.OmBinary, "download-product",
		"--pivnet-api-token", env.PivnetToken,
		"--pivnet-product-slug", "credhub-service-broker",
		"--product-version-regex", `1\.6\..*`,
		"--file-glob", "*.zip",
		"--output-directory", outputDir,
		"--proxy-url", env.ProxyURL,
		"--proxy-username", testinfra.TestUsername,
		"--proxy-password", testinfra.TestPassword,
		"--proxy-auth-type", "spnego",
		"--proxy-krb5-config", env.KRB5Path,
	)

	cmd.Env = os.Environ()

	t.Logf("Running: %s %s", env.OmBinary, strings.Join(cmd.Args[1:], " "))

	output, err := cmd.CombinedOutput()
	t.Logf("Output:\n%s", string(output))

	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	verifyDownloadedFiles(t, outputDir)
}

// testSPNEGOWithConfigFile tests SPNEGO proxy auth using a config file.
func testSPNEGOWithConfigFile(t *testing.T, env *spnegoTestEnv) {
	tmpDir := t.TempDir()

	downloadConfig := fmt.Sprintf(`---
pivnet-product-slug: credhub-service-broker
product-version-regex: "1\\.6\\..*"
file-glob: "*.zip"
proxy-url: %s
proxy-username: %s
proxy-password: %s
proxy-auth-type: spnego
proxy-krb5-config: %s
`, env.ProxyURL, testinfra.TestUsername, testinfra.TestPassword, env.KRB5Path)

	configPath := filepath.Join(tmpDir, "download-config.yml")
	if err := os.WriteFile(configPath, []byte(downloadConfig), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "downloads")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	cmd := exec.Command(env.OmBinary, "download-product",
		"--config", configPath,
		"--output-directory", outputDir,
		"--pivnet-api-token", env.PivnetToken,
	)

	cmd.Env = os.Environ()

	t.Logf("Running: %s %s", env.OmBinary, strings.Join(cmd.Args[1:], " "))

	output, err := cmd.CombinedOutput()
	t.Logf("Output:\n%s", string(output))

	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	verifyDownloadedFiles(t, outputDir)
}

// TestSimpleProxy tests basic proxy functionality without authentication.
func TestSimpleProxy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	pivnetToken := os.Getenv("PIVNET_TOKEN")
	if pivnetToken == "" {
		t.Fatal("PIVNET_TOKEN environment variable must be set")
	}

	omBinary := buildOmBinary(t)
	t.Logf("om binary built at: %s", omBinary)

	// Use shared infrastructure from pkg/testinfra
	env, cleanup := testinfra.StartSimpleProxyInfra(t)
	defer cleanup()

	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "downloads")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	cmd := exec.Command(omBinary, "download-product",
		"--pivnet-api-token", pivnetToken,
		"--pivnet-product-slug", "credhub-service-broker",
		"--product-version-regex", `1\.6\..*`,
		"--file-glob", "*.zip",
		"--output-directory", outputDir,
		"--proxy-url", env.ProxyURL,
	)

	cmd.Env = os.Environ()

	t.Logf("Running: %s %s", omBinary, strings.Join(cmd.Args[1:], " "))

	output, err := cmd.CombinedOutput()
	t.Logf("Output:\n%s", string(output))

	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	verifyDownloadedFiles(t, outputDir)
}

// TestProxyOptionsInHelp verifies proxy options appear in om help output.
func TestProxyOptionsInHelp(t *testing.T) {
	omBinary := buildOmBinary(t)

	cmd := exec.Command(omBinary, "download-product", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	requiredOptions := []string{
		"proxy-url",
		"proxy-username",
		"proxy-password",
		"proxy-auth-type",
		"proxy-krb5-config",
	}

	for _, opt := range requiredOptions {
		if !strings.Contains(outputStr, opt) {
			t.Errorf("missing option in help output: %s", opt)
		}
	}
}

// verifyDownloadedFiles checks that at least one file was downloaded.
func verifyDownloadedFiles(t *testing.T, outputDir string) {
	t.Helper()

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	var downloadedFiles []string
	for _, entry := range entries {
		if !entry.IsDir() {
			downloadedFiles = append(downloadedFiles, entry.Name())
		}
	}

	if len(downloadedFiles) == 0 {
		t.Fatal("Expected at least one downloaded file, got none")
	}

	t.Logf("Downloaded: %v", downloadedFiles)
}

// buildOmBinary builds the om binary from the current codebase.
func buildOmBinary(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get caller information")
	}
	omMainPath := filepath.Join(filepath.Dir(filename), "..", "main.go")

	tmpDir := t.TempDir()
	omBinaryPath := filepath.Join(tmpDir, "om")

	cmd := exec.Command("go", "build", "-o", omBinaryPath, omMainPath)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build om binary: %v\nOutput: %s", err, output)
	}

	return omBinaryPath
}
