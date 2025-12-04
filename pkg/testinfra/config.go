// Package testinfra provides shared test infrastructure for SPNEGO/Kerberos proxy testing.
package testinfra

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"
)

const (
	TestUsername  = "testuser"
	TestPassword  = "testpass123"
	TestRealm     = "EXAMPLE.COM"
	TestAdminPass = "admin123"
	KDCHostname   = "kdc.example.com"
	ProxyHostname = "proxy.example.com"
	KDCImage      = "gcavalcante8808/krb5-server:latest"
	SquidImage    = "ubuntu/squid:latest"
	ProxyPort     = "3128"
)

// Embedded config templates from testdata directory.
// These files are easier to read and maintain than inline strings.
var (
	//go:embed testdata/squid/squid-kerberos.conf.tmpl
	squidKerberosTemplate string

	//go:embed testdata/squid/squid-simple.conf
	squidSimpleConfig string
)

// SquidKerberosTemplateData holds the data for the Kerberos squid config template.
type SquidKerberosTemplateData struct {
	ProxyHostname string
	Realm         string
}

// GetSquidKerberosConfig returns the Squid config with Kerberos/SPNEGO auth enabled.
// Uses template substitution for ProxyHostname and Realm.
func GetSquidKerberosConfig(proxyHostname, realm string) (string, error) {
	tmpl, err := template.New("squid-kerberos").Parse(squidKerberosTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse squid kerberos template: %w", err)
	}

	data := SquidKerberosTemplateData{
		ProxyHostname: proxyHostname,
		Realm:         realm,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute squid kerberos template: %w", err)
	}

	return buf.String(), nil
}

// GetSquidSimpleConfig returns the Squid config that allows all traffic without auth.
func GetSquidSimpleConfig() string {
	return squidSimpleConfig
}

// SquidKerberosConfig is kept for backward compatibility.
// Deprecated: Use GetSquidKerberosConfig() instead for proper template substitution.
var SquidKerberosConfig = func() string {
	config, _ := GetSquidKerberosConfig(ProxyHostname, TestRealm)
	return config
}()

// SquidSimpleConfig is kept for backward compatibility.
// Use GetSquidSimpleConfig() for explicit function call.
var SquidSimpleConfig = squidSimpleConfig

// GetKRB5Config returns a krb5.conf for the client to connect to the KDC.
func GetKRB5Config(kdcHost string, kdcPort string) string {
	return fmt.Sprintf(`[libdefaults]
    default_realm = %s
    dns_lookup_realm = false
    dns_lookup_kdc = false
    ticket_lifetime = 24h
    renew_lifetime = 7d
    forwardable = true
    rdns = false

[realms]
    %s = {
        kdc = %s:%s
        admin_server = %s:749
    }

[domain_realm]
    .example.com = %s
    example.com = %s
`, TestRealm, TestRealm, kdcHost, kdcPort, kdcHost, TestRealm, TestRealm)
}

// GetProxyKRB5Config returns a krb5.conf for the proxy container.
func GetProxyKRB5Config(kdcIP string) string {
	return fmt.Sprintf(`[libdefaults]
    default_realm = %s
    dns_lookup_realm = false
    dns_lookup_kdc = false

[realms]
    %s = {
        kdc = %s:88
        admin_server = %s:749
    }

[domain_realm]
    .example.com = %s
    example.com = %s
`, TestRealm, TestRealm, kdcIP, kdcIP, TestRealm, TestRealm)
}
