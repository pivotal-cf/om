package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SSLCertificateSettings struct {
	CertPem       string `json:"certificate" yaml:"certificate"`
	PrivateKeyPem string `json:"private_key" yaml:"private_key"`
}

type SSLCertificateOutput struct {
	Certificate SSLCertificate `json:"ssl_certificate"`
}

type SSLCertificate struct {
	Certificate string `json:"certificate"`
}

type RBACSettings struct {
	SAMLAdminGroup      string `json:"rbac_saml_admin_group,omitempty" yaml:"rbac_saml_admin_group"`
	SAMLGroupsAttribute string `json:"rbac_saml_groups_attribute,omitempty" yaml:"rbac_saml_groups_attribute"`
	LDAPAdminGroupName  string `json:"ldap_rbac_admin_group_name,omitempty" yaml:"ldap_rbac_admin_group_name"`
}

type BannerSettings struct {
	UIBanner  string `json:"ui_banner_contents" yaml:"ui_banner_contents"`
	SSHBanner string `json:"ssh_banner_contents" yaml:"ssh_banner_contents"`
}

type PivnetSettings struct {
	APIToken string `json:"api_token" yaml:"api_token"`
}

type SyslogSettings struct {
	Enabled string `json:"enabled,omitempty" yaml:"enabled"`
	Address string `json:"address,omitempty" yaml:"address"`
	Port string `json:"port,omitempty" yaml:"port"`
	TransportProtocol string `json:"transport_protocol,omitempty" yaml:"transport_protocol"`
	TLSEnabled string `json:"tls_enabled,omitempty" yaml:"tls_enabled"`
	PermittedPeer string `json:"permitted_peer,omitempty" yaml:"permitted_peer"`
	SSLCACertificate string `json:"ssl_ca_certificate,omitempty" yaml:"ssl_ca_certificate"`
	QueueSize string `json:"queue_size,omitempty" yaml:"queue_size"`
	ForwardDebugLogs string `json:"forward_debug_logs,omitempty" yaml:"forward_debug_logs"`
	CustomRsyslogConfig string `json:"custom_rsyslog_configuration,omitempty" yaml:"custom_rsyslog_configuration"`
}

func (a Api) UpdateSSLCertificate(certBody SSLCertificateSettings) error {
	body, err := json.Marshal(certBody)
	if err != nil {
		return err // not tested
	}

	req, err := http.NewRequest("PUT", "/api/v0/settings/ssl_certificate", bytes.NewReader(body))
	if err != nil {
		return err // not tested
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) GetSSLCertificate() (SSLCertificateOutput, error) {
	var output SSLCertificateOutput

	req, err := http.NewRequest("GET", "/api/v0/settings/ssl_certificate", nil)
	if err != nil {
		return output, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return output, err
	}

	if err = validateStatusOK(resp); err != nil {
		return SSLCertificateOutput{}, err
	}

	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return output, err
	}

	if output.Certificate.Certificate == "" {
		output.Certificate.Certificate = "Ops Manager Self Signed Cert"
	}

	return output, nil
}

func (a Api) DeleteSSLCertificate() error {
	req, err := http.NewRequest("DELETE", "/api/v0/settings/ssl_certificate", nil)
	if err != nil {
		return err // not tested
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) UpdatePivnetToken(pivnetSettings PivnetSettings) error {
	body, err := json.Marshal(pivnetSettings)
	if err != nil {
		return err
	}

	payload := strings.NewReader(fmt.Sprintf(
		`{ "pivotal_network_settings": %s}`, body))

	req, err := http.NewRequest("PUT", "/api/v0/settings/pivotal_network_settings", payload)
	if err != nil {
		return err // not tested
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) EnableRBAC(rbacSettings RBACSettings) error {
	settingsBytes, err := json.Marshal(rbacSettings)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(
		"PUT",
		"/api/v0/settings/rbac",
		bytes.NewReader(settingsBytes),
	)
	if err != nil {
		return err // not tested
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) UpdateBanner(bannerSettings BannerSettings) error {
	body, err := json.Marshal(bannerSettings)
	if err != nil {
		return err // not tested
	}

	req, err := http.NewRequest("PUT", "/api/v0/settings/banner", bytes.NewReader(body))
	if err != nil {
		return err // not tested
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) UpdateSyslogSettings(syslogSettings SyslogSettings) error {
	body, err := json.Marshal(syslogSettings)
	if err != nil {
		return err // not tested
	}

	payload := strings.NewReader(fmt.Sprintf(
		`{ "syslog": %s}`, body))

	req, err := http.NewRequest("PUT", "/api/v0/settings/syslog", payload)
	if err != nil {
		return err // not tested
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
