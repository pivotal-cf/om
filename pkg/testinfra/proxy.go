package testinfra

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartKerberosProxy starts a Squid proxy with Kerberos auth.
func StartKerberosProxy(ctx context.Context, networkName string, keytabPath string, kdc testcontainers.Container) (testcontainers.Container, error) {
	kdcIP, err := kdc.ContainerIP(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get KDC IP: %w", err)
	}

	krb5Conf := GetProxyKRB5Config(kdcIP)
	krb5File, err := os.CreateTemp("", "krb5-proxy-*.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp krb5.conf: %w", err)
	}
	if _, err := krb5File.WriteString(krb5Conf); err != nil {
		krb5File.Close()
		os.Remove(krb5File.Name())
		return nil, fmt.Errorf("failed to write krb5.conf: %w", err)
	}
	krb5File.Close()

	// Generate Squid config from template - use "localhost" since test connects via localhost:mappedPort
	squidConfig, err := GetSquidKerberosConfig("localhost", TestRealm)
	if err != nil {
		os.Remove(krb5File.Name())
		return nil, fmt.Errorf("failed to generate squid config: %w", err)
	}

	squidFile, err := os.CreateTemp("", "squid-kerberos-*.conf")
	if err != nil {
		os.Remove(krb5File.Name())
		return nil, fmt.Errorf("failed to create temp squid.conf: %w", err)
	}
	if _, err := squidFile.WriteString(squidConfig); err != nil {
		squidFile.Close()
		os.Remove(squidFile.Name())
		os.Remove(krb5File.Name())
		return nil, fmt.Errorf("failed to write squid.conf: %w", err)
	}
	squidFile.Close()

	req := testcontainers.ContainerRequest{
		Image:        SquidImage,
		Hostname:     ProxyHostname,
		Networks:     []string{networkName},
		ExposedPorts: []string{ProxyPort + "/tcp"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      keytabPath,
				ContainerFilePath: "/etc/squid/proxy.keytab",
				FileMode:          0644,
			},
			{
				HostFilePath:      krb5File.Name(),
				ContainerFilePath: "/etc/krb5.conf",
				FileMode:          0644,
			},
			{
				HostFilePath:      squidFile.Name(),
				ContainerFilePath: "/etc/squid/squid.conf",
				FileMode:          0644,
			},
		},
		Env: map[string]string{
			"KRB5_KTNAME": "/etc/squid/proxy.keytab",
		},
		WaitingFor: wait.ForListeningPort(ProxyPort + "/tcp").WithStartupTimeout(120 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	os.Remove(krb5File.Name())
	os.Remove(squidFile.Name())

	if err != nil {
		return nil, fmt.Errorf("failed to start Kerberos proxy: %w", err)
	}

	return container, nil
}

// StartSimpleProxy starts a Squid proxy without authentication.
func StartSimpleProxy(ctx context.Context, networkName string) (testcontainers.Container, error) {
	squidFile, err := os.CreateTemp("", "squid-simple-*.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp squid.conf: %w", err)
	}
	if _, err := squidFile.WriteString(GetSquidSimpleConfig()); err != nil {
		squidFile.Close()
		os.Remove(squidFile.Name())
		return nil, fmt.Errorf("failed to write squid.conf: %w", err)
	}
	squidFile.Close()

	req := testcontainers.ContainerRequest{
		Image:        SquidImage,
		Hostname:     ProxyHostname,
		Networks:     []string{networkName},
		ExposedPorts: []string{ProxyPort + "/tcp"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      squidFile.Name(),
				ContainerFilePath: "/etc/squid/squid.conf",
				FileMode:          0644,
			},
		},
		WaitingFor: wait.ForListeningPort(ProxyPort + "/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	os.Remove(squidFile.Name())

	if err != nil {
		return nil, fmt.Errorf("failed to start simple proxy container: %w", err)
	}

	return container, nil
}

// GetProxyURL returns the proxy URL accessible from the host.
func GetProxyURL(ctx context.Context, proxy testcontainers.Container) (string, error) {
	host, err := proxy.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get proxy host: %w", err)
	}

	port, err := proxy.MappedPort(ctx, ProxyPort+"/tcp")
	if err != nil {
		return "", fmt.Errorf("failed to get proxy port: %w", err)
	}

	return fmt.Sprintf("http://%s:%s", host, port.Port()), nil
}

// ExtractKeytabFromKDC copies the keytab from KDC to a local temp file.
func ExtractKeytabFromKDC(ctx context.Context, kdc testcontainers.Container) (string, error) {
	reader, err := kdc.CopyFileFromContainer(ctx, "/tmp/proxy.keytab")
	if err != nil {
		return "", fmt.Errorf("failed to copy keytab from KDC: %w", err)
	}
	defer reader.Close()

	tmpFile, err := os.CreateTemp("", "proxy-*.keytab")
	if err != nil {
		return "", fmt.Errorf("failed to create temp keytab file: %w", err)
	}

	if _, err := io.Copy(tmpFile, reader); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write keytab: %w", err)
	}
	tmpFile.Close()

	return tmpFile.Name(), nil
}
