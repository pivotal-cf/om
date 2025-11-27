// Package testinfra provides shared test infrastructure for SPNEGO/Kerberos proxy testing.
package testinfra

import "fmt"

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

// SquidKerberosConfig is the Squid config with Kerberos/SPNEGO auth enabled.
const SquidKerberosConfig = `# Squid configuration requiring Kerberos/SPNEGO authentication
http_port 3128

# ACLs
acl localnet src all
acl SSL_ports port 443
acl Safe_ports port 80
acl Safe_ports port 443
acl CONNECT method CONNECT

# Deny unsafe ports
http_access deny !Safe_ports
http_access deny CONNECT !SSL_ports

# Kerberos/SPNEGO authentication configuration
# The negotiate_kerberos_auth helper validates SPNEGO tokens
auth_param negotiate program /usr/lib/squid/negotiate_kerberos_auth -d -s HTTP/proxy.example.com@EXAMPLE.COM
auth_param negotiate children 10 startup=2 idle=2
auth_param negotiate keep_alive on

# Require authentication for all requests
acl authenticated proxy_auth REQUIRED
http_access allow authenticated

# Deny everything else
http_access deny all

# Logging
access_log /var/log/squid/access.log squid
cache_log /var/log/squid/cache.log

# Cache settings
cache deny all
coredump_dir /var/spool/squid

# Performance tuning
shutdown_lifetime 1 seconds
`

// SquidSimpleConfig is a Squid config that allows all traffic without auth.
const SquidSimpleConfig = `# Simple Squid configuration for testing - allows all traffic
http_port 3128

# Allow all source networks
acl localnet src all
acl Safe_ports port 80
acl Safe_ports port 443
acl CONNECT method CONNECT

# Allow all traffic for testing
http_access allow all

# Logging - use proper log files instead of stdio
access_log /var/log/squid/access.log squid
cache_log /var/log/squid/cache.log

# Minimal cache
cache deny all

# Disable coredumps
coredump_dir /var/spool/squid
`
