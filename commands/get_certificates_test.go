package commands

import (
	"testing"
)

func TestExtractSerialFromPEM_ValidPEM(t *testing.T) {
	validPEM := `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAIBsX+c6mMNwMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAklOMQswCQYDVQQIDAJUUzELMAkGA1UEBwwCUFMxEjAQBgNVBAoMCVRlc3Qg
Q29tcDAeFw0yNTA3MjkwODAwMDBaFw0yNjA3MjkwODAwMDBaMEUxCzAJBgNVBAYT
AklOMQswCQYDVQQIDAJUUzELMAkGA1UEBwwCUFMxEjAQBgNVBAoMCVRlc3QgQ29t
cDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALcFZx1Og5/cT74Zfz4G
3bFvFJ+ndjKnxy0I9Ju0OJKtQX1A9Ibn+EkGG1XimZmlfY6F9AB7Rx+azphBLRzP
EqIxe6DgCbiq8HUXEh6mQKGiZu7gdzpUwULbFfN7VTxRKL3X64wUgRp9Sblj9wCQ
ZjNULiXG9onCKz5NgORVnqYjLQu5LAL3fdI1rBUpyr4jcChpph5EguPSNSlMY2ni
yl9JtWDwWfuvDypf5UvKOG/0F61XhMbxnSKGLuLPSlfAXrkSk4QFrWrJrTWIPj3u
7P+pGpN2eYOlOE+YIzFydNWrZ7MRBjMHIqLtR8UwMi0Xb5L3YUbUNtrdMG52TJzD
qV0CAwEAAaNQME4wHQYDVR0OBBYEFMN5QmjUnZmyE6A8KCGxXb+jY+9yMB8GA1Ud
IwQYMBaAFMN5QmjUnZmyE6A8KCGxXb+jY+9yMAwGA1UdEwQFMAMBAf8wDQYJKoZI
hvcNAQELBQADggEBAK/zZK0ZoxOpXpOTplN1VyrTSPuUr19/jEsOYz9Th4XhP0vH
PsmjErRhvCrBIbfKrptbl5M3IXtzce57wJKYu4s9RM+5RM6+n7Lh3IpMzFdeyYxw
w5XK3CvuhbzZAP9JhKnITfKvmGV2Ov1j8RcfBhujRwEQW7nKjEk1AWKrNnJ0BPfR
uqgzTt9y+v4TxAjmZ0uK6qMGyuw8uBx1mKQ3Efe3ya0psIPuwUuNn37NvxkC2bhu
9IM2o8I2xHJmvGuRHHl9v1Rxbj+nFJJ7/NEO57HBlKJ5iUb6mSTToDSEak5T9n+G
/XHRPMdb6qU7vMds49x+rNmcRAkU+nGJRuOtBGk=
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
