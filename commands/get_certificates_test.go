package commands

import (
	"testing"
)

func TestExtractSerialFromPEM_ValidPEM(t *testing.T) {
	validPEM := `-----BEGIN CERTIFICATE-----
MIIDaTCCAlGgAwIBAgIUXjSvpwI5RX6NQdS7iwAAAAAAAMwwDQYJKoZIhvcNAQEL
BQAwRzELMAkGA1UEBhMCVVMxDjAMBgNVBAoMBUR1bW15MSgwJgYDVQQDDB9sb2ct
Y2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMB4XDTI1MDcyOTE0NTY0NVoXDTI3
MDcyOTE0NTY0NVowRzELMAkGA1UEBhMCVVMxDjAMBgNVBAoMBUR1bW15MSgwJgYD
VQQDDB9sb2ctY2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxkrPEu1uD1QcvVmHgql60r1u0fl4BbmB+5pQ
9J8wnbSOpefbq6YiTb8auHf/ChpwrQnIVv4NiFYOy4s73CutY1vXSalfhrbzMdug
GGOtZB0LVtVHZi1GGxysx9DteDFHsKuPCCa+LrSLR89b2doP6jZ/031L7rn7J+k2
wevEStmtjLAiekiMx+b4caWKYmhHjmHfgw9r5obEFN3JSKfLNPBqAbEjyTPjx4U6
BxaxJwaABeY6t8iJKXs+pmbDoh5BrXgOriLzyFy3ws8oP+gK9aSLBZk9/37LMql+
fE/JB4EWTz+9LDeKfRANlWGgcBqqlWb6E9kyV8TqC4ncvIHIcwIDAQABo00wSzAq
BgNVHREEIzAhgh9sb2ctY2FjaGUtc3lzbG9nLXNlcnZlci1tZXRyaWNzMB0GA1Ud
DgQWBBSBl0c4tqKJnWkK3+ac0edqmXJRSTANBgkqhkiG9w0BAQsFAAOCAQEAVBMv
/+yW4XkZxqzjTm1ZCryXT8+mtD7tYLSuuvHFWKWsAnVjf1Ve2V8OH5caZEoAeT7b
Z4XHCs7sLlg81HLI9tXMOhlrRATgC+ccnuGom2Ts1e4mqworhz5/uhF35ci6+qGv
Zv8a+d9NK1mm5vIPv4y2jE2bE3+tR8ggtrkxPRoRnvlzFx8C7XK9xVHZH88UOJSl
1lVnJAeOd3VTyy3ADqHRmoh3gTq6lL5CfDn4gfdxAxfaLsuwmusu9I4Zt1uKRA2A
kPg01BLHtNI/U4oTDLku7wLJGFhXBAnLVpomuV5m6xwHZ4eOPvsp6XN1laUBwaAu
svTEb6EMuB8T2B9Rtg==
-----END CERTIFICATE-----`

	serial, err := extractSerialFromPEM(validPEM)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if serial == "" {
		t.Error("expected serial number, got empty string")
	}
}

func TestExtractSerialFromPEM_InvalidPEM(t *testing.T) {
	invalidPEM := "invalid-pem-data"

	serial, err := extractSerialFromPEM(invalidPEM)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if serial != "" {
		t.Errorf("expected empty serial, got: %s", serial)
	}
}

func TestExtractSerialFromPEM_EmptyPEM(t *testing.T) {
	serial, err := extractSerialFromPEM("")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if serial != "" {
		t.Errorf("expected empty serial, got: %s", serial)
	}
}
