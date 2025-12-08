package testinfra

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

// SPNEGOEnv holds the SPNEGO test infrastructure state.
type SPNEGOEnv struct {
	ProxyURL    string // URL to connect to proxy from host
	SPNProxyURL string // URL for SPN construction
	KRB5Path    string // Path to krb5.conf
	TempDir     string

	ctx        context.Context
	network    *testcontainers.DockerNetwork
	kdc        testcontainers.Container
	proxy      testcontainers.Container
	keytabPath string
	logger     Logger
}

type Logger interface {
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}

type testLogger struct {
	t *testing.T
}

func (l *testLogger) Log(args ...interface{}) {
	l.t.Helper()
	l.t.Log(args...)
}

func (l *testLogger) Logf(format string, args ...interface{}) {
	l.t.Helper()
	l.t.Logf(format, args...)
}

type stdLogger struct{}

func (l *stdLogger) Log(args ...interface{}) {
	fmt.Println(args...)
}

func (l *stdLogger) Logf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// StartSPNEGOInfra starts KDC and Kerberos proxy containers for integration tests.
func StartSPNEGOInfra(t *testing.T) (*SPNEGOEnv, func()) {
	t.Helper()

	env, err := startSPNEGOInfraInternal(&testLogger{t}, t.TempDir())
	if err != nil {
		t.Fatalf("Failed to start SPNEGO infrastructure: %v", err)
	}

	cleanup := func() {
		env.Close()
	}

	return env, cleanup
}

// StartSPNEGOInfraStandalone starts KDC and Kerberos proxy for manual testing.
func StartSPNEGOInfraStandalone() (*SPNEGOEnv, error) {
	tmpDir, err := os.MkdirTemp("", "spnego-infra-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	env, err := startSPNEGOInfraInternal(&stdLogger{}, tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}

	return env, nil
}

func startSPNEGOInfraInternal(logger Logger, tmpDir string) (*SPNEGOEnv, error) {
	ctx := context.Background()

	env := &SPNEGOEnv{
		ctx:     ctx,
		TempDir: tmpDir,
		logger:  logger,
	}

	logger.Log("Creating Docker network...")
	testNetwork, err := network.New(ctx,
		network.WithCheckDuplicate(),
		network.WithDriver("bridge"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create test network: %w", err)
	}
	env.network = testNetwork

	logger.Log("Starting KDC container...")
	kdc, err := StartKDC(ctx, testNetwork.Name)
	if err != nil {
		testNetwork.Remove(ctx)
		return nil, fmt.Errorf("failed to start KDC: %w", err)
	}
	env.kdc = kdc
	logger.Log("KDC started successfully")

	logger.Log("Creating test user principal...")
	if err := CreatePrincipal(ctx, kdc, TestUsername, TestPassword); err != nil {
		kdc.Terminate(ctx)
		testNetwork.Remove(ctx)
		return nil, fmt.Errorf("failed to create test principal: %w", err)
	}

	logger.Log("Creating HTTP service principal for proxy...")
	if err := CreateServicePrincipal(ctx, kdc, "HTTP", ProxyHostname); err != nil {
		kdc.Terminate(ctx)
		testNetwork.Remove(ctx)
		return nil, fmt.Errorf("failed to create service principal: %w", err)
	}

	logger.Log("Extracting keytab from KDC...")
	keytabPath, err := ExtractKeytabFromKDC(ctx, kdc)
	if err != nil {
		kdc.Terminate(ctx)
		testNetwork.Remove(ctx)
		return nil, fmt.Errorf("failed to extract keytab: %w", err)
	}
	env.keytabPath = keytabPath

	logger.Log("Starting Squid proxy with Kerberos auth...")
	proxy, err := StartKerberosProxy(ctx, testNetwork.Name, keytabPath, kdc)
	if err != nil {
		os.Remove(keytabPath)
		kdc.Terminate(ctx)
		testNetwork.Remove(ctx)
		return nil, fmt.Errorf("failed to start Kerberos proxy: %w", err)
	}
	env.proxy = proxy
	time.Sleep(3 * time.Second)

	proxyURL, err := GetProxyURL(ctx, proxy)
	if err != nil {
		proxy.Terminate(ctx)
		os.Remove(keytabPath)
		kdc.Terminate(ctx)
		testNetwork.Remove(ctx)
		return nil, fmt.Errorf("failed to get proxy URL: %w", err)
	}
	env.ProxyURL = proxyURL
	logger.Logf("Proxy running at: %s", proxyURL)

	kdcHost, _ := kdc.Host(ctx)
	kdcPort, _ := kdc.MappedPort(ctx, "88/tcp")
	krb5Config := GetKRB5Config(kdcHost, kdcPort.Port())

	krb5Path := filepath.Join(tmpDir, "krb5.conf")
	if err := os.WriteFile(krb5Path, []byte(krb5Config), 0644); err != nil {
		proxy.Terminate(ctx)
		os.Remove(keytabPath)
		kdc.Terminate(ctx)
		testNetwork.Remove(ctx)
		return nil, fmt.Errorf("failed to write krb5.conf: %w", err)
	}
	env.KRB5Path = krb5Path
	env.SPNProxyURL = fmt.Sprintf("http://%s:%s", ProxyHostname, ProxyPort)
	logger.Logf("KRB5_CONFIG: %s", krb5Path)

	return env, nil
}

func (e *SPNEGOEnv) Close() {
	if e.proxy != nil {
		e.logger.Log("Terminating proxy container...")
		e.proxy.Terminate(e.ctx)
	}
	if e.keytabPath != "" {
		os.Remove(e.keytabPath)
	}
	if e.kdc != nil {
		e.logger.Log("Terminating KDC container...")
		e.kdc.Terminate(e.ctx)
	}
	if e.network != nil {
		e.logger.Log("Removing Docker network...")
		e.network.Remove(e.ctx)
	}
}

func (e *SPNEGOEnv) PrintInstructions() {
	fmt.Println()
	fmt.Println("=== SPNEGO Test Infrastructure Running ===")
	fmt.Println()
	fmt.Println("Set environment variables:")
	fmt.Printf("  export KRB5_CONFIG=%s\n", e.KRB5Path)
	fmt.Printf("  export HTTP_PROXY=%s\n", e.ProxyURL)
	fmt.Printf("  export HTTPS_PROXY=%s\n", e.ProxyURL)
	fmt.Println()
	fmt.Println("Run om with:")
	fmt.Printf("  go run ./main.go download-product \\\n")
	fmt.Printf("    --pivnet-api-token $PIVNET_TOKEN \\\n")
	fmt.Printf("    --pivnet-product-slug tanzu-yourkit-buildpack \\\n")
	fmt.Printf("    --product-version-regex \".*\" \\\n")
	fmt.Printf("    --file-glob \"*.tgz\" \\\n")
	fmt.Printf("    --output-directory /tmp/downloads \\\n")
	fmt.Printf("    --proxy-url %s \\\n", e.SPNProxyURL)
	fmt.Printf("    --proxy-username %s \\\n", TestUsername)
	fmt.Printf("    --proxy-password %s \\\n", TestPassword)
	fmt.Printf("    --proxy-domain %s \\\n", TestRealm)
	fmt.Printf("    --proxy-auth-type spnego\n")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop...")
}

type SimpleProxyEnv struct {
	ProxyURL string
	ctx      context.Context
	network  *testcontainers.DockerNetwork
	proxy    testcontainers.Container
	logger   Logger
}

// StartSimpleProxyInfra starts a proxy without authentication.
func StartSimpleProxyInfra(t *testing.T) (*SimpleProxyEnv, func()) {
	t.Helper()

	ctx := context.Background()
	logger := &testLogger{t}

	env := &SimpleProxyEnv{
		ctx:    ctx,
		logger: logger,
	}

	logger.Log("Creating Docker network...")
	testNetwork, err := network.New(ctx,
		network.WithCheckDuplicate(),
		network.WithDriver("bridge"),
	)
	if err != nil {
		t.Fatalf("Failed to create test network: %v", err)
	}
	env.network = testNetwork

	logger.Log("Starting simple proxy...")
	proxy, err := StartSimpleProxy(ctx, testNetwork.Name)
	if err != nil {
		testNetwork.Remove(ctx)
		t.Fatalf("Failed to start proxy: %v", err)
	}
	env.proxy = proxy

	time.Sleep(2 * time.Second)

	proxyURL, err := GetProxyURL(ctx, proxy)
	if err != nil {
		proxy.Terminate(ctx)
		testNetwork.Remove(ctx)
		t.Fatalf("Failed to get proxy URL: %v", err)
	}
	env.ProxyURL = proxyURL
	logger.Logf("Proxy running at: %s", proxyURL)

	cleanup := func() {
		logger.Log("Terminating proxy container...")
		proxy.Terminate(ctx)
		logger.Log("Removing Docker network...")
		testNetwork.Remove(ctx)
	}

	return env, cleanup
}
