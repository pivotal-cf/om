package generator

type SyslogProperties struct {
	Enabled                    string `yaml:"enabled,omitempty"`
	Address                    string `yaml:"address"`
	Port                       string `yaml:"port"`
	TransportProtocol          string `yaml:"transport_protocol,omitempty"`
	TLSEnabled                 string `yaml:"tls_enabled,omitempty"`
	PermittedPeer              string `yaml:"permitted_peer,omitempty"`
	SSLCACertificate           string `yaml:"ssl_ca_certificate,omitempty"`
	QueueSize                  string `yaml:"queue_size,omitempty"`
	ForwardDebugLogs           string `yaml:"forward_debug_logs,omitempty"`
	CustomRSyslogConfiguration string `yaml:"custom_rsyslog_configuration,omitempty"`
}

func CreateSyslogProperties(metadata metadata) *SyslogProperties {
	if !metadata.UsesOpsManagerSyslogProperties() {
		return nil
	}

	return &SyslogProperties{
		Enabled:                    "((syslog_enabled))",
		Address:                    "((syslog_address))",
		Port:                       "((syslog_port))",
		TransportProtocol:          "((syslog_transport_protocol))",
		TLSEnabled:                 "((syslog_tls_enabled))",
		PermittedPeer:              "((syslog_permitted_peer))",
		SSLCACertificate:           "((syslog_ssl_ca_certificate))",
		QueueSize:                  "((syslog_queue_size))",
		ForwardDebugLogs:           "((syslog_forward_debug_logs))",
		CustomRSyslogConfiguration: "((syslog_custom_rsyslog_configuration))",
	}
}