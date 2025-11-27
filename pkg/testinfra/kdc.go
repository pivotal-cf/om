package testinfra

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartKDC starts a KDC container.
func StartKDC(ctx context.Context, networkName string) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:    KDCImage,
		Hostname: KDCHostname,
		Networks: []string{networkName},
		Env: map[string]string{
			"KRB5_REALM":          TestRealm,
			"KRB5_KDC":            KDCHostname,
			"KRB5_ADMIN_PASSWORD": TestAdminPass,
		},
		ExposedPorts: []string{"88/tcp", "88/udp", "749/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Principal \"admin/admin@EXAMPLE.COM\" created"),
			wait.ForListeningPort("88/tcp"),
		).WithStartupTimeout(120 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start KDC container: %w", err)
	}

	return container, nil
}

// CreatePrincipal creates a user principal.
func CreatePrincipal(ctx context.Context, kdc testcontainers.Container, principal, password string) error {
	cmd := []string{
		"kadmin.local", "-q",
		fmt.Sprintf("addprinc -pw %s %s@%s", password, principal, TestRealm),
	}

	exitCode, _, err := kdc.Exec(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create principal %s: %w", principal, err)
	}
	if exitCode != 0 {
		return fmt.Errorf("kadmin.local exited with code %d", exitCode)
	}

	return nil
}

// CreateServicePrincipal creates a service principal and extracts its keytab.
func CreateServicePrincipal(ctx context.Context, kdc testcontainers.Container, service, hostname string) error {
	spn := fmt.Sprintf("%s/%s@%s", service, hostname, TestRealm)

	createCmd := []string{
		"kadmin.local", "-q",
		fmt.Sprintf("addprinc -randkey %s", spn),
	}

	exitCode, _, err := kdc.Exec(ctx, createCmd)
	if err != nil {
		return fmt.Errorf("failed to create service principal %s: %w", spn, err)
	}
	if exitCode != 0 {
		return fmt.Errorf("kadmin.local addprinc exited with code %d", exitCode)
	}

	keytabCmd := []string{
		"kadmin.local", "-q",
		fmt.Sprintf("ktadd -k /tmp/proxy.keytab %s", spn),
	}

	exitCode, _, err = kdc.Exec(ctx, keytabCmd)
	if err != nil {
		return fmt.Errorf("failed to create keytab for %s: %w", spn, err)
	}
	if exitCode != 0 {
		return fmt.Errorf("kadmin.local ktadd exited with code %d", exitCode)
	}

	verifyCmd := []string{"ls", "-la", "/tmp/proxy.keytab"}
	exitCode, _, err = kdc.Exec(ctx, verifyCmd)
	if err != nil {
		return fmt.Errorf("keytab file not found after creation: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("keytab file verification failed with code %d", exitCode)
	}

	return nil
}
